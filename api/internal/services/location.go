// Файл location.go реализует бизнес-логику для работы с локациями.
// Предоставляет операции CRUD, валидацию данных, генерацию slug
// и делегирует пространственные запросы в LocationRepository.
package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
	"kudytudy-api/internal/models"
)

// LocationService — сервис управления локациями.
// Содержит бизнес-логику CRUD и валидацию.
type LocationService struct {
	repo   *database.LocationRepository
	logger *zap.Logger
}

// NewLocationService создаёт экземпляр сервиса локаций.
func NewLocationService(repo *database.LocationRepository, logger *zap.Logger) *LocationService {
	return &LocationService{
		repo:   repo,
		logger: logger,
	}
}

// Create создаёт новую локацию.
// Генерирует slug из названия, валидирует данные.
func (s *LocationService) Create(ctx context.Context, ownerID string, req *models.CreateLocationRequest) (*models.Location, error) {
	if msg := req.Validate(); msg != "" {
		return nil, fmt.Errorf("%s", msg)
	}

	ownerUUID, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, fmt.Errorf("некорректный owner_id: %w", err)
	}

	slug := models.GenerateSlug(req.Name)
	if slug == "" {
		slug = "location"
	}

	// Проверка уникальности slug; при совпадении добавляем суффикс.
	baseSlug := slug
	suffix := 1
	for {
		_, err := s.repo.FindBySlug(ctx, slug)
		if err != nil {
			break
		}
		suffix++
		slug = fmt.Sprintf("%s-%d", baseSlug, suffix)
	}

	accessLevel := models.AccessLevelOpen
	if req.AccessLevel != "" {
		accessLevel = models.AccessLevel(req.AccessLevel)
	}

	densityLevel := models.DensityGreen
	if req.DensityLevel != "" {
		densityLevel = models.DensityLevel(req.DensityLevel)
	}

	loc := &models.Location{
		OwnerID:          ownerUUID,
		Slug:             slug,
		Name:             req.Name,
		DescriptionShort: req.DescriptionShort,
		DescriptionFull:  req.DescriptionFull,
		Category:         req.Category,
		Tags:             req.Tags,
		PricePerNight:    req.PricePerNight,
		Capacity:         req.Capacity,
		AccessLevel:      accessLevel,
		DensityLevel:     densityLevel,
		ChildFriendly:    req.ChildFriendly,
		Address:          req.Address,
		IsPublished:      req.IsPublished,
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		PreviewImageURL:  req.PreviewImageURL,
		GalleryURLs:      req.GalleryURLs,
	}

	if loc.Tags == nil {
		loc.Tags = []string{}
	}

	created, err := s.repo.Create(ctx, loc)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// GetByID возвращает локацию по UUID.
func (s *LocationService) GetByID(ctx context.Context, id string) (*models.Location, error) {
	return s.repo.FindByID(ctx, id)
}

// Update обновляет локацию. Только владелец может обновить.
func (s *LocationService) Update(ctx context.Context, id, ownerID string, req *models.UpdateLocationRequest) (*models.Location, error) {
	return s.repo.Update(ctx, id, ownerID, req)
}

// Delete удаляет локацию. Только владелец может удалить.
func (s *LocationService) Delete(ctx context.Context, id, ownerID string) error {
	return s.repo.Delete(ctx, id, ownerID)
}

// Search выполняет пространственный поиск локаций с фильтрами и пагинацией.
func (s *LocationService) Search(ctx context.Context, filter *models.LocationFilter) (*models.LocationListResponse, error) {
	locations, total, err := s.repo.Search(ctx, filter)
	if err != nil {
		return nil, err
	}

	if locations == nil {
		locations = []models.Location{}
	}

	return &models.LocationListResponse{
		Locations: locations,
		Total:     total,
		Page:      filter.Page,
		PerPage:   filter.PerPage,
	}, nil
}
