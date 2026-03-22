package handlers

import (
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"kudytudy-api/internal/services"
)

// WSHandler handles WebSocket upgrade endpoints.
type WSHandler struct {
	hub    *services.WSHub
	logger *zap.Logger
}

// NewWSHandler creates a new WebSocket handler.
func NewWSHandler(hub *services.WSHub, logger *zap.Logger) *WSHandler {
	return &WSHandler{
		hub:    hub,
		logger: logger.Named("ws_handler"),
	}
}

// HandleWeather upgrades to WebSocket and streams weather updates.
// GET /ws/v1/weather
func (h *WSHandler) HandleWeather(c fiber.Ctx) error {
	return h.doUpgrade(c, services.ChannelWeather, "")
}

// HandleNotifications upgrades to WebSocket for user-specific notifications.
// GET /ws/v1/notifications (requires auth)
func (h *WSHandler) HandleNotifications(c fiber.Ctx) error {
	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "требуется авторизация"})
	}
	userID := userIDRaw.(string)
	channel := services.ChannelNotifications + ":" + userID
	return h.doUpgrade(c, channel, userID)
}

// doUpgrade performs the WebSocket handshake via fasthttp upgrader.
func (h *WSHandler) doUpgrade(c fiber.Ctx, channel, userID string) error {
	upgrader := websocket.FastHTTPUpgrader{
		CheckOrigin: func(ctx *fasthttp.RequestCtx) bool { return true },
	}

	fastHTTPCtx, ok := c.Context().(*fasthttp.RequestCtx)
	if !ok {
		h.logger.Error("failed to assert context to *fasthttp.RequestCtx")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal context error"})
	}

	err := upgrader.Upgrade(fastHTTPCtx, func(conn *websocket.Conn) {
		h.hub.RegisterFastHTTP(channel, conn)
		defer h.hub.UnregisterFastHTTP(channel, conn)

		h.logger.Info("ws client connected",
			zap.String("channel", channel),
			zap.String("user_id", userID),
		)

		// Keep-alive read loop: detect disconnect, handle pongs.
		_ = conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(120 * time.Second))
			return nil
		})

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}

		h.logger.Info("ws client disconnected", zap.String("channel", channel))
	})

	if err != nil {
		h.logger.Warn("ws upgrade failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "websocket upgrade failed"})
	}

	return nil
}
