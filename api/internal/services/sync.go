package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
)

type SyncService struct {
	reviewService    *ReviewService
	telemetryService *TelemetryService
	logger           *zap.Logger
}

func NewSyncService(reviewService *ReviewService, telemetryService *TelemetryService, logger *zap.Logger) *SyncService {
	return &SyncService{
		reviewService:    reviewService,
		telemetryService: telemetryService,
		logger:           logger.Named("sync_service"),
	}
}

// ProcessSyncPayload safely applies batched offline actions array.
func (s *SyncService) ProcessSyncPayload(ctx context.Context, userID string, payload *models.SyncPayload) *models.SyncResponse {
	resp := &models.SyncResponse{Success: true}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		resp.Success = false
		resp.Errors = append(resp.Errors, "некорректный ID пользователя")
		return resp
	}

	// 1. Process Reviews
	for _, rev := range payload.Reviews {
		// Validating empty IDs just in case
		if rev.LocationID != nil && *rev.LocationID == "00000000-0000-0000-0000-000000000000" {
			resp.Errors = append(resp.Errors, "отзыв пропущен: пустой location_id")
			continue
		}
		
		_, err := s.reviewService.CreateReview(ctx, rev, userID)
		if err != nil {
			locID := "unknown"
			if rev.LocationID != nil {
				locID = *rev.LocationID
			}
			resp.Errors = append(resp.Errors, fmt.Sprintf("ошибка синхронизации отзыва для локации %s: %v", locID, err))
		}
	}

	// 2. Process Telemetry
	if s.telemetryService != nil {
		for _, tel := range payload.Telemetry {
			// As telemetry is asynchronously pushed to Redis, we can just call PushEvent
			s.telemetryService.PushEvent(tel.EventType, userUUID, tel.LocationID, tel.Lat, tel.Lon, tel.Metadata)
		}
	} else if len(payload.Telemetry) > 0 {
		s.logger.Warn("телеметрия пропущена: сервис телеметрии не инициализирован")
	}

	if len(resp.Errors) > 0 {
		resp.Success = false
	}

	s.logger.Info("обработан sync payload",
		zap.Int("reviews_count", len(payload.Reviews)),
		zap.Int("telemetry_count", len(payload.Telemetry)),
		zap.Bool("success", resp.Success),
	)

	return resp
}
