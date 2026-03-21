// Файл optional_jwt.go реализует необязательный JWT middleware для Go Fiber.
// В отличие от обязательного JWTAuth, этот middleware не блокирует запрос
// при отсутствии или невалидности токена. Если валидный JWT присутствует —
// user_id и role инжектятся в контекст. Если нет — запрос проходит без идентификации.
// Используется для эндпоинтов с auth-flex семантикой (например, POST /trips/:id/join).
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"kudytudy-api/internal/services"
)

// NewOptionalJWTAuth создаёт middleware для необязательной JWT-аутентификации.
// Поведение:
//   - Если заголовок Authorization отсутствует — пропускает запрос без ошибки.
//   - Если заголовок присутствует и токен валиден — инжектит user_id и role в Locals.
//   - Если заголовок присутствует, но токен невалиден — пропускает запрос без ошибки
//     (логирует предупреждение на уровне Debug).
//
// Это позволяет обработчику определить, авторизован ли пользователь,
// проверяя c.Locals("user_id") на nil.
func NewOptionalJWTAuth(jwtService *services.JWTService, logger *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Извлечение заголовка Authorization.
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// Токен не передан — пропускаем без идентификации.
			return c.Next()
		}

		// Проверка формата Bearer <token>.
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			// Некорректный формат — пропускаем без ошибки.
			logger.Debug("optional JWT: некорректный формат Authorization",
				zap.String("path", c.Path()),
			)
			return c.Next()
		}

		tokenString := parts[1]

		// Валидация access-токена.
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			// Токен невалиден — пропускаем без ошибки (auth-flex).
			logger.Debug("optional JWT: невалидный токен, продолжаем без идентификации",
				zap.String("path", c.Path()),
				zap.Error(err),
			)
			return c.Next()
		}

		// Токен валиден — инжектим данные пользователя в контекст.
		c.Locals(ContextKeyUserID, claims.UserID)
		c.Locals(ContextKeyRole, claims.Role)

		logger.Debug("optional JWT: пользователь идентифицирован",
			zap.String("path", c.Path()),
			zap.String("user_id", claims.UserID),
			zap.String("role", claims.Role),
		)

		return c.Next()
	}
}
