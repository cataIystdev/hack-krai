// Файл jwt.go реализует JWT middleware для Go Fiber.
// Извлекает JWT-токен из заголовка Authorization (Bearer scheme),
// валидирует подпись и claims, и инжектит user_id и role в контекст запроса
// через fiber.Ctx.Locals(). При невалидном токене возвращает 401 Unauthorized.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"kudytudy-api/internal/services"
)

// ContextKeyUserID — ключ для хранения user_id в контексте запроса.
const ContextKeyUserID = "user_id"

// ContextKeyRole — ключ для хранения роли пользователя в контексте запроса.
const ContextKeyRole = "role"

// NewJWTAuth создаёт middleware для JWT-аутентификации.
// Извлекает токен из заголовка Authorization (формат: "Bearer <token>"),
// валидирует его через JWTService и инжектит user_id и role в контекст.
// При отсутствии или невалидности токена возвращает HTTP 401.
func NewJWTAuth(jwtService *services.JWTService, logger *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Извлечение заголовка Authorization.
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Debug("отсутствует заголовок Authorization",
				zap.String("path", c.Path()),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "отсутствует заголовок Authorization",
			})
		}

		// Проверка формата Bearer <token>.
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			logger.Debug("некорректный формат заголовка Authorization",
				zap.String("header", authHeader),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "некорректный формат заголовка Authorization, ожидается: Bearer <token>",
			})
		}

		tokenString := parts[1]

		// Валидация access-токена.
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			logger.Debug("невалидный access-токен",
				zap.String("path", c.Path()),
				zap.Error(err),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "невалидный или истекший токен авторизации",
			})
		}

		// Инъекция данных пользователя в контекст запроса.
		// Доступны через c.Locals(middleware.ContextKeyUserID) и c.Locals(middleware.ContextKeyRole).
		c.Locals(ContextKeyUserID, claims.UserID)
		c.Locals(ContextKeyRole, claims.Role)

		return c.Next()
	}
}
