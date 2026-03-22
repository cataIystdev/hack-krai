package services

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
	"kudytudy-api/internal/models"
)

// StorytellingService — генерация route assets для историй точек маршрута.
type StorytellingService struct {
	routeRepo     *database.RouteRepository
	locationRepo  *database.LocationRepository
	tripRepo      *database.TripRepository
	storage       *StorageService
	weather       *WeatherService
	logger        *zap.Logger
}

func NewStorytellingService(
	routeRepo *database.RouteRepository,
	locationRepo *database.LocationRepository,
	tripRepo *database.TripRepository,
	storage *StorageService,
	weather *WeatherService,
	logger *zap.Logger,
) *StorytellingService {
	return &StorytellingService{
		routeRepo:    routeRepo,
		locationRepo: locationRepo,
		tripRepo:     tripRepo,
		storage:      storage,
		weather:      weather,
		logger:       logger.Named("storytelling_service"),
	}
}

func (s *StorytellingService) GenerateStories(ctx context.Context, routeID, userID uuid.UUID) (*models.GenerateStoriesResponse, error) {
	route, points, locations, err := s.loadAuthorizedRoute(ctx, routeID, userID)
	if err != nil {
		return nil, err
	}

	generated := 0
	for _, point := range points {
		loc, ok := locations[point.LocationID]
		if !ok {
			continue
		}
		storyText := s.buildStoryText(loc, point)
		audioURL := ""
		if s.storage != nil {
			objectName := fmt.Sprintf("stories/%s/%s.txt", routeID.String(), point.ID.String())
			result, upErr := s.storage.Upload(ctx, objectName, bytes.NewReader([]byte(storyText)), int64(len(storyText)), "text/plain")
			if upErr != nil {
				s.logger.Warn("не удалось загрузить story asset", zap.Error(upErr))
			} else if result != nil {
				audioURL = result.URL
			}
		}

		if err := s.routeRepo.UpdatePointStory(ctx, routeID, point.ID, audioURL, storyText); err != nil {
			return nil, err
		}
		generated++
	}

	return &models.GenerateStoriesResponse{
		RouteID:     route.ID.String(),
		PointsCount: len(points),
		Generated:   generated,
	}, nil
}

func (s *StorytellingService) GetStories(ctx context.Context, routeID, userID uuid.UUID) ([]models.RouteStory, error) {
	_, points, locations, err := s.loadAuthorizedRoute(ctx, routeID, userID)
	if err != nil {
		return nil, err
	}

	stories := make([]models.RouteStory, 0, len(points))
	for _, point := range points {
		loc := locations[point.LocationID]
		duration := 60
		if point.StayDurationMin > 0 {
			duration = point.StayDurationMin
		}
		stories = append(stories, models.RouteStory{
			PointID:      point.ID.String(),
			LocationID:   point.LocationID.String(),
			LocationName: loc.Name,
			AudioURL:     point.AudioStoryURL,
			StoryText:    point.StoryText,
			DurationSec:  duration,
			WeatherHint:  point.WeatherCondition,
		})
	}
	return stories, nil
}

func (s *StorytellingService) loadAuthorizedRoute(
	ctx context.Context,
	routeID, userID uuid.UUID,
) (*models.Route, []models.RoutePointDB, map[uuid.UUID]models.Location, error) {
	route, err := s.routeRepo.FindByID(ctx, routeID)
	if err != nil {
		if err == database.ErrRouteNotFound {
			return nil, nil, nil, ErrRouteNotFound
		}
		return nil, nil, nil, err
	}

	trip, err := s.tripRepo.FindByID(ctx, route.TripID)
	if err != nil {
		return nil, nil, nil, err
	}
	if trip.CreatorID != userID {
		isMember, memberErr := s.tripRepo.IsMember(ctx, trip.ID, userID)
		if memberErr != nil {
			return nil, nil, nil, memberErr
		}
		if !isMember {
			return nil, nil, nil, ErrTripForbidden
		}
	}

	points, err := s.routeRepo.FindPointsByRouteID(ctx, routeID)
	if err != nil {
		return nil, nil, nil, err
	}

	locationIDs := make([]string, 0, len(points))
	for _, point := range points {
		locationIDs = append(locationIDs, point.LocationID.String())
	}
	locs, err := s.locationRepo.FindByIDs(ctx, locationIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	locMap := make(map[uuid.UUID]models.Location, len(locs))
	for _, loc := range locs {
		locMap[loc.ID] = loc
	}

	return route, points, locMap, nil
}

func (s *StorytellingService) buildStoryText(loc models.Location, point models.RoutePointDB) string {
	lines := []string{
		fmt.Sprintf("%s — одна из точек вашего маршрута по Краснодарскому краю.", loc.Name),
	}
	if loc.DescriptionShort != "" {
		lines = append(lines, loc.DescriptionShort)
	}
	if loc.Category != "" {
		lines = append(lines, fmt.Sprintf("Это место категории %s, поэтому здесь особенно хорошо задержаться на %d минут.", loc.Category, point.StayDurationMin))
	}
	if point.WeatherCondition != "" {
		lines = append(lines, fmt.Sprintf("Сейчас здесь ожидается погода: %s.", point.WeatherCondition))
	}
	lines = append(lines, "Сделайте паузу, осмотритесь и двигайтесь дальше в темпе маршрута.")
	return strings.Join(lines, " ")
}
