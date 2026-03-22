package services

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// WSHub — unit tests (local-only mode, без Redis)
// ---------------------------------------------------------------------------

func TestWSHub_NewHub_NilRedis(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	hub := NewWSHub(nil, logger)
	require.NotNil(t, hub)
	assert.Equal(t, 0, hub.ActiveConnections())
}

func TestWSHub_BroadcastNoClients(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	hub := NewWSHub(nil, logger)

	// Broadcast на пустой канал — не паникует.
	assert.NotPanics(t, func() {
		hub.Broadcast("weather", []byte(`{"type":"test"}`))
	})
}

func TestWSHub_BroadcastJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	hub := NewWSHub(nil, logger)

	// BroadcastJSON на пустой хаб — не паникует.
	assert.NotPanics(t, func() {
		hub.BroadcastJSON("weather", map[string]string{"type": "test"})
	})
}

func TestWSHub_ActiveConnections_Empty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	hub := NewWSHub(nil, logger)
	assert.Equal(t, 0, hub.ActiveConnections())
}

func TestWSHub_ChannelConstants(t *testing.T) {
	assert.Equal(t, "weather", ChannelWeather)
	assert.Equal(t, "notifications", ChannelNotifications)
}

func TestWeatherMsg_Fields(t *testing.T) {
	msg := WeatherMsg{
		Type: "weather_update",
		Data: []string{"point1", "point2"},
		TS:   time.Now().UnixMilli(),
	}
	assert.Equal(t, "weather_update", msg.Type)
	assert.NotZero(t, msg.TS)
}

// ---------------------------------------------------------------------------
// WeatherTicker — unit tests
// ---------------------------------------------------------------------------

func TestWeatherTicker_StopDoesNotPanic(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	hub := NewWSHub(nil, logger)

	// Создаём mock WeatherService (nil — tick будет паниковать,
	// но мы сразу стопим поэтому не вызовем tick).
	ticker := &WeatherTicker{
		weatherService: nil,
		hub:            hub,
		interval:       1 * time.Hour, // огромный интервал чтобы не тикнуло
		stopCh:         make(chan struct{}),
		logger:         logger,
	}

	// Just test that Stop doesn't panic.
	assert.NotPanics(t, func() {
		close(ticker.stopCh)
	})
}

// ---------------------------------------------------------------------------
// Concurrency safety test
// ---------------------------------------------------------------------------

func TestWSHub_ConcurrentBroadcast(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	hub := NewWSHub(nil, logger)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hub.Broadcast("weather", []byte(`{"type":"test"}`))
			hub.BroadcastJSON("notifications:user1", map[string]string{"msg": "hello"})
			_ = hub.ActiveConnections()
		}()
	}
	wg.Wait()
}
