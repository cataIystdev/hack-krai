// Файл vosk_stt.go реализует STT-клиент через Vosk-server (WebSocket).
// Vosk-server запускается в Docker (alphacep/kaldi-ru) и принимает PCM 16kHz mono
// по WebSocket на порту 2700. Аудиофайлы конвертируются через ffmpeg.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// voskChunkSize — размер одного чанка PCM-данных для отправки по WebSocket (8KB).
const voskChunkSize = 8192

// VoskClient — клиент для распознавания речи через Vosk-server (WebSocket).
type VoskClient struct {
	// wsURL — URL WebSocket подключения к Vosk-server (ws://host:port).
	wsURL string

	// logger — логгер.
	logger *zap.Logger

	// isMock — флаг mock-режима.
	isMock bool
}

// NewVoskClient создаёт клиент распознавания речи через Vosk.
// host и port — адрес Vosk-server (обычно localhost:2700).
// isMock — если true, возвращает тестовый текст без подключения к Vosk.
func NewVoskClient(host string, port int, isMock bool, logger *zap.Logger) *VoskClient {
	return &VoskClient{
		wsURL:  fmt.Sprintf("ws://%s:%d", host, port),
		logger: logger.Named("vosk_stt"),
		isMock: isMock,
	}
}

// Transcribe отправляет аудиофайл на распознавание через Vosk-server и возвращает текст.
// Пайплайн:
//  1. ffmpeg конвертирует аудио (mp3/wav/webm/ogg) в PCM 16kHz mono 16-bit
//  2. PCM-данные отправляются чанками по WebSocket
//  3. Vosk-server возвращает partial и final results
//  4. Все final-результаты конкатенируются в итоговый текст
//
// В mock-режиме возвращает детерминированный тестовый текст.
func (v *VoskClient) Transcribe(ctx context.Context, audioReader io.Reader, filename string, fileSize int64) (string, error) {
	// Mock-режим.
	if v.isMock {
		v.logger.Info("mock-режим: возвращаем тестовую транскрипцию")
		return mockTranscription(), nil
	}

	v.logger.Info("начало распознавания через Vosk",
		zap.String("filename", filename),
		zap.Int64("size", fileSize),
		zap.String("vosk_url", v.wsURL),
	)

	// Шаг 1: Конвертация аудио в PCM 16kHz mono через ffmpeg.
	pcmData, err := v.convertToPCM(ctx, audioReader)
	if err != nil {
		return "", fmt.Errorf("ошибка конвертации аудио в PCM: %w", err)
	}

	v.logger.Debug("аудио сконвертировано в PCM",
		zap.Int("pcm_bytes", len(pcmData)),
	)

	// Шаг 2: Подключение к Vosk-server по WebSocket.
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, v.wsURL, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка подключения к Vosk-server (%s): %w", v.wsURL, err)
	}
	defer conn.Close()

	// Шаг 3: Отправка конфигурации (sample rate).
	configMsg := map[string]any{
		"config": map[string]any{
			"sample_rate": 16000,
		},
	}
	if err := conn.WriteJSON(configMsg); err != nil {
		return "", fmt.Errorf("ошибка отправки конфигурации Vosk: %w", err)
	}

	// Шаг 4: Отправка PCM-данных чанками.
	for offset := 0; offset < len(pcmData); offset += voskChunkSize {
		end := offset + voskChunkSize
		if end > len(pcmData) {
			end = len(pcmData)
		}
		chunk := pcmData[offset:end]

		if err := conn.WriteMessage(websocket.BinaryMessage, chunk); err != nil {
			return "", fmt.Errorf("ошибка отправки аудио-чанка Vosk: %w", err)
		}

		// Читаем partial results (и игнорируем — нас интересуют только final).
		var partial voskResult
		if err := conn.ReadJSON(&partial); err != nil {
			return "", fmt.Errorf("ошибка чтения ответа Vosk: %w", err)
		}
	}

	// Шаг 5: Сигнал окончания аудио (EOF).
	if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"eof": 1}`)); err != nil {
		return "", fmt.Errorf("ошибка отправки EOF Vosk: %w", err)
	}

	// Шаг 6: Чтение финального результата.
	var finalResult voskResult
	if err := conn.ReadJSON(&finalResult); err != nil {
		return "", fmt.Errorf("ошибка чтения финального ответа Vosk: %w", err)
	}

	text := finalResult.Text
	if text == "" {
		text = finalResult.Partial
	}

	v.logger.Info("распознавание завершено",
		zap.Int("text_length", len(text)),
		zap.String("text_preview", truncateStr(text, 100)),
	)

	return text, nil
}

// voskResult — ответ Vosk-server (partial или final).
type voskResult struct {
	// Partial — промежуточный результат.
	Partial string `json:"partial,omitempty"`

	// Text — финальный результат.
	Text string `json:"text,omitempty"`
}

// convertToPCM конвертирует аудио из любого формата в PCM 16kHz mono 16-bit LE
// через ffmpeg. Принимает io.Reader с аудиоданными, возвращает байты PCM.
// Использует временный файл вместо pipe для надёжной обработки webm/matroska.
func (v *VoskClient) convertToPCM(ctx context.Context, audioReader io.Reader) ([]byte, error) {
	// Читаем всё аудио в память.
	audioData, err := io.ReadAll(audioReader)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения аудиоданных: %w", err)
	}

	if len(audioData) == 0 {
		return nil, fmt.Errorf("пустые аудиоданные (0 байт)")
	}

	v.logger.Debug("аудиоданные получены для конвертации",
		zap.Int("size_bytes", len(audioData)),
	)

	// Создаём временный файл для входных данных.
	// Pipe (stdin) ненадёжен для контейнерных форматов (webm/matroska),
	// так как ffmpeg не может seek-ить для чтения EBML-заголовков.
	tmpFile, err := os.CreateTemp("", "audio-*.webm")
	if err != nil {
		return nil, fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(audioData); err != nil {
		return nil, fmt.Errorf("ошибка записи аудио во временный файл: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("ошибка закрытия временного файла: %w", err)
	}

	// ffmpeg: читает из файла, выводит PCM s16le 16kHz mono в stdout.
	// Таймаут 30 секунд для конвертации.
	convertCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(convertCtx, "ffmpeg",
		"-y",                 // перезаписывать выходные файлы
		"-loglevel", "error", // минимальный вывод
		"-i", tmpFile.Name(), // вход из файла (не pipe)
		"-ar", "16000", // sample rate 16kHz
		"-ac", "1", // mono
		"-f", "s16le", // формат: signed 16-bit little-endian PCM
		"-acodec", "pcm_s16le", // кодек
		"pipe:1", // выход в stdout
	)

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		v.logger.Error("ошибка ffmpeg",
			zap.Error(err),
			zap.String("stderr", errBuf.String()),
			zap.Int("input_size", len(audioData)),
		)
		return nil, fmt.Errorf("ffmpeg error: %w (stderr: %s)", err, errBuf.String())
	}

	if outBuf.Len() == 0 {
		return nil, fmt.Errorf("ffmpeg вернул пустые PCM данные (вход: %d байт)", len(audioData))
	}

	return outBuf.Bytes(), nil
}

// truncateStr обрезает строку до maxLen символов.
func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// voskResultFromBytes парсит JSON-ответ Vosk.
func voskResultFromBytes(data []byte) (*voskResult, error) {
	var result voskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
