package models

import (
	"testing"
)

func TestCreateReviewRequest_Validate(t *testing.T) {
	validUUID := "123e4567-e89b-12d3-a456-426614174000"
	invalidUUID := "not-a-uuid"

	tests := []struct {
		name     string
		req      CreateReviewRequest
		expected string
	}{
		{
			name: "valid request",
			req: CreateReviewRequest{
				LocationID: &validUUID,
				Rating:     5,
				Text:       "Отличное место!",
			},
			expected: "",
		},
		{
			name: "rating too low",
			req: CreateReviewRequest{
				LocationID: &validUUID,
				Rating:     0,
				Text:       "Ужас",
			},
			expected: "rating должен быть от 1 до 5",
		},
		{
			name: "rating too high",
			req: CreateReviewRequest{
				LocationID: &validUUID,
				Rating:     6,
				Text:       "Супер",
			},
			expected: "rating должен быть от 1 до 5",
		},
		{
			name: "missing location and booking",
			req: CreateReviewRequest{
				Rating: 4,
				Text:   "Норм",
			},
			expected: "хотя бы location_id или booking_id должны быть указаны",
		},
		{
			name: "invalid location_id uuid",
			req: CreateReviewRequest{
				LocationID: &invalidUUID,
				Rating:     5,
				Text:       "Тест",
			},
			expected: "location_id должен быть валидным UUID",
		},
		{
			name: "invalid booking_id uuid",
			req: CreateReviewRequest{
				BookingID: &invalidUUID,
				Rating:    5,
				Text:      "Тест",
			},
			expected: "booking_id должен быть валидным UUID",
		},
		{
			name: "text too long",
			req: CreateReviewRequest{
				LocationID: &validUUID,
				Rating:     5,
				Text:       string(make([]rune, 2001)),
			},
			expected: "text не должен превышать 2000 символов",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.req.Validate()
			if errStr != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, errStr)
			}
		})
	}
}
