// Файл map.go содержит модели данных для эндпоинта GET /api/v1/map/locations.
// Определяет DTO для фильтрации точек на карте, структуру точки и ответ API.
// Поддерживает фильтрацию по bbox, категории, плотности и статусу рекомендации.
package models

// MapLocationFilter — параметры фильтрации точек для карты.
// Поддерживает bbox-поиск, фильтрацию по категории и плотности,
// а также флаг рекомендованных точек.
type MapLocationFilter struct {
	// MinLat — минимальная широта bbox.
	MinLat *float64 `query:"min_lat"`

	// MaxLat — максимальная широта bbox.
	MaxLat *float64 `query:"max_lat"`

	// MinLon — минимальная долгота bbox.
	MinLon *float64 `query:"min_lon"`

	// MaxLon — максимальная долгота bbox.
	MaxLon *float64 `query:"max_lon"`

	// Category — фильтр по категории (winery, farm, trail и др.).
	Category *string `query:"category"`

	// DensityLevel — фильтр по плотности (red/yellow/green).
	DensityLevel *string `query:"density_level"`

	// IsRecommended — вернуть только рекомендованные точки.
	IsRecommended *bool `query:"is_recommended"`

	// Limit — максимальное количество точек в ответе.
	Limit int `query:"limit"`
}

// HasBBox проверяет, заданы ли все четыре координаты bounding box.
func (f *MapLocationFilter) HasBBox() bool {
	return f.MinLat != nil && f.MaxLat != nil && f.MinLon != nil && f.MaxLon != nil
}

// NormalizeLimit устанавливает лимит по умолчанию и ограничивает максимум.
func (f *MapLocationFilter) NormalizeLimit() {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Limit > 200 {
		f.Limit = 200
	}
}

// Validate проверяет корректность фильтра для карты.
// Возвращает текст ошибки или пустую строку при успешной валидации.
func (f *MapLocationFilter) Validate() string {
	if f.HasBBox() {
		if *f.MinLat < -90 || *f.MaxLat > 90 {
			return "широта должна быть в диапазоне [-90, 90]"
		}
		if *f.MinLon < -180 || *f.MaxLon > 180 {
			return "долгота должна быть в диапазоне [-180, 180]"
		}
		if *f.MinLat >= *f.MaxLat {
			return "min_lat должна быть меньше max_lat"
		}
		if *f.MinLon >= *f.MaxLon {
			return "min_lon должна быть меньше max_lon"
		}
	}
	return ""
}

// MapPoint — точка локации на карте.
// Содержит только поля, необходимые для рендеринга маркера на карте.
type MapPoint struct {
	// ID — уникальный идентификатор локации.
	ID string `json:"id"`

	// Name — название локации.
	Name string `json:"name"`

	// Latitude — широта.
	Latitude float64 `json:"latitude"`

	// Longitude — долгота.
	Longitude float64 `json:"longitude"`

	// Category — категория (winery, farm, trail, nature и др.).
	Category string `json:"category"`

	// DensityLevel — уровень плотности туристов (red/yellow/green).
	DensityLevel string `json:"density_level"`

	// PreviewImageURL — URL preview-изображения для маркера.
	PreviewImageURL string `json:"preview_image_url"`

	// IsRecommended — помечена ли точка как рекомендованная для текущего профиля.
	IsRecommended bool `json:"is_recommended"`

	// RecommendationScore — оценка релевантности (0.0-1.0). Заполняется только для рекомендованных.
	RecommendationScore float32 `json:"recommendation_score,omitempty"`
}

// MapLocationsResponse — ответ на запрос GET /api/v1/map/locations.
type MapLocationsResponse struct {
	// Points — массив точек для отображения на карте.
	Points []MapPoint `json:"points"`

	// Total — общее количество точек в ответе.
	Total int `json:"total"`
}
