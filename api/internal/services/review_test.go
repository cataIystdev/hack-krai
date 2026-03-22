package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
)

// TestReviewService_CreateReview_Validation checks that invalid requests
// are rejected before hitting the database.
func TestReviewService_CreateReview_Validation(t *testing.T) {
	logger := zap.NewNop()
	// Pass nil for repositories; we expect validation to fail and return an error
	// without attempting to dereference the repositories.
	service := NewReviewService(nil, nil, logger)

	req := models.CreateReviewRequest{
		Rating: 0, // Invalid, should be 1-5
		Text:   "Short", // Invalid, should be at least 10 chars
	}

	review, err := service.CreateReview(context.Background(), req, "valid-uuid-not-reached")
	
	// Expect an error related to validation
	assert.Error(t, err)
	assert.Nil(t, review)
	assert.ErrorIs(t, err, ErrReviewValidation)
	assert.Contains(t, err.Error(), "rating должен быть от 1 до 5")
}

// TestReviewService_CreateReview_InvalidUUID checks that an invalid author UUID is rejected.
func TestReviewService_CreateReview_InvalidUUID(t *testing.T) {
	logger := zap.NewNop()
	service := NewReviewService(nil, nil, logger)

	strUUID := "d718b9b8-db8a-49f4-9fb1-df611f7ff36" // missing one char
	req := models.CreateReviewRequest{
		Rating: 5,
		Text:   "This is a properly long review text.",
		LocationID: &strUUID,
	}

	_, err := service.CreateReview(context.Background(), req, "invalid-author-uuid")
	
	assert.ErrorIs(t, err, ErrReviewValidation)
	assert.Contains(t, err.Error(), "location_id должен быть валидным UUID")
}
