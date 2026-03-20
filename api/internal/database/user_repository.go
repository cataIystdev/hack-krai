// Файл user_repository.go реализует репозиторий для работы с таблицей users.
// Предоставляет методы создания, поиска по email и поиска по ID.
// Все запросы выполняются через pgxpool (PostgreSQL connection pool).
package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
)

// ErrUserNotFound — ошибка, возвращаемая при отсутствии пользователя в БД.
var ErrUserNotFound = errors.New("пользователь не найден")

// ErrUserAlreadyExists — ошибка, возвращаемая при попытке создания пользователя
// с email, который уже зарегистрирован.
var ErrUserAlreadyExists = errors.New("пользователь с таким email уже существует")

// UserRepository — репозиторий для работы с таблицей users в PostgreSQL.
// Инкапсулирует SQL-запросы и маппинг результатов на Go-структуры.
type UserRepository struct {
	pg     *PostgresClient
	logger *zap.Logger
}

// NewUserRepository создаёт экземпляр репозитория пользователей.
// Принимает клиент PostgreSQL и логгер.
func NewUserRepository(pg *PostgresClient, logger *zap.Logger) *UserRepository {
	return &UserRepository{
		pg:     pg,
		logger: logger,
	}
}

// Create создаёт нового пользователя в БД.
// Вставляет запись с email, password_hash, role и display_name.
// Возвращает заполненную структуру User с данными из БД (включая сгенерированный UUID).
// При конфликте по email возвращает ErrUserAlreadyExists.
func (r *UserRepository) Create(ctx context.Context, email, passwordHash, displayName, role string) (*models.User, error) {
	query := `
		INSERT INTO users (email, password_hash, role, display_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, password_hash, role, display_name, karma, vibe_vector_id, created_at, updated_at
	`

	var user models.User
	err := r.pg.Pool.QueryRow(ctx, query, email, passwordHash, role, displayName).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.DisplayName,
		&user.Karma,
		&user.VibeVectorID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		// Проверка нарушения уникальности email (код 23505).
		if err.Error() != "" && contains(err.Error(), "duplicate key") {
			r.logger.Debug("попытка создания пользователя с существующим email",
				zap.String("email", email),
			)
			return nil, ErrUserAlreadyExists
		}
		r.logger.Error("ошибка создания пользователя",
			zap.String("email", email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	r.logger.Info("пользователь создан",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
		zap.String("role", string(user.Role)),
	)

	return &user, nil
}

// FindByEmail ищет пользователя по адресу электронной почты.
// Возвращает ErrUserNotFound, если пользователь не найден.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, display_name, karma, vibe_vector_id, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.pg.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.DisplayName,
		&user.Karma,
		&user.VibeVectorID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.logger.Error("ошибка поиска пользователя по email",
			zap.String("email", email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	return &user, nil
}

// FindByID ищет пользователя по UUID.
// Возвращает ErrUserNotFound, если пользователь не найден.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, display_name, karma, vibe_vector_id, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.pg.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.DisplayName,
		&user.Karma,
		&user.VibeVectorID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.logger.Error("ошибка поиска пользователя по ID",
			zap.String("id", id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	return &user, nil
}

// UpdateProfile обновляет display_name пользователя.
// Возвращает обновлённую структуру User.
func (r *UserRepository) UpdateProfile(ctx context.Context, id, displayName string) (*models.User, error) {
	query := `
		UPDATE users SET display_name = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, email, password_hash, role, display_name, karma, vibe_vector_id, created_at, updated_at
	`

	var user models.User
	err := r.pg.Pool.QueryRow(ctx, query, displayName, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.DisplayName,
		&user.Karma,
		&user.VibeVectorID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.logger.Error("ошибка обновления профиля",
			zap.String("id", id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка обновления профиля: %w", err)
	}

	r.logger.Info("профиль обновлён",
		zap.String("user_id", user.ID.String()),
		zap.String("display_name", user.DisplayName),
	)

	return &user, nil
}

// contains — вспомогательная функция для проверки наличия подстроки.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

// searchString — наивный поиск подстроки.
func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
