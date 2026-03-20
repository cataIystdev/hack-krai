// Файл cors.go реализует CORS middleware для HTTP-сервера.
// Настраивает разрешённые источники, методы и заголовки для
// кросс-доменных запросов от фронтенда (Nuxt 3 PWA).
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// NewCORS создаёт и настраивает CORS middleware.
// Разрешает запросы с любых источников в режиме разработки.
// В продакшене следует ограничить AllowOrigins конкретными доменами.
// Возвращает настроенный middleware-обработчик Fiber.
func NewCORS() fiber.Handler {
	return cors.New(cors.Config{
		// AllowOrigins — список разрешённых источников запросов.
		// В режиме разработки разрешены все источники для удобства.
		AllowOrigins: []string{"*"},

		// AllowMethods — разрешённые HTTP-методы для кросс-доменных запросов.
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		},

		// AllowHeaders — разрешённые заголовки в кросс-доменных запросах.
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
		},

		// AllowCredentials — разрешить передачу cookies и заголовков авторизации.
		AllowCredentials: false,

		// MaxAge — время кэширования preflight-ответа в секундах (1 час).
		MaxAge: 3600,
	})
}
