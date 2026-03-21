// Файл map_test.go содержит unit-тесты для моделей карты.
// Проверяет валидацию MapLocationFilter, HasBBox, NormalizeLimit, IsDemo,
// GetProfile и валидацию demo-профилей.
package models

import "testing"

// TestMapLocationFilter_HasBBox проверяет определение bbox-поиска.
func TestMapLocationFilter_HasBBox(t *testing.T) {
	tests := []struct {
		name   string
		filter MapLocationFilter
		want   bool
	}{
		{
			name:   "все четыре координаты заданы",
			filter: MapLocationFilter{MinLat: floatPtr(44.0), MaxLat: floatPtr(45.0), MinLon: floatPtr(37.0), MaxLon: floatPtr(40.0)},
			want:   true,
		},
		{
			name:   "нет MinLat",
			filter: MapLocationFilter{MaxLat: floatPtr(45.0), MinLon: floatPtr(37.0), MaxLon: floatPtr(40.0)},
			want:   false,
		},
		{
			name:   "пустой фильтр",
			filter: MapLocationFilter{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.HasBBox(); got != tt.want {
				t.Errorf("HasBBox() = %v, ожидалось %v", got, tt.want)
			}
		})
	}
}

// TestMapLocationFilter_NormalizeLimit проверяет нормализацию лимита.
func TestMapLocationFilter_NormalizeLimit(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		wantLimit int
	}{
		{"нулевой лимит", 0, 50},
		{"отрицательный лимит", -5, 50},
		{"нормальный лимит", 30, 30},
		{"превышение максимума", 500, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &MapLocationFilter{Limit: tt.limit}
			f.NormalizeLimit()
			if f.Limit != tt.wantLimit {
				t.Errorf("NormalizeLimit() Limit = %d, ожидалось %d", f.Limit, tt.wantLimit)
			}
		})
	}
}

// TestMapLocationFilter_Validate проверяет валидацию фильтра.
func TestMapLocationFilter_Validate(t *testing.T) {
	tests := []struct {
		name      string
		filter    MapLocationFilter
		wantError bool
	}{
		{
			name:      "валидный bbox",
			filter:    MapLocationFilter{MinLat: floatPtr(44.0), MaxLat: floatPtr(45.0), MinLon: floatPtr(37.0), MaxLon: floatPtr(40.0)},
			wantError: false,
		},
		{
			name:      "min_lat >= max_lat",
			filter:    MapLocationFilter{MinLat: floatPtr(46.0), MaxLat: floatPtr(45.0), MinLon: floatPtr(37.0), MaxLon: floatPtr(40.0)},
			wantError: true,
		},
		{
			name:      "без bbox (нет ошибки)",
			filter:    MapLocationFilter{},
			wantError: false,
		},
		{
			name:      "широта вне диапазона",
			filter:    MapLocationFilter{MinLat: floatPtr(-100.0), MaxLat: floatPtr(45.0), MinLon: floatPtr(37.0), MaxLon: floatPtr(40.0)},
			wantError: true,
		},
		{
			name:      "валидный профиль calm_wine_mountains",
			filter:    MapLocationFilter{Profile: strPtr("calm_wine_mountains")},
			wantError: false,
		},
		{
			name:      "валидный профиль active_adventure",
			filter:    MapLocationFilter{Profile: strPtr("active_adventure")},
			wantError: false,
		},
		{
			name:      "валидный профиль family_kids",
			filter:    MapLocationFilter{Profile: strPtr("family_kids")},
			wantError: false,
		},
		{
			name:      "валидный профиль gastro_cultural",
			filter:    MapLocationFilter{Profile: strPtr("gastro_cultural")},
			wantError: false,
		},
		{
			name:      "невалидный профиль",
			filter:    MapLocationFilter{Profile: strPtr("unknown_profile")},
			wantError: true,
		},
		{
			name:      "пустой профиль (нет ошибки)",
			filter:    MapLocationFilter{Profile: strPtr("")},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.filter.Validate()
			hasError := errMsg != ""
			if hasError != tt.wantError {
				t.Errorf("Validate() = %q, wantError = %v", errMsg, tt.wantError)
			}
		})
	}
}

// TestMapLocationFilter_IsDemo проверяет определение demo-режима.
func TestMapLocationFilter_IsDemo(t *testing.T) {
	tests := []struct {
		name   string
		filter MapLocationFilter
		want   bool
	}{
		{
			name:   "demo включен",
			filter: MapLocationFilter{Demo: boolPtr(true)},
			want:   true,
		},
		{
			name:   "demo выключен",
			filter: MapLocationFilter{Demo: boolPtr(false)},
			want:   false,
		},
		{
			name:   "demo не задан",
			filter: MapLocationFilter{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.IsDemo(); got != tt.want {
				t.Errorf("IsDemo() = %v, ожидалось %v", got, tt.want)
			}
		})
	}
}

// TestMapLocationFilter_GetProfile проверяет получение demo-профиля.
func TestMapLocationFilter_GetProfile(t *testing.T) {
	tests := []struct {
		name   string
		filter MapLocationFilter
		want   string
	}{
		{
			name:   "профиль задан",
			filter: MapLocationFilter{Profile: strPtr("active_adventure")},
			want:   "active_adventure",
		},
		{
			name:   "профиль пуст — по умолчанию",
			filter: MapLocationFilter{Profile: strPtr("")},
			want:   "calm_wine_mountains",
		},
		{
			name:   "профиль nil — по умолчанию",
			filter: MapLocationFilter{},
			want:   "calm_wine_mountains",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.GetProfile(); got != tt.want {
				t.Errorf("GetProfile() = %q, ожидалось %q", got, tt.want)
			}
		})
	}
}

// floatPtr — вспомогательная функция для создания указателя на float64.
func floatPtr(v float64) *float64 {
	return &v
}

// boolPtr — вспомогательная функция для создания указателя на bool.
func boolPtr(v bool) *bool {
	return &v
}

// strPtr — вспомогательная функция для создания указателя на string.
func strPtr(v string) *string {
	return &v
}
