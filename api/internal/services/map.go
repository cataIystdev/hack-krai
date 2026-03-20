// Файл map.go реализует бизнес-логику для карты локаций.
// MapService отвечает за выборку точек для отображения маркеров на карте
// с поддержкой bbox-фильтрации, категорий и плотности.
package services

import (
	"context"

	"go.uber.org/zap"

	"kudytudy-api/internal/database"
	"kudytudy-api/internal/models"
)

// MapService — сервис для работы с картой локаций.
// Предоставляет метод получения точек для отображения маркеров на карте.
type MapService struct {
	locationRepo *database.LocationRepository
	logger       *zap.Logger
}

// NewMapService создаёт экземпляр сервиса карты.
func NewMapService(locationRepo *database.LocationRepository, logger *zap.Logger) *MapService {
	return &MapService{
		locationRepo: locationRepo,
		logger:       logger.Named("map_service"),
	}
}

// GetMapLocations возвращает точки для отображения на карте.
// Поддерживает фильтрацию по bbox, категории и плотности.
// Если bbox не задан, возвращает все опубликованные локации до лимита.
func (s *MapService) GetMapLocations(ctx context.Context, filter *models.MapLocationFilter) (*models.MapLocationsResponse, error) {
	// Валидация фильтра.
	if errMsg := filter.Validate(); errMsg != "" {
		return nil, &ValidationError{Message: errMsg}
	}

	points, err := s.locationRepo.SearchForMap(ctx, filter)
	if err != nil {
		s.logger.Error("ошибка получения точек карты", zap.Error(err))
		return nil, err
	}

	// Гарантируем непустой массив в JSON ([] вместо null).
	if points == nil {
		points = []models.MapPoint{}
	}

	return &models.MapLocationsResponse{
		Points: points,
		Total:  len(points),
	}, nil
}

// ValidationError — ошибка валидации входных данных.
type ValidationError struct {
	Message string
}

// Error реализует интерфейс error.
func (e *ValidationError) Error() string {
	return e.Message
}
