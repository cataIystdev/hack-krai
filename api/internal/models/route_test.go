// Файл route_test.go содержит unit-тесты для моделей маршрутов.
// Проверяет валидацию BuildRouteRequest, BuildTripRouteRequest,
// NormalizeDefaults и корректность констант.
package models

import (
	"testing"
)

// TestBuildRouteRequest_Validate проверяет валидацию запроса ручного маршрута.
func TestBuildRouteRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		req       BuildRouteRequest
		wantError bool
	}{
		{
			name:      "валидный запрос",
			req:       BuildRouteRequest{LocationIDs: []string{"id1", "id2"}, Transport: "car"},
			wantError: false,
		},
		{
			name:      "менее 2 локаций",
			req:       BuildRouteRequest{LocationIDs: []string{"id1"}},
			wantError: true,
		},
		{
			name:      "пустой массив",
			req:       BuildRouteRequest{LocationIDs: []string{}},
			wantError: true,
		},
		{
			name:      "более 20 локаций",
			req:       BuildRouteRequest{LocationIDs: make([]string, 21)},
			wantError: true,
		},
		{
			name:      "невалидный транспорт",
			req:       BuildRouteRequest{LocationIDs: []string{"id1", "id2"}, Transport: "helicopter"},
			wantError: true,
		},
		{
			name:      "без транспорта (валидно)",
			req:       BuildRouteRequest{LocationIDs: []string{"id1", "id2"}},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			hasError := err != nil
			if hasError != tt.wantError {
				t.Errorf("Validate() error = %v, wantError = %v", err, tt.wantError)
			}
		})
	}
}

// TestBuildRouteRequest_NormalizeDefaults проверяет установку значений по умолчанию.
func TestBuildRouteRequest_NormalizeDefaults(t *testing.T) {
	tests := []struct {
		name          string
		transport     string
		wantTransport string
	}{
		{"пустой транспорт", "", "car"},
		{"car остаётся car", "car", "car"},
		{"walk остаётся walk", "walk", "walk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &BuildRouteRequest{
				LocationIDs: []string{"id1", "id2"},
				Transport:   tt.transport,
			}
			req.NormalizeDefaults()
			if req.Transport != tt.wantTransport {
				t.Errorf("NormalizeDefaults() Transport = %s, ожидалось %s", req.Transport, tt.wantTransport)
			}
		})
	}
}

// TestBuildTripRouteRequest_Validate проверяет валидацию trip-aware запроса.
func TestBuildTripRouteRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		req       BuildTripRouteRequest
		wantError bool
	}{
		{
			name:      "пустой запрос (валидно — автоподбор)",
			req:       BuildTripRouteRequest{},
			wantError: false,
		},
		{
			name:      "с локациями",
			req:       BuildTripRouteRequest{LocationIDs: []string{"id1", "id2"}},
			wantError: false,
		},
		{
			name:      "более 20 локаций",
			req:       BuildTripRouteRequest{LocationIDs: make([]string, 21)},
			wantError: true,
		},
		{
			name:      "отрицательный max_points",
			req:       BuildTripRouteRequest{MaxPoints: -1},
			wantError: true,
		},
		{
			name:      "max_points > 20",
			req:       BuildTripRouteRequest{MaxPoints: 21},
			wantError: true,
		},
		{
			name:      "max_points = 10 (валидно)",
			req:       BuildTripRouteRequest{MaxPoints: 10},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			hasError := err != nil
			if hasError != tt.wantError {
				t.Errorf("Validate() error = %v, wantError = %v", err, tt.wantError)
			}
		})
	}
}

// TestBuildTripRouteRequest_NormalizeDefaults проверяет установку max_points по формату поездки.
func TestBuildTripRouteRequest_NormalizeDefaults(t *testing.T) {
	tests := []struct {
		name          string
		maxPoints     int
		tripFormat    string
		wantMaxPoints int
	}{
		{"day_trip без max_points", 0, "day_trip", 5},
		{"weekend без max_points", 0, "weekend", 8},
		{"multi_day без max_points", 0, "multi_day", 15},
		{"неизвестный формат", 0, "unknown", 10},
		{"явно задано 7", 7, "day_trip", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &BuildTripRouteRequest{MaxPoints: tt.maxPoints}
			req.NormalizeDefaults(tt.tripFormat)
			if req.MaxPoints != tt.wantMaxPoints {
				t.Errorf("NormalizeDefaults() MaxPoints = %d, ожидалось %d", req.MaxPoints, tt.wantMaxPoints)
			}
		})
	}
}

// TestMaxRoutePointsPerFormat проверяет наличие лимитов для всех форматов.
func TestMaxRoutePointsPerFormat(t *testing.T) {
	expectedFormats := []string{"day_trip", "weekend", "multi_day"}
	for _, format := range expectedFormats {
		if _, ok := MaxRoutePointsPerFormat[format]; !ok {
			t.Errorf("MaxRoutePointsPerFormat не содержит формат %q", format)
		}
	}

	if MaxRoutePointsPerFormat["day_trip"] >= MaxRoutePointsPerFormat["multi_day"] {
		t.Error("day_trip должен иметь меньше точек чем multi_day")
	}
}

// TestRouteConstants проверяет корректность констант статусов, слотов и аудиторий.
func TestRouteConstants(t *testing.T) {
	if RouteStatusDraft != "draft" {
		t.Errorf("RouteStatusDraft = %q, ожидалось draft", RouteStatusDraft)
	}
	if RouteStatusActive != "active" {
		t.Errorf("RouteStatusActive = %q, ожидалось active", RouteStatusActive)
	}
	if RouteStatusCompleted != "completed" {
		t.Errorf("RouteStatusCompleted = %q, ожидалось completed", RouteStatusCompleted)
	}

	if TimeSlotMorning != "morning" {
		t.Errorf("TimeSlotMorning = %q, ожидалось morning", TimeSlotMorning)
	}
	if TimeSlotAfternoon != "afternoon" {
		t.Errorf("TimeSlotAfternoon = %q, ожидалось afternoon", TimeSlotAfternoon)
	}
	if TimeSlotEvening != "evening" {
		t.Errorf("TimeSlotEvening = %q, ожидалось evening", TimeSlotEvening)
	}

	if AudienceAll != "all" {
		t.Errorf("AudienceAll = %q, ожидалось all", AudienceAll)
	}
	if AudienceAdults != "adults" {
		t.Errorf("AudienceAdults = %q, ожидалось adults", AudienceAdults)
	}
	if AudienceChildren != "children" {
		t.Errorf("AudienceChildren = %q, ожидалось children", AudienceChildren)
	}
}
