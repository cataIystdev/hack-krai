package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"kudytudy-api/internal/models"
)

// buildStoryText — чистая функция, тестируем напрямую через пустой receiver.
// tts=nil → fallback-режим без ElevenLabs (только текст).
func newTestStorytellingService() *StorytellingService {
	return &StorytellingService{tts: nil}
}

func TestBuildStoryText_FullLocation(t *testing.T) {
	svc := newTestStorytellingService()
	loc := models.Location{
		Name:             "Абрау-Дюрсо",
		Category:         "winery",
		DescriptionShort: "Знаменитая винодельня на берегу озера.",
	}
	point := models.RoutePointDB{
		StayDurationMin:  90,
		WeatherCondition: "rain",
	}
	text := svc.buildStoryText(loc, point)

	assert.Contains(t, text, "Абрау-Дюрсо")
	assert.Contains(t, text, "Знаменитая винодельня")
	assert.Contains(t, text, "winery")
	assert.Contains(t, text, "90")
	assert.Contains(t, text, "rain")
	assert.Contains(t, text, "маршрута")
}

func TestBuildStoryText_NoDescriptionShort(t *testing.T) {
	svc := newTestStorytellingService()
	loc := models.Location{
		Name:     "Козья ферма",
		Category: "farm",
	}
	point := models.RoutePointDB{StayDurationMin: 60}
	text := svc.buildStoryText(loc, point)

	assert.Contains(t, text, "Козья ферма")
	assert.Contains(t, text, "farm")
	assert.NotEmpty(t, text)
}

func TestBuildStoryText_NoCategory(t *testing.T) {
	svc := newTestStorytellingService()
	loc := models.Location{
		Name:             "Неизвестная точка",
		DescriptionShort: "Описание.",
	}
	point := models.RoutePointDB{StayDurationMin: 45}
	text := svc.buildStoryText(loc, point)

	assert.Contains(t, text, "Неизвестная точка")
	assert.NotContains(t, text, "категории")
}

func TestBuildStoryText_NoWeather(t *testing.T) {
	svc := newTestStorytellingService()
	loc := models.Location{
		Name:     "Парк Галицкого",
		Category: "cultural",
	}
	point := models.RoutePointDB{
		StayDurationMin:  120,
		WeatherCondition: "",
	}
	text := svc.buildStoryText(loc, point)

	assert.NotContains(t, text, "погода")
	assert.Contains(t, text, "Парк Галицкого")
}

func TestBuildStoryText_AlwaysEndsWithMovingOn(t *testing.T) {
	svc := newTestStorytellingService()
	loc := models.Location{Name: "Точка A"}
	point := models.RoutePointDB{}
	text := svc.buildStoryText(loc, point)

	// Always ends with the movement prompt
	assert.True(t, strings.HasSuffix(strings.TrimSpace(text), "двигайтесь дальше в темпе маршрута."))
}

func TestBuildStoryText_DifferentLocations(t *testing.T) {
	svc := newTestStorytellingService()
	cases := []struct {
		name     string
		category string
	}{
		{"Медовые водопады", "nature"},
		{"Подводное погружение", "extreme"},
		{"Сыроварня Коваленко", "gastro"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			loc := models.Location{Name: tc.name, Category: tc.category}
			point := models.RoutePointDB{StayDurationMin: 60}
			text := svc.buildStoryText(loc, point)
			assert.Contains(t, text, tc.name)
			assert.Contains(t, text, tc.category)
			assert.NotEmpty(t, text)
		})
	}
}
