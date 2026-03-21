// Файл route.go содержит модели данных для построения и хранения маршрутов.
// Определяет: BuildRouteRequest (ручной режим), BuildTripRouteRequest (trip-aware),
// RoutePoint (точка маршрута), RoutePreview (превью/ответ), Route и RoutePointDB (DB-модели).
package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Константы статусов маршрута.
const (
	// RouteStatusDraft — маршрут в черновике.
	RouteStatusDraft = "draft"

	// RouteStatusActive — маршрут активен (используется в поездке).
	RouteStatusActive = "active"

	// RouteStatusCompleted — маршрут завершён.
	RouteStatusCompleted = "completed"
)

// Константы тайм-слотов для точек маршрута.
const (
	// TimeSlotMorning — утренний слот (08:00-12:00).
	TimeSlotMorning = "morning"

	// TimeSlotAfternoon — дневной слот (12:00-17:00).
	TimeSlotAfternoon = "afternoon"

	// TimeSlotEvening — вечерний слот (17:00-21:00).
	TimeSlotEvening = "evening"
)

// Константы целевой аудитории для точек маршрута.
const (
	// AudienceAll — подходит всем участникам.
	AudienceAll = "all"

	// AudienceAdults — для взрослых участников.
	AudienceAdults = "adults"

	// AudienceChildren — для детей.
	AudienceChildren = "children"
)

// MaxRoutePointsPerFormat — максимальное количество точек маршрута по формату поездки.
var MaxRoutePointsPerFormat = map[string]int{
	"day_trip": 5,
	"weekend":  8,
	"multi_day": 15,
}

// BuildRouteRequest — DTO для запроса построения маршрута в ручном режиме.
// Принимает список ID локаций, вид транспорта и флаг оптимизации порядка.
type BuildRouteRequest struct {
	// LocationIDs — список UUID локаций для включения в маршрут. Минимум 2 точки.
	LocationIDs []string `json:"location_ids"`

	// Transport — вид транспорта: car, public, walk, bike. По умолчанию car.
	Transport string `json:"transport"`

	// Optimize — если true, порядок точек будет оптимизирован для минимизации пути.
	Optimize bool `json:"optimize"`
}

// Validate проверяет корректность запроса построения маршрута.
func (r *BuildRouteRequest) Validate() error {
	if len(r.LocationIDs) < 2 {
		return errors.New("требуется минимум 2 локации для построения маршрута")
	}
	if len(r.LocationIDs) > 20 {
		return errors.New("максимум 20 локаций в маршруте")
	}
	if r.Transport != "" && !ValidTransports[r.Transport] {
		return errors.New("допустимые виды транспорта: car, public, walk, bike")
	}
	return nil
}

// NormalizeDefaults устанавливает значения по умолчанию для незаполненных полей.
func (r *BuildRouteRequest) NormalizeDefaults() {
	if r.Transport == "" {
		r.Transport = "car"
	}
}

// BuildTripRouteRequest — DTO для запроса построения trip-aware маршрута.
// Если LocationIDs пуст — локации подбираются автоматически на основе параметров поездки.
type BuildTripRouteRequest struct {
	// LocationIDs — список UUID локаций (опционально, если пуст — автоподбор).
	LocationIDs []string `json:"location_ids"`

	// MaxPoints — максимальное количество точек (0 = определяется форматом поездки).
	MaxPoints int `json:"max_points"`

	// Optimize — если true, порядок точек оптимизируется для минимизации пути.
	Optimize bool `json:"optimize"`
}

// Validate проверяет корректность запроса trip-aware маршрута.
func (r *BuildTripRouteRequest) Validate() error {
	if len(r.LocationIDs) > 20 {
		return errors.New("максимум 20 локаций в маршруте")
	}
	if r.MaxPoints < 0 {
		return errors.New("max_points не может быть отрицательным")
	}
	if r.MaxPoints > 20 {
		return errors.New("максимум 20 точек в маршруте")
	}
	return nil
}

// NormalizeDefaults устанавливает значения по умолчанию.
func (r *BuildTripRouteRequest) NormalizeDefaults(tripFormat string) {
	if r.MaxPoints == 0 {
		if max, ok := MaxRoutePointsPerFormat[tripFormat]; ok {
			r.MaxPoints = max
		} else {
			r.MaxPoints = 10
		}
	}
}

// RoutePoint — одна точка маршрута с расчётными данными и метаданными.
type RoutePoint struct {
	// LocationID — UUID локации.
	LocationID string `json:"location_id"`

	// Name — название локации.
	Name string `json:"name"`

	// Latitude — широта.
	Latitude float64 `json:"latitude"`

	// Longitude — долгота.
	Longitude float64 `json:"longitude"`

	// Category — категория локации (winery, farm, trail и др.).
	Category string `json:"category"`

	// PreviewImageURL — URL hero-изображения локации.
	PreviewImageURL string `json:"preview_image_url"`

	// Order — порядковый номер точки в маршруте (начиная с 1).
	Order int `json:"order"`

	// DistanceFromPrevKm — расстояние от предыдущей точки в километрах.
	DistanceFromPrevKm float64 `json:"distance_from_prev_km"`

	// DurationFromPrevMin — время в пути от предыдущей точки в минутах.
	DurationFromPrevMin int `json:"duration_from_prev_min"`

	// StayDurationMin — рекомендуемое время пребывания в данной точке в минутах.
	StayDurationMin int `json:"stay_duration_min"`

	// DayNumber — номер дня поездки (начиная с 1).
	DayNumber int `json:"day_number"`

	// TimeSlot — тайм-слот: morning, afternoon, evening.
	TimeSlot string `json:"time_slot"`

	// TargetAudience — целевая аудитория: all, adults, children.
	TargetAudience string `json:"target_audience"`

	// VibeScore — оценка соответствия vibe-вектору группы (0.0 - 1.0).
	VibeScore float32 `json:"vibe_score,omitempty"`
}

// RoutePreview — полный ответ построенного маршрута.
// Содержит упорядоченный список точек и суммарные данные.
type RoutePreview struct {
	// ID — UUID маршрута (заполняется после сохранения в БД).
	ID string `json:"id,omitempty"`

	// TripID — UUID поездки (для trip-aware маршрутов).
	TripID string `json:"trip_id,omitempty"`

	// Name — название маршрута.
	Name string `json:"name,omitempty"`

	// Status — статус маршрута (draft/active/completed).
	Status string `json:"status,omitempty"`

	// Points — упорядоченный массив точек маршрута.
	Points []RoutePoint `json:"points"`

	// TotalDistanceKm — общее расстояние маршрута в километрах.
	TotalDistanceKm float64 `json:"total_distance_km"`

	// TotalDurationMin — общее время маршрута в минутах (включая пребывание).
	TotalDurationMin int `json:"total_duration_min"`

	// Transport — использованный вид транспорта.
	Transport string `json:"transport"`

	// PointsCount — количество точек в маршруте.
	PointsCount int `json:"points_count"`

	// EstimatedCostRub — расчётная стоимость маршрута в рублях.
	EstimatedCostRub int `json:"estimated_cost_rub,omitempty"`

	// Summary — текстовое описание маршрута.
	Summary string `json:"summary,omitempty"`

	// DaysCount — количество дней маршрута.
	DaysCount int `json:"days_count,omitempty"`

	// CreatedAt — время создания маршрута.
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// Route — DB-модель для таблицы routes.
type Route struct {
	// ID — уникальный идентификатор маршрута.
	ID uuid.UUID `json:"id" db:"id"`

	// TripID — идентификатор поездки.
	TripID uuid.UUID `json:"trip_id" db:"trip_id"`

	// UserID — идентификатор создателя маршрута.
	UserID uuid.UUID `json:"user_id" db:"user_id"`

	// Name — название маршрута.
	Name string `json:"name" db:"name"`

	// Status — статус маршрута (draft/active/completed).
	Status string `json:"status" db:"status"`

	// TotalDistanceKm — общее расстояние в километрах.
	TotalDistanceKm float64 `json:"total_distance_km" db:"total_distance_km"`

	// EstimatedDurationMin — общее расчётное время в минутах.
	EstimatedDurationMin int `json:"estimated_duration_min" db:"estimated_duration_min"`

	// EstimatedCostRub — расчётная стоимость в рублях.
	EstimatedCostRub int `json:"estimated_cost_rub" db:"estimated_cost_rub"`

	// Transport — вид транспорта.
	Transport string `json:"transport" db:"transport"`

	// PointsCount — количество точек.
	PointsCount int `json:"points_count" db:"points_count"`

	// Summary — текстовое описание маршрута.
	Summary string `json:"summary" db:"summary"`

	// CreatedAt — временная метка создания.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RoutePointDB — DB-модель для таблицы route_points.
type RoutePointDB struct {
	// ID — уникальный идентификатор точки.
	ID uuid.UUID `json:"id" db:"id"`

	// RouteID — идентификатор маршрута.
	RouteID uuid.UUID `json:"route_id" db:"route_id"`

	// LocationID — идентификатор локации.
	LocationID uuid.UUID `json:"location_id" db:"location_id"`

	// Position — порядковый номер точки в маршруте.
	Position int `json:"position" db:"position"`

	// DayNumber — номер дня поездки.
	DayNumber int `json:"day_number" db:"day_number"`

	// TimeSlot — тайм-слот (morning/afternoon/evening).
	TimeSlot string `json:"time_slot" db:"time_slot"`

	// TargetAudience — целевая аудитория (all/adults/children).
	TargetAudience string `json:"target_audience" db:"target_audience"`

	// StayDurationMin — рекомендуемое время пребывания в минутах.
	StayDurationMin int `json:"stay_duration_min" db:"stay_duration_min"`

	// DistanceFromPrevKm — расстояние от предыдущей точки в км.
	DistanceFromPrevKm float64 `json:"distance_from_prev_km" db:"distance_from_prev_km"`

	// DurationFromPrevMin — время в пути от предыдущей точки в минутах.
	DurationFromPrevMin int `json:"duration_from_prev_min" db:"duration_from_prev_min"`
}
