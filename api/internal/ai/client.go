// Файл client.go определяет базовый AI-клиент, используемый LLM, Embeddings и Whisper.
// Client инкапсулирует HTTP-клиент, базовый URL, API-ключ и mock-режим.
package ai

import (
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"kudytudy-api/internal/config"
)

// Client — базовый AI-клиент для взаимодействия с OpenAI-совместимым API.
// Инкапсулирует HTTP-клиент, авторизацию и режим работы (mock/real).
type Client struct {
	// httpClient — HTTP-клиент с настроенным таймаутом.
	httpClient *http.Client

	// baseURL — базовый URL API (например, https://api.onlysq.ru/ai/openai/).
	baseURL string

	// apiKey — ключ авторизации.
	apiKey string

	// mock — флаг mock-режима (true, если API-ключ не задан).
	mock bool

	// logger — логгер.
	logger *zap.Logger
}

// NewClient создаёт базовый AI-клиент на основе конфигурации.
// Если APIKey пуст, клиент работает в mock-режиме.
func NewClient(cfg config.AIConfig, logger *zap.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		mock:    cfg.IsMockMode(),
		logger:  logger.Named("ai_client"),
	}
}

// IsMock возвращает true, если клиент работает в mock-режиме.
func (c *Client) IsMock() bool {
	return c.mock
}

// buildURL строит полный URL для API-эндпоинта.
func (c *Client) buildURL(endpoint string) string {
	return c.baseURL + "/v1/" + endpoint
}

// setAuthHeaders устанавливает заголовки авторизации для HTTP-запроса.
func (c *Client) setAuthHeaders(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}
