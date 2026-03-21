// Файл map_service_test.go содержит unit-тесты для MapService.
// Тестирует логику demo-профилей (getDemoProfile) и ValidationError.
package services

import "testing"

// TestGetDemoProfile_KnownProfiles проверяет получение конфигурации всех demo-профилей.
func TestGetDemoProfile_KnownProfiles(t *testing.T) {
	profiles := []struct {
		name            string
		wantDescription string
		wantRecommCount int
	}{
		{
			name:            "calm_wine_mountains",
			wantDescription: "Спокойный отдых: горы, вино, тишина",
			wantRecommCount: 7,
		},
		{
			name:            "active_adventure",
			wantDescription: "Активный отдых: каньоны, рафтинг, горы",
			wantRecommCount: 7,
		},
		{
			name:            "family_kids",
			wantDescription: "Семейный отдых: фермы, дети, природа",
			wantRecommCount: 7,
		},
		{
			name:            "gastro_cultural",
			wantDescription: "Гастрономия и культура: вино, сыр, история",
			wantRecommCount: 7,
		},
	}

	for _, tt := range profiles {
		t.Run(tt.name, func(t *testing.T) {
			profile := getDemoProfile(tt.name)

			if profile.Description != tt.wantDescription {
				t.Errorf("Description = %q, ожидалось %q", profile.Description, tt.wantDescription)
			}

			if len(profile.Recommendations) != tt.wantRecommCount {
				t.Errorf("Recommendations count = %d, ожидалось %d", len(profile.Recommendations), tt.wantRecommCount)
			}

			// Проверка скоров: должны быть в диапазоне (0, 1] и убывать.
			for i, rec := range profile.Recommendations {
				if rec.Score <= 0 || rec.Score > 1.0 {
					t.Errorf("Рекомендация %d (%s): скор %f вне диапазона (0, 1]", i, rec.Name, rec.Score)
				}
				if rec.Name == "" {
					t.Errorf("Рекомендация %d: пустое имя", i)
				}
				if i > 0 && rec.Score > profile.Recommendations[i-1].Score {
					t.Errorf("Рекомендации не отсортированы по убыванию скора: [%d]=%f > [%d]=%f",
						i, rec.Score, i-1, profile.Recommendations[i-1].Score)
				}
			}
		})
	}
}

// TestGetDemoProfile_UnknownFallback проверяет fallback для неизвестного профиля.
func TestGetDemoProfile_UnknownFallback(t *testing.T) {
	profile := getDemoProfile("unknown_profile")
	defaultProfile := getDemoProfile("calm_wine_mountains")

	if profile.Description != defaultProfile.Description {
		t.Errorf("Fallback: Description = %q, ожидалось %q", profile.Description, defaultProfile.Description)
	}

	if len(profile.Recommendations) != len(defaultProfile.Recommendations) {
		t.Errorf("Fallback: Recommendations count = %d, ожидалось %d",
			len(profile.Recommendations), len(defaultProfile.Recommendations))
	}
}

// TestGetDemoProfile_UniqueRecommendations проверяет уникальность рекомендаций в профиле.
func TestGetDemoProfile_UniqueRecommendations(t *testing.T) {
	profileNames := []string{"calm_wine_mountains", "active_adventure", "family_kids", "gastro_cultural"}

	for _, name := range profileNames {
		t.Run(name, func(t *testing.T) {
			profile := getDemoProfile(name)
			seen := make(map[string]bool)

			for _, rec := range profile.Recommendations {
				if seen[rec.Name] {
					t.Errorf("Дубликат рекомендации: %s", rec.Name)
				}
				seen[rec.Name] = true
			}
		})
	}
}

// TestGetDemoProfile_DifferentProfiles проверяет различие между профилями.
func TestGetDemoProfile_DifferentProfiles(t *testing.T) {
	calm := getDemoProfile("calm_wine_mountains")
	active := getDemoProfile("active_adventure")

	// Первые рекомендации должны различаться.
	if calm.Recommendations[0].Name == active.Recommendations[0].Name {
		t.Error("Профили calm_wine_mountains и active_adventure имеют одинаковую топ-рекомендацию")
	}
}

// TestValidationError_Error проверяет реализацию интерфейса error.
func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{Message: "тестовая ошибка"}
	if err.Error() != "тестовая ошибка" {
		t.Errorf("Error() = %q, ожидалось %q", err.Error(), "тестовая ошибка")
	}
}
