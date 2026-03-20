// Файл route.go реализует HTTP-обработчик для эндпоинта POST /api/v1/route/build.
// Принимает список ID локаций и формирует превью маршрута
// с расчётом расстояний и времени.
package handlers

import (
	"github.com/gofiber/fiber/v3"
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
