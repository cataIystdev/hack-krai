package handlers

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
)

type AnalyticsHandler struct {
	repo   *database.AnalyticsRepository
	logger *zap.Logger
}

func NewAnalyticsHandler(repo *database.AnalyticsRepository, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		repo:   repo,
		logger: logger.Named("analytics_handler"),
	}
}

// GetHeatmap returns point clusters showing location popularity.
func (h *AnalyticsHandler) GetHeatmap(c fiber.Ctx) error {
	heatmapData, err := h.repo.GetHeatmap(c.Context())
	if err != nil {
		h.logger.Error("failed to get heatmap data", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, 
			"message": "не удалось получить данные тепловой карты",
		})
	}

	if heatmapData == nil {
		heatmapData = make([]database.HeatmapPoint, 0)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    heatmapData,
	})
}

// GetPredictions returns predictive analytics for upcoming user loads.
func (h *AnalyticsHandler) GetPredictions(c fiber.Ctx) error {
	predictions, err := h.repo.GetPredictions(c.Context())
	if err != nil {
		h.logger.Error("failed to get prediction data", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, 
			"message": "не удалось получить прогноз загруженности",
		})
	}

	if predictions == nil {
		predictions = make([]database.PredictionPoint, 0)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    predictions,
	})
}
