// Файл rbac_test.go содержит unit-тесты для RBAC middleware.
// Тестируемые сценарии: доступ с допустимой ролью, отказ при недопустимой
// роли, отсутствие роли в контексте.
package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// createRBACTestApp создаёт тестовое Fiber-приложение с RBAC middleware.
// Предварительно устанавливает роль через промежуточный обработчик.
func createRBACTestApp(userRole string, allowedRoles ...string) *fiber.App {
	app := fiber.New()

	// Middleware, имитирующий JWT middleware и устанавливающий роль.
	app.Use(func(c fiber.Ctx) error {
		if userRole != "" {
			c.Locals(ContextKeyRole, userRole)
			c.Locals(ContextKeyUserID, "test-user-id")
		}
		return c.Next()
	})

	app.Use(RequireRole(zap.NewNop(), allowedRoles...))

	app.Get("/admin", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"access": "granted"})
	})

	return app
}

// TestRBAC_AllowedRole проверяет успешный доступ с допустимой ролью.
func TestRBAC_AllowedRole(t *testing.T) {
	app := createRBACTestApp("b2g_admin", "b2g_admin")

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// TestRBAC_MultipleAllowedRoles проверяет доступ при указании нескольких ролей.
func TestRBAC_MultipleAllowedRoles(t *testing.T) {
	tests := []struct {
		name     string
		userRole string
		status   int
	}{
		{"admin допущен", "b2g_admin", fiber.StatusOK},
		{"host допущен", "host", fiber.StatusOK},
		{"tourist отклонён", "tourist", fiber.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := createRBACTestApp(tt.userRole, "b2g_admin", "host")
			req := httptest.NewRequest("GET", "/admin", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.status, resp.StatusCode)
		})
	}
}

// TestRBAC_ForbiddenRole проверяет отказ в доступе при недопустимой роли.
func TestRBAC_ForbiddenRole(t *testing.T) {
	app := createRBACTestApp("tourist", "b2g_admin")

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// TestRBAC_NoRoleInContext проверяет отказ при отсутствии роли в контексте.
func TestRBAC_NoRoleInContext(t *testing.T) {
	app := createRBACTestApp("", "b2g_admin")

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
