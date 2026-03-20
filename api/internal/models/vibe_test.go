// Файл vibe_test.go содержит unit-тесты для моделей модуля vibe-профилирования.
// Покрывает валидацию SwipeRequest и корректность констант.
package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSwipeRequest_Validate проверяет валидацию запроса свайпа.
func TestSwipeRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     SwipeRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "валидный запрос right",
			req: SwipeRequest{
				SceneID:   "a1b2c3d4-1111-4000-8000-000000000001",
				Direction: "right",
			},
			wantErr: false,
		},
		{
			name: "валидный запрос left",
			req: SwipeRequest{
				SceneID:   "a1b2c3d4-2222-4000-8000-000000000002",
				Direction: "left",
			},
			wantErr: false,
		},
		{
			name: "пустой scene_id",
			req: SwipeRequest{
				SceneID:   "",
				Direction: "right",
			},
			wantErr: true,
			errMsg:  "scene_id обязателен",
		},
		{
			name: "невалидный UUID scene_id",
			req: SwipeRequest{
				SceneID:   "not-a-uuid",
				Direction: "right",
			},
			wantErr: true,
			errMsg:  "scene_id должен быть валидным UUID",
		},
		{
			name: "пустой direction",
			req: SwipeRequest{
				SceneID:   "a1b2c3d4-1111-4000-8000-000000000001",
				Direction: "",
			},
			wantErr: true,
			errMsg:  "direction обязателен",
		},
		{
			name: "невалидный direction",
			req: SwipeRequest{
				SceneID:   "a1b2c3d4-1111-4000-8000-000000000001",
				Direction: "up",
			},
			wantErr: true,
			errMsg:  "direction должен быть 'right' или 'left'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSwipeDirection_Constants проверяет корректность констант направлений свайпа.
func TestSwipeDirection_Constants(t *testing.T) {
	assert.Equal(t, SwipeDirection("right"), SwipeRight)
	assert.Equal(t, SwipeDirection("left"), SwipeLeft)

	assert.True(t, ValidSwipeDirections[SwipeRight])
	assert.True(t, ValidSwipeDirections[SwipeLeft])
	assert.False(t, ValidSwipeDirections[SwipeDirection("up")])
	assert.False(t, ValidSwipeDirections[SwipeDirection("")])
}
