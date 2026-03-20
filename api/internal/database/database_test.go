// Тесты для пакета database.
// Проверяют создание менеджера (Singleton), PingAll при неинициализированных
// клиентах и корректность ResetSingleton.
package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestGetManagerSingleton проверяет, что GetManager возвращает один и тот же экземпляр.
func TestGetManagerSingleton(t *testing.T) {
	ResetSingleton()
	defer ResetSingleton()

	logger, _ := zap.NewDevelopment()

	m1 := GetManager(logger)
	m2 := GetManager(logger)

	require.NotNil(t, m1, "менеджер не должен быть nil")
	assert.Same(t, m1, m2, "GetManager должен возвращать один и тот же экземпляр")
}

// TestResetSingleton проверяет, что после сброса создаётся новый экземпляр.
func TestResetSingleton(t *testing.T) {
	ResetSingleton()

	logger, _ := zap.NewDevelopment()

	m1 := GetManager(logger)
	ResetSingleton()
	m2 := GetManager(logger)

	assert.NotSame(t, m1, m2, "после ResetSingleton должен быть создан новый экземпляр")
}

// TestPingAllWithNilClients проверяет, что PingAll корректно обрабатывает
// неинициализированные клиенты, возвращая статус "down" для каждого.
func TestPingAllWithNilClients(t *testing.T) {
	ResetSingleton()
	defer ResetSingleton()

	logger, _ := zap.NewDevelopment()
	m := GetManager(logger)

	ctx := context.Background()
	health := m.PingAll(ctx)

	// Все клиенты nil — все сервисы должны быть "down".
	expectedServices := []string{"postgres", "redis", "qdrant", "neo4j", "clickhouse", "minio"}
	for _, svc := range expectedServices {
		h, ok := health[svc]
		assert.True(t, ok, "сервис %q должен присутствовать в результатах", svc)
		assert.Equal(t, "down", h.Status, "сервис %q должен быть down при неинициализированном клиенте", svc)
		assert.NotEmpty(t, h.Error, "сервис %q должен содержать описание ошибки", svc)
	}

	assert.Len(t, health, 6, "должно быть ровно 6 сервисов в результатах")
}

// TestCloseWithNilClients проверяет, что Close не паникует при неинициализированных клиентах.
func TestCloseWithNilClients(t *testing.T) {
	ResetSingleton()
	defer ResetSingleton()

	logger, _ := zap.NewDevelopment()
	m := GetManager(logger)

	// Close не должен паниковать даже если все клиенты nil.
	assert.NotPanics(t, func() {
		m.Close(context.Background())
	}, "Close не должен паниковать при nil-клиентах")
}

// TestServiceHealthJSON проверяет структуру ServiceHealth.
func TestServiceHealthJSON(t *testing.T) {
	health := ServiceHealth{
		Status:    "up",
		LatencyMs: 5,
	}
	assert.Equal(t, "up", health.Status)
	assert.Equal(t, int64(5), health.LatencyMs)
	assert.Empty(t, health.Error)

	healthDown := ServiceHealth{
		Status:    "down",
		LatencyMs: 0,
		Error:     "connection refused",
	}
	assert.Equal(t, "down", healthDown.Status)
	assert.Equal(t, "connection refused", healthDown.Error)
}
