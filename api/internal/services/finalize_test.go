// Файл finalize_test.go содержит unit-тесты для функций обогащения рекомендаций.
// Покрывает: generateReasonShort, computeTagsMatch, FinalizeRequest.NormalizeDefaults,
// JSON round-trip для FinalizeResponse.
package services

import (
	"encoding/json"
	"testing"

	"kudytudy-api/internal/models"
)

// TestGenerateReasonShortHighScore проверяет генерацию причины при высоком score.
func TestGenerateReasonShortHighScore(t *testing.T) {
	result := generateReasonShort(0.95, []string{"природа", "горы"}, "trail")
	if result != "Идеальное совпадение: природа, горы" {
		t.Errorf("ожидалось 'Идеальное совпадение: природа, горы', получено '%s'", result)
	}
}

// TestGenerateReasonShortMediumScore проверяет генерацию причины при среднем score.
func TestGenerateReasonShortMediumScore(t *testing.T) {
	result := generateReasonShort(0.80, []string{"виноградник"}, "winery")
	if result != "Высокое совпадение: виноградник" {
		t.Errorf("ожидалось 'Высокое совпадение: виноградник', получено '%s'", result)
	}
}

// TestGenerateReasonShortLowScore проверяет генерацию причины при низком score.
func TestGenerateReasonShortLowScore(t *testing.T) {
	result := generateReasonShort(0.50, nil, "farm")
	if result != "Подходит вам: farm" {
		t.Errorf("ожидалось 'Подходит вам: farm', получено '%s'", result)
	}
}

// TestGenerateReasonShortNoTagsNoCategory проверяет генерацию без тегов и категории.
func TestGenerateReasonShortNoTagsNoCategory(t *testing.T) {
	result := generateReasonShort(0.65, nil, "")
	if result != "Хорошее совпадение" {
		t.Errorf("ожидалось 'Хорошее совпадение', получено '%s'", result)
	}
}

// TestGenerateReasonShortMaxTags проверяет ограничение до 3 тегов.
func TestGenerateReasonShortMaxTags(t *testing.T) {
	tags := []string{"природа", "горы", "виноградник", "ферма", "каякинг"}
	result := generateReasonShort(0.92, tags, "trail")
	expected := "Идеальное совпадение: природа, горы, виноградник"
	if result != expected {
		t.Errorf("ожидалось '%s', получено '%s'", expected, result)
	}
}

// TestComputeTagsMatchBasic проверяет базовое пересечение тегов.
func TestComputeTagsMatchBasic(t *testing.T) {
	user := []string{"природа", "горы", "виноградник"}
	location := []string{"горы", "ферма", "виноградник", "озеро"}

	result := computeTagsMatch(user, location)
	if len(result) != 2 {
		t.Fatalf("ожидалось 2 совпадения, получено %d: %v", len(result), result)
	}
	// Порядок: горы, виноградник (порядок локации).
	if result[0] != "горы" || result[1] != "виноградник" {
		t.Errorf("ожидалось [горы, виноградник], получено %v", result)
	}
}

// TestComputeTagsMatchEmpty проверяет отсутствие пересечения.
func TestComputeTagsMatchEmpty(t *testing.T) {
	result := computeTagsMatch([]string{"a", "b"}, []string{"c", "d"})
	if len(result) != 0 {
		t.Errorf("ожидался пустой массив, получено %v", result)
	}
}

// TestComputeTagsMatchNilInput проверяет поведение с nil.
func TestComputeTagsMatchNilInput(t *testing.T) {
	result := computeTagsMatch(nil, []string{"a"})
	if len(result) != 0 {
		t.Errorf("ожидался пустой массив, получено %v", result)
	}

	result = computeTagsMatch([]string{"a"}, nil)
	if len(result) != 0 {
		t.Errorf("ожидался пустой массив, получено %v", result)
	}
}

// TestComputeTagsMatchCaseInsensitive проверяет регистронезависимость.
func TestComputeTagsMatchCaseInsensitive(t *testing.T) {
	result := computeTagsMatch([]string{"Природа", "ГОРЫ"}, []string{"природа", "горы"})
	if len(result) != 2 {
		t.Fatalf("ожидалось 2 совпадения, получено %d", len(result))
	}
}

// TestComputeTagsMatchDuplicates проверяет что дубликаты не считаются дважды.
func TestComputeTagsMatchDuplicates(t *testing.T) {
	result := computeTagsMatch([]string{"горы", "горы"}, []string{"горы", "горы", "горы"})
	if len(result) != 1 {
		t.Errorf("ожидалось 1 совпадение (дедупликация), получено %d", len(result))
	}
}

// TestFinalizeRequestNormalizeDefaults проверяет значения по умолчанию.
func TestFinalizeRequestNormalizeDefaults(t *testing.T) {
	tests := []struct {
		name          string
		input         models.FinalizeRequest
		expectedLimit int
	}{
		{"нулевой limit", models.FinalizeRequest{Limit: 0}, 10},
		{"отрицательный limit", models.FinalizeRequest{Limit: -5}, 10},
		{"нормальный limit", models.FinalizeRequest{Limit: 15}, 15},
		{"превышающий limit", models.FinalizeRequest{Limit: 100}, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.input
			req.NormalizeDefaults()
			if req.Limit != tt.expectedLimit {
				t.Errorf("ожидался limit=%d, получено %d", tt.expectedLimit, req.Limit)
			}
		})
	}
}

// TestFinalizeResponseJSON проверяет JSON-сериализацию обогащённого ответа.
func TestFinalizeResponseJSON(t *testing.T) {
	resp := models.FinalizeResponse{
		Recommendations: []models.LocationRecommendation{
			{
				LocationID:       "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				Score:            0.92,
				Name:             "Винодельня Лефкадия",
				Category:         "winery",
				DescriptionShort: "Премиальная винодельня с дегустацией.",
				Tags:             []string{"вино", "природа", "гастрономия"},
				PreviewImageURL:  "https://example.com/lefkadia.jpg",
				Latitude:         44.5,
				Longitude:        38.1,
				DensityLevel:     "green",
				ReasonShort:      "Идеальное совпадение: вино, природа",
				TagsMatch:        []string{"вино", "природа"},
				ChildFriendly:    true,
			},
		},
		TotalFound: 1,
		IsCurated:  false,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("ошибка маршалинга: %v", err)
	}

	// Проверка round-trip.
	var decoded models.FinalizeResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("ошибка анмаршалинга: %v", err)
	}

	if decoded.TotalFound != 1 {
		t.Errorf("total_found: ожидалось 1, получено %d", decoded.TotalFound)
	}
	if decoded.IsCurated {
		t.Errorf("is_curated: ожидалось false")
	}
	if len(decoded.Recommendations) != 1 {
		t.Fatalf("recommendations: ожидалось 1, получено %d", len(decoded.Recommendations))
	}

	rec := decoded.Recommendations[0]
	if rec.ReasonShort != "Идеальное совпадение: вино, природа" {
		t.Errorf("reason_short: ожидалось 'Идеальное совпадение: вино, природа', получено '%s'", rec.ReasonShort)
	}
	if len(rec.TagsMatch) != 2 {
		t.Errorf("tags_match: ожидалось 2, получено %d", len(rec.TagsMatch))
	}
	if !rec.ChildFriendly {
		t.Errorf("child_friendly: ожидалось true")
	}

	// Проверка наличия новых полей в JSON.
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("ошибка анмаршалинга raw: %v", err)
	}

	recs := raw["recommendations"].([]any)
	recMap := recs[0].(map[string]any)

	requiredFields := []string{"reason_short", "tags_match", "child_friendly", "is_curated"}
	for _, field := range requiredFields {
		if field == "is_curated" {
			if _, ok := raw[field]; !ok {
				t.Errorf("отсутствует поле %s в JSON", field)
			}
			continue
		}
		if _, ok := recMap[field]; !ok {
			t.Errorf("отсутствует поле %s в recommendation JSON", field)
		}
	}
}

// TestCuratedResponseJSON проверяет JSON curated ответа.
func TestCuratedResponseJSON(t *testing.T) {
	resp := models.FinalizeResponse{
		Recommendations: []models.LocationRecommendation{},
		TotalFound:      0,
		IsCurated:       true,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("ошибка маршалинга: %v", err)
	}

	var decoded models.FinalizeResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("ошибка анмаршалинга: %v", err)
	}

	if !decoded.IsCurated {
		t.Errorf("is_curated: ожидалось true")
	}
}
