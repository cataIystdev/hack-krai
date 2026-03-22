package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
	"kudytudy-api/internal/models"
)

var ErrRouteNotFound = errors.New("маршрут не найден")

// RebuildRoute пересобирает существующий маршрут на основе текущих точек и weather advisory.
func (s *RouteService) RebuildRoute(ctx context.Context, routeID, userID uuid.UUID, weather *WeatherService, override map[string]any) (*models.RoutePreview, error) {
	if s.routeRepo == nil || s.tripRepo == nil {
		return nil, &ValidationError{Message: "маршрутный сервис недоступен"}
	}

	route, err := s.routeRepo.FindByID(ctx, routeID)
	if err != nil {
		if errors.Is(err, database.ErrRouteNotFound) {
			return nil, ErrRouteNotFound
		}
		return nil, err
	}

	trip, err := s.tripRepo.FindByID(ctx, route.TripID)
	if err != nil {
		return nil, err
	}
	if trip.CreatorID != userID {
		isMember, memberErr := s.tripRepo.IsMember(ctx, trip.ID, userID)
		if memberErr != nil {
			return nil, memberErr
		}
		if !isMember {
			return nil, ErrTripForbidden
		}
	}

	dbPoints, err := s.routeRepo.FindPointsByRouteID(ctx, routeID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(dbPoints))
	for _, point := range dbPoints {
		ids = append(ids, point.LocationID.String())
	}
	locations, err := s.locationRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	locMap := make(map[string]models.Location, len(locations))
	for _, loc := range locations {
		locMap[loc.ID.String()] = loc
	}

	speed := transportSpeedKmH[route.Transport]
	if speed == 0 {
		speed = transportSpeedKmH["car"]
	}

	sortNeeded := false
	rebuildPoints := make([]models.RoutePoint, 0, len(dbPoints))
	for _, dbPoint := range dbPoints {
		loc, ok := locMap[dbPoint.LocationID.String()]
		if !ok {
			continue
		}
		var weatherCondition string
		var tempC *int
		if weather != nil {
			advisory, wErr := weather.GetTravelAdvisory(ctx, loc.Latitude, loc.Longitude)
			if wErr == nil && advisory != nil {
				weatherCondition = advisory.Condition
				tempCopy := advisory.TempC
				tempC = &tempCopy
				if advisory.NeedsRebuild {
					sortNeeded = true
				}
			}
		}
		if blocked, ok := override["mountain_pass_blocked"].(bool); ok && blocked && (loc.Category == "trail" || loc.Category == "extreme") {
			sortNeeded = true
		}

		rebuildPoints = append(rebuildPoints, models.RoutePoint{
			LocationID:          loc.ID.String(),
			Name:                loc.Name,
			Latitude:            loc.Latitude,
			Longitude:           loc.Longitude,
			Category:            loc.Category,
			PreviewImageURL:     loc.PreviewImageURL,
			Order:               len(rebuildPoints) + 1,
			StayDurationMin:     dbPoint.StayDurationMin,
			DayNumber:           dbPoint.DayNumber,
			TimeSlot:            dbPoint.TimeSlot,
			TargetAudience:      dbPoint.TargetAudience,
			AudioStoryURL:       dbPoint.AudioStoryURL,
			StoryText:           dbPoint.StoryText,
			WeatherCondition:    weatherCondition,
			WeatherTempC:        tempC,
		})
	}

	if sortNeeded {
		// В weather-warning режиме выталкиваем risky outdoor категории в конец маршрута.
		sort.SliceStable(rebuildPoints, func(i, j int) bool {
			iRisk := rebuildPoints[i].Category == "trail" || rebuildPoints[i].Category == "extreme"
			jRisk := rebuildPoints[j].Category == "trail" || rebuildPoints[j].Category == "extreme"
			if iRisk == jRisk {
				return rebuildPoints[i].Order < rebuildPoints[j].Order
			}
			return !iRisk && jRisk
		})
		for i := range rebuildPoints {
			rebuildPoints[i].Order = i + 1
		}
	}

	totalDistance := 0.0
	totalDuration := 0
	for i := range rebuildPoints {
		if i > 0 {
			prev := rebuildPoints[i-1]
			curr := &rebuildPoints[i]
			dist := haversineDistance(prev.Latitude, prev.Longitude, curr.Latitude, curr.Longitude)
			curr.DistanceFromPrevKm = math.Round(dist*10) / 10
			curr.DurationFromPrevMin = int(math.Ceil(dist / speed * 60))
			totalDistance += dist
			totalDuration += curr.DurationFromPrevMin
		}
		totalDuration += rebuildPoints[i].StayDurationMin
	}

	dbRebuildPoints := make([]models.RoutePointDB, 0, len(rebuildPoints))
	for _, point := range rebuildPoints {
		locID, parseErr := uuid.Parse(point.LocationID)
		if parseErr != nil {
			continue
		}
		dbRebuildPoints = append(dbRebuildPoints, models.RoutePointDB{
			LocationID:          locID,
			Position:            point.Order,
			DayNumber:           point.DayNumber,
			TimeSlot:            point.TimeSlot,
			TargetAudience:      point.TargetAudience,
			StayDurationMin:     point.StayDurationMin,
			DistanceFromPrevKm:  point.DistanceFromPrevKm,
			DurationFromPrevMin: point.DurationFromPrevMin,
			AudioStoryURL:       point.AudioStoryURL,
			StoryText:           point.StoryText,
			WeatherCondition:    point.WeatherCondition,
			WeatherTempC:        point.WeatherTempC,
		})
	}

	summary := fmt.Sprintf("%s Weather rebuild: %d точек, %.0f км.", route.Name, len(rebuildPoints), totalDistance)
	if err := s.routeRepo.ReplacePoints(ctx, routeID, dbRebuildPoints); err != nil {
		return nil, err
	}
	if err := s.routeRepo.UpdateRouteSummaryAndMetrics(ctx, routeID, math.Round(totalDistance*10)/10, totalDuration, route.EstimatedCostRub, summary); err != nil {
		return nil, err
	}

	s.logger.Info("маршрут перестроен",
		zap.String("route_id", routeID.String()),
		zap.Bool("weather_reorder", sortNeeded),
	)

	return &models.RoutePreview{
		ID:               routeID.String(),
		TripID:           route.TripID.String(),
		Name:             route.Name,
		Status:           route.Status,
		Points:           rebuildPoints,
		TotalDistanceKm:  math.Round(totalDistance*10) / 10,
		TotalDurationMin: totalDuration,
		Transport:        route.Transport,
		PointsCount:      len(rebuildPoints),
		EstimatedCostRub: route.EstimatedCostRub,
		Summary:          summary,
	}, nil
}
