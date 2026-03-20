// Тесты для обработчика health-проверки.
// Проверяют формирование ответа при различных состояниях сервисов БД.
package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
)

// TestHealthCheckAllDown проверяет ответ health-эндпоинта, когда все БД недоступны.
// Ожидается статус 503 и status "degraded".
func TestHealthCheckAllDown(t *testing.T) {
	database.ResetSingleton()
	defer database.ResetSingleton()

	logger, _ := zap.NewDevelopment()
	dbManager := database.GetManager(logger)

	app := fiber.New()
	handler := NewHealthHandler(dbManager, logger)
	app.Get("/api/v1/health", handler.Check)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Статус 503, так как все сервисы "down".
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var response HealthResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	// Проверка структуры ответа.
	assert.Equal(t, "degraded", response.Status)
	assert.NotEmpty(t, response.Timestamp)
	assert.Len(t, response.Services, 6, "должно быть 6 сервисов в ответе")

	// Все сервисы должны быть "down".
	for name, health := range response.Services {
		assert.Equal(t, "down", health.Status, "сервис %q должен быть down", name)
	}
}

// TestHealthResponseStructure проверяет структуру JSON-ответа health-эндпоинта.
func TestHealthResponseStructure(t *testing.T) {
	database.ResetSingleton()
	defer database.ResetSingleton()

	logger, _ := zap.NewDevelopment()
	dbManager := database.GetManager(logger)

	app := fiber.New()
	handler := NewHealthHandler(dbManager, logger)
	app.Get("/api/v1/health", handler.Check)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	// Проверка наличия обязательных полей в ответе.
	assert.Contains(t, response, "status", "ответ должен содержать поле status")
	assert.Contains(t, response, "timestamp", "ответ должен содержать поле timestamp")
	assert.Contains(t, response, "services", "ответ должен содержать поле services")

	// Проверка наличия всех 6 сервисов.
	servicesMap, ok := response["services"].(map[string]interface{})
	require.True(t, ok, "services должен быть объектом")

	expectedServices := []string{"postgres", "redis", "qdrant", "neo4j", "clickhouse", "minio"}
	for _, svc := range expectedServices {
		assert.Contains(t, servicesMap, svc, "сервис %q должен быть в ответе", svc)
	}
}
