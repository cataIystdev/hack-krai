// Файл profile.go реализует HTTP-обработчик профиля пользователя.
// Предоставляет эндпоинт GET /api/v1/profile/me для получения данных
// текущего авторизованного пользователя (без хэша пароля).
package handlers

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"deep-krai-api/internal/database"
	"deep-krai-api/internal/middleware"
)

// ProfileHandler — обработчик запросов профиля пользователя.
type ProfileHandler struct {
	userRepo *database.UserRepository
	logger   *zap.Logger
}

// NewProfileHandler создаёт обработчик профиля.
func NewProfileHandler(userRepo *database.UserRepository, logger *zap.Logger) *ProfileHandler {
	return &ProfileHandler{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetMe обрабатывает GET /api/v1/profile/me.
// Извлекает user_id из контекста (установленный JWT middleware),
// загружает данные пользователя из БД и возвращает публичное представление.
func (h *ProfileHandler) GetMe(c fiber.Ctx) error {
	// Извлечение user_id из контекста (установлен JWT middleware).
	userID, ok := c.Locals(middleware.ContextKeyUserID).(string)
	if !ok || userID == "" {
		h.logger.Error("user_id не найден в контексте запроса")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "пользователь не авторизован",
		})
	}

	// Загрузка данных пользователя из БД.
	user, err := h.userRepo.FindByID(c.Context(), userID)
	if err != nil {
		h.logger.Error("ошибка получения профиля",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "пользователь не найден",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    user.ToPublicResponse(),
	})
}

// UpdateMe обрабатывает PUT /api/v1/profile/me.
// Обновляет профиль текущего авторизованного пользователя.
func (h *ProfileHandler) UpdateMe(c fiber.Ctx) error {
	userID, ok := c.Locals(middleware.ContextKeyUserID).(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "пользователь не авторизован",
		})
	}

	var req struct {
		DisplayName *string `json:"display_name"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат запроса",
		})
	}

	if req.DisplayName == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "необходимо указать поле для обновления",
		})
	}

	user, err := h.userRepo.UpdateProfile(c.Context(), userID, *req.DisplayName)
	if err != nil {
		h.logger.Error("ошибка обновления профиля",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ошибка обновления профиля",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "профиль обновлён",
		"data":    user.ToPublicResponse(),
	})
}
