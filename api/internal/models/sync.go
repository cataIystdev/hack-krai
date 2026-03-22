package models

import "github.com/google/uuid"

// SyncPayload represents a batch of offline actions synchronized from the PWA.
type SyncPayload struct {
	Reviews   []CreateReviewRequest      `json:"reviews,omitempty"`
	Telemetry []TelemetryEventSyncPayload `json:"telemetry,omitempty"`
}

// TelemetryEventSyncPayload represents an offline telemetry event.
type TelemetryEventSyncPayload struct {
	EventType  string                 `json:"event_type"`
	LocationID uuid.UUID              `json:"location_id,omitempty"`
	Lat        float64                `json:"lat,omitempty"`
	Lon        float64                `json:"lon,omitempty"`
	Timestamp  int64                  `json:"timestamp,omitempty"` // Unix ms
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// SyncResponse returns the synchronization outcome, useful for partial successes.
type SyncResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}
