package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
)

var (
	ErrBookingNotFound       = errors.New("бронь не найдена")
	ErrBookingUnavailable    = errors.New("слоты недоступны для бронирования")
	ErrBookingForbidden      = errors.New("нет прав на бронь")
	ErrBookingInvalidStatus  = errors.New("некорректный статус брони для этой операции")
	ErrBookingLocationClosed = errors.New("локация недоступна для бронирования")
)

type bookingQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type BookingRepository struct {
	pg     *PostgresClient
	logger *zap.Logger
}

func NewBookingRepository(pg *PostgresClient, logger *zap.Logger) *BookingRepository {
	return &BookingRepository{
		pg:     pg,
		logger: logger.Named("booking_repository"),
	}
}

func scanBooking(row pgx.Row) (*models.Booking, error) {
	var booking models.Booking
	err := row.Scan(
		&booking.ID,
		&booking.LocationID,
		&booking.UserID,
		&booking.DateFrom,
		&booking.DateTo,
		&booking.GuestsCount,
		&booking.TotalPrice,
		&booking.Status,
		&booking.ContactName,
		&booking.ContactPhone,
		&booking.Comment,
		&booking.HostComment,
		&booking.CancelledBy,
		&booking.ConfirmedAt,
		&booking.CancelledAt,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	return &booking, err
}

func scanBookingResponse(row pgx.Row) (*models.BookingResponse, error) {
	var booking models.BookingResponse
	err := row.Scan(
		&booking.ID,
		&booking.LocationID,
		&booking.UserID,
		&booking.DateFrom,
		&booking.DateTo,
		&booking.GuestsCount,
		&booking.TotalPrice,
		&booking.Status,
		&booking.ContactName,
		&booking.ContactPhone,
		&booking.Comment,
		&booking.HostComment,
		&booking.CancelledBy,
		&booking.ConfirmedAt,
		&booking.CancelledAt,
		&booking.CreatedAt,
		&booking.UpdatedAt,
		&booking.LocationName,
		&booking.LocationSlug,
		&booking.LocationPreviewImage,
	)
	return &booking, err
}

func scanBookingSlot(row pgx.Row) (*models.BookingSlot, error) {
	var slot models.BookingSlot
	err := row.Scan(
		&slot.ID,
		&slot.LocationID,
		&slot.SlotDate,
		&slot.TotalCapacity,
		&slot.AvailableCapacity,
		&slot.IsClosed,
		&slot.CreatedAt,
		&slot.UpdatedAt,
	)
	return &slot, err
}

func (r *BookingRepository) ensureSlotsForRange(ctx context.Context, q bookingQuerier, locationID uuid.UUID, dateFrom, dateTo time.Time) error {
	query := `
		INSERT INTO booking_slots (location_id, slot_date, total_capacity, available_capacity)
		SELECT
			$1,
			gs::date,
			GREATEST(l.capacity, 1),
			GREATEST(l.capacity, 1)
		FROM generate_series($2::date, ($3::date - INTERVAL '1 day'), INTERVAL '1 day') AS gs
		JOIN locations l ON l.id = $1
		WHERE l.is_published = true
		ON CONFLICT (location_id, slot_date) DO NOTHING
	`

	tag, err := q.Exec(ctx, query, locationID, dateFrom, dateTo)
	if err != nil {
		return fmt.Errorf("ошибка подготовки booking slots: %w", err)
	}

	if tag.RowsAffected() == 0 {
		var isPublished bool
		err = q.QueryRow(ctx, `SELECT is_published FROM locations WHERE id = $1`, locationID).Scan(&isPublished)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrLocationNotFound
			}
			return fmt.Errorf("ошибка проверки локации: %w", err)
		}
		if !isPublished {
			return ErrBookingLocationClosed
		}
	}

	return nil
}

func (r *BookingRepository) ListLocationSlots(ctx context.Context, locationID uuid.UUID, monthStart time.Time) ([]models.BookingSlot, error) {
	monthEnd := monthStart.AddDate(0, 1, 0)
	if err := r.ensureSlotsForRange(ctx, r.pg.Pool, locationID, monthStart, monthEnd); err != nil {
		return nil, err
	}

	rows, err := r.pg.Pool.Query(ctx, `
		SELECT id, location_id, slot_date, total_capacity, available_capacity, is_closed, created_at, updated_at
		FROM booking_slots
		WHERE location_id = $1 AND slot_date >= $2 AND slot_date < $3
		ORDER BY slot_date ASC
	`, locationID, monthStart, monthEnd)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения слотов: %w", err)
	}
	defer rows.Close()

	var slots []models.BookingSlot
	for rows.Next() {
		slot, err := scanBookingSlot(rows)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования слота: %w", err)
		}
		slots = append(slots, *slot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по слотам: %w", err)
	}

	return slots, nil
}

func (r *BookingRepository) Create(ctx context.Context, booking *models.Booking) (*models.Booking, error) {
	tx, err := r.pg.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка начала транзакции бронирования: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.ensureSlotsForRange(ctx, tx, booking.LocationID, booking.DateFrom, booking.DateTo); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, `
		SELECT id, location_id, slot_date, total_capacity, available_capacity, is_closed, created_at, updated_at
		FROM booking_slots
		WHERE location_id = $1 AND slot_date >= $2 AND slot_date < $3
		ORDER BY slot_date ASC
		FOR UPDATE
	`, booking.LocationID, booking.DateFrom, booking.DateTo)
	if err != nil {
		return nil, fmt.Errorf("ошибка блокировки слотов: %w", err)
	}
	defer rows.Close()

	expectedNights := int(booking.DateTo.Sub(booking.DateFrom).Hours() / 24)
	var slotIDs []uuid.UUID
	for rows.Next() {
		slot, err := scanBookingSlot(rows)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения слота: %w", err)
		}
		if slot.IsClosed || slot.AvailableCapacity < booking.GuestsCount {
			return nil, ErrBookingUnavailable
		}
		slotIDs = append(slotIDs, slot.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка обхода слотов: %w", err)
	}
	rows.Close()
	if len(slotIDs) != expectedNights {
		return nil, ErrBookingUnavailable
	}

	for _, slotID := range slotIDs {
		_, err := tx.Exec(ctx, `
			UPDATE booking_slots
			SET available_capacity = available_capacity - $1, updated_at = NOW()
			WHERE id = $2
		`, booking.GuestsCount, slotID)
		if err != nil {
			return nil, fmt.Errorf("ошибка уменьшения capacity слота: %w", err)
		}
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO bookings (
			location_id, user_id, date_from, date_to, guests_count, total_price,
			status, contact_name, contact_phone, comment, host_comment
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, location_id, user_id, date_from, date_to, guests_count, total_price,
		          status, contact_name, contact_phone, comment, host_comment, cancelled_by,
		          confirmed_at, cancelled_at, created_at, updated_at
	`,
		booking.LocationID, booking.UserID, booking.DateFrom, booking.DateTo, booking.GuestsCount,
		booking.TotalPrice, booking.Status, booking.ContactName, booking.ContactPhone, booking.Comment,
		booking.HostComment,
	)

	created, err := scanBooking(row)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания брони: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("ошибка commit бронирования: %w", err)
	}

	return created, nil
}

func (r *BookingRepository) GetByID(ctx context.Context, bookingID uuid.UUID) (*models.BookingResponse, error) {
	row := r.pg.Pool.QueryRow(ctx, `
		SELECT
			b.id, b.location_id, b.user_id, b.date_from, b.date_to, b.guests_count, b.total_price,
			b.status, b.contact_name, b.contact_phone, b.comment, b.host_comment, b.cancelled_by,
			b.confirmed_at, b.cancelled_at, b.created_at, b.updated_at,
			l.name, l.slug, COALESCE(l.preview_image_url, '')
		FROM bookings b
		JOIN locations l ON l.id = b.location_id
		WHERE b.id = $1
	`, bookingID)

	booking, err := scanBookingResponse(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("ошибка получения брони: %w", err)
	}

	return booking, nil
}

func (r *BookingRepository) ConfirmOrReject(ctx context.Context, bookingID, hostID uuid.UUID, action models.BookingAction, hostComment string) (*models.BookingResponse, error) {
	tx, err := r.pg.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка начала транзакции обновления брони: %w", err)
	}
	defer tx.Rollback(ctx)

	var booking models.Booking
	var ownerID uuid.UUID
	row := tx.QueryRow(ctx, `
		SELECT
			b.id, b.location_id, b.user_id, b.date_from, b.date_to, b.guests_count, b.total_price,
			b.status, b.contact_name, b.contact_phone, b.comment, b.host_comment, b.cancelled_by,
			b.confirmed_at, b.cancelled_at, b.created_at, b.updated_at,
			l.owner_id
		FROM bookings b
		JOIN locations l ON l.id = b.location_id
		WHERE b.id = $1
		FOR UPDATE
	`, bookingID)
	err = row.Scan(
		&booking.ID, &booking.LocationID, &booking.UserID, &booking.DateFrom, &booking.DateTo, &booking.GuestsCount, &booking.TotalPrice,
		&booking.Status, &booking.ContactName, &booking.ContactPhone, &booking.Comment, &booking.HostComment, &booking.CancelledBy,
		&booking.ConfirmedAt, &booking.CancelledAt, &booking.CreatedAt, &booking.UpdatedAt, &ownerID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("ошибка получения брони для host action: %w", err)
	}

	if ownerID != hostID {
		return nil, ErrBookingForbidden
	}
	if booking.Status != models.BookingStatusPending {
		return nil, ErrBookingInvalidStatus
	}

	switch action {
	case models.BookingActionConfirm:
		_, err = tx.Exec(ctx, `
			UPDATE bookings
			SET status = $1, host_comment = $2, confirmed_at = NOW(), updated_at = NOW()
			WHERE id = $3
		`, models.BookingStatusConfirmed, hostComment, bookingID)
	case models.BookingActionReject:
		_, err = tx.Exec(ctx, `
			UPDATE bookings
			SET status = $1, host_comment = $2, cancelled_by = 'host', cancelled_at = NOW(), updated_at = NOW()
			WHERE id = $3
		`, models.BookingStatusRejected, hostComment, bookingID)
		if err == nil {
			_, err = tx.Exec(ctx, `
				UPDATE booking_slots
				SET available_capacity = LEAST(total_capacity, available_capacity + $1), updated_at = NOW()
				WHERE location_id = $2 AND slot_date >= $3 AND slot_date < $4
			`, booking.GuestsCount, booking.LocationID, booking.DateFrom, booking.DateTo)
		}
	default:
		return nil, ErrBookingInvalidStatus
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления статуса брони: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("ошибка commit host action: %w", err)
	}

	return r.GetByID(ctx, bookingID)
}

func (r *BookingRepository) CancelByUser(ctx context.Context, bookingID, userID uuid.UUID) (*models.BookingResponse, error) {
	tx, err := r.pg.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка начала транзакции отмены: %w", err)
	}
	defer tx.Rollback(ctx)

	var booking models.Booking
	row := tx.QueryRow(ctx, `
		SELECT id, location_id, user_id, date_from, date_to, guests_count, total_price,
		       status, contact_name, contact_phone, comment, host_comment, cancelled_by,
		       confirmed_at, cancelled_at, created_at, updated_at
		FROM bookings
		WHERE id = $1
		FOR UPDATE
	`, bookingID)
	err = row.Scan(
		&booking.ID, &booking.LocationID, &booking.UserID, &booking.DateFrom, &booking.DateTo, &booking.GuestsCount, &booking.TotalPrice,
		&booking.Status, &booking.ContactName, &booking.ContactPhone, &booking.Comment, &booking.HostComment, &booking.CancelledBy,
		&booking.ConfirmedAt, &booking.CancelledAt, &booking.CreatedAt, &booking.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("ошибка получения брони для отмены: %w", err)
	}
	if booking.UserID != userID {
		return nil, ErrBookingForbidden
	}
	if booking.Status != models.BookingStatusPending && booking.Status != models.BookingStatusConfirmed {
		return nil, ErrBookingInvalidStatus
	}

	_, err = tx.Exec(ctx, `
		UPDATE bookings
		SET status = $1, cancelled_by = 'tourist', cancelled_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, models.BookingStatusCancelled, bookingID)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления статуса отмены: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE booking_slots
		SET available_capacity = LEAST(total_capacity, available_capacity + $1), updated_at = NOW()
		WHERE location_id = $2 AND slot_date >= $3 AND slot_date < $4
	`, booking.GuestsCount, booking.LocationID, booking.DateFrom, booking.DateTo)
	if err != nil {
		return nil, fmt.Errorf("ошибка возврата capacity слотов: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("ошибка commit отмены: %w", err)
	}

	return r.GetByID(ctx, bookingID)
}

func (r *BookingRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.BookingResponse, error) {
	rows, err := r.pg.Pool.Query(ctx, `
		SELECT
			b.id, b.location_id, b.user_id, b.date_from, b.date_to, b.guests_count, b.total_price,
			b.status, b.contact_name, b.contact_phone, b.comment, b.host_comment, b.cancelled_by,
			b.confirmed_at, b.cancelled_at, b.created_at, b.updated_at,
			l.name, l.slug, COALESCE(l.preview_image_url, '')
		FROM bookings b
		JOIN locations l ON l.id = b.location_id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка списка броней пользователя: %w", err)
	}
	defer rows.Close()

	var bookings []models.BookingResponse
	for rows.Next() {
		booking, err := scanBookingResponse(rows)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования брони пользователя: %w", err)
		}
		bookings = append(bookings, *booking)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации броней пользователя: %w", err)
	}
	return bookings, nil
}

func (r *BookingRepository) ListByHostID(ctx context.Context, hostID uuid.UUID) ([]models.BookingResponse, error) {
	rows, err := r.pg.Pool.Query(ctx, `
		SELECT
			b.id, b.location_id, b.user_id, b.date_from, b.date_to, b.guests_count, b.total_price,
			b.status, b.contact_name, b.contact_phone, b.comment, b.host_comment, b.cancelled_by,
			b.confirmed_at, b.cancelled_at, b.created_at, b.updated_at,
			l.name, l.slug, COALESCE(l.preview_image_url, '')
		FROM bookings b
		JOIN locations l ON l.id = b.location_id
		WHERE l.owner_id = $1
		ORDER BY b.created_at DESC
	`, hostID)
	if err != nil {
		return nil, fmt.Errorf("ошибка списка броней хоста: %w", err)
	}
	defer rows.Close()

	var bookings []models.BookingResponse
	for rows.Next() {
		booking, err := scanBookingResponse(rows)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования брони хоста: %w", err)
		}
		bookings = append(bookings, *booking)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации броней хоста: %w", err)
	}
	return bookings, nil
}
