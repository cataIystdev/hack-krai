// Файл trip.go реализует бизнес-логику модуля планирования поездок.
// Содержит методы создания поездки, получения деталей с участниками,
// генерации invite-токена, присоединения участника (auth-flex),
// проверки membership-доступа и вычисления merged vibe vector группы.
package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

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

// CollectionGroupVibes — коллекция для merged vibe-векторов групп в Qdrant.
const CollectionGroupVibes = "group_vibes"

// TripService — сервис управления поездками.
type TripService struct {
	tripRepo *database.TripRepository
	userRepo *database.UserRepository
	vibeRepo *database.VibeRepository
	logger   *zap.Logger
}

// NewTripService создаёт новый сервис поездок.
// Параметр vibeRepo может быть nil — при этом MergeGroupVibes не выполняется.
func NewTripService(
	tripRepo *database.TripRepository,
	userRepo *database.UserRepository,
	vibeRepo *database.VibeRepository,
	logger *zap.Logger,
) *TripService {
	return &TripService{
		tripRepo: tripRepo,
		userRepo: userRepo,
		vibeRepo: vibeRepo,
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
		Format:           req.Format,
		Status:           models.TripStatusPlanning,
	}

	// Парсинг vibe_vector_id из DTO, если передан.
	if req.VibeVectorID != "" {
		vid, _ := uuid.Parse(req.VibeVectorID)
		trip.VibeVectorID = &vid
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

	// VibeVectorID создателя подтягивается из профиля пользователя.
	var creatorVibeVectorID *uuid.UUID
	if creator != nil && creator.VibeVectorID != nil {
		creatorVibeVectorID = creator.VibeVectorID
	}

	// Добавление создателя как первого участника.
	creatorMember := &models.TripMember{
		TripID:       createdTrip.ID,
		UserID:       &creatorID,
		DisplayName:  displayName,
		Role:         models.TripRoleCreator,
		VibeVectorID: creatorVibeVectorID,
		Tags:         []string{},
		IsChild:      false,
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

// Update обновляет поездку по переданным полям.
// Доступно только создателю поездки. Обновляются только non-nil поля запроса.
func (s *TripService) Update(ctx context.Context, tripID, userID uuid.UUID, req *models.UpdateTripRequest) (*models.TripWithMembers, error) {
	// Проверка существования поездки и прав.
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		if errors.Is(err, database.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}

	// Только создатель может обновлять поездку.
	if trip.CreatorID != userID {
		return nil, ErrTripForbidden
	}

	// Формирование map обновляемых полей.
	fields := make(map[string]any)

	if req.DateFrom != nil {
		dateFrom, _ := time.Parse(models.DateLayout, *req.DateFrom)
		fields["date_from"] = dateFrom
	}
	if req.DateTo != nil {
		dateTo, _ := time.Parse(models.DateLayout, *req.DateTo)
		fields["date_to"] = dateTo
	}
	if req.BudgetRub != nil {
		fields["budget_rub"] = *req.BudgetRub
	}
	if req.BudgetTier != nil {
		fields["budget_tier"] = *req.BudgetTier
	}
	if req.Transport != nil {
		fields["transport"] = *req.Transport
	}
	if req.GroupSize != nil {
		fields["group_size"] = *req.GroupSize
	}
	if req.GroupComposition != nil {
		fields["group_composition"] = *req.GroupComposition
	}
	if req.Format != nil {
		fields["format"] = *req.Format
	}
	if req.VibeVectorID != nil {
		if *req.VibeVectorID == "" {
			fields["vibe_vector_id"] = nil
		} else {
			vid, _ := uuid.Parse(*req.VibeVectorID)
			fields["vibe_vector_id"] = vid
		}
	}

	// Обновление в БД.
	updatedTrip, err := s.tripRepo.Update(ctx, tripID, fields)
	if err != nil {
		if errors.Is(err, database.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}

	// Загрузка участников.
	members, err := s.tripRepo.FindMembersByTripID(ctx, tripID)
	if err != nil {
		members = []models.TripMember{}
	}

	s.logger.Info("поездка обновлена",
		zap.String("trip_id", tripID.String()),
		zap.String("user_id", userID.String()),
		zap.Int("fields_updated", len(fields)),
	)

	return &models.TripWithMembers{
		Trip:    *updatedTrip,
		Members: members,
	}, nil
}

// GetByID возвращает поездку с полным списком участников.
// Доступ ограничен: только участники поездки или её создатель могут видеть детали.
func (s *TripService) GetByID(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) (*models.TripWithMembers, error) {
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		if errors.Is(err, database.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}

	// Проверка доступа: создатель или участник.
	if err := s.checkMembership(ctx, trip, userID); err != nil {
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
// Auth-flex логика:
//   - Если userID присутствует — подтягивает display_name и vibe_vector_id из профиля users.
//   - Если userID отсутствует — использует данные из запроса (анонимный join).
//
// После присоединения запускает пересчёт merged vibe vector группы.
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

	// Подготовка данных участника.
	displayName := req.DisplayName
	var vibeVectorID *uuid.UUID
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	// Auth-flex: обогащение данных из профиля авторизованного пользователя.
	if userID != nil {
		user, err := s.userRepo.FindByID(ctx, userID.String())
		if err == nil && user != nil {
			// Если display_name не передан в запросе — используем из профиля.
			if displayName == "" {
				displayName = user.DisplayName
			}
			// Подтягиваем vibe_vector_id из профиля пользователя.
			if user.VibeVectorID != nil {
				vibeVectorID = user.VibeVectorID
			}
		} else {
			s.logger.Warn("не удалось загрузить профиль для auth-flex join",
				zap.String("user_id", userID.String()),
				zap.Error(err),
			)
		}
	}

	// Fallback: если display_name всё ещё пуст.
	if displayName == "" {
		displayName = "Участник"
	}

	// Создание записи участника.
	member := &models.TripMember{
		TripID:       tripID,
		UserID:       userID,
		DisplayName:  displayName,
		Role:         models.TripRoleMember,
		VibeVectorID: vibeVectorID,
		Tags:         tags,
		IsChild:      req.IsChild,
	}

	// Атомарная вставка с проверкой лимита group_size и дубликатов.
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

	// Асинхронный пересчёт merged vibe vector (не блокируем ответ).
	go func() {
		bgCtx := context.Background()
		if _, err := s.MergeGroupVibes(bgCtx, tripID); err != nil {
			s.logger.Warn("ошибка пересчёта merged vibe после join",
				zap.String("trip_id", tripID.String()),
				zap.Error(err),
			)
		}
	}()

	return created, nil
}

// GetUserTrips возвращает все поездки пользователя с участниками.
// Используется для GET /api/v1/trips.
func (s *TripService) GetUserTrips(ctx context.Context, userID uuid.UUID) ([]models.TripWithMembers, error) {
	trips, err := s.tripRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]models.TripWithMembers, 0, len(trips))
	for i := range trips {
		members, err := s.tripRepo.FindMembersByTripID(ctx, trips[i].ID)
		if err != nil {
			s.logger.Warn("не удалось загрузить участников поездки",
				zap.String("trip_id", trips[i].ID.String()),
				zap.Error(err),
			)
			members = []models.TripMember{}
		}
		result = append(result, models.TripWithMembers{
			Trip:    trips[i],
			Members: members,
		})
	}

	s.logger.Info("поездки пользователя получены",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(result)),
	)

	return result, nil
}

// GetMembers возвращает список участников поездки.
// Доступ ограничен: только участники поездки или её создатель видят список.
func (s *TripService) GetMembers(ctx context.Context, tripID uuid.UUID, userID uuid.UUID) ([]models.TripMember, error) {
	// Проверка существования поездки.
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		if errors.Is(err, database.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}

	// Проверка доступа: создатель или участник.
	if err := s.checkMembership(ctx, trip, userID); err != nil {
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
// Алгоритм:
//  1. Загрузить всех участников поездки.
//  2. Собрать vibe_vector_id у каждого участника.
//  3. Загрузить векторы из Qdrant (коллекция user_vibes).
//  4. Вычислить взвешенное среднее (дети получают бонус на child_friendly компоненты).
//  5. L2-нормализация результата.
//  6. Upsert merged вектора в Qdrant (коллекция group_vibes).
//  7. Обновить trips.merged_vibe_vector_id.
func (s *TripService) MergeGroupVibes(ctx context.Context, tripID uuid.UUID) (*uuid.UUID, error) {
	// Если vibeRepo не сконфигурирован — пропускаем.
	if s.vibeRepo == nil {
		s.logger.Debug("vibeRepo не сконфигурирован, пропуск MergeGroupVibes",
			zap.String("trip_id", tripID.String()),
		)
		return nil, nil
	}

	// Загрузка участников поездки.
	members, err := s.tripRepo.FindMembersByTripID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки участников: %w", err)
	}

	// Сбор vibe-вектор ID участников, у которых они есть.
	type vibeEntry struct {
		vectorID uuid.UUID
		isChild  bool
	}
	var entries []vibeEntry
	for _, m := range members {
		if m.VibeVectorID != nil {
			entries = append(entries, vibeEntry{
				vectorID: *m.VibeVectorID,
				isChild:  m.IsChild,
			})
		}
	}

	// Если ни у одного участника нет vibe-вектора — возвращаем nil.
	if len(entries) == 0 {
		s.logger.Info("нет vibe-векторов для merge",
			zap.String("trip_id", tripID.String()),
			zap.Int("members_count", len(members)),
		)
		return nil, nil
	}

	// Загрузка векторов из Qdrant.
	var vectors [][]float32
	var childFlags []bool
	for _, e := range entries {
		vec, err := s.vibeRepo.GetVibeVector(ctx, database.CollectionUserVibes, e.vectorID)
		if err != nil {
			s.logger.Warn("не удалось загрузить vibe-вектор участника",
				zap.String("vector_id", e.vectorID.String()),
				zap.Error(err),
			)
			continue
		}
		vectors = append(vectors, vec)
		childFlags = append(childFlags, e.isChild)
	}

	if len(vectors) == 0 {
		s.logger.Info("ни один vibe-вектор не удалось загрузить",
			zap.String("trip_id", tripID.String()),
		)
		return nil, nil
	}

	// Вычисление взвешенного среднего.
	dim := len(vectors[0])
	merged := make([]float32, dim)
	totalWeight := float64(0)

	for i, vec := range vectors {
		if len(vec) != dim {
			continue
		}
		weight := 1.0
		if childFlags[i] {
			// Детские профили получают увеличенный вес (1.2x)
			// для большего влияния на child_friendly компоненты маршрута.
			weight = 1.2
		}
		totalWeight += weight
		for j := 0; j < dim; j++ {
			merged[j] += float32(weight) * vec[j]
		}
	}

	// Нормализация: деление на сумму весов.
	if totalWeight > 0 {
		for j := range merged {
			merged[j] /= float32(totalWeight)
		}
	}

	// L2-нормализация.
	var norm float64
	for _, v := range merged {
		norm += float64(v) * float64(v)
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for j := range merged {
			merged[j] /= float32(norm)
		}
	}

	// Генерация или обновление ID merged вектора.
	// Используем детерминистический UUID на основе tripID для идемпотентности.
	mergedVectorID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("group-vibe-"+tripID.String()))

	// Upsert merged вектора в Qdrant.
	payload := map[string]any{
		"trip_id":        tripID.String(),
		"members_count":  len(vectors),
		"type":           "group_merged",
	}

	err = s.vibeRepo.UpsertVibeVector(ctx, CollectionGroupVibes, mergedVectorID, merged, payload)
	if err != nil {
		s.logger.Warn("не удалось сохранить merged vibe vector в Qdrant",
			zap.String("trip_id", tripID.String()),
			zap.Error(err),
		)
		// Не возвращаем ошибку — merge не критичен для основного flow.
		return nil, nil
	}

	// Обновление trips.merged_vibe_vector_id.
	if err := s.tripRepo.UpdateMergedVibeVector(ctx, tripID, mergedVectorID); err != nil {
		s.logger.Warn("не удалось обновить merged_vibe_vector_id в поездке",
			zap.String("trip_id", tripID.String()),
			zap.Error(err),
		)
		return nil, nil
	}

	s.logger.Info("merged vibe vector вычислен и сохранён",
		zap.String("trip_id", tripID.String()),
		zap.String("merged_vector_id", mergedVectorID.String()),
		zap.Int("vectors_count", len(vectors)),
		zap.Int("dim", dim),
	)

	return &mergedVectorID, nil
}

// checkMembership проверяет, что пользователь является участником или создателем поездки.
// Возвращает ErrTripForbidden если пользователь не имеет доступа.
func (s *TripService) checkMembership(ctx context.Context, trip *models.Trip, userID uuid.UUID) error {
	// Создатель всегда имеет доступ.
	if trip.CreatorID == userID {
		return nil
	}

	// Проверка через таблицу trip_members.
	isMember, err := s.tripRepo.IsMember(ctx, trip.ID, userID)
	if err != nil {
		return fmt.Errorf("ошибка проверки membership: %w", err)
	}

	if !isMember {
		return ErrTripForbidden
	}

	return nil
}

// uuidsToStrings конвертирует срез UUID в срез строк для логирования.
func uuidsToStrings(ids []uuid.UUID) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}
