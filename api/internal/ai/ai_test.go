// Файл ai_test.go содержит unit-тесты для пакета ai.
// Проверяет mock-функции, парсинг ответов и формат данных.
// Оси соответствуют GDD (kudytudy_gdd.md).
package ai

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockTranscription проверяет, что mock-транскрипция не пустая и на русском.
func TestMockTranscription(t *testing.T) {
	text := mockTranscription()
	assert.NotEmpty(t, text)
	assert.Contains(t, text, "природе")
	assert.Contains(t, text, "горы")
}

// TestMockVibeAxes проверяет mock-оси vibe-профиля по GDD.
func TestMockVibeAxes(t *testing.T) {
	axes := mockVibeAxes()
	require.NotNil(t, axes)

	// stress_level в [0, 1].
	assert.GreaterOrEqual(t, axes.StressLevel, 0.0)
	assert.LessOrEqual(t, axes.StressLevel, 1.0)

	// Остальные оси в [-1, 1] (GDD).
	assert.GreaterOrEqual(t, axes.SolitudeVsSocial, -1.0)
	assert.LessOrEqual(t, axes.SolitudeVsSocial, 1.0)
	assert.GreaterOrEqual(t, axes.RelaxVsAdrenaline, -1.0)
	assert.LessOrEqual(t, axes.RelaxVsAdrenaline, 1.0)
	assert.GreaterOrEqual(t, axes.GastroVsNature, -1.0)
	assert.LessOrEqual(t, axes.GastroVsNature, 1.0)
	assert.GreaterOrEqual(t, axes.CultureVsAdventure, -1.0)
	assert.LessOrEqual(t, axes.CultureVsAdventure, 1.0)

	// Теги не пустые.
	assert.NotEmpty(t, axes.ExtractedTags)
	assert.GreaterOrEqual(t, len(axes.ExtractedTags), 3)

	// VibeSummary не пустой.
	assert.NotEmpty(t, axes.VibeSummary)
}

// TestMockEmbedding проверяет mock-эмбеддинг.
func TestMockEmbedding(t *testing.T) {
	vec := mockEmbedding()

	// Размерность 3072.
	assert.Len(t, vec, 3072)

	// Проверка нормализации: длина вектора должна быть ~1.0.
	var norm float64
	for _, v := range vec {
		norm += float64(v) * float64(v)
	}
	norm = math.Sqrt(norm)
	assert.InDelta(t, 1.0, norm, 0.01, "вектор должен быть нормализован")
}

// TestMockSceneEmbedding проверяет уникальность mock-векторов сцен.
func TestMockSceneEmbedding(t *testing.T) {
	vec1 := mockSceneEmbedding(1)
	vec2 := mockSceneEmbedding(2)

	assert.Len(t, vec1, 3072)
	assert.Len(t, vec2, 3072)

	// Вектора разных сцен должны отличаться.
	different := false
	for i := range vec1 {
		if vec1[i] != vec2[i] {
			different = true
			break
		}
	}
	assert.True(t, different, "вектора разных сцен должны отличаться")
}

// TestExtractJSONFromText проверяет парсинг JSON из текста LLM (GDD-формат).
func TestExtractJSONFromText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{
			name: "чистый JSON (GDD формат)",
			text: `{"stress_level": 0.5, "solitude_vs_social": -0.3, "relax_vs_adrenaline": 0.4, "gastro_vs_nature": -0.2, "culture_vs_adventure": 0.6, "extracted_tags": ["горы"], "vibe_summary": "тест"}`,
		},
		{
			name: "JSON в markdown",
			text: "Вот результат:\n```json\n{\"stress_level\": 0.5, \"solitude_vs_social\": -0.3, \"relax_vs_adrenaline\": 0.4, \"gastro_vs_nature\": -0.2, \"culture_vs_adventure\": 0.6, \"extracted_tags\": [\"горы\"], \"vibe_summary\": \"тест\"}\n```",
		},
		{
			name:    "нет JSON",
			text:    "просто текст без JSON",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			axes, err := extractJSONFromText(tt.text)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 0.5, axes.StressLevel)
				assert.Equal(t, []string{"горы"}, axes.ExtractedTags)
			}
		})
	}
}
