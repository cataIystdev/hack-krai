// Файл onboarding_test.go содержит unit-тесты для моделей онбординга хостов.
// Проверяет корректность типов задач, статусов и сообщений прогресса.
package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTaskTypeConstants проверяет корректность констант типов задач.
func TestTaskTypeConstants(t *testing.T) {
	assert.Equal(t, TaskType("onboarding"), TaskTypeOnboarding)
	assert.Equal(t, TaskType("splatting"), TaskTypeSplatting)
	assert.Equal(t, TaskType("story_generation"), TaskTypeStoryGeneration)
	assert.Equal(t, TaskType("voice_profile"), TaskTypeVoiceProfile)
}

// TestTaskStatusConstants проверяет корректность констант статусов.
func TestTaskStatusConstants(t *testing.T) {
	assert.Equal(t, TaskStatus("queued"), TaskStatusQueued)
	assert.Equal(t, TaskStatus("processing"), TaskStatusProcessing)
	assert.Equal(t, TaskStatus("completed"), TaskStatusCompleted)
	assert.Equal(t, TaskStatus("failed"), TaskStatusFailed)
}

// TestProgressMessage проверяет возврат корректных сообщений прогресса.
func TestProgressMessage(t *testing.T) {
	tests := []struct {
		progress int
		expected string
	}{
		{0, "задача в очереди"},
		{10, "распознавание речи..."},
		{20, "распознавание речи..."},
		{30, "анализ описания и извлечение данных..."},
		{50, "анализ описания и извлечение данных..."},
		{60, "генерация вайб-профиля локации..."},
		{70, "генерация вайб-профиля локации..."},
		{80, "создание черновика локации..."},
		{90, "создание черновика локации..."},
		{95, "финализация..."},
		{100, "готово"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			msg := ProgressMessage(tt.progress)
			assert.Equal(t, tt.expected, msg,
				"ProgressMessage(%d) = %q, ожидалось %q", tt.progress, msg, tt.expected)
		})
	}
}

// TestOnboardingResultJSON проверяет JSON-сериализацию OnboardingResult.
func TestOnboardingResultJSON(t *testing.T) {
	result := OnboardingResult{
		NameSuggestion:      "Козья ферма дяди Вани",
		DescriptionShort:    "Крафтовые сыры и горный отдых",
		DescriptionLiterary: "Крафтовые сыры с благородной плесенью по авторской рецептуре.",
		Tags:                []string{"сыр", "козы", "ночлег"},
		Category:            "farm",
		PricePerNight:       5000,
		Amenities:           []string{"домики", "парковка"},
		Capacity:            2,
	}

	assert.NotEmpty(t, result.NameSuggestion)
	assert.Len(t, result.Tags, 3)
	assert.Equal(t, 5000, result.PricePerNight)
	assert.Equal(t, 2, result.Capacity)
	assert.Equal(t, "farm", result.Category)
}
