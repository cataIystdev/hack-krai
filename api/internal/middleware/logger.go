// Файл logger.go реализует middleware для логирования HTTP-запросов.
// Интегрирует Fiber-логгер с Zap для структурированного логирования
// каждого входящего запроса с указанием метода, пути, статуса и Latency.
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// NewRequestLogger создаёт middleware для логирования HTTP-запросов через Zap.
// Для каждого запроса записывает: метод, путь, статус-код, длительность обработки,
// IP-адрес клиента и User-Agent. Уровень логирования зависит от статус-кода ответа:
// 5xx — Error, 4xx — Warn, остальные — Info.
func NewRequestLogger(logger *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Фиксация времени начала обработки запроса.
		start := time.Now()

		// Извлечение параметров запроса до передачи следующему обработчику.
		method := c.Method()
		path := c.Path()
		ip := c.IP()
		userAgent := c.Get("User-Agent")

		// Передача управления следующему обработчику в цепочке.
		err := c.Next()

		// Вычисление длительности обработки запроса.
		latency := time.Since(start)
		statusCode := c.Response().StatusCode()

		// Формирование набора полей для структурированного лога.
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("ip", ip),
			zap.String("user_agent", userAgent),
		}

		// Если обработчик вернул ошибку — добавляем её в лог.
		if err != nil {
			fields = append(fields, zap.Error(err))
		}

		// Выбор уровня логирования в зависимости от статус-кода.
		switch {
		case statusCode >= 500:
			logger.Error("HTTP-запрос", fields...)
		case statusCode >= 400:
			logger.Warn("HTTP-запрос", fields...)
		default:
			logger.Info("HTTP-запрос", fields...)
		}

		return err
	}
}
