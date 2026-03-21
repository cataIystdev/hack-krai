// Файл vibe.go реализует HTTP-обработчики для модуля vibe-профилирования.
// Обрабатывает три эндпоинта:
//   - POST /api/v1/profile/voice — голосовое профилирование (multipart аудио)
//   - POST /api/v1/profile/swipe — свайп сцены (JSON)
//   - POST /api/v1/profile/finalize — финализация и получение рекомендаций
package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
	"kudytudy-api/internal/services"
)

// VibeHandler — обработчики для модуля vibe-профилирования.
type VibeHandler struct {
	vibeService *services.VibeService
	logger      *zap.Logger
}

// NewVibeHandler создаёт обработчики профилирования.
func NewVibeHandler(vibeService *services.VibeService, logger *zap.Logger) *VibeHandler {
	return &VibeHandler{
		vibeService: vibeService,
		logger:      logger.Named("vibe_handler"),
	}
}

// VoiceProfile обрабатывает POST /api/v1/profile/voice.
// Принимает аудиофайл через multipart/form-data (поле "audio").
// Выполняет полный пайплайн: STT -> LLM -> Embeddings -> Qdrant.
// Возвращает оси vibe-профиля и ID вектора.
func (h *VibeHandler) VoiceProfile(c fiber.Ctx) error {
	// Извлечение user_id из JWT claims.
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "не удалось извлечь user_id из токена",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный user_id",
		})
	}

	// Получение аудиофайла из multipart формы.
	fileHeader, err := c.FormFile("audio")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "аудиофайл обязателен (поле 'audio')",
		})
	}

	// Открытие файла для чтения.
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка открытия аудиофайла",
		})
	}
	defer file.Close()

	// Выполнение пайплайна профилирования.
	result, err := h.vibeService.ProcessVoice(
		c.Context(),
		userID,
		file,
		fileHeader.Filename,
		fileHeader.Size,
	)
	if err != nil {
		h.logger.Error("ошибка голосового профилирования",
			zap.String("user_id", userIDStr),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка профилирования: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// Swipe обрабатывает POST /api/v1/profile/swipe.
// Принимает JSON: {"scene_id": "uuid", "direction": "right|left"}.
// Сдвигает vibe-вектор пользователя в Qdrant.
func (h *VibeHandler) Swipe(c fiber.Ctx) error {
	// Извлечение user_id из JWT.
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "не удалось извлечь user_id из токена",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный user_id",
		})
	}

	// Парсинг тела запроса.
	var req models.SwipeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный JSON",
		})
	}

	// Валидация.
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// Выполнение свайпа.
	if err := h.vibeService.Swipe(c.Context(), userID, req.SceneID, models.SwipeDirection(req.Direction)); err != nil {
		h.logger.Error("ошибка свайпа",
			zap.String("user_id", userIDStr),
			zap.String("scene_id", req.SceneID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка свайпа: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"message":   "вектор обновлён",
		"scene_id":  req.SceneID,
		"direction": req.Direction,
	})
}

// Finalize обрабатывает POST /api/v1/profile/finalize.
// Берёт vibe-вектор пользователя из Qdrant и ищет Top-N ближайших локаций.
// Принимает опциональный JSON body с параметрами limit и child_friendly_only.
func (h *VibeHandler) Finalize(c fiber.Ctx) error {
	// Извлечение user_id из JWT.
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "не удалось извлечь user_id из токена",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный user_id",
		})
	}

	// Парсинг опционального тела запроса.
	var req models.FinalizeRequest
	if len(c.Body()) > 0 {
		if err := c.Bind().JSON(&req); err != nil {
			h.logger.Warn("ошибка парсинга FinalizeRequest, используются значения по умолчанию",
				zap.Error(err),
			)
		}
	}
	req.NormalizeDefaults()

	// Финализация и поиск рекомендаций.
	result, err := h.vibeService.Finalize(c.Context(), userID, &req)
	if err != nil {
		h.logger.Error("ошибка финализации профиля",
			zap.String("user_id", userIDStr),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка финализации: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// GetScenes обрабатывает GET /api/v1/profile/scenes.
// Возвращает все сцены свайпа для анкеты.
func (h *VibeHandler) GetScenes(c fiber.Ctx) error {
	scenes, err := h.vibeService.GetScenes(c.Context())
	if err != nil {
		h.logger.Error("ошибка получения сцен", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка получения сцен: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"scenes": scenes,
			"count":  len(scenes),
		},
	})
}
