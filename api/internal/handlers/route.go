// Файл route.go реализует HTTP-обработчики для эндпоинтов маршрутов.
// Поддерживает два режима:
//   - POST /api/v1/route/build — ручное построение по списку ID локаций.
//   - POST /api/v1/trips/{id}/build-route — trip-aware построение
//     с учётом параметров поездки и vibe-профилей участников.
package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
	"kudytudy-api/internal/services"
)

// RouteHandler — обработчик запросов построения маршрутов.
type RouteHandler struct {
	routeService *services.RouteService
	logger       *zap.Logger
}

// NewRouteHandler создаёт обработчик маршрутов.
func NewRouteHandler(routeService *services.RouteService, logger *zap.Logger) *RouteHandler {
	return &RouteHandler{
		routeService: routeService,
		logger:       logger.Named("route_handler"),
	}
}

// BuildRoute обрабатывает POST /api/v1/route/build.
// Принимает JSON с массивом location_ids и возвращает превью маршрута
// с расчётом расстояний между точками по Haversine.
//
// Тело запроса:
//
//	{
//	  "location_ids": ["uuid1", "uuid2", "uuid3"],
//	  "transport": "car",
//	  "optimize": false
//	}
func (h *RouteHandler) BuildRoute(c fiber.Ctx) error {
	var req models.BuildRouteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный JSON",
		})
	}

	// Валидация запроса.
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	req.NormalizeDefaults()

	// Построение маршрута.
	result, err := h.routeService.BuildRoute(c.Context(), &req)
	if err != nil {
		// Проверка ошибки валидации.
		if valErr, ok := err.(*services.ValidationError); ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": valErr.Message,
			})
		}

		h.logger.Error("ошибка построения маршрута", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка построения маршрута",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// BuildTripRoute обрабатывает POST /api/v1/trips/{id}/build-route.
// Строит маршрут с учётом параметров поездки: бюджет, даты, транспорт,
// состав группы, формат, vibe-профили участников.
//
// Тело запроса (все поля опциональны):
//
//	{
//	  "location_ids": ["uuid1", "uuid2"],
//	  "max_points": 8,
//	  "optimize": false
//	}
func (h *RouteHandler) BuildTripRoute(c fiber.Ctx) error {
	// Извлечение ID поездки из URL.
	tripIDStr := c.Params("id")
	if tripIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID поездки обязателен",
		})
	}

	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный UUID поездки",
		})
	}

	// Извлечение user_id из JWT.
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "необходима авторизация",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "некорректный user_id",
		})
	}

	// Парсинг тела запроса.
	var req models.BuildTripRouteRequest
	if err := c.Bind().JSON(&req); err != nil {
		// Пустое тело допустимо — автоподбор локаций.
		req = models.BuildTripRouteRequest{}
	}

	// Валидация.
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// Построение маршрута.
	result, err := h.routeService.BuildTripRoute(c.Context(), tripID, userID, &req)
	if err != nil {
		if valErr, ok := err.(*services.ValidationError); ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": valErr.Message,
			})
		}

		h.logger.Error("ошибка построения trip-aware маршрута",
			zap.String("trip_id", tripIDStr),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка построения маршрута",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}
