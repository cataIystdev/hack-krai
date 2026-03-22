package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// WeatherPoint
// ---------------------------------------------------------------------------

func TestWeatherPointFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	wp := WeatherPoint{
		Point:        "Сочи",
		Latitude:     43.6,
		Longitude:    39.7,
		TempC:        18,
		Condition:    "rain",
		WindSpeed:    12,
		UpdatedAt:    now,
		Severity:     "moderate",
		NeedsRebuild: true,
	}

	assert.Equal(t, "Сочи", wp.Point)
	assert.Equal(t, 43.6, wp.Latitude)
	assert.Equal(t, 39.7, wp.Longitude)
	assert.Equal(t, 18, wp.TempC)
	assert.Equal(t, "rain", wp.Condition)
	assert.Equal(t, 12, wp.WindSpeed)
	assert.Equal(t, now, wp.UpdatedAt)
	assert.Equal(t, "moderate", wp.Severity)
	assert.True(t, wp.NeedsRebuild)
}

func TestWeatherPointJSON(t *testing.T) {
	wp := WeatherPoint{
		Point:     "Горячий Ключ",
		Latitude:  44.6,
		Longitude: 39.1,
		TempC:     22,
		Condition: "clear",
		WindSpeed: 5,
		UpdatedAt: time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(wp)
	require.NoError(t, err)

	var decoded WeatherPoint
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, wp.Point, decoded.Point)
	assert.Equal(t, wp.TempC, decoded.TempC)
	assert.Equal(t, wp.Condition, decoded.Condition)
}

func TestWeatherPointSeverityOmitempty(t *testing.T) {
	wp := WeatherPoint{
		Point:     "Анапа",
		Condition: "clear",
	}
	data, err := json.Marshal(wp)
	require.NoError(t, err)
	// severity and needs_rebuild omitted when empty/false
	assert.NotContains(t, string(data), "\"severity\"")
	assert.NotContains(t, string(data), "\"needs_rebuild\"")
}

func TestWeatherPointNeedsRebuildTrue(t *testing.T) {
	wp := WeatherPoint{
		Point:        "Туапсе",
		Condition:    "storm",
		Severity:     "high",
		NeedsRebuild: true,
	}
	data, err := json.Marshal(wp)
	require.NoError(t, err)
	assert.Contains(t, string(data), "\"needs_rebuild\":true")
	assert.Contains(t, string(data), "\"severity\":\"high\"")
}

// ---------------------------------------------------------------------------
// RouteRebuildRequest
// ---------------------------------------------------------------------------

func TestRouteRebuildRequestJSON(t *testing.T) {
	req := RouteRebuildRequest{
		WeatherOverride: map[string]any{
			"mountain_pass_blocked": true,
			"storm_severity":        "high",
		},
	}
	data, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Contains(t, string(data), "weather_override")

	var decoded RouteRebuildRequest
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, true, decoded.WeatherOverride["mountain_pass_blocked"])
}

func TestRouteRebuildRequestEmptyOverride(t *testing.T) {
	req := RouteRebuildRequest{
		WeatherOverride: map[string]any{},
	}
	assert.Empty(t, req.WeatherOverride)
}

// ---------------------------------------------------------------------------
// RouteStory
// ---------------------------------------------------------------------------

func TestRouteStory(t *testing.T) {
	story := RouteStory{
		PointID:      "point-1",
		LocationID:   "loc-1",
		LocationName: "Абрау-Дюрсо",
		AudioURL:     "https://minio.example.com/stories/xxx.mp3",
		StoryText:    "История о знаменитой винодельне.",
		DurationSec:  90,
		WeatherHint:  "rain",
	}

	assert.NotEmpty(t, story.LocationName)
	assert.NotEmpty(t, story.StoryText)
	assert.Greater(t, story.DurationSec, 0)
	assert.Equal(t, "rain", story.WeatherHint)
}

func TestRouteStoryWeatherHintOmitempty(t *testing.T) {
	story := RouteStory{
		PointID:      "point-2",
		LocationID:   "loc-2",
		LocationName: "Озеро Кардывач",
		StoryText:    "Горное озеро.",
		DurationSec:  60,
	}
	data, err := json.Marshal(story)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "\"weather_hint\"")
	assert.Contains(t, string(data), "\"location_name\":\"Озеро Кардывач\"")
}

func TestRouteStoryZeroDuration(t *testing.T) {
	story := RouteStory{
		PointID:     "point-3",
		StoryText:   "Текст",
		DurationSec: 0,
	}
	// DurationSec=0 is valid struct field value
	assert.Equal(t, 0, story.DurationSec)
}

// ---------------------------------------------------------------------------
// GenerateStoriesResponse
// ---------------------------------------------------------------------------

func TestGenerateStoriesResponse(t *testing.T) {
	resp := GenerateStoriesResponse{
		RouteID:     "route-1",
		PointsCount: 3,
		Generated:   2,
	}
	assert.Equal(t, "route-1", resp.RouteID)
	assert.Equal(t, 3, resp.PointsCount)
	assert.Equal(t, 2, resp.Generated)
}

func TestGenerateStoriesResponseAllGenerated(t *testing.T) {
	resp := GenerateStoriesResponse{
		RouteID:     "route-abc",
		PointsCount: 5,
		Generated:   5,
	}
	assert.Equal(t, resp.PointsCount, resp.Generated)
}

func TestGenerateStoriesResponseJSON(t *testing.T) {
	resp := GenerateStoriesResponse{
		RouteID:     "r1",
		PointsCount: 4,
		Generated:   3,
	}
	data, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(data), "\"route_id\":\"r1\"")
	assert.Contains(t, string(data), "\"points_count\":4")
	assert.Contains(t, string(data), "\"generated\":3")
}
