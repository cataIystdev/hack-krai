package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
)

// ---------------------------------------------------------------------------
// WeatherTicker — unit tests
// ---------------------------------------------------------------------------

func TestNewWeatherTicker_Construction(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	hub := NewWSHub(nil, logger)

	// WeatherService needs to be non-nil for the ticker to call GetRegionWeather.
	// We create a minimal WeatherService (no external deps).
	ws := &WeatherService{logger: logger}

	ticker := NewWeatherTicker(ws, hub, 10*time.Second, logger)
	require.NotNil(t, ticker)

	// Stop immediately — should not panic.
	ticker.Stop()
}

func TestWeatherTicker_TickSendsToHub(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	hub := NewWSHub(nil, logger)

	// Create a WeatherService that returns mock data.
	ws := &WeatherService{logger: logger}

	ticker := &WeatherTicker{
		weatherService: ws,
		hub:            hub,
		interval:       1 * time.Hour,
		stopCh:         make(chan struct{}),
		logger:         logger,
	}

	// Tick should not panic even with no connected clients.
	assert.NotPanics(t, func() {
		ticker.tick()
	})

	ticker.Stop()
}

func TestWeatherService_GetRegionWeather(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ws := &WeatherService{logger: logger}

	points, err := ws.GetRegionWeather(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, points)

	// Verify the weather points have required fields populated.
	for _, pt := range points {
		assert.NotEmpty(t, pt.Name, "Название точки не должно быть пустым")
		assert.NotZero(t, pt.Lat, "Широта должна быть ненулевой")
		assert.NotZero(t, pt.Lon, "Долгота должна быть ненулевой")
	}
}

func TestWeatherService_GetTravelAdvisory(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ws := &WeatherService{logger: logger}

	pt, err := ws.GetTravelAdvisory(context.Background(), 45.0, 39.0)
	require.NoError(t, err)
	require.NotNil(t, pt)
	assert.NotEmpty(t, pt.Condition)
}

func TestWeatherPointModel(t *testing.T) {
	pt := models.WeatherPoint{
		Name:      "Краснодар",
		Lat:       45.0,
		Lon:       39.0,
		Condition: "clear",
	}
	assert.Equal(t, "Краснодар", pt.Name)
	assert.Equal(t, 45.0, pt.Lat)
}
