package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
)

// ---------------------------------------------------------------------------
// TripService — unit tests
// ---------------------------------------------------------------------------

func TestNewTripService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewTripService(nil, nil, nil, logger)
	require.NotNil(t, svc)
}

func TestTripService_ErrorSentinels(t *testing.T) {
	assert.NotNil(t, ErrTripNotFound)
	assert.NotNil(t, ErrTripForbidden)
	assert.NotNil(t, ErrInvalidInviteToken)
	assert.NotNil(t, ErrTripFull)
	assert.NotNil(t, ErrAlreadyMember)
	assert.NotNil(t, ErrInvalidDates)

	// All unique.
	errors := []error{ErrTripNotFound, ErrTripForbidden, ErrInvalidInviteToken, ErrTripFull, ErrAlreadyMember, ErrInvalidDates}
	for i := 0; i < len(errors); i++ {
		for j := i + 1; j < len(errors); j++ {
			assert.NotEqual(t, errors[i], errors[j])
		}
	}
}

func TestTripService_CollectionGroupVibes(t *testing.T) {
	assert.Equal(t, "group_vibes", CollectionGroupVibes)
}

func TestTripService_MergeGroupVibes_NilVibeRepo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewTripService(nil, nil, nil, logger) // vibeRepo = nil

	result, err := svc.MergeGroupVibes(context.Background(), uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, result) // Skipped because vibeRepo is nil.
}

func TestTripService_CheckMembership_CreatorIsAlwaysAllowed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewTripService(nil, nil, nil, logger)

	creatorID := uuid.New()
	trip := &models.Trip{CreatorID: creatorID}

	err := svc.checkMembership(context.Background(), trip, creatorID)
	assert.NoError(t, err)
}

func TestTripService_CheckMembership_NonCreatorNilRepo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewTripService(nil, nil, nil, logger)

	trip := &models.Trip{CreatorID: uuid.New()}
	otherUser := uuid.New()

	// tripRepo is nil, so IsMember call will panic — verifying that non-creator path attempts DB access.
	assert.Panics(t, func() {
		_ = svc.checkMembership(context.Background(), trip, otherUser)
	})
}

func TestUuidsToStrings(t *testing.T) {
	ids := []uuid.UUID{uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")}
	result := uuidsToStrings(ids)
	assert.Len(t, result, 2)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", result[0])
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", result[1])
}

func TestUuidsToStrings_Empty(t *testing.T) {
	result := uuidsToStrings([]uuid.UUID{})
	assert.Empty(t, result)
}
