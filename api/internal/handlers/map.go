// Файл map.go реализует HTTP-обработчик для эндпоинта GET /api/v1/map/locations.
// Возвращает массив точек для отображения на карте с поддержкой
// bbox-фильтрации, фильтрации по категории и плотности.
package handlers

import (
	"github.com/gofiber/fiber/v3"
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
// Параметры запроса (query):
//   - min_lat, max_lat, min_lon, max_lon — bounding box (все 4 обязательны для bbox-поиска)
//   - category — фильтр по категории (winery, farm, trail и др.)
//   - density_level — фильтр по плотности (red, yellow, green)
//   - limit — максимальное количество точек (по умолчанию 50, максимум 200)
func (h *MapHandler) GetMapLocations(c fiber.Ctx) error {
	var filter models.MapLocationFilter
	if err := c.Bind().Query(&filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректные параметры запроса",
		})
	}

	result, err := h.mapService.GetMapLocations(c.Context(), &filter)
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
