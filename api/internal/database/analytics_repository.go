package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TelemetryEvent represents a single event for B2G analytics.
type TelemetryEvent struct {
	EventID    uuid.UUID
	Timestamp  time.Time
	EventType  string
	UserID     uuid.UUID
	LocationID uuid.UUID
	Lat        float64
	Lon        float64
	Metadata   string
}

// PredictionPoint represent an aggregated forecast point.
type PredictionPoint struct {
	LocationID      uuid.UUID `json:"location_id"`
	PredictedVisits int       `json:"predicted_visits"`
}

// HeatmapPoint represents a clustered location popularity.
type HeatmapPoint struct {
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Count int     `json:"count"`
}

// AnalyticsRepository handles queries to ClickHouse.
type AnalyticsRepository struct {
	ch     *ClickHouseClient
	logger *zap.Logger
}

func NewAnalyticsRepository(ch *ClickHouseClient, logger *zap.Logger) *AnalyticsRepository {
	repo := &AnalyticsRepository{
		ch:     ch,
		logger: logger.Named("analytics_repository"),
	}
	// Run schema initialization synchronously
	repo.initSchema(context.Background())
	return repo
}

func (r *AnalyticsRepository) initSchema(ctx context.Context) {
	query := `
CREATE TABLE IF NOT EXISTS events (
    event_id UUID,
    timestamp DateTime,
    event_type String,
    user_id UUID,
    location_id UUID,
    lat Float64,
    lon Float64,
    metadata String
) ENGINE = MergeTree()
ORDER BY (event_type, timestamp)
PARTITION BY toYYYYMM(timestamp);
`
	err := r.ch.Conn.Exec(ctx, query)
	if err != nil {
		r.logger.Error("failed to create events table in ClickHouse", zap.Error(err))
	} else {
		r.logger.Info("ClickHouse schema verified: events table ready")
	}
}

// InsertEvents batch-inserts telemetry into ClickHouse.
func (r *AnalyticsRepository) InsertEvents(ctx context.Context, events []TelemetryEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch, err := r.ch.Conn.PrepareBatch(ctx, "INSERT INTO events")
	if err != nil {
		return fmt.Errorf("prepare batch failed: %w", err)
	}

	for _, e := range events {
		err := batch.Append(
			e.EventID,
			e.Timestamp,
			e.EventType,
			e.UserID,
			e.LocationID,
			e.Lat,
			e.Lon,
			e.Metadata,
		)
		if err != nil {
			return fmt.Errorf("batch append failed: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("batch send failed: %w", err)
	}

	r.logger.Info("events batch inserted to ClickHouse", zap.Int("count", len(events)))
	return nil
}

// GetHeatmap computes a geographic heatmap of events.
func (r *AnalyticsRepository) GetHeatmap(ctx context.Context) ([]HeatmapPoint, error) {
	// Group points by rounding coordinates to cluster them geographically.
	// Filter out invalid/zero coordinates.
	query := `
SELECT 
    round(lat, 3) as rounded_lat, 
    round(lon, 3) as rounded_lon, 
    count() as count
FROM events
WHERE lat != 0 AND lon != 0
GROUP BY rounded_lat, rounded_lon
ORDER BY count DESC
LIMIT 1000
`
	rows, err := r.ch.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("heatmap query failed: %w", err)
	}
	defer rows.Close()

	var points []HeatmapPoint
	for rows.Next() {
		var p HeatmapPoint
		var lat, lon float64
		var count uint64
		if err := rows.Scan(&lat, &lon, &count); err != nil {
			return nil, err
		}
		p.Lat = lat
		p.Lon = lon
		p.Count = int(count)
		points = append(points, p)
	}
	return points, nil
}

// GetPredictions returns a basic forecast of location visits based on recent events (e.g. past 7 days).
func (r *AnalyticsRepository) GetPredictions(ctx context.Context) ([]PredictionPoint, error) {
	// A forecasting algorithm using historical counts to predict load.
	query := `
SELECT 
    location_id, 
    count() * 1.5 as predicted_visits
FROM events
WHERE timestamp >= now() - INTERVAL 7 DAY
  AND location_id != '00000000-0000-0000-0000-000000000000'
GROUP BY location_id
ORDER BY predicted_visits DESC
LIMIT 50
`
	rows, err := r.ch.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("predictions query failed: %w", err)
	}
	defer rows.Close()

	var predictions []PredictionPoint
	for rows.Next() {
		var p PredictionPoint
		var locID uuid.UUID
		var hits float64
		if err := rows.Scan(&locID, &hits); err != nil {
			return nil, err
		}
		p.LocationID = locID
		p.PredictedVisits = int(hits)
		predictions = append(predictions, p)
	}
	return predictions, nil
}
