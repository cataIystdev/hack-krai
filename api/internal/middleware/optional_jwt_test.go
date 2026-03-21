// Файл optional_jwt_test.go содержит unit-тесты для OptionalJWTAuth middleware.
// Проверяет три сценария: отсутствие токена, невалидный токен, валидный токен.
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kudytudy-api/internal/config"
	"kudytudy-api/internal/services"
)

// newTestJWTService создаёт JWTService с тестовой конфигурацией.
func newTestJWTService() *services.JWTService {
	cfg := config.JWTConfig{
		SecretKey:             "test-secret-key-for-unit-tests-only",
		AccessTokenTTLMinutes: 15,
		RefreshTokenTTLHours:  168,
		Issuer:                "test-issuer",
	}
	return services.NewJWTService(cfg, zap.NewNop())
}

// TestOptionalJWTAuth_NoToken проверяет, что запрос без токена
// проходит без ошибки и без идентификации.
func TestOptionalJWTAuth_NoToken(t *testing.T) {
	jwtService := newTestJWTService()
	logger := zap.NewNop()

	app := fiber.New()
	app.Use(NewOptionalJWTAuth(jwtService, logger))
	app.Get("/test", func(c fiber.Ctx) error {
		userID := c.Locals(ContextKeyUserID)
		return c.JSON(fiber.Map{"user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode,
		"запрос без токена должен пройти со статусом 200")
}

// TestOptionalJWTAuth_InvalidToken проверяет, что запрос с невалидным токеном
// проходит без ошибки и без идентификации.
func TestOptionalJWTAuth_InvalidToken(t *testing.T) {
	jwtService := newTestJWTService()
	logger := zap.NewNop()

	app := fiber.New()
	app.Use(NewOptionalJWTAuth(jwtService, logger))
	app.Get("/test", func(c fiber.Ctx) error {
		userID := c.Locals(ContextKeyUserID)
		return c.JSON(fiber.Map{"user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode,
		"запрос с невалидным токеном должен пройти со статусом 200")
}

// TestOptionalJWTAuth_ValidToken проверяет, что запрос с валидным токеном
// проходит с идентификацией (user_id и role в Locals).
func TestOptionalJWTAuth_ValidToken(t *testing.T) {
	jwtService := newTestJWTService()
	logger := zap.NewNop()

	// Генерация валидного access-токена.
	token, _, err := jwtService.GenerateAccessToken("test-user-id", "tourist")
	require.NoError(t, err)

	var capturedUserID, capturedRole interface{}

	app := fiber.New()
	app.Use(NewOptionalJWTAuth(jwtService, logger))
	app.Get("/test", func(c fiber.Ctx) error {
		capturedUserID = c.Locals(ContextKeyUserID)
		capturedRole = c.Locals(ContextKeyRole)
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "test-user-id", capturedUserID,
		"user_id должен быть извлечён из валидного токена")
	assert.Equal(t, "tourist", capturedRole,
		"role должна быть извлечена из валидного токена")
}

// TestOptionalJWTAuth_BadFormat проверяет, что запрос с некорректным форматом
// Authorization (без Bearer) проходит без ошибки.
func TestOptionalJWTAuth_BadFormat(t *testing.T) {
	jwtService := newTestJWTService()
	logger := zap.NewNop()

	app := fiber.New()
	app.Use(NewOptionalJWTAuth(jwtService, logger))
	app.Get("/test", func(c fiber.Ctx) error {
		userID := c.Locals(ContextKeyUserID)
		return c.JSON(fiber.Map{"user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic some-creds")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode,
		"запрос с Basic auth должен пройти без ошибки")
}
