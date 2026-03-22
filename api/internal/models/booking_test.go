package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// CreateBookingRequest.Validate()
// ---------------------------------------------------------------------------

func TestCreateBookingRequestValidate(t *testing.T) {
	validBase := func() CreateBookingRequest {
		return CreateBookingRequest{
			LocationID:   "550e8400-e29b-41d4-a716-446655440000",
			DateFrom:     "2026-04-10",
			DateTo:       "2026-04-12",
			GuestsCount:  2,
			ContactName:  "Иван",
			ContactPhone: "+79990000000",
		}
	}

	t.Run("valid request", func(t *testing.T) {
		req := validBase()
		assert.Equal(t, "", req.Validate())
	})

	t.Run("empty location_id", func(t *testing.T) {
		req := validBase()
		req.LocationID = ""
		assert.Contains(t, req.Validate(), "location_id")
	})

	t.Run("invalid uuid location_id", func(t *testing.T) {
		req := validBase()
		req.LocationID = "bad-uuid"
		assert.Contains(t, req.Validate(), "UUID")
	})

	t.Run("empty date_from", func(t *testing.T) {
		req := validBase()
		req.DateFrom = ""
		assert.Contains(t, req.Validate(), "date_from")
	})

	t.Run("empty date_to", func(t *testing.T) {
		req := validBase()
		req.DateTo = ""
		assert.Contains(t, req.Validate(), "date_to")
	})

	t.Run("bad format date_from", func(t *testing.T) {
		req := validBase()
		req.DateFrom = "10-04-2026"
		assert.Contains(t, req.Validate(), "date_from")
	})

	t.Run("bad format date_to", func(t *testing.T) {
		req := validBase()
		req.DateTo = "2026/04/12"
		assert.Contains(t, req.Validate(), "date_to")
	})

	t.Run("date_to equals date_from", func(t *testing.T) {
		req := validBase()
		req.DateTo = req.DateFrom
		assert.NotEmpty(t, req.Validate())
	})

	t.Run("date_to before date_from", func(t *testing.T) {
		req := validBase()
		req.DateTo = "2026-04-09"
		assert.NotEmpty(t, req.Validate())
	})

	t.Run("guests_count zero", func(t *testing.T) {
		req := validBase()
		req.GuestsCount = 0
		assert.Contains(t, req.Validate(), "guests_count")
	})

	t.Run("guests_count negative", func(t *testing.T) {
		req := validBase()
		req.GuestsCount = -1
		assert.Contains(t, req.Validate(), "guests_count")
	})

	t.Run("empty contact_name", func(t *testing.T) {
		req := validBase()
		req.ContactName = ""
		assert.Contains(t, req.Validate(), "contact_name")
	})

	t.Run("empty contact_phone", func(t *testing.T) {
		req := validBase()
		req.ContactPhone = ""
		assert.Contains(t, req.Validate(), "contact_phone")
	})

	t.Run("comment is optional", func(t *testing.T) {
		req := validBase()
		req.Comment = ""
		assert.Equal(t, "", req.Validate())
	})
}

// ---------------------------------------------------------------------------
// CreateBookingRequest.ParseDates()
// ---------------------------------------------------------------------------

func TestCreateBookingRequestParseDates(t *testing.T) {
	req := &CreateBookingRequest{
		LocationID:   "550e8400-e29b-41d4-a716-446655440000",
		DateFrom:     "2026-04-10",
		DateTo:       "2026-04-15",
		GuestsCount:  2,
		ContactName:  "Иван",
		ContactPhone: "+79990000000",
	}
	require.Equal(t, "", req.Validate())

	from, to := req.ParseDates()
	assert.Equal(t, 2026, from.Year())
	assert.Equal(t, time.April, from.Month())
	assert.Equal(t, 10, from.Day())
	assert.Equal(t, 15, to.Day())
	assert.True(t, to.After(from))

	// Nights calculation matches service logic
	nights := int(to.Sub(from).Hours() / 24)
	assert.Equal(t, 5, nights)
}

// ---------------------------------------------------------------------------
// BookingActionRequest.Validate()
// ---------------------------------------------------------------------------

func TestBookingActionRequestValidate(t *testing.T) {
	t.Run("confirm is valid", func(t *testing.T) {
		assert.Equal(t, "", (&BookingActionRequest{Action: "confirm"}).Validate())
	})
	t.Run("reject is valid", func(t *testing.T) {
		assert.Equal(t, "", (&BookingActionRequest{Action: "reject"}).Validate())
	})
	t.Run("approve is invalid", func(t *testing.T) {
		assert.NotEmpty(t, (&BookingActionRequest{Action: "approve"}).Validate())
	})
	t.Run("empty is invalid", func(t *testing.T) {
		assert.NotEmpty(t, (&BookingActionRequest{Action: ""}).Validate())
	})
	t.Run("cancel is invalid for host action", func(t *testing.T) {
		assert.NotEmpty(t, (&BookingActionRequest{Action: "cancel"}).Validate())
	})
	t.Run("host_comment is optional", func(t *testing.T) {
		req := &BookingActionRequest{Action: "confirm", HostComment: "всё отлично"}
		assert.Equal(t, "", req.Validate())
	})
}

// ---------------------------------------------------------------------------
// LocationSlotsFilter
// ---------------------------------------------------------------------------

func TestLocationSlotsFilter(t *testing.T) {
	t.Run("normalize default month", func(t *testing.T) {
		filter := &LocationSlotsFilter{}
		filter.NormalizeDefaultMonth(time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC))
		assert.Equal(t, "2026-03", filter.Month)
	})

	t.Run("does not overwrite explicit month", func(t *testing.T) {
		filter := &LocationSlotsFilter{Month: "2026-05"}
		filter.NormalizeDefaultMonth(time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC))
		assert.Equal(t, "2026-05", filter.Month)
	})

	t.Run("valid month format", func(t *testing.T) {
		filter := &LocationSlotsFilter{Month: "2026-03"}
		assert.Equal(t, "", filter.Validate())
	})

	t.Run("invalid month format slash", func(t *testing.T) {
		filter := &LocationSlotsFilter{Month: "2026/03"}
		assert.NotEmpty(t, filter.Validate())
	})

	t.Run("invalid month format full date", func(t *testing.T) {
		filter := &LocationSlotsFilter{Month: "2026-03-15"}
		assert.NotEmpty(t, filter.Validate())
	})

	t.Run("empty month is invalid", func(t *testing.T) {
		filter := &LocationSlotsFilter{Month: ""}
		assert.NotEmpty(t, filter.Validate())
	})

	t.Run("parse month first day", func(t *testing.T) {
		filter := &LocationSlotsFilter{Month: "2026-07"}
		m := filter.ParseMonth()
		assert.Equal(t, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), m)
	})
}

// ---------------------------------------------------------------------------
// Booking constants
// ---------------------------------------------------------------------------

func TestBookingConstants(t *testing.T) {
	t.Run("booking statuses", func(t *testing.T) {
		assert.Equal(t, BookingStatus("pending"), BookingStatusPending)
		assert.Equal(t, BookingStatus("confirmed"), BookingStatusConfirmed)
		assert.Equal(t, BookingStatus("rejected"), BookingStatusRejected)
		assert.Equal(t, BookingStatus("cancelled"), BookingStatusCancelled)
	})

	t.Run("booking actions", func(t *testing.T) {
		assert.Equal(t, BookingAction("confirm"), BookingActionConfirm)
		assert.Equal(t, BookingAction("reject"), BookingActionReject)
	})
}
