package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
	"kudytudy-api/internal/services"
)

type BookingHandler struct {
	bookingService *services.BookingService
	logger         *zap.Logger
}

func NewBookingHandler(bookingService *services.BookingService, logger *zap.Logger) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		logger:         logger.Named("booking_handler"),
	}
}

func (h *BookingHandler) ListLocationSlots(c fiber.Ctx) error {
	locationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "некорректный формат ID локации"})
	}

	var filter models.LocationSlotsFilter
	if err := c.Bind().Query(&filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "некорректные query-параметры"})
	}
	filter.NormalizeDefaultMonth(time.Now())
	if msg := filter.Validate(); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": msg})
	}

	slots, err := h.bookingService.ListLocationSlots(c.Context(), locationID, filter.ParseMonth())
	if err != nil {
		return h.handleBookingError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"location_id": locationID,
			"month":       filter.Month,
			"slots":       slots,
			"count":       len(slots),
		},
	})
}

func (h *BookingHandler) Create(c fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	var req models.CreateBookingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "некорректный формат запроса"})
	}
	if msg := req.Validate(); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": msg})
	}

	booking, err := h.bookingService.Create(c.Context(), userID, &req)
	if err != nil {
		return h.handleBookingError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "бронь успешно создана",
		"data":    booking,
	})
}

func (h *BookingHandler) ConfirmOrReject(c fiber.Ctx) error {
	hostID, err := extractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	bookingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "некорректный формат ID брони"})
	}

	var req models.BookingActionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "некорректный формат запроса"})
	}
	if msg := req.Validate(); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": msg})
	}

	booking, err := h.bookingService.ConfirmOrReject(c.Context(), bookingID, hostID, &req)
	if err != nil {
		return h.handleBookingError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "статус брони обновлён",
		"data":    booking,
	})
}

func (h *BookingHandler) Cancel(c fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	bookingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "некорректный формат ID брони"})
	}

	booking, err := h.bookingService.Cancel(c.Context(), bookingID, userID)
	if err != nil {
		return h.handleBookingError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "бронь отменена",
		"data":    booking,
	})
}

func (h *BookingHandler) ListMyBookings(c fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	bookings, err := h.bookingService.ListMyBookings(c.Context(), userID)
	if err != nil {
		return h.handleBookingError(c, err)
	}
	return c.JSON(fiber.Map{"success": true, "data": bookings, "count": len(bookings)})
}

func (h *BookingHandler) ListHostBookings(c fiber.Ctx) error {
	hostID, err := extractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	bookings, err := h.bookingService.ListHostBookings(c.Context(), hostID)
	if err != nil {
		return h.handleBookingError(c, err)
	}
	return c.JSON(fiber.Map{"success": true, "data": bookings, "count": len(bookings)})
}

func (h *BookingHandler) handleBookingError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, services.ErrBookingNotFound), errors.Is(err, services.ErrLocationNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "ресурс не найден"})
	case errors.Is(err, services.ErrBookingForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "нет прав на эту операцию"})
	case errors.Is(err, services.ErrBookingUnavailable):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "выбранные даты недоступны для бронирования"})
	case errors.Is(err, services.ErrBookingInvalidStatus):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "операция недоступна для текущего статуса брони"})
	default:
		h.logger.Error("внутренняя ошибка booking", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "внутренняя ошибка сервера"})
	}
}

func extractUserID(c fiber.Ctx) (uuid.UUID, error) {
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return uuid.Nil, errors.New("не удалось определить пользователя")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("некорректный идентификатор пользователя")
	}
	return userID, nil
}
