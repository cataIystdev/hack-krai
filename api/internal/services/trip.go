// Файл trip.go реализует бизнес-логику модуля планирования поездок.
// Содержит методы создания поездки, получения деталей с участниками,
// генерации invite-токена, присоединения участника и заглушку Group Vibe Merge.
package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
	"kudytudy-api/internal/models"
)

// Ошибки сервиса поездок.
var (
	// ErrTripNotFound — поездка не найдена.
	ErrTripNotFound = errors.New("поездка не найдена")

	// ErrTripForbidden — нет прав на операцию с поездкой.
	ErrTripForbidden = errors.New("нет прав на эту операцию")

	// ErrInvalidInviteToken — неверный invite-токен.
	ErrInvalidInviteToken = errors.New("неверный invite-токен")

	// ErrTripFull — превышено максимальное количество участников.
	ErrTripFull = errors.New("превышено максимальное количество участников")

	// ErrAlreadyMember — пользователь уже является участником поездки.
	ErrAlreadyMember = errors.New("пользователь уже является участником поездки")

	// ErrInvalidDates — некорректные даты поездки.
	ErrInvalidDates = errors.New("некорректные даты поездки")
)

// TripService — сервис управления поездками.
type TripService struct {
	tripRepo *database.TripRepository
	userRepo *database.UserRepository
	logger   *zap.Logger
}

// NewTripService создаёт новый сервис поездок.
func NewTripService(
	tripRepo *database.TripRepository,
	userRepo *database.UserRepository,
	logger *zap.Logger,
) *TripService {
	return &TripService{
		tripRepo: tripRepo,
		userRepo: userRepo,
		logger:   logger.Named("trip_service"),
	}
}

// Create создаёт новую поездку и добавляет создателя как первого участника.
// Процесс: валидация дат, создание записи в trips, добавление создателя в trip_members
// с ролью creator.
func (s *TripService) Create(ctx context.Context, creatorID uuid.UUID, req *models.CreateTripRequest) (*models.TripWithMembers, error) {
	// Нормализация дефолтных значений.
	req.NormalizeDefaults()

	// Парсинг дат.
	dateFrom, dateTo := req.ParseDates()

	// Создание объекта поездки.
	trip := &models.Trip{
		CreatorID:        creatorID,
		DateFrom:         dateFrom,
		DateTo:           dateTo,
		BudgetRub:        req.BudgetRub,
		BudgetTier:       req.BudgetTier,
		Transport:        req.Transport,
		GroupSize:        req.GroupSize,
		GroupComposition: req.GroupComposition,
		Status:           models.TripStatusPlanning,
	}

	// Вставка поездки в БД.
	createdTrip, err := s.tripRepo.Create(ctx, trip)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания поездки: %w", err)
	}

	// Получение данных создателя для display_name.
	creator, err := s.userRepo.FindByID(ctx, creatorID.String())
	if err != nil {
		s.logger.Warn("не удалось получить данные создателя, используем fallback",
			zap.String("creator_id", creatorID.String()),
			zap.Error(err),
		)
	}

	displayName := "Создатель"
	if creator != nil && creator.DisplayName != "" {
		displayName = creator.DisplayName
	}

	// Добавление создателя как первого участника.
	creatorMember := &models.TripMember{
		TripID:      createdTrip.ID,
		UserID:      &creatorID,
		DisplayName: displayName,
		Role:        models.TripRoleCreator,
		Tags:        []string{},
		IsChild:     false,
	}

	member, err := s.tripRepo.AddMember(ctx, creatorMember)
	if err != nil {
		s.logger.Error("ошибка добавления создателя в участники",
			zap.String("trip_id", createdTrip.ID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка добавления создателя: %w", err)
	}

	s.logger.Info("поездка создана",
		zap.String("trip_id", createdTrip.ID.String()),
		zap.String("creator_id", creatorID.String()),
		zap.String("invite_token", createdTrip.InviteToken.String()),
	)

	return &models.TripWithMembers{
		Trip:    *createdTrip,
		Members: []models.TripMember{*member},
	}, nil
}

// GetByID возвращает поездку с полным списком участников.
func (s *TripService) GetByID(ctx context.Context, tripID uuid.UUID) (*models.TripWithMembers, error) {
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		if errors.Is(err, database.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}

	members, err := s.tripRepo.FindMembersByTripID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	if members == nil {
		members = []models.TripMember{}
	}

	return &models.TripWithMembers{
		Trip:    *trip,
		Members: members,
	}, nil
}

// GenerateInviteToken генерирует новый invite-токен для поездки.
// Доступно только создателю поездки (userID == trip.creator_id).
// Возвращает новый токен.
func (s *TripService) GenerateInviteToken(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) (uuid.UUID, error) {
	// Проверка существования поездки и прав.
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		if errors.Is(err, database.ErrTripNotFound) {
			return uuid.Nil, ErrTripNotFound
		}
		return uuid.Nil, err
	}

	// Только создатель может генерировать invite-токен.
	if trip.CreatorID != userID {
		return uuid.Nil, ErrTripForbidden
	}

	// Генерация нового UUID.
	newToken := uuid.New()

	// Обновление токена в БД.
	if err := s.tripRepo.UpdateInviteToken(ctx, tripID, newToken); err != nil {
		return uuid.Nil, err
	}

	s.logger.Info("invite-токен обновлён",
		zap.String("trip_id", tripID.String()),
		zap.String("new_token", newToken.String()),
	)

	return newToken, nil
}

// Join присоединяет участника к поездке по invite-токену.
// Доступно без авторизации: неавторизованные участники имеют user_id = nil.
// Проверяет валидность токена и атомарно вставляет участника с проверкой group_size.
func (s *TripService) Join(ctx context.Context, tripID uuid.UUID, req *models.JoinTripRequest, userID *uuid.UUID) (*models.TripMember, error) {
	// Парсинг invite-токена.
	inviteToken, err := uuid.Parse(req.InviteToken)
	if err != nil {
		return nil, ErrInvalidInviteToken
	}

	// Поиск поездки по ID.
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		if errors.Is(err, database.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}

	// Проверка валидности invite-токена.
	if trip.InviteToken != inviteToken {
		return nil, ErrInvalidInviteToken
	}

	// Подготовка тегов.
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	// Создание записи участника.
	member := &models.TripMember{
		TripID:      tripID,
		UserID:      userID,
		DisplayName: req.DisplayName,
		Role:        models.TripRoleMember,
		Tags:        tags,
		IsChild:     req.IsChild,
	}

	// Атомарная вставка с проверкой лимита group_size.
	// INSERT выполняется только если count < group_size (один SQL-запрос).
	created, err := s.tripRepo.AddMemberAtomic(ctx, member, trip.GroupSize)
	if err != nil {
		if errors.Is(err, database.ErrMemberAlreadyExists) {
			return nil, ErrAlreadyMember
		}
		if errors.Is(err, database.ErrTripFull) {
			return nil, ErrTripFull
		}
		return nil, err
	}

	s.logger.Info("участник присоединился к поездке",
		zap.String("trip_id", tripID.String()),
		zap.String("member_id", created.ID.String()),
		zap.String("display_name", created.DisplayName),
	)

	return created, nil
}

// GetMembers возвращает список участников поездки.
func (s *TripService) GetMembers(ctx context.Context, tripID uuid.UUID) ([]models.TripMember, error) {
	// Проверка существования поездки.
	_, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		if errors.Is(err, database.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}

	members, err := s.tripRepo.FindMembersByTripID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	if members == nil {
		members = []models.TripMember{}
	}

	return members, nil
}

// MergeGroupVibes вычисляет средневзвешенный vibe-вектор группы.
// Текущая реализация — заглушка, возвращающая nil.
// В будущем будет загружать vibe_vector_id всех участников из Qdrant,
// вычислять взвешенное среднее (дети получают дополнительный вес на child_friendly)
// и сохранять результат в Qdrant с обновлением trips.merged_vibe_vector_id.
func (s *TripService) MergeGroupVibes(ctx context.Context, tripID uuid.UUID) (*uuid.UUID, error) {
	// Получение участников с vibe-векторами.
	members, err := s.tripRepo.FindMembersByTripID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	// Сбор vibe-вектор ID участников, у которых они есть.
	var vibeVectorIDs []uuid.UUID
	for _, m := range members {
		if m.VibeVectorID != nil {
			vibeVectorIDs = append(vibeVectorIDs, *m.VibeVectorID)
		}
	}

	// Если ни у одного участника нет vibe-вектора — возвращаем nil.
	if len(vibeVectorIDs) == 0 {
		s.logger.Info("нет vibe-векторов для merge",
			zap.String("trip_id", tripID.String()),
			zap.Int("members_count", len(members)),
		)
		return nil, nil
	}

	// Заглушка: в будущем здесь будет:
	// 1. Загрузка векторов из Qdrant по vibeVectorIDs
	// 2. Вычисление средневзвешенного вектора (детские профили получают
	//    дополнительный вес на измерения child_friendly)
	// 3. Upsert merged вектора в Qdrant
	// 4. Обновление trips.merged_vibe_vector_id

	s.logger.Info("GroupVibeMerge: заглушка, вектора собраны",
		zap.String("trip_id", tripID.String()),
		zap.Int("vectors_count", len(vibeVectorIDs)),
		zap.Strings("vector_ids", uuidsToStrings(vibeVectorIDs)),
	)

	return nil, nil
}

// uuidsToStrings конвертирует срез UUID в срез строк для логирования.
func uuidsToStrings(ids []uuid.UUID) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}
