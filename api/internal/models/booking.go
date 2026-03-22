package models

import (
	"time"

	"github.com/google/uuid"
)

// BookingStatus — статус бронирования.
type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusRejected  BookingStatus = "rejected"
	BookingStatusCancelled BookingStatus = "cancelled"
)

// BookingAction — действие хоста над бронью.
type BookingAction string

const (
	BookingActionConfirm BookingAction = "confirm"
	BookingActionReject  BookingAction = "reject"
)

// BookingSlot — дневной слот доступности локации.
type BookingSlot struct {
	ID                uuid.UUID `json:"id" db:"id"`
	LocationID        uuid.UUID `json:"location_id" db:"location_id"`
	SlotDate          time.Time `json:"slot_date" db:"slot_date"`
	TotalCapacity     int       `json:"total_capacity" db:"total_capacity"`
	AvailableCapacity int       `json:"available_capacity" db:"available_capacity"`
	IsClosed          bool      `json:"is_closed" db:"is_closed"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Booking — запись бронирования.
type Booking struct {
	ID           uuid.UUID      `json:"id" db:"id"`
	LocationID   uuid.UUID      `json:"location_id" db:"location_id"`
	UserID       uuid.UUID      `json:"user_id" db:"user_id"`
	DateFrom     time.Time      `json:"date_from" db:"date_from"`
	DateTo       time.Time      `json:"date_to" db:"date_to"`
	GuestsCount  int            `json:"guests_count" db:"guests_count"`
	TotalPrice   int            `json:"total_price" db:"total_price"`
	Status       BookingStatus  `json:"status" db:"status"`
	ContactName  string         `json:"contact_name" db:"contact_name"`
	ContactPhone string         `json:"contact_phone" db:"contact_phone"`
	Comment      string         `json:"comment" db:"comment"`
	HostComment  string         `json:"host_comment" db:"host_comment"`
	CancelledBy  *string        `json:"cancelled_by,omitempty" db:"cancelled_by"`
	ConfirmedAt  *time.Time     `json:"confirmed_at,omitempty" db:"confirmed_at"`
	CancelledAt  *time.Time     `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// BookingResponse — бронирование с данными локации для UI.
type BookingResponse struct {
	Booking
	LocationName         string `json:"location_name"`
	LocationSlug         string `json:"location_slug"`
	LocationPreviewImage string `json:"location_preview_image_url"`
}

// CreateBookingRequest — создание брони.
type CreateBookingRequest struct {
	LocationID   string `json:"location_id"`
	DateFrom     string `json:"date_from"`
	DateTo       string `json:"date_to"`
	GuestsCount  int    `json:"guests_count"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	Comment      string `json:"comment"`
}

// Validate проверяет тело запроса на создание брони.
func (r *CreateBookingRequest) Validate() string {
	if r.LocationID == "" {
		return "location_id обязателен"
	}
	if _, err := uuid.Parse(r.LocationID); err != nil {
		return "location_id должен быть валидным UUID"
	}
	if r.DateFrom == "" {
		return "date_from обязателен"
	}
	if r.DateTo == "" {
		return "date_to обязателен"
	}
	dateFrom, err := time.Parse(DateLayout, r.DateFrom)
	if err != nil {
		return "некорректный формат date_from (ожидается YYYY-MM-DD)"
	}
	dateTo, err := time.Parse(DateLayout, r.DateTo)
	if err != nil {
		return "некорректный формат date_to (ожидается YYYY-MM-DD)"
	}
	if !dateTo.After(dateFrom) {
		return "date_to должна быть позже date_from"
	}
	if r.GuestsCount <= 0 {
		return "guests_count должен быть не менее 1"
	}
	if r.ContactName == "" {
		return "contact_name обязателен"
	}
	if r.ContactPhone == "" {
		return "contact_phone обязателен"
	}
	return ""
}

// ParseDates парсит даты бронирования после Validate.
func (r *CreateBookingRequest) ParseDates() (time.Time, time.Time) {
	dateFrom, _ := time.Parse(DateLayout, r.DateFrom)
	dateTo, _ := time.Parse(DateLayout, r.DateTo)
	return dateFrom, dateTo
}

// BookingActionRequest — действие хоста: confirm/reject.
type BookingActionRequest struct {
	Action      string `json:"action"`
	HostComment string `json:"host_comment"`
}

// Validate проверяет действие хоста.
func (r *BookingActionRequest) Validate() string {
	if r.Action != string(BookingActionConfirm) && r.Action != string(BookingActionReject) {
		return "action должна быть confirm или reject"
	}
	return ""
}

// LocationSlotsFilter — параметры получения слотов.
type LocationSlotsFilter struct {
	Month string `query:"month"`
}

// NormalizeDefaultMonth подставляет текущий месяц, если параметр не передан.
func (f *LocationSlotsFilter) NormalizeDefaultMonth(now time.Time) {
	if f.Month == "" {
		f.Month = now.Format("2006-01")
	}
}

// Validate проверяет формат month=YYYY-MM.
func (f *LocationSlotsFilter) Validate() string {
	if f.Month == "" {
		return "month обязателен"
	}
	if _, err := time.Parse("2006-01", f.Month); err != nil {
		return "некорректный формат month (ожидается YYYY-MM)"
	}
	return ""
}

// ParseMonth возвращает первый день месяца.
func (f *LocationSlotsFilter) ParseMonth() time.Time {
	monthStart, _ := time.Parse("2006-01", f.Month)
	return monthStart
}
