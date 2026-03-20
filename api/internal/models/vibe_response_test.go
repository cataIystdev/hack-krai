// Файл vibe_response_test.go содержит тесты для VoiceProfileResponse.
// Проверяет корректность JSON-сериализации screenshot-ready формата ответа.
package models

import (
	"encoding/json"
	"testing"
)

// TestVoiceProfileResponseJSON проверяет JSON-сериализацию VoiceProfileResponse.
// Убеждается, что все поля корректно сериализуются и десериализуются,
// включая вложенную структуру axes.
func TestVoiceProfileResponseJSON(t *testing.T) {
	resp := VoiceProfileResponse{
		Transcription: "Мне нравится отдыхать на природе",
		Axes: VibeAxesResponse{
			StressLevel:        0.7,
			SolitudeVsSocial:   -0.5,
			RelaxVsAdrenaline:  -0.3,
			GastroVsNature:     0.4,
			CultureVsAdventure: 0.3,
		},
		ExtractedTags:     []string{"горы", "лес", "ферма"},
		VibeSummary:       "Природный турист с элементами приключений",
		VibePassportTitle: "Исследователь Кубани",
		VectorID:          "550e8400-e29b-41d4-a716-446655440000",
		ProcessingTimeMs:  1842,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("ошибка маршалинга VoiceProfileResponse: %v", err)
	}

	// Десериализация обратно для проверки.
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("ошибка демаршалинга JSON: %v", err)
	}

	// Проверка наличия ключевых полей.
	requiredFields := []string{
		"transcription",
		"axes",
		"extracted_tags",
		"vibe_summary",
		"vibe_passport_title",
		"vector_id",
		"processing_time_ms",
	}
	for _, field := range requiredFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("отсутствует обязательное поле %q в JSON-ответе", field)
		}
	}

	// Проверка вложенной структуры axes.
	axesMap, ok := parsed["axes"].(map[string]any)
	if !ok {
		t.Fatal("поле axes должно быть объектом")
	}

	axesFields := []string{
		"stress_level",
		"solitude_vs_social",
		"relax_vs_adrenaline",
		"gastro_vs_nature",
		"culture_vs_adventure",
	}
	for _, field := range axesFields {
		if _, ok := axesMap[field]; !ok {
			t.Errorf("отсутствует поле %q в axes", field)
		}
	}

	// Проверка значений.
	if parsed["vibe_passport_title"] != "Исследователь Кубани" {
		t.Errorf("vibe_passport_title = %v, ожидалось 'Исследователь Кубани'", parsed["vibe_passport_title"])
	}

	if parsed["processing_time_ms"].(float64) != 1842.0 {
		t.Errorf("processing_time_ms = %v, ожидалось 1842", parsed["processing_time_ms"])
	}
}

// TestVoiceProfileResponseRoundTrip проверяет полный цикл сериализации/десериализации.
func TestVoiceProfileResponseRoundTrip(t *testing.T) {
	original := VoiceProfileResponse{
		Transcription: "Тестовый текст транскрипции",
		Axes: VibeAxesResponse{
			StressLevel:        0.5,
			SolitudeVsSocial:   0.0,
			RelaxVsAdrenaline:  -1.0,
			GastroVsNature:     1.0,
			CultureVsAdventure: -0.5,
		},
		ExtractedTags:     []string{"спа", "релакс", "тишина"},
		VibeSummary:       "Ценитель спокойного отдыха",
		VibePassportTitle: "Ценитель спокойствия",
		VectorID:          "test-uuid",
		ProcessingTimeMs:  2500,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("ошибка маршалинга: %v", err)
	}

	var restored VoiceProfileResponse
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("ошибка демаршалинга: %v", err)
	}

	if restored.Transcription != original.Transcription {
		t.Errorf("transcription = %q, ожидалось %q", restored.Transcription, original.Transcription)
	}
	if restored.Axes.StressLevel != original.Axes.StressLevel {
		t.Errorf("axes.stress_level = %v, ожидалось %v", restored.Axes.StressLevel, original.Axes.StressLevel)
	}
	if restored.VibePassportTitle != original.VibePassportTitle {
		t.Errorf("vibe_passport_title = %q, ожидалось %q", restored.VibePassportTitle, original.VibePassportTitle)
	}
	if restored.ProcessingTimeMs != original.ProcessingTimeMs {
		t.Errorf("processing_time_ms = %d, ожидалось %d", restored.ProcessingTimeMs, original.ProcessingTimeMs)
	}
	if len(restored.ExtractedTags) != len(original.ExtractedTags) {
		t.Errorf("кол-во extracted_tags = %d, ожидалось %d", len(restored.ExtractedTags), len(original.ExtractedTags))
	}
}
