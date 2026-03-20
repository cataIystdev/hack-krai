// Файл recovery.go реализует middleware для перехвата паник (panic recovery).
// Предотвращает падение сервера при необработанных паниках в обработчиках,
// логирует стектрейс через Zap и возвращает клиенту ответ 500.
package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// NewRecovery создаёт middleware для перехвата паник.
// При возникновении паники логирует стектрейс через Zap и возвращает
// клиенту JSON-ответ с кодом 500 Internal Server Error.
// Без этого middleware необработанная паника приведёт к падению всего сервера.
func NewRecovery(logger *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		// Отложенная функция перехватывает панику (если она произошла).
		defer func() {
			if r := recover(); r != nil {
				// Получение стектрейса для отладки.
				stack := debug.Stack()

				// Логирование паники со стектрейсом.
				logger.Error("перехвачена паника в обработчике",
					zap.Any("panic", r),
					zap.String("stack", string(stack)),
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
				)

				// Формирование ответа клиенту.
				err = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Internal Server Error",
					"message": fmt.Sprintf("непредвиденная ошибка: %v", r),
				})
			}
		}()

		// Передача управления следующему обработчику в цепочке.
		return c.Next()
	}
}
