// Пакет models содержит структуры данных, соответствующие таблицам БД.
// Структуры используются для маппинга результатов SQL-запросов
// и формирования ответов API.
package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole — тип перечисления ролей пользователя.
// Соответствует PostgreSQL типу user_role.
type UserRole string

const (
	// RoleTourist — обычный пользователь (турист). Роль по умолчанию.
	RoleTourist UserRole = "tourist"

	// RoleHost — хозяин локации (фермер, владелец агроусадьбы).
	RoleHost UserRole = "host"

	// RoleB2GAdmin — администратор B2G-платформы.
	RoleB2GAdmin UserRole = "b2g_admin"
)

// ValidRoles — набор допустимых ролей для валидации.
var ValidRoles = map[UserRole]bool{
	RoleTourist:  true,
	RoleHost:     true,
	RoleB2GAdmin: true,
}

// IsValidRole проверяет, является ли переданная строка допустимой ролью.
func IsValidRole(role string) bool {
	return ValidRoles[UserRole(role)]
}

// User — структура пользователя, соответствующая таблице users в PostgreSQL.
// Содержит все поля записи, включая хэш пароля (не возвращается клиенту).
type User struct {
	// ID — уникальный идентификатор пользователя (UUID v4).
	ID uuid.UUID `json:"id" db:"id"`

	// Email — адрес электронной почты для аутентификации.
	Email string `json:"email" db:"email"`

	// PasswordHash — bcrypt-хэш пароля. Не сериализуется в JSON.
	PasswordHash string `json:"-" db:"password_hash"`

	// Role — роль пользователя в системе.
	Role UserRole `json:"role" db:"role"`

	// DisplayName — отображаемое имя пользователя.
	DisplayName string `json:"display_name" db:"display_name"`

	// Karma — очки кармы пользователя.
	Karma int `json:"karma" db:"karma"`

	// VibeVectorID — идентификатор вектора vibe-профиля в Qdrant.
	// NULL до прохождения onboarding-опроса.
	VibeVectorID *uuid.UUID `json:"vibe_vector_id,omitempty" db:"vibe_vector_id"`

	// CreatedAt — временная метка создания записи.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt — временная метка последнего обновления.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserPublicResponse — публичное представление пользователя без чувствительных данных.
// Используется в ответах API для исключения хэша пароля.
type UserPublicResponse struct {
	// ID — уникальный идентификатор пользователя.
	ID uuid.UUID `json:"id"`

	// Email — адрес электронной почты.
	Email string `json:"email"`

	// Role — роль пользователя.
	Role UserRole `json:"role"`

	// DisplayName — отображаемое имя.
	DisplayName string `json:"display_name"`

	// Karma — очки кармы.
	Karma int `json:"karma"`

	// VibeVectorID — ID вектора vibe-профиля.
	VibeVectorID *uuid.UUID `json:"vibe_vector_id,omitempty"`

	// CreatedAt — дата создания записи.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt — дата обновления записи.
	UpdatedAt time.Time `json:"updated_at"`
}

// ToPublicResponse формирует публичное представление пользователя.
// Исключает хэш пароля и другие чувствительные данные.
func (u *User) ToPublicResponse() *UserPublicResponse {
	return &UserPublicResponse{
		ID:           u.ID,
		Email:        u.Email,
		Role:         u.Role,
		DisplayName:  u.DisplayName,
		Karma:        u.Karma,
		VibeVectorID: u.VibeVectorID,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
