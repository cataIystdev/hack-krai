// Файл map.go реализует HTTP-обработчик для эндпоинта GET /api/v1/map/locations.
// Возвращает массив точек для отображения на карте с поддержкой
// bbox-фильтрации, demo-режима и обогащения рекомендациями.
// Эндпоинт публичный, но опционально принимает JWT для персонализации.
package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
	"kudytudy-api/internal/services"
)

// MapHandler — обработчик запросов карты локаций.
type MapHandler struct {
	mapService *services.MapService
	logger     *zap.Logger
}

// NewMapHandler создаёт обработчик карты.
func NewMapHandler(mapService *services.MapService, logger *zap.Logger) *MapHandler {
	return &MapHandler{
		mapService: mapService,
		logger:     logger.Named("map_handler"),
	}
}

// GetMapLocations обрабатывает GET /api/v1/map/locations.
// Возвращает массив точек для отображения маркеров на карте.
//
// Режимы работы:
//   - По умолчанию: PostGIS поиск с bbox/category/density фильтрами.
//   - Demo (?demo=true): curated набор с предустановленными рекомендациями.
//   - Hybrid: если передан JWT, обогащает точки скорами из Qdrant.
//
// Параметры запроса (query):
//   - min_lat, max_lat, min_lon, max_lon — bounding box (все 4 для bbox-поиска)
//   - category — фильтр по категории (winery, farm, trail и др.)
//   - density_level — фильтр по плотности (red, yellow, green)
//   - is_recommended — только рекомендованные точки
//   - limit — максимальное количество точек (по умолчанию 50, максимум 200)
//   - demo — включить demo-режим (true/false)
//   - profile — demo-профиль (calm_wine_mountains, active_adventure, family_kids, gastro_cultural)
func (h *MapHandler) GetMapLocations(c fiber.Ctx) error {
	var filter models.MapLocationFilter
	if err := c.Bind().Query(&filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректные параметры запроса",
		})
	}

	// Опциональное извлечение user_id из JWT (эндпоинт публичный, JWT не обязателен).
	var userID *uuid.UUID
	if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
		if parsed, err := uuid.Parse(uid); err == nil {
			userID = &parsed
		}
	}

	result, err := h.mapService.GetMapLocations(c.Context(), &filter, userID)
	if err != nil {
		// Проверка ошибки валидации.
		if valErr, ok := err.(*services.ValidationError); ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": valErr.Message,
			})
		}

		h.logger.Error("ошибка получения точек карты", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка получения точек карты",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}
