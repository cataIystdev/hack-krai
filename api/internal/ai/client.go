// Файл client.go реализует базовый HTTP-клиент для OnlySQ API.
// OnlySQ API совместим с форматом OpenAI API (тот же формат запросов и ответов).
// Клиент используется как основа для Whisper, LLM и Embeddings клиентов.
package ai

import (
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"deep-krai-api/internal/config"
)

// Client — базовый HTTP-клиент для OnlySQ / OpenAI-совместимого API.
// Содержит настройки подключения, HTTP-клиент и логгер.
type Client struct {
	// baseURL — базовый URL API (например, https://api.onlysq.ru/ai/openai/).
	baseURL string

	// apiKey — ключ авторизации (Bearer token).
	apiKey string

	// httpClient — HTTP-клиент с настроенными таймаутами.
	httpClient *http.Client

	// logger — логгер для записи операций.
	logger *zap.Logger

	// isMock — флаг mock-режима (без реального API).
	isMock bool
}

// NewClient создаёт базовый AI-клиент из конфигурации.
// Если API-ключ пустой — клиент работает в mock-режиме.
func NewClient(cfg config.AIConfig, logger *zap.Logger) *Client {
	baseURL := cfg.BaseURL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		logger: logger.Named("ai_client"),
		isMock: cfg.IsMockMode(),
	}
}

// IsMock возвращает true, если клиент работает в mock-режиме.
func (c *Client) IsMock() bool {
	return c.isMock
}

// buildURL формирует полный URL для API-вызова.
func (c *Client) buildURL(path string) string {
	path = strings.TrimPrefix(path, "/")
	return c.baseURL + path
}

// setAuthHeaders устанавливает заголовки авторизации и Content-Type.
func (c *Client) setAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
}
