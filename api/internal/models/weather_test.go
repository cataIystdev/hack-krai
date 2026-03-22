package models

import "testing"

func TestGenerateStoriesResponse(t *testing.T) {
	resp := GenerateStoriesResponse{
		RouteID:     "route-1",
		PointsCount: 3,
		Generated:   2,
	}

	if resp.RouteID != "route-1" || resp.PointsCount != 3 || resp.Generated != 2 {
		t.Fatalf("неожиданный GenerateStoriesResponse: %+v", resp)
	}
}

func TestRouteStory(t *testing.T) {
	story := RouteStory{
		PointID:      "point-1",
		LocationID:   "loc-1",
		LocationName: "Абрау-Дюрсо",
		AudioURL:     "https://example.com/story.txt",
		StoryText:    "История о локации",
		DurationSec:  90,
		WeatherHint:  "rain",
	}

	if story.LocationName == "" || story.StoryText == "" || story.DurationSec <= 0 {
		t.Fatalf("story содержит некорректные данные: %+v", story)
	}
}
