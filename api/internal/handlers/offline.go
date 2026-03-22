package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
	
	"kudytudy-api/internal/services"
)

type OfflineHandler struct {
	offlineService *services.OfflineService
	logger         *zap.Logger
}

func NewOfflineHandler(offlineService *services.OfflineService, logger *zap.Logger) *OfflineHandler {
	return &OfflineHandler{
		offlineService: offlineService,
		logger:         logger.Named("offline_handler"),
	}
}

// DownloadOfflineBundle streams an in-memory ZIP containing route data for offline PWA usage.
func (h *OfflineHandler) DownloadOfflineBundle(c fiber.Ctx) error {
	routeIDStr := c.Params("id")
	routeID, err := uuid.Parse(routeIDStr)
	if err != nil {
		h.logger.Warn("неверный формат id маршрута", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid route id"})
	}

	userIDRaw := c.Locals("user_id")
	var userID string
	if userIDRaw != nil {
		userID = userIDRaw.(string)
	}

	zipBytes, err := h.offlineService.BuildZipBundle(c.Context(), userID, routeID)
	if err != nil {
		h.logger.Error("ошибка при генерации офлайн-бандла", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate offline bundle"})
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", `attachment; filename="route_bundle.zip"`)
	return c.Send(zipBytes)
}
