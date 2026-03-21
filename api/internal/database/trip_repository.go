// Файл trip_repository.go реализует слой доступа к данным для таблиц trips и trip_members.
// Инкапсулирует SQL-запросы: создание поездок и участников, поиск по ID и invite-токену,
// обновление токенов, vibe-векторов и частичное обновление поездок, подсчёт участников.
package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

// tripColumns — список столбцов таблицы trips для SELECT-запросов.
const tripColumns = `id, creator_id, date_from, date_to, budget_rub, budget_tier, transport,
	group_size, group_composition, format, invite_token, vibe_vector_id,
	merged_vibe_vector_id, status, created_at, updated_at`

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

// scanTrip сканирует строку результата в структуру models.Trip.
// Используется для избежания дублирования Scan-полей в разных методах.
func scanTrip(row pgx.Row) (*models.Trip, error) {
	var trip models.Trip
	err := row.Scan(
		&trip.ID,
		&trip.CreatorID,
		&trip.DateFrom,
		&trip.DateTo,
		&trip.BudgetRub,
		&trip.BudgetTier,
		&trip.Transport,
		&trip.GroupSize,
		&trip.GroupComposition,
		&trip.Format,
		&trip.InviteToken,
		&trip.VibeVectorID,
		&trip.MergedVibeVectorID,
		&trip.Status,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

// Create создаёт новую поездку в таблице trips.
// Возвращает созданную поездку со всеми полями, включая сгенерированные ID и invite_token.
func (r *TripRepository) Create(ctx context.Context, trip *models.Trip) (*models.Trip, error) {
	query := `
		INSERT INTO trips (creator_id, date_from, date_to, budget_rub, budget_tier, transport,
		                   group_size, group_composition, format, vibe_vector_id, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING ` + tripColumns

	result, err := scanTrip(r.pg.Pool.QueryRow(ctx, query,
		trip.CreatorID,
		trip.DateFrom,
		trip.DateTo,
		trip.BudgetRub,
		trip.BudgetTier,
		trip.Transport,
		trip.GroupSize,
		trip.GroupComposition,
		trip.Format,
		trip.VibeVectorID,
		trip.Status,
	))
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
	return result, nil
}

// FindByID находит поездку по её UUID.
// Возвращает ErrTripNotFound если поездка не существует.
func (r *TripRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
	query := `SELECT ` + tripColumns + ` FROM trips WHERE id = $1`

	result, err := scanTrip(r.pg.Pool.QueryRow(ctx, query, id))
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

	return result, nil
}

// FindByInviteToken находит поездку по её invite-токену.
// Используется при присоединении участника по ссылке-приглашению.
// Возвращает ErrTripNotFound если поездка не существует.
func (r *TripRepository) FindByInviteToken(ctx context.Context, token uuid.UUID) (*models.Trip, error) {
	query := `SELECT ` + tripColumns + ` FROM trips WHERE invite_token = $1`

	result, err := scanTrip(r.pg.Pool.QueryRow(ctx, query, token))
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

	return result, nil
}

// Update частично обновляет поездку по переданным полям.
// Принимает map[string]any с именами столбцов и значениями.
// Возвращает обновлённую поездку или ErrTripNotFound.
func (r *TripRepository) Update(ctx context.Context, tripID uuid.UUID, fields map[string]any) (*models.Trip, error) {
	if len(fields) == 0 {
		return r.FindByID(ctx, tripID)
	}

	setParts := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields)+1)
	paramIdx := 1

	for col, val := range fields {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", col, paramIdx))
		args = append(args, val)
		paramIdx++
	}

	args = append(args, tripID)
	query := fmt.Sprintf(`UPDATE trips SET %s WHERE id = $%d RETURNING %s`,
		strings.Join(setParts, ", "), paramIdx, tripColumns)

	result, err := scanTrip(r.pg.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTripNotFound
		}
		r.logger.Error("ошибка обновления поездки",
			zap.String("trip_id", tripID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка обновления поездки: %w", err)
	}

	r.logger.Info("поездка обновлена",
		zap.String("trip_id", result.ID.String()),
		zap.Int("fields_updated", len(fields)),
	)
	return result, nil
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

// FindByUserID возвращает все поездки, в которых пользователь является участником.
// Выполняет JOIN trips с trip_members по user_id.
// Возвращает поездки, отсортированные по дате создания (новые первые).
func (r *TripRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Trip, error) {
	query := `
		SELECT t.id, t.creator_id, t.date_from, t.date_to, t.budget_rub, t.budget_tier,
		       t.transport, t.group_size, t.group_composition, t.format, t.invite_token,
		       t.vibe_vector_id, t.merged_vibe_vector_id, t.status, t.created_at, t.updated_at
		FROM trips t
		INNER JOIN trip_members tm ON t.id = tm.trip_id
		WHERE tm.user_id = $1
		ORDER BY t.created_at DESC
	`

	rows, err := r.pg.Pool.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("ошибка получения поездок пользователя",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка получения поездок пользователя: %w", err)
	}
	defer rows.Close()

	var trips []models.Trip
	for rows.Next() {
		var t models.Trip
		if err := rows.Scan(
			&t.ID,
			&t.CreatorID,
			&t.DateFrom,
			&t.DateTo,
			&t.BudgetRub,
			&t.BudgetTier,
			&t.Transport,
			&t.GroupSize,
			&t.GroupComposition,
			&t.Format,
			&t.InviteToken,
			&t.VibeVectorID,
			&t.MergedVibeVectorID,
			&t.Status,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования поездки: %w", err)
		}
		trips = append(trips, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации поездок: %w", err)
	}

	return trips, nil
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
