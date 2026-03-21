// Файл llm.go реализует клиент для извлечения осей vibe-профиля из текста через LLM.
// Отправляет текст транскрипции в /v1/chat/completions с system prompt,
// задающим структуру JSON-ответа (оси предпочтений туриста).
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"kudytudy-api/internal/config"
)

// vibeSystemPrompt — системный промпт для LLM, задающий формат ответа.
// Оси и формат JSON соответствуют GDD (kudytudy_gdd.md).
const vibeSystemPrompt = `Ты — AI-профайлер туристической платформы КудыТуды. Проанализируй текст пользователя
и определи его эмоциональное состояние и предпочтения. Верни ТОЛЬКО JSON (без markdown, без пояснений):

{
  "stress_level": 0.0-1.0,
  "solitude_vs_social": -1.0 to 1.0,
  "relax_vs_adrenaline": -1.0 to 1.0,
  "gastro_vs_nature": -1.0 to 1.0,
  "culture_vs_adventure": -1.0 to 1.0,
  "extracted_tags": ["тег1", "тег2", ...],
  "vibe_summary": "Одно предложение о настроении"
}

Описание осей:
- stress_level: 0.0 = спокоен и расслаблен, 1.0 = очень устал/напряжён, нужна перезарядка
- solitude_vs_social: -1.0 = полное уединение и тишина, 1.0 = шумная компания, тусовки, общение
- relax_vs_adrenaline: -1.0 = релакс, спа, пляж, медитация, 1.0 = экстрим, адреналин, рафтинг
- gastro_vs_nature: -1.0 = гастрономия, вино, рестораны, дегустации, 1.0 = чистая природа, лес, горы
- culture_vs_adventure: -1.0 = музеи, история, архитектура, крепости, 1.0 = походы, исследование, каякинг
- extracted_tags: 3-8 тегов из текста (вино, горы, море, ферма, тишина, лес, экстрим и т.д.)
- vibe_summary: одно предложение на русском, описывающее настроение и предпочтения`

// LLMClient — клиент для извлечения осей vibe-профиля через LLM API.
type LLMClient struct {
	// client — базовый AI-клиент.
	client *Client

	// model — идентификатор модели (gpt-4o-mini).
	model string

	// logger — логгер.
	logger *zap.Logger
}

// NewLLMClient создаёт клиент для извлечения осей.
func NewLLMClient(client *Client, cfg config.AIConfig, logger *zap.Logger) *LLMClient {
	return &LLMClient{
		client: client,
		model:  cfg.LLMModel,
		logger: logger.Named("llm"),
	}
}

// ExtractVibeAxes извлекает оси vibe-профиля из текста транскрипции.
// Отправляет текст в LLM с system prompt, парсит JSON-ответ.
// В mock-режиме возвращает детерминированные оси.
func (l *LLMClient) ExtractVibeAxes(ctx context.Context, text string) (*VibeAxes, error) {
	// Mock-режим.
	if l.client.IsMock() {
		l.logger.Info("mock-режим: возвращаем тестовые оси")
		return mockVibeAxes(), nil
	}

	l.logger.Info("извлечение осей vibe-профиля",
		zap.Int("text_length", len(text)),
		zap.String("model", l.model),
	)

	// Формирование запроса.
	reqBody := ChatCompletionRequest{
		Model: l.model,
		Messages: []ChatMessage{
			{Role: "system", Content: vibeSystemPrompt},
			{Role: "user", Content: text},
		},
		Temperature: 0.3,
		MaxTokens:   1000,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации запроса LLM: %w", err)
	}

	// Создание HTTP-запроса.
	url := l.client.buildURL("chat/completions")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания HTTP-запроса: %w", err)
	}

	l.client.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	// Выполнение запроса.
	resp, err := l.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка вызова LLM API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа LLM API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		l.logger.Error("ошибка LLM API",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return nil, fmt.Errorf("LLM API вернул статус %d: %s", resp.StatusCode, string(respBody))
	}

	// Парсинг ответа chat/completions.
	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа LLM API: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM API вернул пустой массив choices")
	}

	// Извлечение JSON из ответа модели.
	content := chatResp.Choices[0].Message.Content

	// Парсинг JSON осей из текста ответа.
	var axes VibeAxes
	if err := json.Unmarshal([]byte(content), &axes); err != nil {
		// Попытка найти JSON в ответе (модель может обернуть в markdown).
		axes, err = extractJSONFromText(content)
		if err != nil {
			l.logger.Error("не удалось извлечь JSON из ответа LLM",
				zap.String("content", content),
				zap.Error(err),
			)
			return nil, fmt.Errorf("ошибка парсинга осей из ответа LLM: %w", err)
		}
	}

	l.logger.Info("оси vibe-профиля извлечены",
		zap.Float64("stress_level", axes.StressLevel),
		zap.Float64("solitude_vs_social", axes.SolitudeVsSocial),
		zap.Float64("relax_vs_adrenaline", axes.RelaxVsAdrenaline),
		zap.Int("tags_count", len(axes.ExtractedTags)),
	)

	return &axes, nil
}

// extractJSONFromText пытается извлечь JSON из текста, который может содержать
// markdown-обёртку (```json ... ```) или другой текст вокруг JSON.
func extractJSONFromText(text string) (VibeAxes, error) {
	var axes VibeAxes

	// Поиск JSON между фигурными скобками.
	start := -1
	depth := 0
	for i, ch := range text {
		if ch == '{' {
			if depth == 0 {
				start = i
			}
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 && start >= 0 {
				jsonStr := text[start : i+1]
				if err := json.Unmarshal([]byte(jsonStr), &axes); err == nil {
					return axes, nil
				}
			}
		}
	}

	return axes, fmt.Errorf("JSON не найден в тексте ответа")
}

// onboardingSystemPrompt — системный промпт для извлечения
// структурированных данных локации из голосового описания хоста.
// Формат соответствует GDD Feature 5 (Zero-UI Онбординг).
const onboardingSystemPrompt = `Ты — AI-ассистент туристической платформы КудыТуды. Владелец локации (фермер, хозяин усадьбы и т.д.) описал своё место голосом. Проанализируй текст и извлеки структурированные данные для создания карточки локации.

Верни ТОЛЬКО JSON (без markdown, без пояснений):

{
  "name_suggestion": "Красивое название локации (придумай на основе описания)",
  "description_short": "Краткое описание для карточки (1-2 предложения)",
  "description_literary": "Полное литературное описание (3-7 предложений, переписанное из разговорной речи в приятный текст)",
  "tags": ["тег1", "тег2", "тег3 и больше"],
  "category": "категория",
  "price_per_night": 0,
  "amenities": ["удобство1", "удобство2 и больше"],
  "capacity": 0
}

Правила:
- name_suggestion: придумай привлекательное название на основе описания, как для туристического сайта
- description_short: 1-2 предложения, ёмко и привлекательно
- description_literary: переведи разговорную речь в литературный текст, убери мат, жаргон, сделай приятным для чтения
- tags: 3-8 тегов на русском (вино, сыр, горы, козы, ферма, тишина, природа, ночлег, дегустация и т.д.)
- category: одна из: farm, winery, trail, guesthouse, restaurant, camping, excursion, workshop, other
- price_per_night: цена за ночь в рублях (целое число), 0 если не упоминается
- amenities: удобства, извлечённые из речи (парковка, Wi-Fi, домики, баня и т.д.)
- capacity: максимальное количество гостей (целое число), 0 если не упоминается`

// ExtractLocationData извлекает структурированные данные локации
// из текста транскрипции голосового описания хоста.
// Используется в pipeline онбординга (GDD Feature 5).
// В mock-режиме возвращает детерминированные данные "Козья ферма дяди Вани".
func (l *LLMClient) ExtractLocationData(ctx context.Context, text string) (*LocationData, error) {
	// Mock-режим.
	if l.client.IsMock() {
		l.logger.Info("mock-режим: возвращаем тестовые данные локации")
		return mockLocationData(), nil
	}

	l.logger.Info("извлечение данных локации из описания хоста",
		zap.Int("text_length", len(text)),
		zap.String("model", l.model),
	)

	// Формирование запроса.
	reqBody := ChatCompletionRequest{
		Model: l.model,
		Messages: []ChatMessage{
			{Role: "system", Content: onboardingSystemPrompt},
			{Role: "user", Content: text},
		},
		Temperature: 0.4,
		MaxTokens:   1500,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации запроса LLM: %w", err)
	}

	// Создание HTTP-запроса.
	url := l.client.buildURL("chat/completions")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания HTTP-запроса: %w", err)
	}

	l.client.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	// Выполнение запроса.
	resp, err := l.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка вызова LLM API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа LLM API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		l.logger.Error("ошибка LLM API при извлечении данных локации",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return nil, fmt.Errorf("LLM API вернул статус %d: %s", resp.StatusCode, string(respBody))
	}

	// Парсинг ответа chat/completions.
	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа LLM API: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM API вернул пустой массив choices")
	}

	content := chatResp.Choices[0].Message.Content

	// Парсинг JSON из ответа модели.
	var data LocationData
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		// Попытка найти JSON в ответе (модель может обернуть в markdown).
		data, err = extractLocationJSONFromText(content)
		if err != nil {
			l.logger.Error("не удалось извлечь JSON данных локации из ответа LLM",
				zap.String("content", content),
				zap.Error(err),
			)
			return nil, fmt.Errorf("ошибка парсинга данных локации из ответа LLM: %w", err)
		}
	}

	l.logger.Info("данные локации извлечены",
		zap.String("name", data.NameSuggestion),
		zap.String("category", data.Category),
		zap.Int("tags_count", len(data.Tags)),
		zap.Int("price", data.PricePerNight),
	)

	return &data, nil
}

// LocationData — структурированные данные локации, извлечённые LLM.
// Повторяет формат из models.OnboardingResult для использования в AI-слое.
type LocationData struct {
	// NameSuggestion — предложенное название.
	NameSuggestion string `json:"name_suggestion"`

	// DescriptionShort — краткое описание.
	DescriptionShort string `json:"description_short"`

	// DescriptionLiterary — литературное описание.
	DescriptionLiterary string `json:"description_literary"`

	// Tags — теги.
	Tags []string `json:"tags"`

	// Category — категория.
	Category string `json:"category"`

	// PricePerNight — цена за ночь.
	PricePerNight int `json:"price_per_night"`

	// Amenities — удобства.
	Amenities []string `json:"amenities,omitempty"`

	// Capacity — вместимость.
	Capacity int `json:"capacity"`
}

// extractLocationJSONFromText извлекает JSON LocationData из текста ответа LLM.
func extractLocationJSONFromText(text string) (LocationData, error) {
	var data LocationData
	start := -1
	depth := 0
	for i, ch := range text {
		if ch == '{' {
			if depth == 0 {
				start = i
			}
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 && start >= 0 {
				jsonStr := text[start : i+1]
				if err := json.Unmarshal([]byte(jsonStr), &data); err == nil {
					return data, nil
				}
			}
		}
	}
	return data, fmt.Errorf("JSON данных локации не найден в тексте ответа")
}
