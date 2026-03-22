package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
	"kudytudy-api/internal/models"
)

var (
	ErrLocationNotFound    = errors.New("локация не найдена")
	ErrBookingNotFound      = errors.New("бронь не найдена")
	ErrBookingForbidden     = errors.New("нет прав на эту бронь")
	ErrBookingUnavailable   = errors.New("нет доступных слотов на выбранные даты")
	ErrBookingInvalidStatus = errors.New("операция недоступна для текущего статуса брони")
)

type BookingService struct {
	bookingRepo  *database.BookingRepository
	locationRepo *database.LocationRepository
	logger       *zap.Logger
}

func NewBookingService(bookingRepo *database.BookingRepository, locationRepo *database.LocationRepository, logger *zap.Logger) *BookingService {
	return &BookingService{
		bookingRepo:  bookingRepo,
		locationRepo: locationRepo,
		logger:       logger.Named("booking_service"),
	}
}

func (s *BookingService) ListLocationSlots(ctx context.Context, locationID uuid.UUID, monthStart time.Time) ([]models.BookingSlot, error) {
	if _, err := s.locationRepo.FindByID(ctx, locationID.String()); err != nil {
		if errors.Is(err, database.ErrLocationNotFound) {
			return nil, ErrLocationNotFound
		}
		return nil, err
	}
	slots, err := s.bookingRepo.ListLocationSlots(ctx, locationID, monthStart)
	if err != nil {
		if errors.Is(err, database.ErrBookingLocationClosed) {
			return nil, ErrBookingUnavailable
		}
		return nil, err
	}
	return slots, nil
}

func (s *BookingService) Create(ctx context.Context, userID uuid.UUID, req *models.CreateBookingRequest) (*models.BookingResponse, error) {
	locationID, _ := uuid.Parse(req.LocationID)
	dateFrom, dateTo := req.ParseDates()
	today := time.Now().Truncate(24 * time.Hour)

	if dateFrom.Before(today) {
		return nil, ErrBookingUnavailable
	}

	location, err := s.locationRepo.FindByID(ctx, locationID.String())
	if err != nil {
		if errors.Is(err, database.ErrLocationNotFound) {
			return nil, ErrLocationNotFound
		}
		return nil, err
	}
	if !location.IsPublished {
		return nil, ErrBookingUnavailable
	}
	if location.Capacity <= 0 {
		return nil, ErrBookingUnavailable
	}
	if req.GuestsCount > location.Capacity && location.Capacity > 0 {
		return nil, ErrBookingUnavailable
	}

	nights := int(dateTo.Sub(dateFrom).Hours() / 24)
	totalPrice := location.PricePerNight * nights

	booking := &models.Booking{
		LocationID:   locationID,
		UserID:       userID,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		GuestsCount:  req.GuestsCount,
		TotalPrice:   totalPrice,
		Status:       models.BookingStatusPending,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		Comment:      req.Comment,
		HostComment:  "",
	}

	created, err := s.bookingRepo.Create(ctx, booking)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrBookingUnavailable), errors.Is(err, database.ErrBookingLocationClosed):
			return nil, ErrBookingUnavailable
		case errors.Is(err, database.ErrLocationNotFound):
			return nil, ErrLocationNotFound
		default:
			return nil, err
		}
	}

	resp, err := s.bookingRepo.GetByID(ctx, created.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("бронь создана",
		zap.String("booking_id", created.ID.String()),
		zap.String("location_id", locationID.String()),
		zap.String("user_id", userID.String()),
	)

	return resp, nil
}

func (s *BookingService) ConfirmOrReject(ctx context.Context, bookingID, hostID uuid.UUID, req *models.BookingActionRequest) (*models.BookingResponse, error) {
	resp, err := s.bookingRepo.ConfirmOrReject(ctx, bookingID, hostID, models.BookingAction(req.Action), req.HostComment)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrBookingNotFound):
			return nil, ErrBookingNotFound
		case errors.Is(err, database.ErrBookingForbidden):
			return nil, ErrBookingForbidden
		case errors.Is(err, database.ErrBookingInvalidStatus):
			return nil, ErrBookingInvalidStatus
		default:
			return nil, err
		}
	}
	return resp, nil
}

func (s *BookingService) Cancel(ctx context.Context, bookingID, userID uuid.UUID) (*models.BookingResponse, error) {
	resp, err := s.bookingRepo.CancelByUser(ctx, bookingID, userID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrBookingNotFound):
			return nil, ErrBookingNotFound
		case errors.Is(err, database.ErrBookingForbidden):
			return nil, ErrBookingForbidden
		case errors.Is(err, database.ErrBookingInvalidStatus):
			return nil, ErrBookingInvalidStatus
		default:
			return nil, err
		}
	}
	return resp, nil
}

func (s *BookingService) ListMyBookings(ctx context.Context, userID uuid.UUID) ([]models.BookingResponse, error) {
	return s.bookingRepo.ListByUserID(ctx, userID)
}

func (s *BookingService) ListHostBookings(ctx context.Context, hostID uuid.UUID) ([]models.BookingResponse, error) {
	return s.bookingRepo.ListByHostID(ctx, hostID)
}

func (s *BookingService) GetByID(ctx context.Context, bookingID uuid.UUID) (*models.BookingResponse, error) {
	resp, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, database.ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("ошибка получения брони: %w", err)
	}
	return resp, nil
}
