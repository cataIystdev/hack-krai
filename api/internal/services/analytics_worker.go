package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
)

const (
	TelemetryConsumerGroup = "analytics_clickhouse_sink"
	TelemetryConsumerName  = "worker-1"
)

// AnalyticsWorker reads events from Redis Streams and batch-inserts them into ClickHouse.
type AnalyticsWorker struct {
	redisClient *database.RedisClient
	clickhouse  *database.AnalyticsRepository
	logger      *zap.Logger
}

func NewAnalyticsWorker(redisClient *database.RedisClient, clickhouse *database.AnalyticsRepository, logger *zap.Logger) *AnalyticsWorker {
	return &AnalyticsWorker{
		redisClient: redisClient,
		clickhouse:  clickhouse,
		logger:      logger.Named("analytics_worker"),
	}
}

// Start begins the event consumption loop.
func (w *AnalyticsWorker) Start(ctx context.Context) {
	// Ensure the consumer group exists, reading from ID 0 if creating freshly.
	err := w.redisClient.Client.XGroupCreateMkStream(ctx, TelemetryStreamKey, TelemetryConsumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		w.logger.Error("failed to create redis consumer group", zap.Error(err))
	}

	w.logger.Info("started analytics worker")

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopping analytics worker...")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *AnalyticsWorker) processBatch(ctx context.Context) {
	// Read up to 100 events from the stream.
	streams, err := w.redisClient.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    TelemetryConsumerGroup,
		Consumer: TelemetryConsumerName,
		Streams:  []string{TelemetryStreamKey, ">"}, // ">" means unread messages
		Count:    100,
		Block:    1 * time.Second, // block for up to 1 sec
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return // no new messages
		}
		w.logger.Error("xreadgroup failed", zap.Error(err))
		return
	}

	for _, stream := range streams {
		var eventsToInsert []database.TelemetryEvent
		var msgIDs []string

		for _, msg := range stream.Messages {
			msgIDs = append(msgIDs, msg.ID)
			
			eventJSON, ok := msg.Values["event"].(string)
			if !ok {
				continue
			}

			var event database.TelemetryEvent
			if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
				w.logger.Error("failed to unmarshal recorded event", zap.Error(err))
				continue
			}
			eventsToInsert = append(eventsToInsert, event)
		}

		if len(eventsToInsert) > 0 {
			err = w.clickhouse.InsertEvents(ctx, eventsToInsert)
			if err != nil {
				w.logger.Error("failed to sink events to ClickHouse (will retry next loop)", zap.Error(err))
				// Error occurred, we DO NOT XAck so messages will be kept pending and retried
				continue 
			}

			// Acknowledge successfully processed messages to remove them from pending list
			w.redisClient.Client.XAck(ctx, TelemetryStreamKey, TelemetryConsumerGroup, msgIDs...)
			w.logger.Info("processed and sinked telemetry batch", zap.Int("count", len(eventsToInsert)))
		}
	}
}
