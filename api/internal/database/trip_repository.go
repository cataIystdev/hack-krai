// Файл trip_repository.go реализует слой доступа к данным для таблиц trips и trip_members.
// Инкапсулирует SQL-запросы: создание поездок и участников, поиск по ID и invite-токену,
// обновление токенов и vibe-векторов, подсчёт участников.
package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
)

// Ошибки репозитория поездок.
var (
	// ErrTripNotFound — поездка не найдена по указанному ID.
	ErrTripNotFound = errors.New("поездка не найдена")

	// ErrMemberAlreadyExists — пользователь уже является участником поездки.
	ErrMemberAlreadyExists = errors.New("пользователь уже является участником поездки")

	// ErrTripFull — достигнут лимит участников поездки (group_size).
	ErrTripFull = errors.New("достигнут лимит участников поездки")
)

// TripRepository — репозиторий для работы с таблицами trips и trip_members.
type TripRepository struct {
	pg     *PostgresClient
	logger *zap.Logger
}

// NewTripRepository создаёт новый репозиторий поездок.
func NewTripRepository(pg *PostgresClient, logger *zap.Logger) *TripRepository {
	return &TripRepository{
		pg:     pg,
		logger: logger.Named("trip_repository"),
	}
}

// Create создаёт новую поездку в таблице trips.
// Возвращает созданную поездку со всеми полями, включая сгенерированные ID и invite_token.
func (r *TripRepository) Create(ctx context.Context, trip *models.Trip) (*models.Trip, error) {
	query := `
		INSERT INTO trips (creator_id, date_from, date_to, budget_rub, budget_tier, transport, group_size, group_composition, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, creator_id, date_from, date_to, budget_rub, budget_tier, transport,
		          group_size, group_composition, invite_token, merged_vibe_vector_id, status,
		          created_at, updated_at
	`

	var result models.Trip
	err := r.pg.Pool.QueryRow(ctx, query,
		trip.CreatorID,
		trip.DateFrom,
		trip.DateTo,
		trip.BudgetRub,
		trip.BudgetTier,
		trip.Transport,
		trip.GroupSize,
		trip.GroupComposition,
		trip.Status,
	).Scan(
		&result.ID,
		&result.CreatorID,
		&result.DateFrom,
		&result.DateTo,
		&result.BudgetRub,
		&result.BudgetTier,
		&result.Transport,
		&result.GroupSize,
		&result.GroupComposition,
		&result.InviteToken,
		&result.MergedVibeVectorID,
		&result.Status,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("ошибка создания поездки",
			zap.String("creator_id", trip.CreatorID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка создания поездки: %w", err)
	}

	r.logger.Info("поездка создана",
		zap.String("trip_id", result.ID.String()),
		zap.String("creator_id", result.CreatorID.String()),
	)
	return &result, nil
}

// FindByID находит поездку по её UUID.
// Возвращает ErrTripNotFound если поездка не существует.
func (r *TripRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
	query := `
		SELECT id, creator_id, date_from, date_to, budget_rub, budget_tier, transport,
		       group_size, group_composition, invite_token, merged_vibe_vector_id, status,
		       created_at, updated_at
		FROM trips WHERE id = $1
	`

	var trip models.Trip
	err := r.pg.Pool.QueryRow(ctx, query, id).Scan(
		&trip.ID,
		&trip.CreatorID,
		&trip.DateFrom,
		&trip.DateTo,
		&trip.BudgetRub,
		&trip.BudgetTier,
		&trip.Transport,
		&trip.GroupSize,
		&trip.GroupComposition,
		&trip.InviteToken,
		&trip.MergedVibeVectorID,
		&trip.Status,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTripNotFound
		}
		r.logger.Error("ошибка поиска поездки",
			zap.String("trip_id", id.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка поиска поездки: %w", err)
	}

	return &trip, nil
}

// FindByInviteToken находит поездку по её invite-токену.
// Используется при присоединении участника по ссылке-приглашению.
// Возвращает ErrTripNotFound если поездка не существует.
func (r *TripRepository) FindByInviteToken(ctx context.Context, token uuid.UUID) (*models.Trip, error) {
	query := `
		SELECT id, creator_id, date_from, date_to, budget_rub, budget_tier, transport,
		       group_size, group_composition, invite_token, merged_vibe_vector_id, status,
		       created_at, updated_at
		FROM trips WHERE invite_token = $1
	`

	var trip models.Trip
	err := r.pg.Pool.QueryRow(ctx, query, token).Scan(
		&trip.ID,
		&trip.CreatorID,
		&trip.DateFrom,
		&trip.DateTo,
		&trip.BudgetRub,
		&trip.BudgetTier,
		&trip.Transport,
		&trip.GroupSize,
		&trip.GroupComposition,
		&trip.InviteToken,
		&trip.MergedVibeVectorID,
		&trip.Status,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTripNotFound
		}
		r.logger.Error("ошибка поиска поездки по invite_token",
			zap.String("invite_token", token.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка поиска поездки по invite_token: %w", err)
	}

	return &trip, nil
}

// UpdateInviteToken обновляет invite-токен поездки на новый UUID.
// Используется для перегенерации ссылки-приглашения.
// Возвращает ErrTripNotFound если поездка не существует.
func (r *TripRepository) UpdateInviteToken(ctx context.Context, tripID uuid.UUID, newToken uuid.UUID) error {
	query := `UPDATE trips SET invite_token = $1 WHERE id = $2`

	result, err := r.pg.Pool.Exec(ctx, query, newToken, tripID)
	if err != nil {
		r.logger.Error("ошибка обновления invite_token",
			zap.String("trip_id", tripID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("ошибка обновления invite_token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrTripNotFound
	}

	r.logger.Info("invite_token обновлён",
		zap.String("trip_id", tripID.String()),
		zap.String("new_token", newToken.String()),
	)
	return nil
}

// UpdateMergedVibeVector обновляет merged_vibe_vector_id поездки.
// Вызывается после пересчёта средневзвешенного вектора группы.
func (r *TripRepository) UpdateMergedVibeVector(ctx context.Context, tripID uuid.UUID, vectorID uuid.UUID) error {
	query := `UPDATE trips SET merged_vibe_vector_id = $1 WHERE id = $2`

	result, err := r.pg.Pool.Exec(ctx, query, vectorID, tripID)
	if err != nil {
		r.logger.Error("ошибка обновления merged_vibe_vector_id",
			zap.String("trip_id", tripID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("ошибка обновления merged_vibe_vector_id: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrTripNotFound
	}

	return nil
}

// AddMember добавляет нового участника в поездку.
// Возвращает ErrMemberAlreadyExists при попытке повторного присоединения
// авторизованного пользователя (нарушение UNIQUE constraint на trip_id + user_id).
func (r *TripRepository) AddMember(ctx context.Context, member *models.TripMember) (*models.TripMember, error) {
	query := `
		INSERT INTO trip_members (trip_id, user_id, display_name, role, vibe_vector_id, tags, is_child)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, trip_id, user_id, display_name, role, vibe_vector_id, tags, is_child, joined_at
	`

	var result models.TripMember
	err := r.pg.Pool.QueryRow(ctx, query,
		member.TripID,
		member.UserID,
		member.DisplayName,
		member.Role,
		member.VibeVectorID,
		member.Tags,
		member.IsChild,
	).Scan(
		&result.ID,
		&result.TripID,
		&result.UserID,
		&result.DisplayName,
		&result.Role,
		&result.VibeVectorID,
		&result.Tags,
		&result.IsChild,
		&result.JoinedAt,
	)
	if err != nil {
		if contains(err.Error(), "duplicate key") || contains(err.Error(), "idx_trip_members_unique_user") {
			return nil, ErrMemberAlreadyExists
		}
		r.logger.Error("ошибка добавления участника",
			zap.String("trip_id", member.TripID.String()),
			zap.String("display_name", member.DisplayName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка добавления участника: %w", err)
	}

	r.logger.Info("участник добавлен в поездку",
		zap.String("trip_id", result.TripID.String()),
		zap.String("member_id", result.ID.String()),
		zap.String("display_name", result.DisplayName),
	)
	return &result, nil
}

// FindMembersByTripID возвращает всех участников поездки, отсортированных по дате присоединения.
func (r *TripRepository) FindMembersByTripID(ctx context.Context, tripID uuid.UUID) ([]models.TripMember, error) {
	query := `
		SELECT id, trip_id, user_id, display_name, role, vibe_vector_id, tags, is_child, joined_at
		FROM trip_members
		WHERE trip_id = $1
		ORDER BY joined_at ASC
	`

	rows, err := r.pg.Pool.Query(ctx, query, tripID)
	if err != nil {
		r.logger.Error("ошибка получения участников поездки",
			zap.String("trip_id", tripID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка получения участников: %w", err)
	}
	defer rows.Close()

	var members []models.TripMember
	for rows.Next() {
		var m models.TripMember
		if err := rows.Scan(
			&m.ID,
			&m.TripID,
			&m.UserID,
			&m.DisplayName,
			&m.Role,
			&m.VibeVectorID,
			&m.Tags,
			&m.IsChild,
			&m.JoinedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования участника: %w", err)
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации участников: %w", err)
	}

	return members, nil
}

// CountMembers возвращает количество участников поездки.
// Используется для проверки ограничения group_size при присоединении.
func (r *TripRepository) CountMembers(ctx context.Context, tripID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM trip_members WHERE trip_id = $1`

	var count int
	err := r.pg.Pool.QueryRow(ctx, query, tripID).Scan(&count)
	if err != nil {
		r.logger.Error("ошибка подсчёта участников",
			zap.String("trip_id", tripID.String()),
			zap.Error(err),
		)
		return 0, fmt.Errorf("ошибка подсчёта участников: %w", err)
	}

	return count, nil
}

// AddMemberAtomic атомарно добавляет участника в поездку с проверкой лимита group_size.
// Использует INSERT ... SELECT WHERE (SELECT count(*) ...) < maxGroupSize,
// что гарантирует атомарность проверки и вставки на уровне одного SQL-запроса.
// Возвращает ErrTripFull если лимит уже достигнут (0 rows affected).
// Возвращает ErrMemberAlreadyExists при нарушении UNIQUE constraint (trip_id + user_id).
func (r *TripRepository) AddMemberAtomic(ctx context.Context, member *models.TripMember, maxGroupSize int) (*models.TripMember, error) {
	query := `
		INSERT INTO trip_members (trip_id, user_id, display_name, role, vibe_vector_id, tags, is_child)
		SELECT $1, $2, $3, $4, $5, $6, $7
		WHERE (SELECT COUNT(*) FROM trip_members WHERE trip_id = $1) < $8
		RETURNING id, trip_id, user_id, display_name, role, vibe_vector_id, tags, is_child, joined_at
	`

	var result models.TripMember
	err := r.pg.Pool.QueryRow(ctx, query,
		member.TripID,
		member.UserID,
		member.DisplayName,
		member.Role,
		member.VibeVectorID,
		member.Tags,
		member.IsChild,
		maxGroupSize,
	).Scan(
		&result.ID,
		&result.TripID,
		&result.UserID,
		&result.DisplayName,
		&result.Role,
		&result.VibeVectorID,
		&result.Tags,
		&result.IsChild,
		&result.JoinedAt,
	)
	if err != nil {
		// Проверка нарушения UNIQUE constraint (повторное присоединение).
		if contains(err.Error(), "duplicate key") || contains(err.Error(), "idx_trip_members_unique_user") {
			return nil, ErrMemberAlreadyExists
		}
		// Проверка отсутствия вставленной строки (лимит достигнут).
		if err.Error() == "no rows in result set" {
			return nil, ErrTripFull
		}
		r.logger.Error("ошибка атомарного добавления участника",
			zap.String("trip_id", member.TripID.String()),
			zap.String("display_name", member.DisplayName),
			zap.Int("max_group_size", maxGroupSize),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка атомарного добавления участника: %w", err)
	}

	r.logger.Info("участник атомарно добавлен в поездку",
		zap.String("trip_id", result.TripID.String()),
		zap.String("member_id", result.ID.String()),
		zap.String("display_name", result.DisplayName),
	)
	return &result, nil
}
