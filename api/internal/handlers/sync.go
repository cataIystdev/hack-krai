package handlers

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
	"kudytudy-api/internal/services"
)

type SyncHandler struct {
	syncService *services.SyncService
	logger      *zap.Logger
}

func NewSyncHandler(syncService *services.SyncService, logger *zap.Logger) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
		logger:      logger.Named("sync_handler"),
	}
}

// SyncOfflineData receives batched actions and executes them asynchronously
func (h *SyncHandler) SyncOfflineData(c fiber.Ctx) error {
	var payload models.SyncPayload
	if err := c.Bind().Body(&payload); err != nil {
		h.logger.Warn("неверный формат данных для синхронизации", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "неверный формат данных массива синхронизации"})
	}

	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "требуется авторизация"})
	}
	userID := userIDRaw.(string)

	resp := h.syncService.ProcessSyncPayload(c.Context(), userID, &payload)

	status := fiber.StatusOK
	if !resp.Success {
		// 207 Multi-Status could be proper here if we have partial success,
		// but using 200 with an error list is simpler to consume for PWA.
		status = fiber.StatusMultiStatus 
	}

	return c.Status(status).JSON(resp)
}
