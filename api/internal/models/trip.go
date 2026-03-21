// Файл trip.go содержит структуры данных для таблиц trips и trip_members.
// Определяет модели поездки и участника, DTO для создания поездки и
// присоединения к ней, константы статусов и ролей, методы валидации.
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TripStatus — статус поездки.
type TripStatus string

const (
	// TripStatusPlanning — поездка в стадии планирования.
	TripStatusPlanning TripStatus = "planning"

	// TripStatusActive — поездка активна (в процессе).
	TripStatusActive TripStatus = "active"

	// TripStatusCompleted — поездка завершена.
	TripStatusCompleted TripStatus = "completed"

	// TripStatusCancelled — поездка отменена.
	TripStatusCancelled TripStatus = "cancelled"
)

// TripMemberRole — роль участника в поездке.
type TripMemberRole string

const (
	// TripRoleCreator — создатель поездки.
	TripRoleCreator TripMemberRole = "creator"

	// TripRoleMember — обычный участник.
	TripRoleMember TripMemberRole = "member"
)

// Trip — структура поездки, соответствующая таблице trips в PostgreSQL.
// Содержит даты, бюджет, транспорт, состав группы и invite-токен.
type Trip struct {
	// ID — уникальный идентификатор поездки (UUID v4).
	ID uuid.UUID `json:"id" db:"id"`

	// CreatorID — UUID создателя поездки.
	CreatorID uuid.UUID `json:"creator_id" db:"creator_id"`

	// DateFrom — дата начала поездки.
	DateFrom time.Time `json:"date_from" db:"date_from"`

	// DateTo — дата окончания поездки.
	DateTo time.Time `json:"date_to" db:"date_to"`

	// BudgetRub — общий бюджет поездки в рублях.
	BudgetRub int `json:"budget_rub" db:"budget_rub"`

	// BudgetTier — уровень бюджета (economy/comfort/premium).
	BudgetTier string `json:"budget_tier" db:"budget_tier"`

	// Transport — вид транспорта (car/public/walk/bike).
	Transport string `json:"transport" db:"transport"`

	// GroupSize — планируемое количество участников.
	GroupSize int `json:"group_size" db:"group_size"`

	// GroupComposition — состав группы в формате JSON.
	// Пример: {"adults": 2, "children": [{"age": 8}, {"age": 12}]}.
	GroupComposition json.RawMessage `json:"group_composition" db:"group_composition"`

	// Format — формат поездки (day_trip/weekend/multi_day).
	Format string `json:"format" db:"format"`

	// InviteToken — уникальный токен для приглашения участников.
	InviteToken uuid.UUID `json:"invite_token" db:"invite_token"`

	// VibeVectorID — ID vibe-вектора создателя в Qdrant (nullable).
	VibeVectorID *uuid.UUID `json:"vibe_vector_id,omitempty" db:"vibe_vector_id"`

	// MergedVibeVectorID — ID средневзвешенного vibe-вектора группы в Qdrant.
	MergedVibeVectorID *uuid.UUID `json:"merged_vibe_vector_id,omitempty" db:"merged_vibe_vector_id"`

	// Status — статус поездки (planning/active/completed/cancelled).
	Status TripStatus `json:"status" db:"status"`

	// CreatedAt — временная метка создания.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt — временная метка последнего обновления.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TripMember — структура участника поездки, соответствующая таблице trip_members.
// Хранит роль, теги предпочтений и ссылку на vibe-вектор участника.
type TripMember struct {
	// ID — уникальный идентификатор записи участника.
	ID uuid.UUID `json:"id" db:"id"`

	// TripID — идентификатор поездки.
	TripID uuid.UUID `json:"trip_id" db:"trip_id"`

	// UserID — идентификатор пользователя (nullable для неавторизованных).
	UserID *uuid.UUID `json:"user_id,omitempty" db:"user_id"`

	// DisplayName — отображаемое имя участника.
	DisplayName string `json:"display_name" db:"display_name"`

	// Role — роль в поездке (creator/member).
	Role TripMemberRole `json:"role" db:"role"`

	// VibeVectorID — ID vibe-вектора участника в Qdrant (nullable).
	VibeVectorID *uuid.UUID `json:"vibe_vector_id,omitempty" db:"vibe_vector_id"`

	// Tags — теги предпочтений участника.
	Tags []string `json:"tags" db:"tags"`

	// IsChild — является ли участник ребёнком.
	IsChild bool `json:"is_child" db:"is_child"`

	// JoinedAt — временная метка присоединения.
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

// TripWithMembers — расширенная структура поездки со списком участников.
// Используется для ответа GET /api/v1/trips/{id}.
type TripWithMembers struct {
	// Trip — данные поездки.
	Trip

	// Members — список участников поездки.
	Members []TripMember `json:"members"`
}

// CreateTripRequest — DTO для создания новой поездки.
// Все обязательные поля: date_from, date_to.
type CreateTripRequest struct {
	// DateFrom — дата начала поездки (формат: YYYY-MM-DD).
	DateFrom string `json:"date_from"`

	// DateTo — дата окончания поездки (формат: YYYY-MM-DD).
	DateTo string `json:"date_to"`

	// BudgetRub — общий бюджет в рублях (опционально, по умолчанию 0).
	BudgetRub int `json:"budget_rub"`

	// BudgetTier — уровень бюджета: economy, comfort, premium (по умолчанию comfort).
	BudgetTier string `json:"budget_tier"`

	// Transport — вид транспорта: car, public, walk, bike (по умолчанию car).
	Transport string `json:"transport"`

	// GroupSize — планируемое количество участников (по умолчанию 1).
	GroupSize int `json:"group_size"`

	// GroupComposition — состав группы в формате JSON (опционально).
	GroupComposition json.RawMessage `json:"group_composition"`

	// Format — формат поездки: day_trip, weekend, multi_day (по умолчанию multi_day).
	Format string `json:"format"`

	// VibeVectorID — ID vibe-вектора создателя из Qdrant (опционально).
	VibeVectorID string `json:"vibe_vector_id"`
}

// DateLayout — формат даты для парсинга (ISO 8601 date).
const DateLayout = "2006-01-02"

// ValidBudgetTiers — допустимые уровни бюджета.
var ValidBudgetTiers = map[string]bool{
	"economy": true,
	"comfort": true,
	"premium": true,
}

// ValidTransports — допустимые виды транспорта.
var ValidTransports = map[string]bool{
	"car":    true,
	"public": true,
	"walk":   true,
	"bike":   true,
}

// ValidFormats — допустимые форматы поездки.
var ValidFormats = map[string]bool{
	"day_trip":  true,
	"weekend":   true,
	"multi_day": true,
}

// Validate проверяет корректность данных для создания поездки.
// Возвращает текст ошибки или пустую строку при успехе.
func (r *CreateTripRequest) Validate() string {
	if r.DateFrom == "" {
		return "дата начала обязательна"
	}
	if r.DateTo == "" {
		return "дата окончания обязательна"
	}

	dateFrom, err := time.Parse(DateLayout, r.DateFrom)
	if err != nil {
		return "некорректный формат даты начала (ожидается YYYY-MM-DD)"
	}

	dateTo, err := time.Parse(DateLayout, r.DateTo)
	if err != nil {
		return "некорректный формат даты окончания (ожидается YYYY-MM-DD)"
	}

	if dateTo.Before(dateFrom) {
		return "дата окончания не может быть раньше даты начала"
	}

	if r.BudgetRub < 0 {
		return "бюджет не может быть отрицательным"
	}

	if r.BudgetTier != "" && !ValidBudgetTiers[r.BudgetTier] {
		return "допустимые уровни бюджета: economy, comfort, premium"
	}

	if r.Transport != "" && !ValidTransports[r.Transport] {
		return "допустимые виды транспорта: car, public, walk, bike"
	}

	if r.GroupSize < 0 {
		return "количество участников не может быть отрицательным"
	}

	if r.Format != "" && !ValidFormats[r.Format] {
		return "допустимые форматы поездки: day_trip, weekend, multi_day"
	}

	if r.VibeVectorID != "" {
		if _, err := uuid.Parse(r.VibeVectorID); err != nil {
			return "vibe_vector_id должен быть валидным UUID"
		}
	}

	return ""
}

// ParseDates парсит строковые даты в time.Time.
// Вызывается после успешной валидации.
func (r *CreateTripRequest) ParseDates() (time.Time, time.Time) {
	dateFrom, _ := time.Parse(DateLayout, r.DateFrom)
	dateTo, _ := time.Parse(DateLayout, r.DateTo)
	return dateFrom, dateTo
}

// NormalizeDefaults устанавливает значения по умолчанию для незаполненных полей.
func (r *CreateTripRequest) NormalizeDefaults() {
	if r.BudgetTier == "" {
		r.BudgetTier = "comfort"
	}
	if r.Transport == "" {
		r.Transport = "car"
	}
	if r.GroupSize <= 0 {
		r.GroupSize = 1
	}
	if r.GroupComposition == nil {
		r.GroupComposition = json.RawMessage("{}")
	}
	if r.Format == "" {
		r.Format = "multi_day"
	}
}

// JoinTripRequest — DTO для присоединения к поездке по invite-ссылке.
// Эндпоинт POST /api/v1/trips/{id}/join доступен без авторизации.
type JoinTripRequest struct {
	// InviteToken — токен приглашения (обязательный).
	InviteToken string `json:"invite_token"`

	// DisplayName — отображаемое имя участника (обязательный).
	DisplayName string `json:"display_name"`

	// Tags — теги предпочтений (опционально).
	Tags []string `json:"tags"`

	// IsChild — является ли участник ребёнком (по умолчанию false).
	IsChild bool `json:"is_child"`
}

// Validate проверяет корректность данных для присоединения к поездке.
// Возвращает текст ошибки или пустую строку при успехе.
func (r *JoinTripRequest) Validate() string {
	if r.InviteToken == "" {
		return "invite_token обязателен"
	}
	if r.DisplayName == "" {
		return "display_name обязателен"
	}

	// Проверка формата UUID.
	if _, err := uuid.Parse(r.InviteToken); err != nil {
		return "invite_token должен быть валидным UUID"
	}

	return ""
}

// UpdateTripRequest — DTO для частичного обновления поездки.
// Все поля — указатели (nil = не обновлять). PUT /api/v1/trips/{id}.
type UpdateTripRequest struct {
	// DateFrom — новая дата начала (формат: YYYY-MM-DD).
	DateFrom *string `json:"date_from"`

	// DateTo — новая дата окончания (формат: YYYY-MM-DD).
	DateTo *string `json:"date_to"`

	// BudgetRub — новый бюджет в рублях.
	BudgetRub *int `json:"budget_rub"`

	// BudgetTier — новый уровень бюджета.
	BudgetTier *string `json:"budget_tier"`

	// Transport — новый вид транспорта.
	Transport *string `json:"transport"`

	// GroupSize — новое количество участников.
	GroupSize *int `json:"group_size"`

	// GroupComposition — новый состав группы.
	GroupComposition *json.RawMessage `json:"group_composition"`

	// Format — новый формат поездки.
	Format *string `json:"format"`

	// VibeVectorID — новый ID vibe-вектора.
	VibeVectorID *string `json:"vibe_vector_id"`
}

// Validate проверяет корректность данных для обновления поездки.
// Проверяет только переданные (не nil) поля.
func (r *UpdateTripRequest) Validate() string {
	if r.DateFrom != nil {
		if *r.DateFrom == "" {
			return "дата начала не может быть пустой"
		}
		if _, err := time.Parse(DateLayout, *r.DateFrom); err != nil {
			return "некорректный формат даты начала (ожидается YYYY-MM-DD)"
		}
	}

	if r.DateTo != nil {
		if *r.DateTo == "" {
			return "дата окончания не может быть пустой"
		}
		if _, err := time.Parse(DateLayout, *r.DateTo); err != nil {
			return "некорректный формат даты окончания (ожидается YYYY-MM-DD)"
		}
	}

	// Проверка порядка дат, если обе переданы.
	if r.DateFrom != nil && r.DateTo != nil {
		dateFrom, _ := time.Parse(DateLayout, *r.DateFrom)
		dateTo, _ := time.Parse(DateLayout, *r.DateTo)
		if dateTo.Before(dateFrom) {
			return "дата окончания не может быть раньше даты начала"
		}
	}

	if r.BudgetRub != nil && *r.BudgetRub < 0 {
		return "бюджет не может быть отрицательным"
	}

	if r.BudgetTier != nil && !ValidBudgetTiers[*r.BudgetTier] {
		return "допустимые уровни бюджета: economy, comfort, premium"
	}

	if r.Transport != nil && !ValidTransports[*r.Transport] {
		return "допустимые виды транспорта: car, public, walk, bike"
	}

	if r.GroupSize != nil && *r.GroupSize < 1 {
		return "количество участников должно быть не менее 1"
	}

	if r.Format != nil && !ValidFormats[*r.Format] {
		return "допустимые форматы поездки: day_trip, weekend, multi_day"
	}

	if r.VibeVectorID != nil && *r.VibeVectorID != "" {
		if _, err := uuid.Parse(*r.VibeVectorID); err != nil {
			return "vibe_vector_id должен быть валидным UUID"
		}
	}

	return ""
}
