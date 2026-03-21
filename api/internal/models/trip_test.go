// Файл trip_test.go содержит unit-тесты для моделей и DTO модуля поездок.
// Тестирует валидацию CreateTripRequest, UpdateTripRequest, JoinTripRequest,
// нормализацию дефолтов, парсинг дат и корректность констант.
package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreateTripRequestValidation тестирует валидацию DTO создания поездки.
func TestCreateTripRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateTripRequest
		wantErr string
	}{
		{
			name:    "пустая дата начала",
			req:     CreateTripRequest{DateTo: "2026-04-13"},
			wantErr: "дата начала обязательна",
		},
		{
			name:    "пустая дата окончания",
			req:     CreateTripRequest{DateFrom: "2026-04-10"},
			wantErr: "дата окончания обязательна",
		},
		{
			name:    "некорректный формат даты начала",
			req:     CreateTripRequest{DateFrom: "10-04-2026", DateTo: "2026-04-13"},
			wantErr: "некорректный формат даты начала (ожидается YYYY-MM-DD)",
		},
		{
			name:    "некорректный формат даты окончания",
			req:     CreateTripRequest{DateFrom: "2026-04-10", DateTo: "13/04/2026"},
			wantErr: "некорректный формат даты окончания (ожидается YYYY-MM-DD)",
		},
		{
			name:    "дата окончания раньше начала",
			req:     CreateTripRequest{DateFrom: "2026-04-13", DateTo: "2026-04-10"},
			wantErr: "дата окончания не может быть раньше даты начала",
		},
		{
			name:    "отрицательный бюджет",
			req:     CreateTripRequest{DateFrom: "2026-04-10", DateTo: "2026-04-13", BudgetRub: -1},
			wantErr: "бюджет не может быть отрицательным",
		},
		{
			name:    "некорректный уровень бюджета",
			req:     CreateTripRequest{DateFrom: "2026-04-10", DateTo: "2026-04-13", BudgetTier: "luxury"},
			wantErr: "допустимые уровни бюджета: economy, comfort, premium",
		},
		{
			name:    "некорректный вид транспорта",
			req:     CreateTripRequest{DateFrom: "2026-04-10", DateTo: "2026-04-13", Transport: "helicopter"},
			wantErr: "допустимые виды транспорта: car, public, walk, bike",
		},
		{
			name:    "отрицательное количество участников",
			req:     CreateTripRequest{DateFrom: "2026-04-10", DateTo: "2026-04-13", GroupSize: -1},
			wantErr: "количество участников не может быть отрицательным",
		},
		{
			name:    "некорректный формат поездки",
			req:     CreateTripRequest{DateFrom: "2026-04-10", DateTo: "2026-04-13", Format: "expedition"},
			wantErr: "допустимые форматы поездки: day_trip, weekend, multi_day",
		},
		{
			name:    "некорректный vibe_vector_id",
			req:     CreateTripRequest{DateFrom: "2026-04-10", DateTo: "2026-04-13", VibeVectorID: "not-a-uuid"},
			wantErr: "vibe_vector_id должен быть валидным UUID",
		},
		{
			name: "корректный запрос (минимальный)",
			req:  CreateTripRequest{DateFrom: "2026-04-10", DateTo: "2026-04-13"},
		},
		{
			name: "корректный запрос (полный)",
			req: CreateTripRequest{
				DateFrom:     "2026-04-10",
				DateTo:       "2026-04-13",
				BudgetRub:    50000,
				BudgetTier:   "comfort",
				Transport:    "car",
				GroupSize:    4,
				Format:       "multi_day",
				VibeVectorID: "550e8400-e29b-41d4-a716-446655440000",
			},
		},
		{
			name: "однодневная поездка с форматом day_trip",
			req:  CreateTripRequest{DateFrom: "2026-04-10", DateTo: "2026-04-10", Format: "day_trip"},
		},
		{
			name: "выходные с форматом weekend",
			req:  CreateTripRequest{DateFrom: "2026-04-10", DateTo: "2026-04-12", Format: "weekend"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.req.Validate()
			if tt.wantErr != "" {
				assert.Equal(t, tt.wantErr, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

// TestCreateTripRequestParseDates тестирует парсинг дат в time.Time.
func TestCreateTripRequestParseDates(t *testing.T) {
	req := CreateTripRequest{DateFrom: "2026-04-10", DateTo: "2026-04-13"}
	dateFrom, dateTo := req.ParseDates()

	assert.Equal(t, 2026, dateFrom.Year())
	assert.Equal(t, 4, int(dateFrom.Month()))
	assert.Equal(t, 10, dateFrom.Day())

	assert.Equal(t, 2026, dateTo.Year())
	assert.Equal(t, 4, int(dateTo.Month()))
	assert.Equal(t, 13, dateTo.Day())
}

// TestCreateTripRequestNormalizeDefaults тестирует установку значений по умолчанию.
func TestCreateTripRequestNormalizeDefaults(t *testing.T) {
	req := CreateTripRequest{}
	req.NormalizeDefaults()

	assert.Equal(t, "comfort", req.BudgetTier)
	assert.Equal(t, "car", req.Transport)
	assert.Equal(t, 1, req.GroupSize)
	assert.Equal(t, json.RawMessage("{}"), req.GroupComposition)
	assert.Equal(t, "multi_day", req.Format)
}

// TestCreateTripRequestNormalizeDefaults_NoOverwrite тестирует, что NormalizeDefaults
// не перезаписывает уже установленные значения.
func TestCreateTripRequestNormalizeDefaults_NoOverwrite(t *testing.T) {
	req := CreateTripRequest{
		BudgetTier:       "economy",
		Transport:        "bike",
		GroupSize:        5,
		GroupComposition: json.RawMessage(`{"adults":3}`),
		Format:           "weekend",
	}
	req.NormalizeDefaults()

	assert.Equal(t, "economy", req.BudgetTier)
	assert.Equal(t, "bike", req.Transport)
	assert.Equal(t, 5, req.GroupSize)
	assert.Equal(t, json.RawMessage(`{"adults":3}`), req.GroupComposition)
	assert.Equal(t, "weekend", req.Format)
}

// TestJoinTripRequestValidation тестирует валидацию DTO присоединения к поездке.
func TestJoinTripRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     JoinTripRequest
		wantErr string
	}{
		{
			name:    "пустой invite_token",
			req:     JoinTripRequest{DisplayName: "Маша"},
			wantErr: "invite_token обязателен",
		},
		{
			name:    "пустой display_name",
			req:     JoinTripRequest{InviteToken: "550e8400-e29b-41d4-a716-446655440000"},
			wantErr: "display_name обязателен",
		},
		{
			name:    "некорректный формат invite_token",
			req:     JoinTripRequest{InviteToken: "not-a-uuid", DisplayName: "Маша"},
			wantErr: "invite_token должен быть валидным UUID",
		},
		{
			name: "корректный запрос (минимальный)",
			req:  JoinTripRequest{InviteToken: "550e8400-e29b-41d4-a716-446655440000", DisplayName: "Маша"},
		},
		{
			name: "корректный запрос (полный)",
			req: JoinTripRequest{
				InviteToken: "550e8400-e29b-41d4-a716-446655440000",
				DisplayName: "Ребёнок Петя",
				Tags:        []string{"зоопарк", "парк"},
				IsChild:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.req.Validate()
			if tt.wantErr != "" {
				assert.Equal(t, tt.wantErr, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

// TestUpdateTripRequestValidation тестирует валидацию DTO обновления поездки.
func TestUpdateTripRequestValidation(t *testing.T) {
	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name    string
		req     UpdateTripRequest
		wantErr string
	}{
		{
			name: "пустой запрос (нет изменений) — ок",
			req:  UpdateTripRequest{},
		},
		{
			name:    "пустая дата начала",
			req:     UpdateTripRequest{DateFrom: strPtr("")},
			wantErr: "дата начала не может быть пустой",
		},
		{
			name:    "некорректный формат даты начала",
			req:     UpdateTripRequest{DateFrom: strPtr("10-04-2026")},
			wantErr: "некорректный формат даты начала (ожидается YYYY-MM-DD)",
		},
		{
			name:    "пустая дата окончания",
			req:     UpdateTripRequest{DateTo: strPtr("")},
			wantErr: "дата окончания не может быть пустой",
		},
		{
			name:    "некорректный формат даты окончания",
			req:     UpdateTripRequest{DateTo: strPtr("13/04/2026")},
			wantErr: "некорректный формат даты окончания (ожидается YYYY-MM-DD)",
		},
		{
			name:    "дата окончания раньше начала",
			req:     UpdateTripRequest{DateFrom: strPtr("2026-04-13"), DateTo: strPtr("2026-04-10")},
			wantErr: "дата окончания не может быть раньше даты начала",
		},
		{
			name:    "отрицательный бюджет",
			req:     UpdateTripRequest{BudgetRub: intPtr(-1)},
			wantErr: "бюджет не может быть отрицательным",
		},
		{
			name:    "некорректный уровень бюджета",
			req:     UpdateTripRequest{BudgetTier: strPtr("luxury")},
			wantErr: "допустимые уровни бюджета: economy, comfort, premium",
		},
		{
			name:    "некорректный вид транспорта",
			req:     UpdateTripRequest{Transport: strPtr("helicopter")},
			wantErr: "допустимые виды транспорта: car, public, walk, bike",
		},
		{
			name:    "нулевое количество участников",
			req:     UpdateTripRequest{GroupSize: intPtr(0)},
			wantErr: "количество участников должно быть не менее 1",
		},
		{
			name:    "некорректный формат",
			req:     UpdateTripRequest{Format: strPtr("expedition")},
			wantErr: "допустимые форматы поездки: day_trip, weekend, multi_day",
		},
		{
			name:    "некорректный vibe_vector_id",
			req:     UpdateTripRequest{VibeVectorID: strPtr("not-a-uuid")},
			wantErr: "vibe_vector_id должен быть валидным UUID",
		},
		{
			name: "корректное обновление (только бюджет)",
			req:  UpdateTripRequest{BudgetRub: intPtr(75000), BudgetTier: strPtr("premium")},
		},
		{
			name: "корректное обновление (полное)",
			req: UpdateTripRequest{
				DateFrom:     strPtr("2026-05-01"),
				DateTo:       strPtr("2026-05-05"),
				BudgetRub:    intPtr(100000),
				BudgetTier:   strPtr("premium"),
				Transport:    strPtr("public"),
				GroupSize:    intPtr(6),
				Format:       strPtr("multi_day"),
				VibeVectorID: strPtr("550e8400-e29b-41d4-a716-446655440000"),
			},
		},
		{
			name: "очистка vibe_vector_id (пустая строка) — ок",
			req:  UpdateTripRequest{VibeVectorID: strPtr("")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.req.Validate()
			if tt.wantErr != "" {
				assert.Equal(t, tt.wantErr, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

// TestTripStatusConstants тестирует корректность констант статусов поездки.
func TestTripStatusConstants(t *testing.T) {
	assert.Equal(t, TripStatus("planning"), TripStatusPlanning)
	assert.Equal(t, TripStatus("active"), TripStatusActive)
	assert.Equal(t, TripStatus("completed"), TripStatusCompleted)
	assert.Equal(t, TripStatus("cancelled"), TripStatusCancelled)
}

// TestTripMemberRoleConstants тестирует корректность констант ролей участников.
func TestTripMemberRoleConstants(t *testing.T) {
	assert.Equal(t, TripMemberRole("creator"), TripRoleCreator)
	assert.Equal(t, TripMemberRole("member"), TripRoleMember)
}

// TestValidBudgetTiers тестирует валидность карты допустимых уровней бюджета.
func TestValidBudgetTiers(t *testing.T) {
	assert.True(t, ValidBudgetTiers["economy"])
	assert.True(t, ValidBudgetTiers["comfort"])
	assert.True(t, ValidBudgetTiers["premium"])
	assert.False(t, ValidBudgetTiers["luxury"])
	assert.False(t, ValidBudgetTiers[""])
}

// TestValidTransports тестирует валидность карты допустимых видов транспорта.
func TestValidTransports(t *testing.T) {
	assert.True(t, ValidTransports["car"])
	assert.True(t, ValidTransports["public"])
	assert.True(t, ValidTransports["walk"])
	assert.True(t, ValidTransports["bike"])
	assert.False(t, ValidTransports["helicopter"])
	assert.False(t, ValidTransports[""])
}

// TestValidFormats тестирует валидность карты допустимых форматов поездки.
func TestValidFormats(t *testing.T) {
	assert.True(t, ValidFormats["day_trip"])
	assert.True(t, ValidFormats["weekend"])
	assert.True(t, ValidFormats["multi_day"])
	assert.False(t, ValidFormats["expedition"])
	assert.False(t, ValidFormats[""])
}
