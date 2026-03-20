#!/usr/bin/env python3
"""Vosk WebSocket server compatible with websockets 8.1"""
import json, os, sys, asyncio
import concurrent.futures, logging
import websockets
from vosk import Model, KaldiRecognizer

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("vosk-server")

model = None
pool = None

def process_chunk(rec, message):
    if message == '{"eof" : 1}' or message == '{"eof": 1}':
        return rec.FinalResult(), True
    elif isinstance(message, bytes):
        if rec.AcceptWaveform(message):
            return rec.Result(), False
        else:
            return rec.PartialResult(), False
    return '{}', False

async def recognize(websocket, path):
    global model, pool
    rec = None
    sample_rate = float(os.environ.get('VOSK_SAMPLE_RATE', 16000))
    logger.info('Connection from %s', websocket.remote_address)
    
    try:
        async for message in websocket:
            if isinstance(message, str):
                jobj = json.loads(message)
                if 'config' in jobj:
                    sample_rate = jobj['config'].get('sample_rate', sample_rate)
                    rec = KaldiRecognizer(model, sample_rate)
                    continue
                if 'eof' in jobj:
                    if rec:
                        result = rec.FinalResult()
                    else:
                        result = json.dumps({"text": ""})
                    await websocket.send(result)
                    break
            elif isinstance(message, bytes):
                if rec is None:
                    rec = KaldiRecognizer(model, sample_rate)
                loop = asyncio.get_running_loop()
                response, stop = await loop.run_in_executor(pool, process_chunk, rec, message)
                await websocket.send(response)
                if stop:
                    break
    except websockets.exceptions.ConnectionClosed:
        logger.info("Client disconnected")

def main():
    global model, pool
    model_path = os.environ.get('VOSK_MODEL_PATH', 'model')
    interface = os.environ.get('VOSK_SERVER_INTERFACE', '0.0.0.0')
    port = int(os.environ.get('VOSK_SERVER_PORT', 2700))
    
    logger.info(f"Loading model from {model_path}...")
    model = Model(model_path)
    logger.info("Model loaded!")
    
    pool = concurrent.futures.ThreadPoolExecutor((os.cpu_count() or 1))
    
    logger.info(f"Starting server on {interface}:{port}")
    # websockets 8.x compatible serve
    start_server = websockets.serve(recognize, interface, port)
    asyncio.get_event_loop().run_until_complete(start_server)
    logger.info(f"Server listening on {interface}:{port}")
    asyncio.get_event_loop().run_forever()

if __name__ == '__main__':
    main()
