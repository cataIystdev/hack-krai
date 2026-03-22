package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// BookingService — unit tests (constructor & error sentinel checks)
// ---------------------------------------------------------------------------

func TestNewBookingService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewBookingService(nil, nil, logger)
	require.NotNil(t, svc)
}

func TestBookingErrorSentinels(t *testing.T) {
	assert.NotNil(t, ErrLocationNotFound)
	assert.NotNil(t, ErrBookingNotFound)
	assert.NotNil(t, ErrBookingForbidden)
	assert.NotNil(t, ErrBookingUnavailable)
	assert.NotNil(t, ErrBookingInvalidStatus)

	// All errors must be unique.
	errors := []error{
		ErrLocationNotFound,
		ErrBookingNotFound,
		ErrBookingForbidden,
		ErrBookingUnavailable,
		ErrBookingInvalidStatus,
	}
	for i := 0; i < len(errors); i++ {
		for j := i + 1; j < len(errors); j++ {
			assert.NotEqual(t, errors[i], errors[j],
				"error sentinels must be unique: %v == %v", errors[i], errors[j])
		}
	}
}

func TestBookingErrorMessages(t *testing.T) {
	assert.Contains(t, ErrLocationNotFound.Error(), "локация")
	assert.Contains(t, ErrBookingNotFound.Error(), "бронь")
	assert.Contains(t, ErrBookingForbidden.Error(), "прав")
	assert.Contains(t, ErrBookingUnavailable.Error(), "слот")
	assert.Contains(t, ErrBookingInvalidStatus.Error(), "статус")
}

func TestBookingService_NilRepos(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewBookingService(nil, nil, logger)
	assert.Nil(t, svc.bookingRepo)
	assert.Nil(t, svc.locationRepo)
}
