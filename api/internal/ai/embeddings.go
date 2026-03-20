// Файл embeddings.go реализует клиент для генерации векторных эмбеддингов текста.
// Отправляет текст осей vibe-профиля в /v1/embeddings и возвращает вектор 3072 измерений
// для сохранения в Qdrant.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// EmbeddingsClient взаимодействует с API для генерации векторов (Embeddings).
// Настраивается в config.go (переопределяется EmbeddingsBaseURL).
type EmbeddingsClient struct {
	client  *Client
	baseURL string
	model   string
	logger  *zap.Logger
}

// NewEmbeddingsClient создает новый экземпляр EmbeddingsClient.
func NewEmbeddingsClient(client *Client, model string, embeddingsBaseURL string, logger *zap.Logger) *EmbeddingsClient {
	return &EmbeddingsClient{
		client:  client,
		baseURL: embeddingsBaseURL,
		model:   model,
		logger:  logger.Named("embeddings"),
	}
}

// Generate генерирует вектор эмбеддинга для текста.
// Возвращает []float32 размерностью 3072 (для text-embedding-3-large).
// В mock-режиме возвращает детерминированный вектор.
func (e *EmbeddingsClient) Generate(ctx context.Context, text string) ([]float32, error) {
	// Mock-режим.
	if e.client.IsMock() {
		e.logger.Info("mock-режим: возвращаем тестовый вектор")
		return mockEmbedding(), nil
	}

	e.logger.Info("генерация эмбеддинга",
		zap.Int("text_length", len(text)),
		zap.String("model", e.model),
	)

	// Формирование запроса.
	reqBody := EmbeddingRequest{
		Model: e.model,
		Input: text,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации запроса эмбеддингов: %w", err)
	}

	// Создание HTTP-запроса.
	var url string
	if e.baseURL != "" {
		url = strings.TrimRight(e.baseURL, "/") + "/embeddings"
	} else {
		url = e.client.buildURL("embeddings")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания HTTP-запроса: %w", err)
	}

	e.client.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	// Выполнение запроса.
	resp, err := e.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка вызова Embeddings API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа Embeddings API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		e.logger.Error("ошибка Embeddings API",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return nil, fmt.Errorf("Embeddings API вернул статус %d: %s", resp.StatusCode, string(respBody))
	}

	var embResp EmbeddingResponse
	if err := json.Unmarshal(respBody, &embResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа Embeddings API: %w", err)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("Embeddings API вернул пустой массив data")
	}

	vector := embResp.Data[0].Embedding

	e.logger.Info("эмбеддинг сгенерирован",
		zap.Int("vector_size", len(vector)),
	)

	return vector, nil
}
