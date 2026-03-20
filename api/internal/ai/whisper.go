// Файл whisper.go реализует клиент для Whisper STT API (распознавание речи).
// Отправляет аудиофайл через multipart/form-data на /v1/audio/transcriptions
// и возвращает распознанный текст.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"go.uber.org/zap"

	"deep-krai-api/internal/config"
)

// WhisperClient — клиент для распознавания речи через Whisper API.
type WhisperClient struct {
	// client — базовый AI-клиент.
	client *Client

	// model — идентификатор модели Whisper (whisper-1).
	model string

	// logger — логгер.
	logger *zap.Logger
}

// NewWhisperClient создаёт клиент распознавания речи.
func NewWhisperClient(client *Client, cfg config.AIConfig, logger *zap.Logger) *WhisperClient {
	return &WhisperClient{
		client: client,
		model:  cfg.WhisperModel,
		logger: logger.Named("whisper"),
	}
}

// Transcribe отправляет аудиофайл на распознавание и возвращает текст.
// Параметры:
//   - ctx: контекст для отмены операции.
//   - audioReader: поток аудиоданных.
//   - filename: имя файла (для определения формата: mp3, wav, webm и т.д.).
//   - fileSize: размер файла в байтах.
//
// В mock-режиме возвращает детерминированный текст.
func (w *WhisperClient) Transcribe(ctx context.Context, audioReader io.Reader, filename string, fileSize int64) (string, error) {
	// Mock-режим: возвращаем тестовый текст.
	if w.client.IsMock() {
		w.logger.Info("mock-режим: возвращаем тестовую транскрипцию")
		return mockTranscription(), nil
	}

	w.logger.Info("отправка аудио на распознавание",
		zap.String("filename", filename),
		zap.Int64("size", fileSize),
		zap.String("model", w.model),
	)

	// Формирование multipart/form-data запроса.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Добавление аудиофайла.
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("ошибка создания multipart формы: %w", err)
	}

	if _, err := io.Copy(part, audioReader); err != nil {
		return "", fmt.Errorf("ошибка копирования аудиоданных: %w", err)
	}

	// Добавление параметров.
	if err := writer.WriteField("model", w.model); err != nil {
		return "", fmt.Errorf("ошибка записи поля model: %w", err)
	}
	if err := writer.WriteField("language", "ru"); err != nil {
		return "", fmt.Errorf("ошибка записи поля language: %w", err)
	}
	if err := writer.WriteField("response_format", "json"); err != nil {
		return "", fmt.Errorf("ошибка записи поля response_format: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("ошибка закрытия multipart writer: %w", err)
	}

	// Создание HTTP-запроса.
	url := w.client.buildURL("audio/transcriptions")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return "", fmt.Errorf("ошибка создания HTTP-запроса: %w", err)
	}

	w.client.setAuthHeaders(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Выполнение запроса.
	resp, err := w.client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка вызова Whisper API: %w", err)
	}
	defer resp.Body.Close()

	// Чтение и обработка ответа.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа Whisper API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		w.logger.Error("ошибка Whisper API",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return "", fmt.Errorf("Whisper API вернул статус %d: %s", resp.StatusCode, string(respBody))
	}

	var result TranscriptionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("ошибка парсинга ответа Whisper API: %w", err)
	}

	w.logger.Info("транскрипция завершена",
		zap.Int("text_length", len(result.Text)),
	)

	return result.Text, nil
}
