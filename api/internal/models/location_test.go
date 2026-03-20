// Тесты для модели Location: валидация, генерация slug, фильтры, пагинация.
package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGenerateSlug проверяет генерацию slug из названия.
func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"кириллица", "Винодельня Абрау-Дюрсо", "vinodelnya-abrau-dyurso"},
		{"латиница", "Happy Farm", "happy-farm"},
		{"спецсимволы", "Козья ферма №1!", "kozya-ferma-1"},
		{"пробелы", "  Горы  и  Море  ", "gory-i-more"},
		{"цифры", "33 водопада", "33-vodopada"},
		{"пустая строка", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlug(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCreateLocationRequestValidate проверяет валидацию DTO создания локации.
func TestCreateLocationRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateLocationRequest
		wantErr string
	}{
		{
			name:    "без названия",
			req:     CreateLocationRequest{Latitude: 44.0, Longitude: 38.0},
			wantErr: "название обязательно",
		},
		{
			name:    "широта вне диапазона",
			req:     CreateLocationRequest{Name: "Test", Latitude: 91.0, Longitude: 38.0},
			wantErr: "широта должна быть в диапазоне [-90, 90]",
		},
		{
			name:    "долгота вне диапазона",
			req:     CreateLocationRequest{Name: "Test", Latitude: 44.0, Longitude: 181.0},
			wantErr: "долгота должна быть в диапазоне [-180, 180]",
		},
		{
			name:    "нулевые координаты",
			req:     CreateLocationRequest{Name: "Test", Latitude: 0, Longitude: 0},
			wantErr: "координаты обязательны",
		},
		{
			name:    "отрицательная стоимость",
			req:     CreateLocationRequest{Name: "Test", Latitude: 44.0, Longitude: 38.0, PricePerNight: -100},
			wantErr: "стоимость не может быть отрицательной",
		},
		{
			name:    "корректные данные",
			req:     CreateLocationRequest{Name: "Ферма", Latitude: 44.5, Longitude: 38.5},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.req.Validate()
			assert.Equal(t, tt.wantErr, result)
		})
	}
}

// TestLocationFilterHasRadiusSearch проверяет определение радиусного поиска.
func TestLocationFilterHasRadiusSearch(t *testing.T) {
	lat, lon, radius := 44.5, 38.5, 50.0
	filter := &LocationFilter{Lat: &lat, Lon: &lon, RadiusKm: &radius}
	assert.True(t, filter.HasRadiusSearch())

	filterEmpty := &LocationFilter{}
	assert.False(t, filterEmpty.HasRadiusSearch())
}

// TestLocationFilterHasBBoxSearch проверяет определение bbox-поиска.
func TestLocationFilterHasBBoxSearch(t *testing.T) {
	minLat, maxLat, minLon, maxLon := 43.0, 46.0, 36.0, 41.0
	filter := &LocationFilter{MinLat: &minLat, MaxLat: &maxLat, MinLon: &minLon, MaxLon: &maxLon}
	assert.True(t, filter.HasBBoxSearch())

	filterPartial := &LocationFilter{MinLat: &minLat}
	assert.False(t, filterPartial.HasBBoxSearch())
}

// TestLocationFilterNormalizePagination проверяет нормализацию пагинации.
func TestLocationFilterNormalizePagination(t *testing.T) {
	filter := &LocationFilter{}
	filter.NormalizePagination()
	assert.Equal(t, 1, filter.Page)
	assert.Equal(t, 20, filter.PerPage)

	filterBig := &LocationFilter{PerPage: 500}
	filterBig.NormalizePagination()
	assert.Equal(t, 100, filterBig.PerPage)
}

// TestLocationFilterOffset проверяет вычисление смещения.
func TestLocationFilterOffset(t *testing.T) {
	filter := &LocationFilter{Page: 3, PerPage: 10}
	assert.Equal(t, 20, filter.Offset())
}
