package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewTelemetryService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewTelemetryService(nil, logger)
	require.NotNil(t, svc)
}

func TestTelemetryStreamKeyConstant(t *testing.T) {
	assert.Equal(t, "analytics:telemetry_events", TelemetryStreamKey)
}

func TestNewTelemetryService_NilRedis(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	// Constructor should not panic even with nil redis.
	svc := NewTelemetryService(nil, logger)
	assert.NotNil(t, svc)
}
