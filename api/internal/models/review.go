package models

import (
	"time"

	"github.com/google/uuid"
)

// Review — запись отзыва.
type Review struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	AuthorID    uuid.UUID  `json:"author_id" db:"author_id"`
	LocationID  *uuid.UUID `json:"location_id,omitempty" db:"location_id"`
	BookingID   *uuid.UUID `json:"booking_id,omitempty" db:"booking_id"`
	Rating      int        `json:"rating" db:"rating"`
	Text        string     `json:"text" db:"text"`
	IsFromHost  bool       `json:"is_from_host" db:"is_from_host"`
	KarmaDelta  int        `json:"karma_delta" db:"karma_delta"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// ReviewResponse — расширенная модель отзыва с данными об авторе для UI.
type ReviewResponse struct {
	Review
	AuthorName  string `json:"author_name"`
	AuthorRole  string `json:"author_role"`
}

// CreateReviewRequest — DTO для создания отзыва.
type CreateReviewRequest struct {
	LocationID *string `json:"location_id,omitempty"`
	BookingID  *string `json:"booking_id,omitempty"`
	Rating     int     `json:"rating"`
	Text       string  `json:"text"`
	IsFromHost bool    `json:"is_from_host"`
}

// Validate проверяет тело запроса на создание отзыва.
func (r *CreateReviewRequest) Validate() string {
	if r.Rating < 1 || r.Rating > 5 {
		return "rating должен быть от 1 до 5"
	}
	if r.LocationID == nil && r.BookingID == nil {
		return "хотя бы location_id или booking_id должны быть указаны"
	}
	if r.LocationID != nil && *r.LocationID != "" {
		if _, err := uuid.Parse(*r.LocationID); err != nil {
			return "location_id должен быть валидным UUID"
		}
	}
	if r.BookingID != nil && *r.BookingID != "" {
		if _, err := uuid.Parse(*r.BookingID); err != nil {
			return "booking_id должен быть валидным UUID"
		}
	}
	if len(r.Text) > 2000 {
		return "text не должен превышать 2000 символов"
	}
	return ""
}
