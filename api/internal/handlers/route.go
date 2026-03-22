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
	routeService        *services.RouteService
	storytellingService *services.StorytellingService
	weatherService      *services.WeatherService
	logger              *zap.Logger
}

// NewRouteHandler создаёт обработчик маршрутов.
func NewRouteHandler(
	routeService *services.RouteService,
	storytellingService *services.StorytellingService,
	weatherService *services.WeatherService,
	logger *zap.Logger,
) *RouteHandler {
	return &RouteHandler{
		routeService:        routeService,
		storytellingService: storytellingService,
		weatherService:      weatherService,
		logger:              logger.Named("route_handler"),
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

// GenerateStories обрабатывает POST /api/v1/route/:id/generate-stories.
func (h *RouteHandler) GenerateStories(c fiber.Ctx) error {
	if h.storytellingService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"success": false, "message": "storytelling сервис недоступен"})
	}

	routeID, userID, err := h.extractRouteAndUser(c)
	if err != nil {
		return err
	}

	result, svcErr := h.storytellingService.GenerateStories(c.Context(), routeID, userID)
	if svcErr != nil {
		return h.handlePhase12Error(c, svcErr)
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// GetStories обрабатывает GET /api/v1/route/:id/stories.
func (h *RouteHandler) GetStories(c fiber.Ctx) error {
	if h.storytellingService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"success": false, "message": "storytelling сервис недоступен"})
	}

	routeID, userID, errResp := h.extractRouteAndUser(c)
	if errResp != nil {
		return errResp
	}

	stories, svcErr := h.storytellingService.GetStories(c.Context(), routeID, userID)
	if svcErr != nil {
		return h.handlePhase12Error(c, svcErr)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stories,
		"count":   len(stories),
	})
}

// Rebuild обрабатывает POST /api/v1/route/:id/rebuild.
func (h *RouteHandler) Rebuild(c fiber.Ctx) error {
	routeID, userID, errResp := h.extractRouteAndUser(c)
	if errResp != nil {
		return errResp
	}

	var req models.RouteRebuildRequest
	if err := c.Bind().JSON(&req); err != nil {
		req = models.RouteRebuildRequest{}
	}

	result, svcErr := h.routeService.RebuildRoute(c.Context(), routeID, userID, h.weatherService, req.WeatherOverride)
	if svcErr != nil {
		return h.handlePhase12Error(c, svcErr)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

func (h *RouteHandler) extractRouteAndUser(c fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	routeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "некорректный UUID маршрута"})
	}

	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return uuid.Nil, uuid.Nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "необходима авторизация"})
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "некорректный user_id"})
	}

	return routeID, userID, nil
}

func (h *RouteHandler) handlePhase12Error(c fiber.Ctx, err error) error {
	switch err {
	case services.ErrRouteNotFound:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "маршрут не найден"})
	case services.ErrTripForbidden:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "нет доступа к маршруту"})
	default:
		h.logger.Error("ошибка phase12 route flow", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "внутренняя ошибка сервера"})
	}
}
