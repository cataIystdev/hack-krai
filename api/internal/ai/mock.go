// Файл mock.go содержит детерминированные mock-функции для AI-клиентов.
// Используются в dev/staging режиме при отсутствии API-ключа.
// Mock-ответы имитируют реальные данные для тестирования полного пайплайна.
package ai

import (
	"math"
)

// mockTranscription возвращает детерминированную транскрипцию для тестирования.
// Текст описывает типичного туриста, предпочитающего природный отдых.
func mockTranscription() string {
	return "Мне нравится отдыхать на природе, подальше от города. " +
		"Люблю горы, лес, тишину. Хочу попробовать местную кухню, " +
		"посетить фермы и виноградники. Бюджет средний, не экономлю, " +
		"но и не трачу слишком много. Путешествую с семьёй, есть ребёнок 8 лет. " +
		"Хочется спокойного отдыха, но с элементами приключений — " +
		"пешие маршруты, каякинг, велосипеды."
}

// mockVibeAxes возвращает детерминированные оси vibe-профиля.
// Профиль описывает туриста, предпочитающего природный отдых с семьёй.
// Оси соответствуют GDD (kudytudy_gdd.md).
func mockVibeAxes() *VibeAxes {
	return &VibeAxes{
		StressLevel:        0.7,
		SolitudeVsSocial:   -0.5,
		RelaxVsAdrenaline:  -0.3,
		GastroVsNature:     0.4,
		CultureVsAdventure: 0.3,
		ExtractedTags:      []string{"горы", "лес", "ферма", "виноградник", "каякинг", "велосипед", "местная кухня"},
		VibeSummary:        "Семейный турист, предпочитающий природный отдых с элементами приключений. Интересуется местной гастрономией и агротуризмом.",
	}
}

// mockVibePassportTitle возвращает заголовок vibe-паспорта для demo-режима.
// Заголовок описывает архетип туриста на основе mock-осей.
func mockVibePassportTitle() string {
	return "Исследователь Кубани"
}

// mockEmbedding возвращает детерминированный вектор 3072 измерений.
// Вектор нормализован (единичная длина) для корректной работы Cosine Similarity.
// Значения генерируются синусоидой для создания реалистичного распределения.
func mockEmbedding() []float32 {
	const size = 3072
	vector := make([]float32, size)

	// Генерация детерминированного вектора через синусоидальную функцию.
	var norm float64
	for i := 0; i < size; i++ {
		val := math.Sin(float64(i)*0.1) * math.Cos(float64(i)*0.03)
		vector[i] = float32(val)
		norm += val * val
	}

	// Нормализация до единичной длины.
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range vector {
			vector[i] = float32(float64(vector[i]) / norm)
		}
	}

	return vector
}

// mockSceneEmbedding генерирует детерминированный вектор для сцены свайпа.
// Каждая сцена получает уникальный вектор на основе seed.
func mockSceneEmbedding(seed int) []float32 {
	const size = 3072
	vector := make([]float32, size)

	var norm float64
	for i := 0; i < size; i++ {
		val := math.Sin(float64(i+seed*100)*0.07) * math.Cos(float64(i+seed*50)*0.05)
		vector[i] = float32(val)
		norm += val * val
	}

	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range vector {
			vector[i] = float32(float64(vector[i]) / norm)
		}
	}

	return vector
}
