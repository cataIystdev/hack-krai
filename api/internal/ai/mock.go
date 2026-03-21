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

// mockLocationData возвращает детерминированные данные локации для demo-режима.
// Данные соответствуют примеру из GDD Feature 5 (Козья ферма дяди Вани).
func mockLocationData() *LocationData {
	return &LocationData{
		NameSuggestion:      "Козья ферма дяди Вани",
		DescriptionShort:    "Крафтовые сыры и горный отдых с видом на закат",
		DescriptionLiterary: "Крафтовые сыры с благородной плесенью по авторской рецептуре. 12 альпийских коз на горном пастбище. Два уютных гостевых домика с видом на закат. Идеальное место для тех, кто ищет тишину, природу и настоящий деревенский колорит.",
		Tags:                []string{"сыр", "козы", "ночлег", "тишина", "горы", "ферма"},
		Category:            "farm",
		PricePerNight:       5000,
		Amenities:           []string{"домики", "парковка", "дегустация"},
		Capacity:            4,
	}
}

// mockOnboardingTranscription возвращает детерминированную транскрипцию
// голосового описания хоста для demo-режима.
func mockOnboardingTranscription() string {
	return "ну я дядя ваня делаю сыр с плесенью у меня козы двенадцать голов " +
		"два домика есть для гостей пять тысяч за ночь " +
		"приезжайте у нас тишина горы красота закаты офигенные " +
		"парковка есть можно дегустацию провести"
}

