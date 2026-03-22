package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
)

var (
	ErrReviewNotFound = errors.New("отзыв не найден")
)

type reviewQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type ReviewRepository struct {
	pg     *PostgresClient
	logger *zap.Logger
}

func NewReviewRepository(pg *PostgresClient, logger *zap.Logger) *ReviewRepository {
	return &ReviewRepository{
		pg:     pg,
		logger: logger.Named("review_repository"),
	}
}

func scanReviewResponse(row pgx.Row) (*models.ReviewResponse, error) {
	var review models.ReviewResponse
	err := row.Scan(
		&review.ID,
		&review.AuthorID,
		&review.LocationID,
		&review.BookingID,
		&review.Rating,
		&review.Text,
		&review.IsFromHost,
		&review.KarmaDelta,
		&review.CreatedAt,
		&review.UpdatedAt,
		&review.AuthorName,
		&review.AuthorRole,
	)
	return &review, err
}

func (r *ReviewRepository) Create(ctx context.Context, db pgx.Tx, review *models.Review) error {
	query := `
		INSERT INTO reviews (
			author_id, location_id, booking_id, rating, text, is_from_host, karma_delta
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id, created_at, updated_at
	`
	
	var runner reviewQuerier = r.pg.Pool
	if db != nil {
		runner = db
	}

	err := runner.QueryRow(
		ctx, query,
		review.AuthorID,
		review.LocationID,
		review.BookingID,
		review.Rating,
		review.Text,
		review.IsFromHost,
		review.KarmaDelta,
	).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create review", zap.Error(err))
		return err
	}
	return nil
}

func (r *ReviewRepository) GetByLocationID(ctx context.Context, locationID uuid.UUID, limit, offset int) ([]models.ReviewResponse, error) {
	query := `
		SELECT 
			r.id, r.author_id, r.location_id, r.booking_id, 
			r.rating, r.text, r.is_from_host, r.karma_delta, 
			r.created_at, r.updated_at,
			COALESCE(u.display_name, 'Аноним'), u.role
		FROM reviews r
		JOIN users u ON u.id = r.author_id
		WHERE r.location_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pg.Pool.Query(ctx, query, locationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.ReviewResponse
	for rows.Next() {
		rev, err := scanReviewResponse(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, *rev)
	}

	return reviews, rows.Err()
}

func (r *ReviewRepository) GetByAuthorID(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]models.ReviewResponse, error) {
	query := `
		SELECT 
			r.id, r.author_id, r.location_id, r.booking_id, 
			r.rating, r.text, r.is_from_host, r.karma_delta, 
			r.created_at, r.updated_at,
			COALESCE(u.display_name, 'Аноним'), u.role
		FROM reviews r
		JOIN users u ON u.id = r.author_id
		WHERE r.author_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pg.Pool.Query(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.ReviewResponse
	for rows.Next() {
		rev, err := scanReviewResponse(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, *rev)
	}

	return reviews, rows.Err()
}
