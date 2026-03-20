// Файл jwt_test.go содержит unit-тесты для JWT middleware.
// Тестируемые сценарии: отсутствие заголовка, некорректный формат,
// невалидный токен, успешная аутентификация.
package middleware

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"deep-krai-api/internal/config"
	"deep-krai-api/internal/services"
)

// testJWTService создаёт JWT-сервис с тестовыми настройками.
func testJWTService() *services.JWTService {
	cfg := config.JWTConfig{
		SecretKey:             "test-secret-key-for-jwt-testing-32chars-minimum",
		AccessTokenTTLMinutes: 15,
		RefreshTokenTTLHours:  168,
		Issuer:                "test-issuer",
	}
	return services.NewJWTService(cfg, zap.NewNop())
}

// createTestApp создаёт тестовое Fiber-приложение с JWT middleware.
func createTestApp(jwtSvc *services.JWTService) *fiber.App {
	app := fiber.New()
	app.Use(NewJWTAuth(jwtSvc, zap.NewNop()))
	app.Get("/protected", func(c fiber.Ctx) error {
		userID := c.Locals(ContextKeyUserID).(string)
		role := c.Locals(ContextKeyRole).(string)
		return c.JSON(fiber.Map{
			"user_id": userID,
			"role":    role,
		})
	})
	return app
}

// TestJWTMiddleware_NoAuthHeader проверяет возврат 401 при отсутствии заголовка.
func TestJWTMiddleware_NoAuthHeader(t *testing.T) {
	app := createTestApp(testJWTService())

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestJWTMiddleware_InvalidFormat проверяет возврат 401 при некорректном формате.
func TestJWTMiddleware_InvalidFormat(t *testing.T) {
	app := createTestApp(testJWTService())

	tests := []struct {
		name   string
		header string
	}{
		{"только токен без Bearer", "some-token-string"},
		{"Basic вместо Bearer", "Basic some-token"},
		{"пустой Bearer", "Bearer "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.header)
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		})
	}
}

// TestJWTMiddleware_InvalidToken проверяет возврат 401 при невалидном токене.
func TestJWTMiddleware_InvalidToken(t *testing.T) {
	app := createTestApp(testJWTService())

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.string")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestJWTMiddleware_ExpiredToken проверяет возврат 401 при истекшем токене.
func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	jwtSvc := testJWTService()
	app := createTestApp(jwtSvc)

	// Создание истекшего токена.
	cfg := config.JWTConfig{
		SecretKey:             "test-secret-key-for-jwt-testing-32chars-minimum",
		AccessTokenTTLMinutes: 15,
		RefreshTokenTTLHours:  168,
		Issuer:                "test-issuer",
	}
	claims := services.TokenClaims{
		UserID:    "user-123",
		Role:      "tourist",
		TokenType: services.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestJWTMiddleware_ValidToken проверяет успешную аутентификацию
// и инъекцию user_id и role в контекст.
func TestJWTMiddleware_ValidToken(t *testing.T) {
	jwtSvc := testJWTService()
	app := createTestApp(jwtSvc)

	accessToken, _, err := jwtSvc.GenerateAccessToken("user-test-123", "host")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyStr := string(body)
	assert.Contains(t, bodyStr, "user-test-123")
	assert.Contains(t, bodyStr, "host")
}

// TestJWTMiddleware_RefreshTokenRejected проверяет, что refresh-токен
// не принимается как access-токен.
func TestJWTMiddleware_RefreshTokenRejected(t *testing.T) {
	jwtSvc := testJWTService()
	app := createTestApp(jwtSvc)

	// Генерация refresh-токена.
	refreshToken, _, err := jwtSvc.GenerateRefreshToken("user-123", "tourist")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
