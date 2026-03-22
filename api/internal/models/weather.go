package models

import "time"

// WeatherPoint — погодная точка региона.
type WeatherPoint struct {
	Point       string    `json:"point"`
	Latitude    float64   `json:"lat"`
	Longitude   float64   `json:"lon"`
	TempC       int       `json:"temp"`
	Condition   string    `json:"condition"`
	WindSpeed   int       `json:"wind_speed"`
	UpdatedAt   time.Time `json:"updated_at"`
	Severity    string    `json:"severity,omitempty"`
	NeedsRebuild bool     `json:"needs_rebuild,omitempty"`
}

// RouteRebuildRequest — ручной override для live rebuild.
type RouteRebuildRequest struct {
	WeatherOverride map[string]any `json:"weather_override"`
}

// RouteStory — история для точки маршрута.
type RouteStory struct {
	PointID       string `json:"point_id"`
	LocationID    string `json:"location_id"`
	LocationName  string `json:"location_name"`
	AudioURL      string `json:"audio_url"`
	StoryText     string `json:"story_text"`
	DurationSec   int    `json:"duration_sec"`
	WeatherHint   string `json:"weather_hint,omitempty"`
	DebugInfo     string `json:"debug_info,omitempty"`
}

// GenerateStoriesResponse — результат генерации историй.
type GenerateStoriesResponse struct {
	RouteID      string `json:"route_id"`
	PointsCount  int    `json:"points_count"`
	Generated    int    `json:"generated"`
}
