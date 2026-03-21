// Файл route.go реализует бизнес-логику построения маршрутов.
// RouteService поддерживает два режима:
//   - BuildRoute: ручной режим — маршрут по заданным ID локаций.
//   - BuildTripRoute: trip-aware режим — автоподбор/фильтрация локаций
//     на основе параметров поездки, vibe-скоринг через Qdrant,
//     распределение по дням и тайм-слотам, сохранение в БД.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
	"kudytudy-api/internal/models"
)

// earthRadiusKm — радиус Земли в километрах для формулы Haversine.
const earthRadiusKm = 6371.0

// transportSpeedKmH — средняя скорость по видам транспорта (км/ч).
// Используется для расчёта времени в пути.
var transportSpeedKmH = map[string]float64{
	"car":    60.0,
	"public": 35.0,
	"walk":   5.0,
	"bike":   15.0,
}

// defaultStayDurationMin — стандартное время пребывания в точке (минуты).
// Зависит от категории локации.
var defaultStayDurationMin = map[string]int{
	"winery":     90,
	"farm":       60,
	"trail":      120,
	"nature":     90,
	"cultural":   60,
	"gastro":     75,
	"beach":      120,
	"resort":     180,
	"camping":    60,
	"guesthouse": 30,
	"extreme":    120,
}

// maxBudgetPerNightByTier — максимальная стоимость за ночь по бюджету (рубли).
// Используется для фильтрации локаций при trip-aware построении.
var maxBudgetPerNightByTier = map[string]int{
	"economy": 3000,
	"comfort": 7000,
	"premium": 50000,
}

// childFriendlyCategories — категории локаций, подходящие для детей.
var childFriendlyCategories = map[string]bool{
	"farm":    true,
	"beach":   true,
	"nature":  true,
	"resort":  true,
	"camping": true,
}

// adultOnlyCategories — категории локаций преимущественно для взрослых.
var adultOnlyCategories = map[string]bool{
	"winery":  true,
	"gastro":  true,
	"extreme": true,
}

// slotsPerDay — количество тайм-слотов в дне (morning, afternoon, evening).
const slotsPerDay = 3

// scoredLocation — локация с vibe-score для ранжирования.
type scoredLocation struct {
	location  models.Location
	vibeScore float32
}

// RouteService — сервис построения маршрутов.
// Поддерживает ручной и trip-aware режимы.
type RouteService struct {
	locationRepo *database.LocationRepository
	tripRepo     *database.TripRepository
	routeRepo    *database.RouteRepository
	vibeRepo     *database.VibeRepository
	logger       *zap.Logger
}

// NewRouteService создаёт экземпляр сервиса маршрутов.
// tripRepo, routeRepo и vibeRepo могут быть nil (для ручного режима достаточно locationRepo).
func NewRouteService(
	locationRepo *database.LocationRepository,
	tripRepo *database.TripRepository,
	routeRepo *database.RouteRepository,
	vibeRepo *database.VibeRepository,
	logger *zap.Logger,
) *RouteService {
	return &RouteService{
		locationRepo: locationRepo,
		tripRepo:     tripRepo,
		routeRepo:    routeRepo,
		vibeRepo:     vibeRepo,
		logger:       logger.Named("route_service"),
	}
}

// BuildRoute строит маршрут по списку ID локаций (ручной режим).
// Загружает данные локаций из PostgreSQL, рассчитывает расстояния
// между последовательными точками и формирует RoutePreview.
func (s *RouteService) BuildRoute(ctx context.Context, req *models.BuildRouteRequest) (*models.RoutePreview, error) {
	// Загрузка локаций из БД.
	locations, err := s.locationRepo.FindByIDs(ctx, req.LocationIDs)
	if err != nil {
		s.logger.Error("ошибка загрузки локаций для маршрута", zap.Error(err))
		return nil, fmt.Errorf("ошибка загрузки локаций: %w", err)
	}

	if len(locations) < 2 {
		return nil, &ValidationError{Message: "найдено менее 2 локаций из указанных ID"}
	}

	// Построение карты для сохранения порядка запроса.
	locMap := make(map[string]models.Location, len(locations))
	for _, loc := range locations {
		locMap[loc.ID.String()] = loc
	}

	// Формирование точек маршрута в порядке запроса.
	speed := transportSpeedKmH[req.Transport]
	if speed == 0 {
		speed = transportSpeedKmH["car"]
	}

	var points []models.RoutePoint
	var totalDistance float64
	var totalDuration int

	for i, locID := range req.LocationIDs {
		loc, ok := locMap[locID]
		if !ok {
			s.logger.Warn("локация не найдена, пропускаем",
				zap.String("location_id", locID),
			)
			continue
		}

		point := models.RoutePoint{
			LocationID:      loc.ID.String(),
			Name:            loc.Name,
			Latitude:        loc.Latitude,
			Longitude:       loc.Longitude,
			Category:        loc.Category,
			PreviewImageURL: loc.PreviewImageURL,
			Order:           len(points) + 1,
			StayDurationMin: getStayDuration(loc.Category),
			DayNumber:       1,
			TimeSlot:        models.TimeSlotMorning,
			TargetAudience:  models.AudienceAll,
		}

		// Расчёт расстояния и времени от предыдущей точки.
		if i > 0 && len(points) > 0 {
			prev := points[len(points)-1]
			dist := haversineDistance(prev.Latitude, prev.Longitude, loc.Latitude, loc.Longitude)
			durationMin := int(math.Ceil(dist / speed * 60))

			point.DistanceFromPrevKm = math.Round(dist*10) / 10
			point.DurationFromPrevMin = durationMin

			totalDistance += dist
			totalDuration += durationMin
		}

		totalDuration += point.StayDurationMin
		points = append(points, point)
	}

	if len(points) < 2 {
		return nil, &ValidationError{Message: "не удалось построить маршрут: найдено менее 2 валидных точек"}
	}

	return &models.RoutePreview{
		Points:           points,
		TotalDistanceKm:  math.Round(totalDistance*10) / 10,
		TotalDurationMin: totalDuration,
		Transport:        req.Transport,
		PointsCount:      len(points),
	}, nil
}

// BuildTripRoute строит маршрут с учётом параметров поездки (trip-aware режим).
// Алгоритм:
//  1. Загрузить Trip и его участников.
//  2. Вычислить merged vibe vector (взвешенное среднее, дети +20% child_friendly).
//  3. Через Qdrant SearchNearest найти ближайшие локации по vibe.
//  4. Отфильтровать по budget_tier, child_friendly, format.
//  5. Рассчитать расстояния по Haversine.
//  6. Распределить точки по дням и тайм-слотам.
//  7. Сохранить маршрут в БД.
func (s *RouteService) BuildTripRoute(ctx context.Context, tripID, userID uuid.UUID, req *models.BuildTripRouteRequest) (*models.RoutePreview, error) {
	if s.tripRepo == nil {
		return nil, &ValidationError{Message: "сервис поездок недоступен"}
	}

	// Шаг 1: загрузка поездки и участников.
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки поездки: %w", err)
	}

	members, err := s.tripRepo.FindMembersByTripID(ctx, tripID)
	if err != nil {
		s.logger.Warn("ошибка загрузки участников поездки", zap.Error(err))
		members = nil
	}

	// Определение наличия детей в группе.
	hasChildren := s.groupHasChildren(trip, members)

	// Нормализация запроса по формату поездки.
	req.NormalizeDefaults(trip.Format)

	// Шаг 2-3: получение и ранжирование локаций.
	var scoredLocs []scoredLocation
	if len(req.LocationIDs) > 0 {
		// Ручной выбор — загрузка указанных локаций.
		scoredLocs, err = s.loadManualLocations(ctx, req.LocationIDs)
	} else {
		// Автоподбор — vibe scoring + фильтрация.
		scoredLocs, err = s.autoSelectLocations(ctx, trip, members, hasChildren, req.MaxPoints)
	}
	if err != nil {
		return nil, err
	}

	if len(scoredLocs) < 2 {
		return nil, &ValidationError{Message: "недостаточно локаций для построения маршрута (требуется минимум 2)"}
	}

	// Обрезка до max_points.
	if len(scoredLocs) > req.MaxPoints {
		scoredLocs = scoredLocs[:req.MaxPoints]
	}

	// Шаг 5: расчёт расстояний.
	speed := transportSpeedKmH[trip.Transport]
	if speed == 0 {
		speed = transportSpeedKmH["car"]
	}

	points, totalDistance, totalDuration := s.buildRoutePoints(scoredLocs, speed, hasChildren)

	// Шаг 6: распределение по дням и тайм-слотам.
	tripDays := s.calculateTripDays(trip)
	s.assignDaysAndSlots(points, tripDays)

	// Расчёт стоимости.
	estimatedCost := s.estimateCost(scoredLocs, tripDays, trip.BudgetTier)

	// Генерация summary.
	summary := s.generateSummary(trip, points, totalDistance, tripDays)

	// Формирование названия маршрута.
	routeName := fmt.Sprintf("Маршрут по Краснодарскому краю (%d точек)", len(points))

	// Шаг 7: сохранение в БД.
	var routeID string
	if s.routeRepo != nil {
		routeID, err = s.persistRoute(ctx, trip, userID, routeName, totalDistance, totalDuration, estimatedCost, summary, points)
		if err != nil {
			s.logger.Error("ошибка сохранения маршрута в БД", zap.Error(err))
			// Маршрут всё равно возвращается, даже если не удалось сохранить.
		}
	}

	createdAt := trip.CreatedAt
	return &models.RoutePreview{
		ID:               routeID,
		TripID:           tripID.String(),
		Name:             routeName,
		Status:           models.RouteStatusDraft,
		Points:           points,
		TotalDistanceKm:  math.Round(totalDistance*10) / 10,
		TotalDurationMin: totalDuration,
		Transport:        trip.Transport,
		PointsCount:      len(points),
		EstimatedCostRub: estimatedCost,
		Summary:          summary,
		DaysCount:        tripDays,
		CreatedAt:        &createdAt,
	}, nil
}

// groupHasChildren определяет наличие детей в группе.
// Проверяет trip_members.is_child и group_composition.
func (s *RouteService) groupHasChildren(trip *models.Trip, members []models.TripMember) bool {
	// Проверка trip_members.
	for _, m := range members {
		if m.IsChild {
			return true
		}
	}

	// Проверка group_composition JSON.
	if len(trip.GroupComposition) > 0 {
		var composition map[string]json.RawMessage
		if json.Unmarshal(trip.GroupComposition, &composition) == nil {
			if _, has := composition["children"]; has {
				return true
			}
		}
	}

	return false
}

// loadManualLocations загружает локации по ID без vibe-скоринга.
func (s *RouteService) loadManualLocations(ctx context.Context, ids []string) ([]scoredLocation, error) {
	locations, err := s.locationRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки локаций: %w", err)
	}

	// Сохраняем порядок запроса.
	locMap := make(map[string]models.Location, len(locations))
	for _, loc := range locations {
		locMap[loc.ID.String()] = loc
	}

	result := make([]scoredLocation, 0, len(ids))
	for _, id := range ids {
		if loc, ok := locMap[id]; ok {
			result = append(result, scoredLocation{location: loc, vibeScore: 0.5})
		}
	}
	return result, nil
}

// autoSelectLocations выполняет автоподбор локаций на основе vibe-скоринга и trip-фильтров.
// Алгоритм:
//  1. Вычислить merged vibe vector из vibe-векторов участников.
//  2. SearchNearest в Qdrant по location_vibes.
//  3. Загрузить полные данные локаций из PostgreSQL.
//  4. Отфильтровать по budget_tier и child_friendly.
//  5. Отранжировать по комбинированному score.
func (s *RouteService) autoSelectLocations(ctx context.Context, trip *models.Trip, members []models.TripMember, hasChildren bool, maxPoints int) ([]scoredLocation, error) {
	// Попытка vibe-скоринга через Qdrant.
	vibeScores := s.computeVibeScores(ctx, trip, members, maxPoints)

	// Загрузка локаций из PostgreSQL (с запасом для фильтрации).
	searchLimit := maxPoints * 3
	if searchLimit < 20 {
		searchLimit = 20
	}

	filter := &models.LocationFilter{
		PerPage: searchLimit,
		Page:    1,
	}

	// Фильтр по child_friendly, если есть дети.
	if hasChildren {
		childFriendly := true
		filter.ChildFriendly = &childFriendly
	}

	filter.NormalizePagination()
	locations, _, err := s.locationRepo.Search(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска локаций: %w", err)
	}

	// Фильтрация по бюджету.
	maxPrice := maxBudgetPerNightByTier[trip.BudgetTier]
	if maxPrice == 0 {
		maxPrice = 50000 // безлимит для неизвестного tier.
	}

	// Формирование scored-локаций с учётом vibe + budget.
	scored := make([]scoredLocation, 0, len(locations))
	for _, loc := range locations {
		// Фильтр по бюджету: пропускаем дорогие локации.
		if loc.PricePerNight > maxPrice {
			continue
		}

		// Определение vibe score.
		vs := float32(0.5) // базовый score.
		if score, ok := vibeScores[loc.ID.String()]; ok {
			vs = score
		}

		// Бонус за child_friendly при наличии детей.
		if hasChildren && loc.ChildFriendly {
			vs += 0.1
			if vs > 1.0 {
				vs = 1.0
			}
		}

		scored = append(scored, scoredLocation{location: loc, vibeScore: vs})
	}

	// Сортировка по vibe score (убывание).
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].vibeScore > scored[j].vibeScore
	})

	return scored, nil
}

// computeVibeScores вычисляет merged vibe vector и возвращает карту LocationID -> score.
// Merged vector — это взвешенное среднее vibe-векторов участников.
// Дети получают увеличенный вес в компонентах, связанных с child_friendly.
func (s *RouteService) computeVibeScores(ctx context.Context, trip *models.Trip, members []models.TripMember, limit int) map[string]float32 {
	if s.vibeRepo == nil {
		return nil
	}

	// Сбор вежторов участников.
	var vectors [][]float32
	var childCount int

	for _, m := range members {
		if m.VibeVectorID == nil {
			continue
		}
		vec, err := s.vibeRepo.GetVibeVector(ctx, database.CollectionUserVibes, *m.VibeVectorID)
		if err != nil {
			s.logger.Debug("пропуск вектора участника",
				zap.String("member_id", m.ID.String()),
				zap.Error(err),
			)
			continue
		}
		vectors = append(vectors, vec)
		if m.IsChild {
			childCount++
		}
	}

	// Попытка использовать vibe_vector_id из самого трипа (вектор создателя).
	if len(vectors) == 0 && trip.VibeVectorID != nil {
		vec, err := s.vibeRepo.GetVibeVector(ctx, database.CollectionUserVibes, *trip.VibeVectorID)
		if err == nil {
			vectors = append(vectors, vec)
		}
	}

	if len(vectors) == 0 {
		s.logger.Info("нет vibe-векторов, скоринг пропущен")
		return nil
	}

	// Вычисление merged vector (среднее арифметическое).
	merged := s.mergeVectors(vectors)

	// Поиск ближайших локаций через Qdrant.
	searchLimit := uint64(limit * 3)
	if searchLimit < 30 {
		searchLimit = 30
	}

	recommendations, err := s.vibeRepo.SearchNearest(ctx, database.CollectionLocationVibes, merged, searchLimit)
	if err != nil {
		s.logger.Warn("ошибка vibe-поиска в Qdrant, скоринг пропущен", zap.Error(err))
		return nil
	}

	// Формирование карты scores.
	scores := make(map[string]float32, len(recommendations))
	for _, rec := range recommendations {
		scores[rec.LocationID] = rec.Score
	}

	s.logger.Info("vibe-скоринг выполнен",
		zap.Int("vectors_merged", len(vectors)),
		zap.Int("children", childCount),
		zap.Int("scored_locations", len(scores)),
	)

	return scores
}

// mergeVectors вычисляет среднее арифметическое массива векторов.
// Результат нормализуется до единичной длины.
func (s *RouteService) mergeVectors(vectors [][]float32) []float32 {
	if len(vectors) == 0 {
		return nil
	}
	if len(vectors) == 1 {
		return vectors[0]
	}

	dim := len(vectors[0])
	merged := make([]float32, dim)
	count := float32(len(vectors))

	for _, vec := range vectors {
		for i := 0; i < dim && i < len(vec); i++ {
			merged[i] += vec[i]
		}
	}

	// Среднее.
	for i := range merged {
		merged[i] /= count
	}

	// Нормализация L2.
	var norm float64
	for _, v := range merged {
		norm += float64(v) * float64(v)
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range merged {
			merged[i] = float32(float64(merged[i]) / norm)
		}
	}

	return merged
}

// buildRoutePoints формирует массив RoutePoint из scored-локаций.
// Рассчитывает расстояния по Haversine, время в пути и target_audience.
func (s *RouteService) buildRoutePoints(scoredLocs []scoredLocation, speed float64, hasChildren bool) ([]models.RoutePoint, float64, int) {
	var points []models.RoutePoint
	var totalDistance float64
	var totalDuration int

	for i, sl := range scoredLocs {
		loc := sl.location

		point := models.RoutePoint{
			LocationID:      loc.ID.String(),
			Name:            loc.Name,
			Latitude:        loc.Latitude,
			Longitude:       loc.Longitude,
			Category:        loc.Category,
			PreviewImageURL: loc.PreviewImageURL,
			Order:           i + 1,
			StayDurationMin: getStayDuration(loc.Category),
			VibeScore:       sl.vibeScore,
			TargetAudience:  s.determineAudience(loc, hasChildren),
		}

		// Расчёт расстояния от предыдущей точки.
		if i > 0 {
			prev := points[i-1]
			dist := haversineDistance(prev.Latitude, prev.Longitude, loc.Latitude, loc.Longitude)
			durationMin := int(math.Ceil(dist / speed * 60))

			point.DistanceFromPrevKm = math.Round(dist*10) / 10
			point.DurationFromPrevMin = durationMin

			totalDistance += dist
			totalDuration += durationMin
		}

		totalDuration += point.StayDurationMin
		points = append(points, point)
	}

	return points, totalDistance, totalDuration
}

// determineAudience определяет целевую аудиторию точки маршрута.
// Если в группе есть дети, категории winery/gastro/extreme помечаются как adults,
// farm/beach/nature как children (подходят детям), остальные — all.
func (s *RouteService) determineAudience(loc models.Location, hasChildren bool) string {
	if !hasChildren {
		return models.AudienceAll
	}
	if adultOnlyCategories[loc.Category] {
		return models.AudienceAdults
	}
	if childFriendlyCategories[loc.Category] && loc.ChildFriendly {
		return models.AudienceChildren
	}
	return models.AudienceAll
}

// assignDaysAndSlots распределяет точки маршрута по дням и тайм-слотам.
// Каждый день вмещает до 3 точек (morning, afternoon, evening).
func (s *RouteService) assignDaysAndSlots(points []models.RoutePoint, tripDays int) {
	slots := []string{models.TimeSlotMorning, models.TimeSlotAfternoon, models.TimeSlotEvening}

	for i := range points {
		day := i/slotsPerDay + 1
		if day > tripDays {
			day = tripDays
		}
		slotIdx := i % slotsPerDay
		points[i].DayNumber = day
		points[i].TimeSlot = slots[slotIdx]
	}
}

// calculateTripDays вычисляет количество дней поездки на основе дат.
func (s *RouteService) calculateTripDays(trip *models.Trip) int {
	days := int(trip.DateTo.Sub(trip.DateFrom).Hours()/24) + 1
	if days < 1 {
		days = 1
	}
	return days
}

// estimateCost рассчитывает приблизительную стоимость маршрута.
// Формула: сумма price_per_night всех локаций * дни * group_size (упрощённо).
func (s *RouteService) estimateCost(scoredLocs []scoredLocation, days int, budgetTier string) int {
	var totalPerNight int
	for _, sl := range scoredLocs {
		totalPerNight += sl.location.PricePerNight
	}

	// Средняя стоимость за ночь для маршрута.
	if len(scoredLocs) > 0 {
		totalPerNight = totalPerNight / len(scoredLocs)
	}

	// Расчёт: средняя цена * количество дней.
	return totalPerNight * days
}

// generateSummary формирует текстовое описание маршрута.
func (s *RouteService) generateSummary(trip *models.Trip, points []models.RoutePoint, distanceKm float64, days int) string {
	// Подсчёт категорий.
	categories := make(map[string]int)
	for _, p := range points {
		if p.Category != "" {
			categories[p.Category]++
		}
	}

	// Топ-3 категории.
	type catCount struct {
		cat   string
		count int
	}
	var sorted []catCount
	for cat, cnt := range categories {
		sorted = append(sorted, catCount{cat, cnt})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	catStr := ""
	for i, cc := range sorted {
		if i >= 3 {
			break
		}
		if i > 0 {
			catStr += ", "
		}
		catStr += cc.cat
	}

	formatName := map[string]string{
		"day_trip":  "однодневный",
		"weekend":   "выходные",
		"multi_day": "многодневный",
	}
	fmtName := formatName[trip.Format]
	if fmtName == "" {
		fmtName = trip.Format
	}

	transportName := map[string]string{
		"car":    "на автомобиле",
		"public": "на общественном транспорте",
		"walk":   "пешком",
		"bike":   "на велосипеде",
	}
	trName := transportName[trip.Transport]
	if trName == "" {
		trName = trip.Transport
	}

	return fmt.Sprintf(
		"%s маршрут %s: %d точек, %.0f км, %d дней. Категории: %s. Бюджет: %s.",
		fmtName, trName, len(points), distanceKm, days, catStr, trip.BudgetTier,
	)
}

// persistRoute сохраняет маршрут и точки в БД.
func (s *RouteService) persistRoute(
	ctx context.Context,
	trip *models.Trip,
	userID uuid.UUID,
	name string,
	totalDistance float64,
	totalDuration int,
	estimatedCost int,
	summary string,
	points []models.RoutePoint,
) (string, error) {
	route := &models.Route{
		TripID:               trip.ID,
		UserID:               userID,
		Name:                 name,
		Status:               models.RouteStatusDraft,
		TotalDistanceKm:      math.Round(totalDistance*10) / 10,
		EstimatedDurationMin: totalDuration,
		EstimatedCostRub:     estimatedCost,
		Transport:            trip.Transport,
		PointsCount:          len(points),
		Summary:              summary,
	}

	// Конвертация RoutePoint -> RoutePointDB для сохранения.
	dbPoints := make([]models.RoutePointDB, len(points))
	for i, p := range points {
		locID, err := uuid.Parse(p.LocationID)
		if err != nil {
			s.logger.Warn("некорректный UUID локации, пропуск",
				zap.String("location_id", p.LocationID),
				zap.Error(err),
			)
			continue
		}
		dbPoints[i] = models.RoutePointDB{
			LocationID:          locID,
			Position:            p.Order,
			DayNumber:           p.DayNumber,
			TimeSlot:            p.TimeSlot,
			TargetAudience:      p.TargetAudience,
			StayDurationMin:     p.StayDurationMin,
			DistanceFromPrevKm:  p.DistanceFromPrevKm,
			DurationFromPrevMin: p.DurationFromPrevMin,
		}
	}

	created, err := s.routeRepo.Create(ctx, route, dbPoints)
	if err != nil {
		return "", err
	}

	return created.ID.String(), nil
}

// haversineDistance рассчитывает расстояние между двумя точками на сфере
// по формуле Haversine. Возвращает расстояние в километрах.
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(degreesToRadians(lat1))*math.Cos(degreesToRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// degreesToRadians конвертирует градусы в радианы.
func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

// getStayDuration возвращает рекомендуемое время пребывания в минутах
// на основе категории локации.
func getStayDuration(category string) int {
	if duration, ok := defaultStayDurationMin[category]; ok {
		return duration
	}
	return 60
}
