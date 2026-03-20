// Файл models.go содержит структуры данных для взаимодействия с AI API.
// Определяет типы запросов и ответов для Whisper STT, LLM chat completions
// и Embeddings API в формате, совместимом с OpenAI API.
package ai

// VibeAxes — структура осей vibe-профиля пользователя.
// Извлекается из речи через LLM и описывает предпочтения туриста.
// Оси соответствуют GDD (kudytudy_gdd.md, раздел Voice-to-Vibe).
type VibeAxes struct {
	// StressLevel — уровень стресса (0.0 = спокоен, 1.0 = очень устал/напряжён).
	StressLevel float64 `json:"stress_level"`

	// SolitudeVsSocial — уединение vs компания (-1.0 = полное уединение, 1.0 = шумная компания).
	SolitudeVsSocial float64 `json:"solitude_vs_social"`

	// RelaxVsAdrenaline — релакс vs адреналин (-1.0 = полный релакс/спа, 1.0 = экстрим/адреналин).
	RelaxVsAdrenaline float64 `json:"relax_vs_adrenaline"`

	// GastroVsNature — гастрономия vs природа (-1.0 = чистая гастрономия, 1.0 = чистая природа).
	GastroVsNature float64 `json:"gastro_vs_nature"`

	// CultureVsAdventure — культура vs приключения (-1.0 = музеи/история, 1.0 = походы/исследование).
	CultureVsAdventure float64 `json:"culture_vs_adventure"`

	// ExtractedTags — теги предпочтений, извлечённые из речи (вино, горы, тишина, ферма...).
	ExtractedTags []string `json:"extracted_tags"`

	// VibeSummary — краткое резюме настроения на русском языке.
	VibeSummary string `json:"vibe_summary"`
}

// ChatCompletionRequest — запрос к /v1/chat/completions (OpenAI-совместимый).
type ChatCompletionRequest struct {
	// Model — идентификатор модели (gpt-4o-mini, gemini-2.5-flash и т.д.).
	Model string `json:"model"`

	// Messages — массив сообщений диалога.
	Messages []ChatMessage `json:"messages"`

	// Temperature — степень случайности ответа (0.0-2.0).
	Temperature float64 `json:"temperature,omitempty"`

	// MaxTokens — максимальное количество токенов в ответе.
	MaxTokens int `json:"max_tokens,omitempty"`
}

// ChatMessage — одно сообщение в диалоге.
type ChatMessage struct {
	// Role — роль отправителя: system, user, assistant.
	Role string `json:"role"`

	// Content — текст сообщения.
	Content string `json:"content"`
}

// ChatCompletionResponse — ответ от /v1/chat/completions.
type ChatCompletionResponse struct {
	// Choices — массив вариантов ответа.
	Choices []ChatChoice `json:"choices"`

	// Usage — статистика использования токенов.
	Usage TokenUsage `json:"usage"`
}

// ChatChoice — один вариант ответа модели.
type ChatChoice struct {
	// Message — ответное сообщение модели.
	Message ChatMessage `json:"message"`

	// FinishReason — причина завершения: stop, length, content_filter.
	FinishReason string `json:"finish_reason"`
}

// EmbeddingRequest — запрос к /v1/embeddings (OpenAI-совместимый).
type EmbeddingRequest struct {
	// Model — идентификатор модели эмбеддингов.
	Model string `json:"model"`

	// Input — текст для генерации вектора.
	Input string `json:"input"`
}

// EmbeddingResponse — ответ от /v1/embeddings.
type EmbeddingResponse struct {
	// Data — массив с результатами эмбеддинга.
	Data []EmbeddingData `json:"data"`

	// Usage — статистика использования токенов.
	Usage TokenUsage `json:"usage"`
}

// EmbeddingData — один результат эмбеддинга.
type EmbeddingData struct {
	// Embedding — вектор эмбеддинга ([]float32 для экономии памяти).
	Embedding []float32 `json:"embedding"`

	// Index — индекс входного текста.
	Index int `json:"index"`
}

// TranscriptionRequest — параметры запроса к /v1/audio/transcriptions.
// Файл передаётся через multipart/form-data, а не JSON.
type TranscriptionRequest struct {
	// Model — идентификатор модели (whisper-1).
	Model string

	// Language — язык аудио (ru для русского).
	Language string
}

// TranscriptionResponse — ответ от /v1/audio/transcriptions.
type TranscriptionResponse struct {
	// Text — распознанный текст.
	Text string `json:"text"`
}

// TokenUsage — статистика использования токенов.
type TokenUsage struct {
	// PromptTokens — токены в запросе.
	PromptTokens int `json:"prompt_tokens"`

	// CompletionTokens — токены в ответе.
	CompletionTokens int `json:"completion_tokens"`

	// TotalTokens — общее количество токенов.
	TotalTokens int `json:"total_tokens"`
}
