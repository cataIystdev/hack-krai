package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kudytudy-api/internal/services"
)

// ---------------------------------------------------------------------------
// AnalyticsHandler — unit tests
// ---------------------------------------------------------------------------

func TestNewAnalyticsHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	handler := NewAnalyticsHandler(nil, logger)
	require.NotNil(t, handler)
	assert.Nil(t, handler.repo) // nil repo is acceptable — CI/CD uses real DB.
}

// ---------------------------------------------------------------------------
// SyncHandler — unit tests
// ---------------------------------------------------------------------------

func TestNewSyncHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	syncService := services.NewSyncService(nil, nil, logger)
	handler := NewSyncHandler(syncService, logger)
	require.NotNil(t, handler)
}

// ---------------------------------------------------------------------------
// OfflineHandler — unit tests
// ---------------------------------------------------------------------------

func TestNewOfflineHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	offlineService := services.NewOfflineService(nil, nil, logger)
	handler := NewOfflineHandler(offlineService, logger)
	require.NotNil(t, handler)
}
