import os
import uvicorn
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Union
from sentence_transformers import SentenceTransformer

# Инициализация приложения и модели (загружается при старте)
app = FastAPI(title="Local Embeddings API (OpenAI Compatible)")

MODEL_NAME = os.environ.get("EMBEDDINGS_MODEL", "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2")
print(f"Загрузка модели {MODEL_NAME}...")
try:
    model = SentenceTransformer(MODEL_NAME)
    print("Модель успешно загружена!")
except Exception as e:
    print(f"Ошибка загрузки модели: {e}")
    exit(1)

class EmbeddingRequest(BaseModel):
    input: Union[str, List[str]]
    model: str = "local-minilm"

@app.post("/v1/embeddings")
@app.post("/embeddings")
async def create_embeddings(req: EmbeddingRequest):
    # Преобразуем строку в список, если нужно
    texts = [req.input] if isinstance(req.input, str) else req.input
    if not texts:
        raise HTTPException(status_code=400, detail="input cannot be empty")

    try:
        # Генерируем векторы
        embeddings = model.encode(texts, normalize_embeddings=True)
        
        # Формируем ответ в формате OpenAI
        data = []
        for i, emb in enumerate(embeddings):
            data.append({
                "object": "embedding",
                "embedding": emb.tolist(),
                "index": i
            })
            
        return {
            "object": "list",
            "data": data,
            "model": req.model,
            "usage": {
                "prompt_tokens": sum(len(t) for t in texts), # Упрощенный подсчет
                "total_tokens": sum(len(t) for t in texts)
            }
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    port = int(os.environ.get("PORT", "7997"))
    uvicorn.run(app, host="0.0.0.0", port=port)
