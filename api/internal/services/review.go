package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
	"kudytudy-api/internal/models"
)

var (
	ErrReviewValidation = errors.New("ошибка валидации отзыва")
)

type ReviewService struct {
	reviewRepo *database.ReviewRepository
	userRepo   *database.UserRepository
	logger     *zap.Logger
}

func NewReviewService(reviewRepo *database.ReviewRepository, userRepo *database.UserRepository, logger *zap.Logger) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		userRepo:   userRepo,
		logger:     logger.Named("review_service"),
	}
}

func (s *ReviewService) CreateReview(ctx context.Context, req models.CreateReviewRequest, authorID string) (*models.Review, error) {
	if errStr := req.Validate(); errStr != "" {
		return nil, fmt.Errorf("%w: %s", ErrReviewValidation, errStr)
	}

	authorUUID, err := uuid.Parse(authorID)
	if err != nil {
		return nil, fmt.Errorf("invalid author id: %w", err)
	}

	var locID, bookID *uuid.UUID
	if req.LocationID != nil && *req.LocationID != "" {
		uid, _ := uuid.Parse(*req.LocationID)
		locID = &uid
	}
	if req.BookingID != nil && *req.BookingID != "" {
		uid, _ := uuid.Parse(*req.BookingID)
		bookID = &uid
	}

	// Calculate karma delta:
	// If a user reviews a location, they get +5 karma for activity (per GDD).
	// If a host reviews a user, the host gets +2, but normally host gives karma to user.
	// For simplicity in MVP, author gets the karma. 
	karmaDelta := 0
	if req.IsFromHost {
		karmaDelta = 2
	} else {
		karmaDelta = 5
	}

	review := &models.Review{
		AuthorID:   authorUUID,
		LocationID: locID,
		BookingID:  bookID,
		Rating:     req.Rating,
		Text:       req.Text,
		IsFromHost: req.IsFromHost,
		KarmaDelta: karmaDelta,
	}

	err = s.reviewRepo.Create(ctx, nil, review)
	if err != nil {
		return nil, fmt.Errorf("failed to save review: %w", err)
	}

	// Award karma to the author
	if karmaDelta != 0 {
		_, err = s.userRepo.UpdateKarma(ctx, authorID, karmaDelta)
		if err != nil {
			s.logger.Error("failed to award karma to reviewer", zap.String("author_id", authorID), zap.Error(err))
		}
	}

	return review, nil
}

func (s *ReviewService) GetLocationReviews(ctx context.Context, locationID string, limit, offset int) ([]models.ReviewResponse, error) {
	locUUID, err := uuid.Parse(locationID)
	if err != nil {
		return nil, fmt.Errorf("invalid location id: %w", err)
	}
	return s.reviewRepo.GetByLocationID(ctx, locUUID, limit, offset)
}

func (s *ReviewService) GetUserReviews(ctx context.Context, authorID string, limit, offset int) ([]models.ReviewResponse, error) {
	authorUUID, err := uuid.Parse(authorID)
	if err != nil {
		return nil, fmt.Errorf("invalid author id: %w", err)
	}
	return s.reviewRepo.GetByAuthorID(ctx, authorUUID, limit, offset)
}
