package services

import (
	"context"
	"time"

	"go.uber.org/zap"

	"kudytudy-api/internal/models"
)

// WeatherService — mock-friendly погодный сервис для региональных точек.
type WeatherService struct {
	logger *zap.Logger
}

func NewWeatherService(logger *zap.Logger) *WeatherService {
	return &WeatherService{logger: logger.Named("weather_service")}
}

func (s *WeatherService) GetRegionWeather(ctx context.Context) ([]models.WeatherPoint, error) {
	_ = ctx
	now := time.Now().UTC()
	points := []models.WeatherPoint{
		s.makePoint("Сочи", 43.5855, 39.7231, now, 17, "rain", 11),
		s.makePoint("Горячий Ключ", 44.6342, 39.1350, now, 19, "clear", 5),
		s.makePoint("Абрау-Дюрсо", 44.6995, 37.6003, now, 16, "clouds", 7),
		s.makePoint("Лаго-Наки", 44.0886, 40.0170, now, 8, "storm", 18),
		s.makePoint("Краснодар", 45.0355, 38.9753, now, 21, "clear", 4),
	}
	return points, nil
}

func (s *WeatherService) GetTravelAdvisory(ctx context.Context, lat, lon float64) (*models.WeatherPoint, error) {
	points, err := s.GetRegionWeather(ctx)
	if err != nil {
		return nil, err
	}

	var best models.WeatherPoint
	bestDist := 1e18
	for _, point := range points {
		dist := (point.Latitude-lat)*(point.Latitude-lat) + (point.Longitude-lon)*(point.Longitude-lon)
		if dist < bestDist {
			best = point
			bestDist = dist
		}
	}
	return &best, nil
}

func (s *WeatherService) makePoint(name string, lat, lon float64, updatedAt time.Time, temp int, condition string, wind int) models.WeatherPoint {
	point := models.WeatherPoint{
		Point:      name,
		Latitude:   lat,
		Longitude:  lon,
		TempC:      temp,
		Condition:  condition,
		WindSpeed:  wind,
		UpdatedAt:  updatedAt,
	}
	if condition == "storm" || wind >= 15 {
		point.Severity = "warning"
		point.NeedsRebuild = true
	} else {
		point.Severity = "normal"
	}
	return point
}
