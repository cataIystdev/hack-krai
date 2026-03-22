package services

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// WeatherTicker periodically fetches weather data and broadcasts it via WSHub.
type WeatherTicker struct {
	weatherService *WeatherService
	hub            *WSHub
	interval       time.Duration
	stopCh         chan struct{}
	logger         *zap.Logger
}

// NewWeatherTicker creates and starts a weather ticker.
func NewWeatherTicker(weatherService *WeatherService, hub *WSHub, interval time.Duration, logger *zap.Logger) *WeatherTicker {
	t := &WeatherTicker{
		weatherService: weatherService,
		hub:            hub,
		interval:       interval,
		stopCh:         make(chan struct{}),
		logger:         logger.Named("weather_ticker"),
	}

	go t.run()
	return t
}

func (t *WeatherTicker) run() {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	// Immediate first tick.
	t.tick()

	for {
		select {
		case <-ticker.C:
			t.tick()
		case <-t.stopCh:
			return
		}
	}
}

func (t *WeatherTicker) tick() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	points, err := t.weatherService.GetRegionWeather(ctx)
	if err != nil {
		t.logger.Warn("weather tick failed", zap.Error(err))
		return
	}

	msg := WeatherMsg{
		Type: "weather_update",
		Data: points,
		TS:   time.Now().UnixMilli(),
	}

	t.hub.BroadcastJSON(ChannelWeather, msg)

	t.logger.Debug("weather broadcast sent",
		zap.Int("points", len(points)),
		zap.Int("ws_clients", t.hub.ActiveConnections()),
	)
}

// Stop halts the ticker.
func (t *WeatherTicker) Stop() {
	close(t.stopCh)
}
