// Файл route_test.go содержит unit-тесты для моделей маршрутов.
// Проверяет валидацию BuildRouteRequest и NormalizeDefaults.
package models

import (
	"testing"
)

// TestBuildRouteRequest_Validate проверяет валидацию запроса маршрута.
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
