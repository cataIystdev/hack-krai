// Файл location.go содержит структуры данных для таблицы locations.
// Определяет модель локации с пространственными координатами,
// DTO для создания/обновления, структуру фильтра для пространственного поиска
// и публичное представление для ответов API.
package models

import (
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// AccessLevel — тип перечисления уровней доступа к локации.
// Соответствует PostgreSQL типу access_level_type.
type AccessLevel string

const (
	// AccessLevelOpen — локация видна всем пользователям.
	AccessLevelOpen AccessLevel = "open"

	// AccessLevelSemiOpen — видна зарегистрированным с заполненным профилем.
	AccessLevelSemiOpen AccessLevel = "semi_open"

	// AccessLevelHidden — только для пользователей с достаточной кармой.
	AccessLevelHidden AccessLevel = "hidden"
)

// DensityLevel — тип перечисления уровней туристической плотности.
// Соответствует PostgreSQL типу density_level_type.
type DensityLevel string

const (
	// DensityRed — высокая загруженность круглый год.
	DensityRed DensityLevel = "red"

	// DensityYellow — сезонная загруженность.
	DensityYellow DensityLevel = "yellow"

	// DensityGreen — скрытое место (Hidden Gem).
	DensityGreen DensityLevel = "green"
)

// Location — структура локации, соответствующая таблице locations в PostgreSQL.
// Координаты хранятся как отдельные поля Latitude/Longitude, а не как raw GEOMETRY.
type Location struct {
	// ID — уникальный идентификатор локации (UUID v4).
	ID uuid.UUID `json:"id" db:"id"`

	// OwnerID — UUID владельца (хоста).
	OwnerID uuid.UUID `json:"owner_id" db:"owner_id"`

	// Slug — URL-дружественный идентификатор для маршрутов.
	Slug string `json:"slug" db:"slug"`

	// Name — название локации.
	Name string `json:"name" db:"name"`

	// DescriptionShort — краткое описание (для карточек).
	DescriptionShort string `json:"description_short" db:"description_short"`

	// DescriptionFull — полное описание (литературный текст).
	DescriptionFull string `json:"description_full" db:"description_full"`

	// Category — категория локации (свободный текст: winery, farm, trail и др.).
	Category string `json:"category" db:"category"`

	// Tags — массив тегов для фильтрации и поиска.
	Tags []string `json:"tags" db:"tags"`

	// PricePerNight — стоимость проживания за ночь (рубли).
	PricePerNight int `json:"price_per_night" db:"price_per_night"`

	// Capacity — максимальная вместимость (гостей).
	Capacity int `json:"capacity" db:"capacity"`

	// AccessLevel — уровень доступа (open/semi_open/hidden).
	AccessLevel AccessLevel `json:"access_level" db:"access_level"`

	// DensityLevel — уровень плотности туристов (red/yellow/green).
	DensityLevel DensityLevel `json:"density_level" db:"density_level"`

	// ChildFriendly — подходит ли для детей.
	ChildFriendly bool `json:"child_friendly" db:"child_friendly"`

	// SplatURL — URL на 3D-сцену (.splat) в MinIO.
	SplatURL *string `json:"splat_url,omitempty" db:"splat_url"`

	// PreviewImageURL — URL hero-изображения локации для карточек и карты.
	PreviewImageURL string `json:"preview_image_url" db:"preview_image_url"`

	// GalleryURLs — массив URL дополнительных фотографий для экрана детали локации.
	GalleryURLs []string `json:"gallery_urls" db:"gallery_urls"`

	// VibeVectorID — ID вектора vibe-профиля в Qdrant.
	VibeVectorID *uuid.UUID `json:"vibe_vector_id,omitempty" db:"vibe_vector_id"`

	// Address — полный адрес (населённый пункт, район).
	Address string `json:"address" db:"address"`

	// IsPublished — опубликована ли локация.
	IsPublished bool `json:"is_published" db:"is_published"`

	// Latitude — широта (извлекается через ST_Y(geo)).
	Latitude float64 `json:"latitude" db:"latitude"`

	// Longitude — долгота (извлекается через ST_X(geo)).
	Longitude float64 `json:"longitude" db:"longitude"`

	// CreatedAt — временная метка создания.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt — временная метка обновления.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateLocationRequest — DTO для создания новой локации.
type CreateLocationRequest struct {
	Name             string   `json:"name"`
	DescriptionShort string   `json:"description_short"`
	DescriptionFull  string   `json:"description_full"`
	Category         string   `json:"category"`
	Tags             []string `json:"tags"`
	PricePerNight    int      `json:"price_per_night"`
	Capacity         int      `json:"capacity"`
	AccessLevel      string   `json:"access_level"`
	DensityLevel     string   `json:"density_level"`
	ChildFriendly    bool     `json:"child_friendly"`
	Address          string   `json:"address"`
	IsPublished      bool     `json:"is_published"`
	Latitude         float64  `json:"latitude"`
	Longitude        float64  `json:"longitude"`
}

// Validate проверяет корректность данных для создания локации.
func (r *CreateLocationRequest) Validate() string {
	if r.Name == "" {
		return "название обязательно"
	}
	if r.Latitude < -90 || r.Latitude > 90 {
		return "широта должна быть в диапазоне [-90, 90]"
	}
	if r.Longitude < -180 || r.Longitude > 180 {
		return "долгота должна быть в диапазоне [-180, 180]"
	}
	if r.Latitude == 0 && r.Longitude == 0 {
		return "координаты обязательны"
	}
	if r.PricePerNight < 0 {
		return "стоимость не может быть отрицательной"
	}
	if r.Capacity < 0 {
		return "вместимость не может быть отрицательной"
	}
	return ""
}

// UpdateLocationRequest — DTO для обновления локации.
// Все поля опциональны (обновляются только переданные).
type UpdateLocationRequest struct {
	Name             *string  `json:"name,omitempty"`
	DescriptionShort *string  `json:"description_short,omitempty"`
	DescriptionFull  *string  `json:"description_full,omitempty"`
	Category         *string  `json:"category,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	PricePerNight    *int     `json:"price_per_night,omitempty"`
	Capacity         *int     `json:"capacity,omitempty"`
	AccessLevel      *string  `json:"access_level,omitempty"`
	DensityLevel     *string  `json:"density_level,omitempty"`
	ChildFriendly    *bool    `json:"child_friendly,omitempty"`
	Address          *string  `json:"address,omitempty"`
	IsPublished      *bool    `json:"is_published,omitempty"`
	Latitude         *float64 `json:"latitude,omitempty"`
	Longitude        *float64 `json:"longitude,omitempty"`
}

// LocationFilter — параметры фильтрации и пространственного поиска локаций.
type LocationFilter struct {
	// Поиск по радиусу (ST_DWithin). Все три поля обязательны для активации.
	Lat      *float64 `query:"lat"`
	Lon      *float64 `query:"lon"`
	RadiusKm *float64 `query:"radius_km"`

	// Поиск по bbox (ST_Within + ST_MakeEnvelope). Все четыре поля обязательны.
	MinLat *float64 `query:"min_lat"`
	MaxLat *float64 `query:"max_lat"`
	MinLon *float64 `query:"min_lon"`
	MaxLon *float64 `query:"max_lon"`

	// Фильтры по атрибутам.
	Category      *string `query:"category"`
	ChildFriendly *bool   `query:"child_friendly"`
	DensityLevel  *string `query:"density_level"`

	// Пагинация.
	Page    int `query:"page"`
	PerPage int `query:"per_page"`
}

// HasRadiusSearch проверяет, заданы ли параметры поиска по радиусу.
func (f *LocationFilter) HasRadiusSearch() bool {
	return f.Lat != nil && f.Lon != nil && f.RadiusKm != nil
}

// HasBBoxSearch проверяет, заданы ли параметры поиска по bounding box.
func (f *LocationFilter) HasBBoxSearch() bool {
	return f.MinLat != nil && f.MaxLat != nil && f.MinLon != nil && f.MaxLon != nil
}

// NormalizePagination устанавливает значения по умолчанию для пагинации.
func (f *LocationFilter) NormalizePagination() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PerPage <= 0 {
		f.PerPage = 20
	}
	if f.PerPage > 100 {
		f.PerPage = 100
	}
}

// Offset вычисляет смещение для SQL OFFSET на основе текущей страницы.
func (f *LocationFilter) Offset() int {
	return (f.Page - 1) * f.PerPage
}

// LocationListResponse — ответ на запрос списка локаций с пагинацией.
type LocationListResponse struct {
	// Locations — массив локаций.
	Locations []Location `json:"locations"`

	// Total — общее количество записей, удовлетворяющих фильтрам.
	Total int `json:"total"`

	// Page — текущая страница.
	Page int `json:"page"`

	// PerPage — количество записей на странице.
	PerPage int `json:"per_page"`
}

// GenerateSlug формирует URL-дружественный slug из названия локации.
// Транслитерирует кириллицу в латиницу и заменяет пробелы/спецсимволы на дефисы.
func GenerateSlug(name string) string {
	translitMap := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d",
		'е': "e", 'ё': "yo", 'ж': "zh", 'з': "z", 'и': "i",
		'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n",
		'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t",
		'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch",
		'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "",
		'э': "e", 'ю': "yu", 'я': "ya",
	}

	var builder strings.Builder
	lower := strings.ToLower(name)
	prevDash := false

	for _, r := range lower {
		if mapped, ok := translitMap[r]; ok {
			builder.WriteString(mapped)
			prevDash = false
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			prevDash = false
		} else if !prevDash && builder.Len() > 0 {
			builder.WriteByte('-')
			prevDash = true
		}
	}

	result := strings.TrimRight(builder.String(), "-")
	if len(result) > 200 {
		result = result[:200]
	}
	return result
}

// UpdateProfileRequest — DTO для обновления профиля пользователя.
type UpdateProfileRequest struct {
	// DisplayName — новое отображаемое имя.
	DisplayName *string `json:"display_name,omitempty"`
}
