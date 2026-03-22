// Файл onboarding.go содержит HTTP-обработчики для Zero-UI онбординга хостов.
// Реализует эндпоинты для POST /host/onboard, GET /host/tasks/:id
// и GET /host/locations (GDD Feature 5).
package handlers

import (
	"fmt"
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/services"
)

// OnboardingHandler — обработчик запросов онбординга хостов.
type OnboardingHandler struct {
	service *services.OnboardingService
	logger  *zap.Logger
}

// NewOnboardingHandler создаёт экземпляр обработчика онбординга.
func NewOnboardingHandler(service *services.OnboardingService, logger *zap.Logger) *OnboardingHandler {
	return &OnboardingHandler{
		service: service,
		logger:  logger.Named("onboarding_handler"),
	}
}

// Onboard обрабатывает POST /host/onboard — запуск pipeline онбординга.
// Принимает multipart/form-data:
//   - audio (обязательно) — голосовое описание локации
//   - media[] (опционально) — URL медиафайлов (фото/видео)
//   - latitude (опционально) — широта
//   - longitude (опционально) — долгота
//   - address (опционально) — адрес локации
//
// Возвращает OnboardResponse с task_id, статусом и извлечёнными данными.
func (h *OnboardingHandler) Onboard(c fiber.Ctx) error {
	// Извлечение user_id из JWT context.
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "требуется авторизация",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный user_id",
		})
	}

	// Извлечение аудиофайла из multipart.
	audioFile, err := c.FormFile("audio")
	if err != nil || audioFile == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "поле audio обязательно (голосовое описание локации)",
		})
	}

	// Чтение аудиоданных.
	file, err := audioFile.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка чтения аудиофайла",
		})
	}
	defer file.Close()

	audioData, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка чтения аудиоданных",
		})
	}

	// Извлечение опциональных полей.
	var mediaURLs []string
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		if urls, ok := form.Value["media_urls"]; ok {
			mediaURLs = urls
		}
	}

	var latitude, longitude *float64
	if latStr := c.FormValue("latitude"); latStr != "" {
		var lat float64
		if _, err := parseFloat(latStr, &lat); err == nil {
			latitude = &lat
		}
	}
	if lonStr := c.FormValue("longitude"); lonStr != "" {
		var lon float64
		if _, err := parseFloat(lonStr, &lon); err == nil {
			longitude = &lon
		}
	}

	address := c.FormValue("address")

	// Запуск pipeline.
	resp, err := h.service.StartOnboarding(
		c.Context(), userID, audioData, audioFile.Filename,
		mediaURLs, latitude, longitude, address,
	)
	if err != nil {
		h.logger.Error("ошибка онбординга", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка обработки онбординга",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": resp.Message,
		"data":    resp,
	})
}

// GetTaskStatus обрабатывает GET /host/tasks/:id — статус задачи.
func (h *OnboardingHandler) GetTaskStatus(c fiber.Ctx) error {
	// Извлечение user_id из JWT context.
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "требуется авторизация",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный user_id",
		})
	}

	taskIDStr := c.Params("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный task_id",
		})
	}

	resp, err := h.service.GetTaskStatus(c.Context(), taskID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// GetHostLocations обрабатывает GET /host/locations — локации хоста.
func (h *OnboardingHandler) GetHostLocations(c fiber.Ctx) error {
	// Извлечение user_id из JWT context.
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "требуется авторизация",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный user_id",
		})
	}

	locations, err := h.service.GetHostLocations(c.Context(), userID)
	if err != nil {
		h.logger.Error("ошибка получения локаций хоста", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка получения локаций",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    locations,
		"total":   len(locations),
	})
}

// parseFloat парсит строку как float64.
func parseFloat(s string, result *float64) (bool, error) {
	var val float64
	_, err := fmt.Sscanf(s, "%f", &val)
	if err != nil {
		return false, err
	}
	*result = val
	return true, nil
}

// StartSplatting обрабатывает POST /host/splat — запуск генерации 3D-сцены.
// Принимает multipart/form-data:
//   - video (обязательно) — видеофайл для обработки
//   - location_id (обязательно) — UUID локации для привязки 3D-сцены
//
// Загружает видео в MinIO, запускает mock pipeline Gaussian Splatting.
func (h *OnboardingHandler) StartSplatting(c fiber.Ctx) error {
	// Извлечение user_id из JWT context.
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "требуется авторизация",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный user_id",
		})
	}

	// Извлечение location_id из form.
	locationIDStr := c.FormValue("location_id")
	if locationIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "поле location_id обязательно",
		})
	}

	locationID, err := uuid.Parse(locationIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный location_id",
		})
	}

	// Извлечение видеофайла из multipart.
	videoFile, err := c.FormFile("video")
	if err != nil || videoFile == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "поле video обязательно (видеофайл локации)",
		})
	}

	// Чтение видеоданных.
	file, err := videoFile.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка чтения видеофайла",
		})
	}
	defer file.Close()

	videoData, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка чтения видеоданных",
		})
	}

	// Запуск pipeline.
	resp, err := h.service.StartSplatting(c.Context(), userID, locationID, videoData, videoFile.Filename)
	if err != nil {
		h.logger.Error("ошибка запуска splatting", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": resp.Message,
		"data":    resp,
	})
}

