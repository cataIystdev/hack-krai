package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
)

// ---------------------------------------------------------------------------
// SyncService — unit tests
// ---------------------------------------------------------------------------

func TestNewSyncService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewSyncService(nil, nil, logger)
	require.NotNil(t, svc)
}

func TestSyncService_ProcessSyncPayload_InvalidUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewSyncService(nil, nil, logger)

	resp := svc.ProcessSyncPayload(context.Background(), "not-a-uuid", &models.SyncPayload{})
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Errors[0], "некорректный ID")
}

func TestSyncService_ProcessSyncPayload_EmptyPayload(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewSyncService(nil, nil, logger)

	// Valid UUID, but empty payload — should succeed (nothing to process).
	resp := svc.ProcessSyncPayload(context.Background(), "550e8400-e29b-41d4-a716-446655440000", &models.SyncPayload{})
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Empty(t, resp.Errors)
}

func TestSyncService_ProcessSyncPayload_NilTelemetryService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	// No telemetry service — telemetry events should be silently skipped.
	svc := NewSyncService(nil, nil, logger)

	payload := &models.SyncPayload{
		Telemetry: []models.TelemetryEventSyncPayload{
			{EventType: "page_view", Lat: 45.0, Lon: 39.0},
		},
	}

	resp := svc.ProcessSyncPayload(context.Background(), "550e8400-e29b-41d4-a716-446655440000", payload)
	require.NotNil(t, resp)
	// Telemetry skipped, no errors because missing telemetry service is logged but not an error.
	assert.True(t, resp.Success)
}

func TestSyncService_ProcessSyncPayload_ReviewWithNilLocationID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewSyncService(nil, nil, logger)

	nilLocID := "00000000-0000-0000-0000-000000000000"
	payload := &models.SyncPayload{
		Reviews: []models.CreateReviewRequest{
			{LocationID: &nilLocID, Rating: 5, Comment: "test"},
		},
	}

	resp := svc.ProcessSyncPayload(context.Background(), "550e8400-e29b-41d4-a716-446655440000", payload)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Errors[0], "пустой location_id")
}

func TestSyncService_ProcessSyncPayload_MultipleErrors(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewSyncService(nil, nil, logger)

	nilID1 := "00000000-0000-0000-0000-000000000000"
	nilID2 := "00000000-0000-0000-0000-000000000000"
	payload := &models.SyncPayload{
		Reviews: []models.CreateReviewRequest{
			{LocationID: &nilID1, Rating: 5},
			{LocationID: &nilID2, Rating: 3},
		},
	}

	resp := svc.ProcessSyncPayload(context.Background(), "550e8400-e29b-41d4-a716-446655440000", payload)
	assert.False(t, resp.Success)
	assert.Len(t, resp.Errors, 2) // Both reviews have nil location_id.
}
