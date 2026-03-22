package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
)

const TelemetryStreamKey = "analytics:telemetry_events"

// TelemetryService pushes internal events to Redis Streams for background processing.
type TelemetryService struct {
	redisClient *database.RedisClient
	logger      *zap.Logger
}

func NewTelemetryService(redisClient *database.RedisClient, logger *zap.Logger) *TelemetryService {
	return &TelemetryService{
		redisClient: redisClient,
		logger:      logger.Named("telemetry_service"),
	}
}

// PushEvent writes a telemetry event to Redis Streams asynchronously to avoid delaying the response.
func (s *TelemetryService) PushEvent(eventType string, userID, locationID uuid.UUID, lat, lon float64, metadata map[string]any) {
	// Execute in a detached goroutine
	go func() {
		// Timeouts specifically for the push logic
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var metaString string
		if metadata != nil {
			metaBytes, _ := json.Marshal(metadata)
			metaString = string(metaBytes)
		}
		
		event := database.TelemetryEvent{
			EventID:    uuid.New(),
			Timestamp:  time.Now(),
			EventType:  eventType,
			UserID:     userID,
			LocationID: locationID,
			Lat:        lat,
			Lon:        lon,
			Metadata:   metaString,
		}

		eventJSON, err := json.Marshal(event)
		if err != nil {
			s.logger.Error("failed to marshal telemetry event", zap.Error(err))
			return
		}

		err = s.redisClient.Client.XAdd(bgCtx, &redis.XAddArgs{
			Stream: TelemetryStreamKey,
			Values: map[string]interface{}{"event": string(eventJSON)},
		}).Err()
		
		if err != nil {
			s.logger.Error("failed to push telemetry event to redis streams", zap.Error(err))
		}
	}()
}
