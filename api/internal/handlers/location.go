// Файл location.go реализует HTTP-обработчики CRUD-операций для локаций.
// Предоставляет эндпоинты создания, получения, обновления, удаления и поиска
// локаций с пространственными фильтрами PostGIS.
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
	"kudytudy-api/internal/middleware"
	"kudytudy-api/internal/models"
	"kudytudy-api/internal/services"
)

// LocationHandler — обработчик HTTP-запросов для управления локациями.
type LocationHandler struct {
	service *services.LocationService
	logger  *zap.Logger
}

// NewLocationHandler создаёт обработчик локаций.
func NewLocationHandler(service *services.LocationService, logger *zap.Logger) *LocationHandler {
	return &LocationHandler{
		service: service,
		logger:  logger,
	}
}

// Create обрабатывает POST /api/v1/locations.
// Создаёт новую локацию. Доступно только хостам и администраторам.
func (h *LocationHandler) Create(c fiber.Ctx) error {
	userID, ok := c.Locals(middleware.ContextKeyUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "пользователь не авторизован",
		})
	}

	var req models.CreateLocationRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат запроса",
		})
	}

	location, err := h.service.Create(c.Context(), userID, &req)
	if err != nil {
		if err == database.ErrLocationSlugExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "локация с таким slug уже существует",
			})
		}
		h.logger.Error("ошибка создания локации", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "локация успешно создана",
		"data":    location,
	})
}

// GetByID обрабатывает GET /api/v1/locations/:id.
// Возвращает локацию по UUID.
func (h *LocationHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "id обязателен",
		})
	}

	location, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if err == database.ErrLocationNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "локация не найдена",
			})
		}
		h.logger.Error("ошибка получения локации", zap.String("id", id), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "внутренняя ошибка сервера",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    location,
	})
}

// GetSplat обрабатывает GET /api/v1/locations/:id/splat.
// Возвращает URL 3D-сцены (.splat Gaussian Splatting) для конкретной локации.
// Отдельный эндпоинт позволяет фронтенду загружать тяжёлый 3D-контент лениво,
// не включая его в основной payload location detail.
func (h *LocationHandler) GetSplat(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "id обязателен",
		})
	}

	location, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if err == database.ErrLocationNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "локация не найдена",
			})
		}
		h.logger.Error("ошибка получения splat-данных", zap.String("id", id), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "внутренняя ошибка сервера",
		})
	}

	// Формирование ответа с данными 3D-сцены.
	splatData := fiber.Map{
		"location_id":   location.ID,
		"location_name": location.Name,
		"has_splat":     location.SplatURL != nil && *location.SplatURL != "",
	}

	if location.SplatURL != nil && *location.SplatURL != "" {
		splatData["splat_url"] = *location.SplatURL
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    splatData,
	})
}

// Обновляет локацию. Доступно только владельцу (owner_id).
func (h *LocationHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	userID, ok := c.Locals(middleware.ContextKeyUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "пользователь не авторизован",
		})
	}

	var req models.UpdateLocationRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат запроса",
		})
	}

	location, err := h.service.Update(c.Context(), id, userID, &req)
	if err != nil {
		if err == database.ErrLocationNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "локация не найдена",
			})
		}
		if err == database.ErrLocationForbidden {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "нет прав на изменение этой локации",
			})
		}
		h.logger.Error("ошибка обновления локации", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "внутренняя ошибка сервера",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "локация обновлена",
		"data":    location,
	})
}

// Delete обрабатывает DELETE /api/v1/locations/:id.
// Удаляет локацию. Доступно только владельцу (owner_id).
func (h *LocationHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	userID, ok := c.Locals(middleware.ContextKeyUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "пользователь не авторизован",
		})
	}

	err := h.service.Delete(c.Context(), id, userID)
	if err != nil {
		if err == database.ErrLocationNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "локация не найдена",
			})
		}
		if err == database.ErrLocationForbidden {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "нет прав на удаление этой локации",
			})
		}
		h.logger.Error("ошибка удаления локации", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "внутренняя ошибка сервера",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "локация удалена",
	})
}

// List обрабатывает GET /api/v1/locations.
// Выполняет пространственный поиск с фильтрами и пагинацией.
// Параметры: lat, lon, radius_km (радиус) ИЛИ min_lat, max_lat, min_lon, max_lon (bbox).
// Дополнительно: category, child_friendly, density_level, page, per_page.
func (h *LocationHandler) List(c fiber.Ctx) error {
	filter := &models.LocationFilter{}

	// Параметры радиуса.
	if lat := c.Query("lat"); lat != "" {
		v, _ := strconv.ParseFloat(lat, 64)
		filter.Lat = &v
	}
	if lon := c.Query("lon"); lon != "" {
		v, _ := strconv.ParseFloat(lon, 64)
		filter.Lon = &v
	}
	if radius := c.Query("radius_km"); radius != "" {
		v, _ := strconv.ParseFloat(radius, 64)
		filter.RadiusKm = &v
	}

	// Параметры bbox.
	if v := c.Query("min_lat"); v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		filter.MinLat = &f
	}
	if v := c.Query("max_lat"); v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		filter.MaxLat = &f
	}
	if v := c.Query("min_lon"); v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		filter.MinLon = &f
	}
	if v := c.Query("max_lon"); v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		filter.MaxLon = &f
	}

	// Фильтры по атрибутам.
	if cat := c.Query("category"); cat != "" {
		filter.Category = &cat
	}
	if cf := c.Query("child_friendly"); cf != "" {
		v := cf == "true"
		filter.ChildFriendly = &v
	}
	if dl := c.Query("density_level"); dl != "" {
		filter.DensityLevel = &dl
	}

	// Пагинация.
	if p := c.Query("page"); p != "" {
		filter.Page, _ = strconv.Atoi(p)
	}
	if pp := c.Query("per_page"); pp != "" {
		filter.PerPage, _ = strconv.Atoi(pp)
	}

	result, err := h.service.Search(c.Context(), filter)
	if err != nil {
		h.logger.Error("ошибка поиска локаций", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка поиска локаций",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}
