// Тесты для middleware пакета.
// Проверяют работу CORS, Request Logger и Recovery middleware.
package middleware

import (
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestNewCORS проверяет, что CORS middleware корректно создаётся
// и добавляет заголовки в ответ на preflight-запрос.
func TestNewCORS(t *testing.T) {
	app := fiber.New()
	app.Use(NewCORS())
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Выполнение preflight OPTIONS-запроса.
	req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Проверка наличия CORS-заголовков в ответе.
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Methods"))
}

// TestNewRequestLogger проверяет, что Request Logger middleware
// не мешает нормальной обработке запросов.
func TestNewRequestLogger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	app := fiber.New()
	app.Use(NewRequestLogger(logger))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", string(body))
}

// TestNewRequestLoggerWith404 проверяет логирование запроса с несуществующим маршрутом.
func TestNewRequestLoggerWith404(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	app := fiber.New()
	app.Use(NewRequestLogger(logger))

	req, _ := http.NewRequest(http.MethodGet, "/nonexistent", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestNewRecovery проверяет, что Recovery middleware перехватывает панику
// и возвращает ответ 500 вместо падения сервера.
func TestNewRecovery(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	app := fiber.New()
	app.Use(NewRecovery(logger))
	app.Get("/panic", func(c fiber.Ctx) error {
		panic("тестовая паника")
	})

	req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestNewRecoveryNoPanic проверяет, что Recovery middleware не влияет
// на нормальную обработку запросов (без паники).
func TestNewRecoveryNoPanic(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	app := fiber.New()
	app.Use(NewRecovery(logger))
	app.Get("/ok", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req, _ := http.NewRequest(http.MethodGet, "/ok", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", string(body))
}
