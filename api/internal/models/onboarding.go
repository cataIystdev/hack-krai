// Файл onboarding.go определяет модели данных для Zero-UI онбординга хостов.
// Содержит структуры задач AI-обработки (AITask), запросов/ответов онбординга
// и результатов извлечения данных локации через LLM.
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TaskType — тип задачи AI-обработки.
type TaskType string

const (
	// TaskTypeOnboarding — онбординг хоста (voice + media -> location draft).
	TaskTypeOnboarding TaskType = "onboarding"

	// TaskTypeSplatting — генерация 3D-сцены из видео (Gaussian Splatting).
	TaskTypeSplatting TaskType = "splatting"

	// TaskTypeStoryGeneration — генерация аудио-истории для POI.
	TaskTypeStoryGeneration TaskType = "story_generation"

	// TaskTypeVoiceProfile — голосовое профилирование туриста.
	TaskTypeVoiceProfile TaskType = "voice_profile"
)

// TaskStatus — статус выполнения задачи.
type TaskStatus string

const (
	// TaskStatusQueued — задача в очереди, ожидает обработки.
	TaskStatusQueued TaskStatus = "queued"

	// TaskStatusProcessing — задача обрабатывается.
	TaskStatusProcessing TaskStatus = "processing"

	// TaskStatusCompleted — задача успешно завершена.
	TaskStatusCompleted TaskStatus = "completed"

	// TaskStatusFailed — задача завершилась с ошибкой.
	TaskStatusFailed TaskStatus = "failed"
)

// AITask — задача AI-обработки.
// Используется для отслеживания прогресса pipeline онбординга.
type AITask struct {
	// ID — уникальный идентификатор задачи.
	ID uuid.UUID `json:"id" db:"id"`

	// UserID — идентификатор пользователя, создавшего задачу.
	UserID uuid.UUID `json:"user_id" db:"user_id"`

	// Type — тип задачи (onboarding, splatting и др.).
	Type TaskType `json:"type" db:"type"`

	// Status — текущий статус (queued, processing, completed, failed).
	Status TaskStatus `json:"status" db:"status"`

	// Progress — прогресс выполнения (0-100).
	Progress int `json:"progress" db:"progress"`

	// InputData — входные данные (JSONB): audio_url, media_urls, координаты.
	InputData json.RawMessage `json:"input_data,omitempty" db:"input_data"`

	// OutputData — результат выполнения (JSONB): location_id, extracted_data.
	OutputData json.RawMessage `json:"output_data,omitempty" db:"output_data"`

	// Error — текст ошибки (при status = 'failed').
	Error *string `json:"error,omitempty" db:"error"`

	// CreatedAt — временная метка создания.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt — временная метка последнего обновления.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// CompletedAt — временная метка завершения.
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// OnboardingInputData — входные данные для задачи онбординга.
// Сохраняется в ai_tasks.input_data (JSONB).
type OnboardingInputData struct {
	// AudioURL — URL загруженного аудиофайла в MinIO.
	AudioURL string `json:"audio_url"`

	// MediaURLs — URL загруженных медиафайлов (фото/видео двора).
	MediaURLs []string `json:"media_urls,omitempty"`

	// Latitude — широта локации (опционально, задаётся хостом).
	Latitude *float64 `json:"latitude,omitempty"`

	// Longitude — долгота локации (опционально, задаётся хостом).
	Longitude *float64 `json:"longitude,omitempty"`

	// Address — адрес локации (опционально, задаётся хостом).
	Address string `json:"address,omitempty"`
}

// OnboardingOutputData — результат задачи онбординга.
// Сохраняется в ai_tasks.output_data (JSONB).
type OnboardingOutputData struct {
	// LocationID — UUID созданного черновика локации.
	LocationID uuid.UUID `json:"location_id"`

	// Transcription — распознанный текст из аудио.
	Transcription string `json:"transcription"`

	// ExtractedData — структурированные данные, извлечённые LLM из транскрипции.
	ExtractedData *OnboardingResult `json:"extracted_data"`
}

// OnboardingResult — структурированные данные локации, извлечённые LLM.
// Соответствует формату из GDD Feature 5.
type OnboardingResult struct {
	// NameSuggestion — предложенное название локации.
	NameSuggestion string `json:"name_suggestion"`

	// DescriptionShort — краткое описание для карточки.
	DescriptionShort string `json:"description_short"`

	// DescriptionLiterary — литературное описание (переписанное из разговорной речи).
	DescriptionLiterary string `json:"description_literary"`

	// Tags — извлечённые теги (3-8 штук).
	Tags []string `json:"tags"`

	// Category — категория локации (farm, winery, trail, guesthouse и др.).
	Category string `json:"category"`

	// PricePerNight — стоимость за ночь (рубли, 0 если не упоминается).
	PricePerNight int `json:"price_per_night"`

	// Amenities — удобства, извлечённые из речи.
	Amenities []string `json:"amenities,omitempty"`

	// Capacity — вместимость (гостей, 0 если не упоминается).
	Capacity int `json:"capacity"`
}

// OnboardResponse — ответ на запрос онбординга.
type OnboardResponse struct {
	// TaskID — UUID созданной задачи для опроса статуса.
	TaskID uuid.UUID `json:"task_id"`

	// Status — текущий статус задачи.
	Status TaskStatus `json:"status"`

	// Progress — прогресс (0-100).
	Progress int `json:"progress"`

	// Message — человекочитаемое сообщение о текущем шаге.
	Message string `json:"message"`

	// LocationID — UUID созданной локации (заполняется при status=completed).
	LocationID *uuid.UUID `json:"location_id,omitempty"`

	// ExtractedData — извлечённые данные (заполняется при status=completed).
	ExtractedData *OnboardingResult `json:"extracted_data,omitempty"`
}

// ProgressMessage возвращает человекочитаемое описание текущего шага pipeline.
func ProgressMessage(progress int) string {
	switch {
	case progress <= 0:
		return "задача в очереди"
	case progress <= 20:
		return "распознавание речи..."
	case progress <= 50:
		return "анализ описания и извлечение данных..."
	case progress <= 70:
		return "генерация вайб-профиля локации..."
	case progress <= 90:
		return "создание черновика локации..."
	case progress < 100:
		return "финализация..."
	default:
		return "готово"
	}
}
