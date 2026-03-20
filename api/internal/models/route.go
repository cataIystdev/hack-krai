// Файл route.go содержит модели данных для эндпоинта POST /api/v1/route/build.
// Определяет DTO запроса построения маршрута, структуру точки маршрута
// и превью маршрута с расчётными данными.
package models

import "errors"

// BuildRouteRequest — DTO для запроса построения маршрута.
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

// RoutePoint — одна точка маршрута с расчётными данными.
type RoutePoint struct {
	// LocationID — UUID локации.
	LocationID string `json:"location_id"`

	// Name — название локации.
	Name string `json:"name"`

	// Latitude — широта.
	Latitude float64 `json:"latitude"`

	// Longitude — долгота.
	Longitude float64 `json:"longitude"`

	// Order — порядковый номер точки в маршруте (начиная с 1).
	Order int `json:"order"`

	// DistanceFromPrevKm — расстояние от предыдущей точки в километрах.
	// Для первой точки равно 0.
	DistanceFromPrevKm float64 `json:"distance_from_prev_km"`

	// DurationFromPrevMin — время в пути от предыдущей точки в минутах.
	// Для первой точки равно 0.
	DurationFromPrevMin int `json:"duration_from_prev_min"`

	// StayDurationMin — рекомендуемое время пребывания в данной точке в минутах.
	StayDurationMin int `json:"stay_duration_min"`
}

// RoutePreview — предпросмотр построенного маршрута.
// Содержит упорядоченный список точек с расчётами расстояний и времени,
// а также суммарные данные маршрута.
type RoutePreview struct {
	// Points — упорядоченный массив точек маршрута.
	Points []RoutePoint `json:"points"`

	// TotalDistanceKm — общее расстояние маршрута в километрах.
	TotalDistanceKm float64 `json:"total_distance_km"`

	// TotalDurationMin — общее время маршрута в минутах (включая пребывание в точках).
	TotalDurationMin int `json:"total_duration_min"`

	// Transport — использованный вид транспорта.
	Transport string `json:"transport"`

	// PointsCount — количество точек в маршруте.
	PointsCount int `json:"points_count"`
}
