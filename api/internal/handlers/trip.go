// Файл trip.go реализует HTTP-обработчики модуля планирования поездок.
// Обеспечивает создание поездок, получение деталей, генерацию invite-ссылок,
// присоединение участников и получение списка участников.
package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
	"kudytudy-api/internal/services"
)

// TripHandler — обработчик запросов модуля поездок.
type TripHandler struct {
	tripService *services.TripService
	logger      *zap.Logger
}

// NewTripHandler создаёт обработчик поездок.
func NewTripHandler(tripService *services.TripService, logger *zap.Logger) *TripHandler {
	return &TripHandler{
		tripService: tripService,
		logger:      logger.Named("trip_handler"),
	}
}

// Create обрабатывает POST /api/v1/trips.
// Создаёт новую поездку и автоматически добавляет создателя как первого участника.
// Требует JWT-аутентификации.
func (h *TripHandler) Create(c fiber.Ctx) error {
	// Получение ID пользователя из JWT-контекста.
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "не удалось определить пользователя",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "некорректный идентификатор пользователя",
		})
	}

	// Парсинг тела запроса.
	var req models.CreateTripRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Debug("ошибка парсинга тела запроса создания поездки", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат запроса",
		})
	}

	// Валидация.
	if msg := req.Validate(); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": msg,
		})
	}

	// Создание поездки через сервис.
	result, err := h.tripService.Create(c.Context(), userID, &req)
	if err != nil {
		h.logger.Error("ошибка создания поездки", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка создания поездки",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "поездка успешно создана",
		"data":    result,
	})
}

// ListTrips обрабатывает GET /api/v1/trips.
// Возвращает все поездки текущего пользователя с участниками.
// Требует JWT-аутентификации.
func (h *TripHandler) ListTrips(c fiber.Ctx) error {
	// Получение ID пользователя из JWT-контекста.
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "не удалось определить пользователя",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "некорректный идентификатор пользователя",
		})
	}

	trips, err := h.tripService.GetUserTrips(c.Context(), userID)
	if err != nil {
		h.logger.Error("ошибка получения поездок пользователя",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка получения поездок",
		})
	}

	return c.JSON(fiber.Map{
		"data":  trips,
		"count": len(trips),
	})
}

// GetByID обрабатывает GET /api/v1/trips/:id.
// Возвращает детали поездки со списком участников.
// Требует JWT-аутентификации.
func (h *TripHandler) GetByID(c fiber.Ctx) error {
	// Парсинг ID поездки из URL.
	tripIDStr := c.Params("id")
	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат ID поездки",
		})
	}

	// Получение поездки через сервис.
	result, err := h.tripService.GetByID(c.Context(), tripID)
	if err != nil {
		return h.handleTripError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// GenerateInvite обрабатывает POST /api/v1/trips/:id/invite.
// Генерирует новый invite-токен для поездки.
// Доступно только создателю поездки. Требует JWT-аутентификации.
func (h *TripHandler) GenerateInvite(c fiber.Ctx) error {
	// Получение ID пользователя из JWT-контекста.
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "не удалось определить пользователя",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "некорректный идентификатор пользователя",
		})
	}

	// Парсинг ID поездки.
	tripIDStr := c.Params("id")
	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат ID поездки",
		})
	}

	// Генерация нового токена.
	newToken, err := h.tripService.GenerateInviteToken(c.Context(), tripID, userID)
	if err != nil {
		return h.handleTripError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "invite-токен обновлён",
		"data": fiber.Map{
			"invite_token": newToken.String(),
			"invite_url":   "https://deepkrai.ru/trip/" + tripID.String() + "/join?token=" + newToken.String(),
		},
	})
}

// Join обрабатывает POST /api/v1/trips/:id/join.
// Присоединяет участника к поездке по invite-токену.
// Эндпоинт доступен БЕЗ авторизации — неавторизованные пользователи
// передают display_name и tags напрямую.
func (h *TripHandler) Join(c fiber.Ctx) error {
	// Парсинг ID поездки.
	tripIDStr := c.Params("id")
	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат ID поездки",
		})
	}

	// Парсинг тела запроса.
	var req models.JoinTripRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Debug("ошибка парсинга тела запроса присоединения", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат запроса",
		})
	}

	// Валидация.
	if msg := req.Validate(); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": msg,
		})
	}

	// Проверка наличия JWT (необязательно).
	var userID *uuid.UUID
	if userIDStr, ok := c.Locals("user_id").(string); ok && userIDStr != "" {
		if parsed, err := uuid.Parse(userIDStr); err == nil {
			userID = &parsed
		}
	}

	// Присоединение через сервис.
	member, err := h.tripService.Join(c.Context(), tripID, &req, userID)
	if err != nil {
		return h.handleTripError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "участник присоединился к поездке",
		"data":    member,
	})
}

// ListMembers обрабатывает GET /api/v1/trips/:id/members.
// Возвращает список участников поездки.
// Требует JWT-аутентификации.
func (h *TripHandler) ListMembers(c fiber.Ctx) error {
	// Парсинг ID поездки.
	tripIDStr := c.Params("id")
	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат ID поездки",
		})
	}

	// Получение участников через сервис.
	members, err := h.tripService.GetMembers(c.Context(), tripID)
	if err != nil {
		return h.handleTripError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"members": members,
			"count":   len(members),
		},
	})
}

// handleTripError обрабатывает ошибки сервиса поездок
// и формирует соответствующие HTTP-ответы.
func (h *TripHandler) handleTripError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, services.ErrTripNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "поездка не найдена",
		})
	case errors.Is(err, services.ErrTripForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "нет прав на эту операцию",
		})
	case errors.Is(err, services.ErrInvalidInviteToken):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "неверный invite-токен",
		})
	case errors.Is(err, services.ErrTripFull):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "превышено максимальное количество участников",
		})
	case errors.Is(err, services.ErrAlreadyMember):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "пользователь уже является участником поездки",
		})
	default:
		h.logger.Error("внутренняя ошибка модуля поездок", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "внутренняя ошибка сервера",
		})
	}
}
