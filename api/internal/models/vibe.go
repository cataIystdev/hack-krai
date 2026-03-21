// Файл vibe.go содержит модели данных для модуля vibe-профилирования.
// Определяет структуры для сцен свайпа, запросов/ответов API профилирования
// и рекомендаций по локациям.
package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// SwipeDirection — направление свайпа.
type SwipeDirection string

const (
	// SwipeRight — свайп вправо (нравится). Вектор сдвигается к сцене.
	SwipeRight SwipeDirection = "right"

	// SwipeLeft — свайп влево (не нравится). Вектор сдвигается от сцены.
	SwipeLeft SwipeDirection = "left"
)

// ValidSwipeDirections — набор допустимых направлений свайпа.
var ValidSwipeDirections = map[SwipeDirection]bool{
	SwipeRight: true,
	SwipeLeft:  true,
}

// SwipeScene — сцена для свайп-анкеты.
// Каждая сцена представляет визуальный образ с описанием и ассоциированным vibe-вектором.
type SwipeScene struct {
	// ID — уникальный идентификатор сцены.
	ID uuid.UUID `json:"id" db:"id"`

	// Title — заголовок сцены.
	Title string `json:"title" db:"title"`

	// Description — описание сцены.
	Description string `json:"description" db:"description"`

	// ImageURL — URL изображения сцены.
	ImageURL string `json:"image_url" db:"image_url"`

	// DisplayOrder — порядок отображения.
	DisplayOrder int `json:"display_order" db:"display_order"`

	// CreatedAt — временная метка создания.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// SwipeRequest — запрос свайпа сцены.
type SwipeRequest struct {
	// SceneID — UUID сцены.
	SceneID string `json:"scene_id"`

	// Direction — направление свайпа: right (нравится) или left (не нравится).
	Direction string `json:"direction"`
}

// Validate проверяет корректность запроса свайпа.
func (r *SwipeRequest) Validate() error {
	if r.SceneID == "" {
		return errors.New("scene_id обязателен")
	}
	if _, err := uuid.Parse(r.SceneID); err != nil {
		return errors.New("scene_id должен быть валидным UUID")
	}
	if r.Direction == "" {
		return errors.New("direction обязателен")
	}
	if !ValidSwipeDirections[SwipeDirection(r.Direction)] {
		return errors.New("direction должен быть 'right' или 'left'")
	}
	return nil
}

// VibeAxesResponse — оси vibe-профиля пользователя в JSON-ответе.
// Вложенная структура для читаемого JSON формата на фронтенде.
type VibeAxesResponse struct {
	// StressLevel — уровень стресса (0.0-1.0).
	StressLevel float64 `json:"stress_level"`

	// SolitudeVsSocial — уединение vs компания (-1.0 to 1.0).
	SolitudeVsSocial float64 `json:"solitude_vs_social"`

	// RelaxVsAdrenaline — релакс vs адреналин (-1.0 to 1.0).
	RelaxVsAdrenaline float64 `json:"relax_vs_adrenaline"`

	// GastroVsNature — гастрономия vs природа (-1.0 to 1.0).
	GastroVsNature float64 `json:"gastro_vs_nature"`

	// CultureVsAdventure — культура vs приключения (-1.0 to 1.0).
	CultureVsAdventure float64 `json:"culture_vs_adventure"`
}

// VoiceProfileResponse — ответ на запрос голосового профилирования.
// Screenshot-ready формат: фронтенд может рендерить vibe passport без обработки.
type VoiceProfileResponse struct {
	// Transcription — распознанный текст из аудио.
	Transcription string `json:"transcription"`

	// Axes — оси vibe-профиля (вложенный объект).
	Axes VibeAxesResponse `json:"axes"`

	// ExtractedTags — теги предпочтений, извлечённые из речи.
	ExtractedTags []string `json:"extracted_tags"`

	// VibeSummary — краткое резюме настроения на русском.
	VibeSummary string `json:"vibe_summary"`

	// VibePassportTitle — заголовок vibe-паспорта для отображения на экране.
	// Генерируется AI или mock: "Исследователь Кубани", "Гастрономический путешественник".
	VibePassportTitle string `json:"vibe_passport_title"`

	// VectorID — ID вектора в Qdrant (user_id).
	VectorID string `json:"vector_id"`

	// ProcessingTimeMs — время обработки пайплайна в миллисекундах.
	// Включает STT + LLM + Embeddings + Qdrant upsert.
	ProcessingTimeMs int64 `json:"processing_time_ms"`
}

// LocationRecommendation — рекомендация локации на основе vibe-профиля.
// Содержит все поля, необходимые фронтенду для рендеринга карточки рекомендации,
// включая reason_short (причина рекомендации) и tags_match (совпавшие теги).
type LocationRecommendation struct {
	// LocationID — UUID локации.
	LocationID string `json:"location_id"`

	// Score — оценка сходства (cosine similarity, 0.0-1.0).
	Score float32 `json:"score"`

	// Name — название локации.
	Name string `json:"name"`

	// Category — категория локации (winery, farm, trail и др.).
	Category string `json:"category"`

	// DescriptionShort — краткое описание для карточки.
	DescriptionShort string `json:"description_short"`

	// Tags — теги локации для отображения.
	Tags []string `json:"tags"`

	// PreviewImageURL — URL hero-изображения локации.
	PreviewImageURL string `json:"preview_image_url"`

	// SplatURL — URL 3D-сцены (.splat) для immersive preview.
	SplatURL string `json:"splat_url,omitempty"`

	// Latitude — широта для отображения на карте.
	Latitude float64 `json:"latitude"`

	// Longitude — долгота для отображения на карте.
	Longitude float64 `json:"longitude"`

	// DensityLevel — уровень туристической плотности (red/yellow/green).
	DensityLevel string `json:"density_level"`

	// ReasonShort — краткая причина рекомендации на русском языке.
	// Генерируется на основе cosine score и совпадения тегов.
	// Пример: "Высокое совпадение: природа, спокойствие".
	ReasonShort string `json:"reason_short"`

	// TagsMatch — теги, совпавшие между профилем пользователя и локацией.
	// Пустой массив, если пересечения нет.
	TagsMatch []string `json:"tags_match"`

	// ChildFriendly — подходит ли локация для детей.
	ChildFriendly bool `json:"child_friendly"`
}

// FinalizeRequest — опциональные параметры запроса финализации профиля.
// Все поля опциональны. При пустом теле запроса используются значения по умолчанию.
type FinalizeRequest struct {
	// Limit — максимальное количество рекомендаций (по умолчанию 10, максимум 50).
	Limit int `json:"limit,omitempty"`

	// ChildFriendlyOnly — фильтровать только child-friendly локации.
	ChildFriendlyOnly bool `json:"child_friendly_only,omitempty"`
}

// NormalizeDefaults устанавливает значения по умолчанию для FinalizeRequest.
func (r *FinalizeRequest) NormalizeDefaults() {
	if r.Limit <= 0 {
		r.Limit = 10
	}
	if r.Limit > 50 {
		r.Limit = 50
	}
}

// FinalizeResponse — ответ на запрос финализации профиля.
type FinalizeResponse struct {
	// Recommendations — Top-N ближайших локаций.
	Recommendations []LocationRecommendation `json:"recommendations"`

	// TotalFound — общее количество найденных совпадений.
	TotalFound int `json:"total_found"`

	// IsCurated — true, если рекомендации из curated demo набора
	// (вектор пользователя не найден, используется fallback).
	IsCurated bool `json:"is_curated"`
}
