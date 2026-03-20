// Файл rbac.go реализует middleware ограничения доступа по ролям (RBAC).
// Проверяет роль пользователя из контекста запроса (установленную JWT middleware)
// и разрешает доступ только пользователям с указанными ролями.
// При несоответствии роли возвращает HTTP 403 Forbidden.
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// RequireRole создаёт middleware для ограничения доступа по ролям.
// Принимает список допустимых ролей. Если роль текущего пользователя
// (из контекста, установленного JWT middleware) не входит в список,
// возвращает HTTP 403 Forbidden.
// Должен использоваться после JWT middleware в цепочке обработки.
func RequireRole(logger *zap.Logger, roles ...string) fiber.Handler {
	// Формирование множества допустимых ролей для O(1) проверки.
	allowedRoles := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowedRoles[role] = true
	}

	return func(c fiber.Ctx) error {
		// Извлечение роли из контекста.
		role, ok := c.Locals(ContextKeyRole).(string)
		if !ok || role == "" {
			logger.Warn("роль не найдена в контексте запроса",
				zap.String("path", c.Path()),
			)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "доступ запрещён: роль не определена",
			})
		}

		// Проверка роли в множестве допустимых.
		if !allowedRoles[role] {
			logger.Debug("доступ запрещён по роли",
				zap.String("path", c.Path()),
				zap.String("user_role", role),
				zap.Strings("allowed_roles", roles),
			)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "доступ запрещён: недостаточно прав",
			})
		}

		return c.Next()
	}
}
