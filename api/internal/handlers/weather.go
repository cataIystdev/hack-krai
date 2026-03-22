package handlers

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"kudytudy-api/internal/services"
)

type WeatherHandler struct {
	weatherService *services.WeatherService
	logger         *zap.Logger
}

func NewWeatherHandler(weatherService *services.WeatherService, logger *zap.Logger) *WeatherHandler {
	return &WeatherHandler{
		weatherService: weatherService,
		logger:         logger.Named("weather_handler"),
	}
}

func (h *WeatherHandler) GetRegion(c fiber.Ctx) error {
	points, err := h.weatherService.GetRegionWeather(c.Context())
	if err != nil {
		h.logger.Error("ошибка получения погоды региона", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка получения погоды",
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    points,
		"count":   len(points),
	})
}
