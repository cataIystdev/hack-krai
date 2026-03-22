package services

import (
	"context"
	"math"
	"testing"
	"time"

	"go.uber.org/zap"

	"kudytudy-api/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// haversineDistance
// ---------------------------------------------------------------------------

func TestHaversineDistance_SamePoint(t *testing.T) {
	d := haversineDistance(45.0, 39.0, 45.0, 39.0)
	assert.InDelta(t, 0.0, d, 0.001)
}

func TestHaversineDistance_SochiToKrasnodar(t *testing.T) {
	// Сочи — Краснодар ~172 км по прямой (по воздуху, не по дороге)
	d := haversineDistance(43.5855, 39.7231, 45.0355, 38.9753)
	assert.InDelta(t, 172.0, d, 15.0, "Сочи-Краснодар ~172 км по прямой")
}

func TestHaversineDistance_Symmetry(t *testing.T) {
	d1 := haversineDistance(43.5855, 39.7231, 44.6342, 39.135)
	d2 := haversineDistance(44.6342, 39.135, 43.5855, 39.7231)
	assert.InDelta(t, d1, d2, 0.001, "Расстояние должно быть симметричным")
}

func TestHaversineDistance_Positive(t *testing.T) {
	d := haversineDistance(43.0, 39.0, 44.0, 40.0)
	assert.Greater(t, d, 0.0)
}

func TestHaversineDistance_CloseCitiesKubani(t *testing.T) {
	// Краснодар — Горячий Ключ ~50 км
	d := haversineDistance(45.0355, 38.9753, 44.6342, 39.135)
	assert.InDelta(t, 50.0, d, 10.0)
}

// ---------------------------------------------------------------------------
// degreesToRadians
// ---------------------------------------------------------------------------

func TestDegreesToRadians_Zero(t *testing.T) {
	assert.InDelta(t, 0.0, degreesToRadians(0), 1e-9)
}

func TestDegreesToRadians_180(t *testing.T) {
	assert.InDelta(t, math.Pi, degreesToRadians(180), 1e-9)
}

func TestDegreesToRadians_90(t *testing.T) {
	assert.InDelta(t, math.Pi/2, degreesToRadians(90), 1e-9)
}

func TestDegreesToRadians_Negative(t *testing.T) {
	assert.InDelta(t, -math.Pi/2, degreesToRadians(-90), 1e-9)
}

// ---------------------------------------------------------------------------
// getStayDuration
// ---------------------------------------------------------------------------

func TestGetStayDuration_KnownCategories(t *testing.T) {
	cases := []struct {
		category string
		expected int
	}{
		{"winery", 90},
		{"farm", 60},
		{"trail", 120},
		{"nature", 90},
		{"cultural", 60},
		{"gastro", 75},
		{"beach", 120},
		{"resort", 180},
		{"camping", 60},
		{"guesthouse", 30},
		{"extreme", 120},
	}
	for _, tc := range cases {
		t.Run(tc.category, func(t *testing.T) {
			assert.Equal(t, tc.expected, getStayDuration(tc.category))
		})
	}
}

func TestGetStayDuration_UnknownCategory(t *testing.T) {
	assert.Equal(t, 60, getStayDuration("unknown"))
	assert.Equal(t, 60, getStayDuration(""))
}

// ---------------------------------------------------------------------------
// RouteService.mergeVectors
// ---------------------------------------------------------------------------

func newRouteService() *RouteService {
	logger, _ := zap.NewDevelopment()
	return NewRouteService(nil, nil, nil, nil, logger)
}

func TestMergeVectors_Empty(t *testing.T) {
	svc := newRouteService()
	result := svc.mergeVectors(nil)
	assert.Nil(t, result)
}

func TestMergeVectors_Single(t *testing.T) {
	svc := newRouteService()
	vec := []float32{0.5, 0.5, 0.5}
	result := svc.mergeVectors([][]float32{vec})
	assert.Equal(t, vec, result)
}

func TestMergeVectors_TwoEqual(t *testing.T) {
	svc := newRouteService()
	v1 := []float32{1.0, 0.0, 0.0}
	v2 := []float32{1.0, 0.0, 0.0}
	result := svc.mergeVectors([][]float32{v1, v2})
	require.NotNil(t, result)
	// Merged и нормализован — L2 норма = 1
	var norm float64
	for _, v := range result {
		norm += float64(v) * float64(v)
	}
	assert.InDelta(t, 1.0, math.Sqrt(norm), 1e-5, "Результат должен быть нормализован")
}

func TestMergeVectors_AverageCorrect(t *testing.T) {
	svc := newRouteService()
	v1 := []float32{1.0, 0.0}
	v2 := []float32{0.0, 1.0}
	result := svc.mergeVectors([][]float32{v1, v2})
	require.NotNil(t, result)
	require.Len(t, result, 2)
	// Среднее (0.5, 0.5), нормализованное = (0.707, 0.707)
	assert.InDelta(t, 0.707, float64(result[0]), 0.01)
	assert.InDelta(t, 0.707, float64(result[1]), 0.01)
}

// ---------------------------------------------------------------------------
// RouteService.determineAudience
// ---------------------------------------------------------------------------

func TestDetermineAudience_NoChildren(t *testing.T) {
	svc := newRouteService()
	cases := []string{"winery", "farm", "beach", "trail", "extreme", "gastro", "nature"}
	for _, cat := range cases {
		t.Run(cat, func(t *testing.T) {
			loc := models.Location{Category: cat}
			assert.Equal(t, models.AudienceAll, svc.determineAudience(loc, false))
		})
	}
}

func TestDetermineAudience_WithChildren_AdultOnly(t *testing.T) {
	svc := newRouteService()
	for _, cat := range []string{"winery", "gastro", "extreme"} {
		t.Run(cat, func(t *testing.T) {
			loc := models.Location{Category: cat}
			assert.Equal(t, models.AudienceAdults, svc.determineAudience(loc, true))
		})
	}
}

func TestDetermineAudience_WithChildren_ChildFriendly(t *testing.T) {
	svc := newRouteService()
	for _, cat := range []string{"farm", "beach", "nature", "resort", "camping"} {
		t.Run(cat, func(t *testing.T) {
			loc := models.Location{Category: cat, ChildFriendly: true}
			assert.Equal(t, models.AudienceChildren, svc.determineAudience(loc, true))
		})
	}
}

func TestDetermineAudience_WithChildren_NotChildFriendly(t *testing.T) {
	svc := newRouteService()
	loc := models.Location{Category: "farm", ChildFriendly: false}
	// farm is child-friendly category but ChildFriendly flag is false
	assert.Equal(t, models.AudienceAll, svc.determineAudience(loc, true))
}

// ---------------------------------------------------------------------------
// RouteService.assignDaysAndSlots
// ---------------------------------------------------------------------------

func TestAssignDaysAndSlots_OneDay(t *testing.T) {
	svc := newRouteService()
	points := make([]models.RoutePoint, 3)
	svc.assignDaysAndSlots(points, 1)
	assert.Equal(t, 1, points[0].DayNumber)
	assert.Equal(t, models.TimeSlotMorning, points[0].TimeSlot)
	assert.Equal(t, models.TimeSlotAfternoon, points[1].TimeSlot)
	assert.Equal(t, models.TimeSlotEvening, points[2].TimeSlot)
}

func TestAssignDaysAndSlots_MultipleDays(t *testing.T) {
	svc := newRouteService()
	// 7 точек на 3 дня: 3+3+1
	points := make([]models.RoutePoint, 7)
	svc.assignDaysAndSlots(points, 3)
	assert.Equal(t, 1, points[0].DayNumber)  // day 1 morning
	assert.Equal(t, 2, points[3].DayNumber)  // day 2 morning
	assert.Equal(t, 3, points[6].DayNumber)  // day 3 morning
}

func TestAssignDaysAndSlots_OverflowClamped(t *testing.T) {
	svc := newRouteService()
	// 10 точек на 2 дня — лишние идут в последний день
	points := make([]models.RoutePoint, 10)
	svc.assignDaysAndSlots(points, 2)
	for _, p := range points {
		assert.LessOrEqual(t, p.DayNumber, 2)
		assert.GreaterOrEqual(t, p.DayNumber, 1)
	}
}

// ---------------------------------------------------------------------------
// RouteService.estimateCost
// ---------------------------------------------------------------------------

func TestEstimateCost_Basic(t *testing.T) {
	svc := newRouteService()
	scored := []scoredLocation{
		{location: models.Location{PricePerNight: 3000}},
		{location: models.Location{PricePerNight: 5000}},
	}
	// avg = (3000+5000)/2 = 4000, days = 3 → 12000
	cost := svc.estimateCost(scored, 3, "comfort")
	assert.Equal(t, 12000, cost)
}

func TestEstimateCost_Empty(t *testing.T) {
	svc := newRouteService()
	cost := svc.estimateCost(nil, 3, "economy")
	assert.Equal(t, 0, cost)
}

func TestEstimateCost_SingleLocation(t *testing.T) {
	svc := newRouteService()
	scored := []scoredLocation{
		{location: models.Location{PricePerNight: 7000}},
	}
	cost := svc.estimateCost(scored, 2, "comfort")
	assert.Equal(t, 14000, cost)
}

// ---------------------------------------------------------------------------
// RouteService.calculateTripDays
// ---------------------------------------------------------------------------

func TestCalculateTripDays_ThreeDays(t *testing.T) {
	svc := newRouteService()
	trip := &models.Trip{
		DateFrom: time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
		DateTo:   time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC),
	}
	assert.Equal(t, 3, svc.calculateTripDays(trip))
}

func TestCalculateTripDays_SameDay(t *testing.T) {
	svc := newRouteService()
	trip := &models.Trip{
		DateFrom: time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
		DateTo:   time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
	}
	assert.Equal(t, 1, svc.calculateTripDays(trip))
}

func TestCalculateTripDays_OneDay(t *testing.T) {
	svc := newRouteService()
	trip := &models.Trip{
		DateFrom: time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
		DateTo:   time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC),
	}
	assert.Equal(t, 2, svc.calculateTripDays(trip))
}

// ---------------------------------------------------------------------------
// RouteService.generateSummary
// ---------------------------------------------------------------------------

func TestGenerateSummary_ContainsKeyInfo(t *testing.T) {
	svc := newRouteService()
	trip := &models.Trip{
		Format:    "weekend",
		Transport: "car",
		BudgetTier: "comfort",
	}
	points := []models.RoutePoint{
		{Category: "winery"},
		{Category: "winery"},
		{Category: "nature"},
	}
	summary := svc.generateSummary(trip, points, 150.5, 2)
	assert.Contains(t, summary, "выходные")
	assert.Contains(t, summary, "на автомобиле")
	assert.Contains(t, summary, "winery")
	assert.Contains(t, summary, "comfort")
	assert.Contains(t, summary, "150")
	assert.Contains(t, summary, "2")
}

func TestGenerateSummary_UnknownFormatTransport(t *testing.T) {
	svc := newRouteService()
	trip := &models.Trip{
		Format:     "custom_format",
		Transport:  "rocket",
		BudgetTier: "economy",
	}
	points := []models.RoutePoint{{Category: "trail"}}
	summary := svc.generateSummary(trip, points, 0, 1)
	// Falls back to raw values
	assert.Contains(t, summary, "custom_format")
	assert.Contains(t, summary, "rocket")
}

// ---------------------------------------------------------------------------
// WeatherService.makePoint — severity logic
// ---------------------------------------------------------------------------

func TestWeatherServiceMakePoint_Storm(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewWeatherService(logger)
	point := svc.makePoint("Лаго-Наки", 44.0, 40.0, time.Now(), 8, "storm", 10)
	assert.Equal(t, "warning", point.Severity)
	assert.True(t, point.NeedsRebuild)
}

func TestWeatherServiceMakePoint_HighWind(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewWeatherService(logger)
	point := svc.makePoint("Туапсе", 44.0, 39.0, time.Now(), 15, "clouds", 15)
	assert.Equal(t, "warning", point.Severity)
	assert.True(t, point.NeedsRebuild)
}

func TestWeatherServiceMakePoint_Clear(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewWeatherService(logger)
	point := svc.makePoint("Краснодар", 45.0, 39.0, time.Now(), 22, "clear", 4)
	assert.Equal(t, "normal", point.Severity)
	assert.False(t, point.NeedsRebuild)
}

// ---------------------------------------------------------------------------
// WeatherService.GetTravelAdvisory — nearest point
// ---------------------------------------------------------------------------

func TestGetTravelAdvisory_ReturnsNearest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewWeatherService(logger)
	// Краснодар coords: 45.0355, 38.9753 — должен вернуть ближайшую точку
	advisory, err := svc.GetTravelAdvisory(context.TODO(), 45.0, 38.9)
	require.NoError(t, err)
	require.NotNil(t, advisory)
	assert.Equal(t, "Краснодар", advisory.Point)
}

func TestGetTravelAdvisory_SochiNearest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewWeatherService(logger)
	// Сочи coords: 43.5855, 39.7231
	advisory, err := svc.GetTravelAdvisory(context.TODO(), 43.5, 39.8)
	require.NoError(t, err)
	assert.Equal(t, "Сочи", advisory.Point)
}

func TestGetTravelAdvisory_AlwaysReturnsPoint(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewWeatherService(logger)
	advisory, err := svc.GetTravelAdvisory(context.TODO(), 0.0, 0.0)
	require.NoError(t, err)
	assert.NotNil(t, advisory)
	assert.NotEmpty(t, advisory.Point)
}

func TestGetRegionWeather_ReturnsAllPoints(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewWeatherService(logger)
	points, err := svc.GetRegionWeather(context.TODO())
	require.NoError(t, err)
	assert.Len(t, points, 5)
	for _, p := range points {
		assert.NotEmpty(t, p.Point)
		assert.NotEmpty(t, p.Condition)
	}
}
