package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OfflineService struct {
	routeService    *RouteService
	locationService *LocationService
	logger          *zap.Logger
}

func NewOfflineService(routeService *RouteService, locationService *LocationService, logger *zap.Logger) *OfflineService {
	return &OfflineService{
		routeService:    routeService,
		locationService: locationService,
		logger:          logger.Named("offline_service"),
	}
}

// BuildZipBundle returns an in-memory ZIP archive containing the route JSON and associated locations.
func (s *OfflineService) BuildZipBundle(ctx context.Context, currUserID string, routeID uuid.UUID) ([]byte, error) {
	// Fetch the full route (requires user id for permissions if needed, assuming the routeService allows checking).
	// We'll use FindByID which requires auth user to see private trips, etc.
	route, err := s.routeService.routeRepo.FindByID(ctx, routeID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения маршрута: %w", err)
	}

	points, err := s.routeService.routeRepo.FindPointsByRouteID(ctx, routeID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения точек маршрута: %w", err)
	}

	// We also need to fetch the detailed information of all locations in this route.
	locations := make(map[string]interface{})
	for _, pt := range points {
		loc, err := s.locationService.GetByID(ctx, pt.LocationID.String())
		if err == nil && loc != nil {
			locations[pt.LocationID.String()] = loc
		}
	}

	// Package full route with its points for the client
	routeData := map[string]interface{}{
		"route":  route,
		"points": points,
	}

	// Serialize structures to JSON
	routeJSON, err := json.Marshal(routeData)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации маршрута: %w", err)
	}

	locationsJSON, err := json.Marshal(locations)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации локаций: %w", err)
	}

	// Create a new zip archive buffer
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add route.json
	fRoute, err := zipWriter.Create("route.json")
	if err != nil {
		return nil, err
	}
	if _, err := fRoute.Write(routeJSON); err != nil {
		return nil, err
	}

	// Add locations.json
	fLocs, err := zipWriter.Create("locations.json")
	if err != nil {
		return nil, err
	}
	if _, err := fLocs.Write(locationsJSON); err != nil {
		return nil, err
	}

	// Close the writer to flush everything to buf
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("ошибка закрытия zip архива: %w", err)
	}

	s.logger.Info("сгенерирован offline-bundle для маршрута", zap.String("route_id", routeID.String()), zap.Int("size_bytes", buf.Len()))

	return buf.Bytes(), nil
}
