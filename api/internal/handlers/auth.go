// Файл auth.go реализует HTTP-обработчики аутентификации и авторизации.
// Предоставляет эндпоинты: регистрация (POST /auth/register),
// авторизация (POST /auth/login), обновление токена (POST /auth/refresh).
package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"deep-krai-api/internal/services"
)

// AuthHandler — обработчик запросов аутентификации.
type AuthHandler struct {
	authService *services.AuthService
	logger      *zap.Logger
}

// NewAuthHandler создаёт обработчик аутентификации.
func NewAuthHandler(authService *services.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// RegisterRequest — тело запроса регистрации.
type RegisterRequest struct {
	// Email — адрес электронной почты нового пользователя.
	Email string `json:"email"`

	// Password — пароль (минимум 8 символов).
	Password string `json:"password"`

	// DisplayName — отображаемое имя (необязательно).
	DisplayName string `json:"display_name"`

	// Role — роль пользователя: "tourist" (по умолчанию) или "host".
	Role string `json:"role"`
}

// LoginRequest — тело запроса авторизации.
type LoginRequest struct {
	// Email — адрес электронной почты.
	Email string `json:"email"`

	// Password — пароль.
	Password string `json:"password"`
}

// RefreshRequest — тело запроса обновления токена.
type RefreshRequest struct {
	// RefreshToken — действующий refresh-токен.
	RefreshToken string `json:"refresh_token"`
}

// Register обрабатывает POST /api/v1/auth/register.
// Регистрирует нового пользователя, хэширует пароль через bcrypt (cost=12),
// создаёт запись в БД и возвращает пару JWT-токенов.
func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Debug("ошибка парсинга тела запроса регистрации", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат запроса",
		})
	}

	// Валидация обязательных полей.
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "email и password обязательны",
		})
	}

	// Роль по умолчанию — tourist. Допустимы только tourist и host.
	role := req.Role
	if role == "" {
		role = "tourist"
	}
	if role != "tourist" && role != "host" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "допустимые роли: tourist, host",
		})
	}

	result, err := h.authService.Register(c.Context(), req.Email, req.Password, req.DisplayName, role)
	if err != nil {
		return h.handleAuthError(c, err)
	}

	h.logger.Info("пользователь зарегистрирован через API",
		zap.String("email", req.Email),
	)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "пользователь успешно зарегистрирован",
		"data":    result,
	})
}

// Login обрабатывает POST /api/v1/auth/login.
// Проверяет учётные данные и возвращает пару JWT-токенов при успехе.
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Debug("ошибка парсинга тела запроса авторизации", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат запроса",
		})
	}

	// Валидация обязательных полей.
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "email и password обязательны",
		})
	}

	result, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return h.handleAuthError(c, err)
	}

	h.logger.Info("пользователь авторизован через API",
		zap.String("email", req.Email),
	)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "авторизация успешна",
		"data":    result,
	})
}

// Refresh обрабатывает POST /api/v1/auth/refresh.
// Принимает refresh-токен и возвращает новую пару токенов (access + refresh).
func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req RefreshRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Debug("ошибка парсинга тела запроса обновления токена", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат запроса",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "refresh_token обязателен",
		})
	}

	tokens, err := h.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		h.logger.Debug("ошибка обновления токена", zap.Error(err))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "невалидный или истекший refresh-токен",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "токены обновлены",
		"data": fiber.Map{
			"tokens": tokens,
		},
	})
}

// handleAuthError обрабатывает ошибки сервиса аутентификации
// и формирует соответствующие HTTP-ответы.
func (h *AuthHandler) handleAuthError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, services.ErrInvalidCredentials):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "неверный email или пароль",
		})
	case errors.Is(err, services.ErrEmailAlreadyRegistered):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "пользователь с таким email уже зарегистрирован",
		})
	case errors.Is(err, services.ErrPasswordTooShort):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "пароль должен содержать не менее 8 символов",
		})
	case errors.Is(err, services.ErrInvalidEmail):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "некорректный формат email",
		})
	default:
		h.logger.Error("внутренняя ошибка аутентификации", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "внутренняя ошибка сервера",
		})
	}
}
