// Файл map.go реализует бизнес-логику для карты локаций.
// MapService отвечает за выборку точек для отображения маркеров на карте
// с поддержкой трёх режимов:
//   - bbox: пространственный поиск по bounding box через PostGIS
//   - demo: curated набор точек с предустановленными рекомендациями для 4 профилей
//   - hybrid: обогащение точек скорами рекомендаций из Qdrant (при наличии vibe-вектора)
package services

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
	"kudytudy-api/internal/models"
)

// MapService — сервис для работы с картой локаций.
// Предоставляет метод получения точек для отображения маркеров на карте
// в трёх режимах: bbox, demo и hybrid (с рекомендациями из Qdrant).
type MapService struct {
	locationRepo *database.LocationRepository
	vibeRepo     *database.VibeRepository
	logger       *zap.Logger
}

// NewMapService создаёт экземпляр сервиса карты.
// vibeRepo может быть nil — в этом случае режим рекомендаций недоступен.
func NewMapService(locationRepo *database.LocationRepository, vibeRepo *database.VibeRepository, logger *zap.Logger) *MapService {
	return &MapService{
		locationRepo: locationRepo,
		vibeRepo:     vibeRepo,
		logger:       logger.Named("map_service"),
	}
}

// GetMapLocations возвращает точки для отображения на карте.
// Поддерживает три режима:
//   - Demo (?demo=true): curated набор с предустановленными рекомендациями для выбранного профиля.
//   - Hybrid: если передан userID и у пользователя есть vibe-вектор, обогащает точки скорами.
//   - Bbox: стандартный PostGIS поиск.
func (s *MapService) GetMapLocations(ctx context.Context, filter *models.MapLocationFilter, userID *uuid.UUID) (*models.MapLocationsResponse, error) {
	// Валидация фильтра.
	if errMsg := filter.Validate(); errMsg != "" {
		return nil, &ValidationError{Message: errMsg}
	}

	// Demo-режим: curated набор с предустановленными рекомендациями.
	if filter.IsDemo() {
		return s.getDemoLocations(ctx, filter)
	}

	// Стандартный режим: PostGIS поиск.
	points, err := s.locationRepo.SearchForMap(ctx, filter)
	if err != nil {
		s.logger.Error("ошибка получения точек карты", zap.Error(err))
		return nil, err
	}

	// Гарантируем непустой массив в JSON ([] вместо null).
	if points == nil {
		points = []models.MapPoint{}
	}

	// Hybrid: обогащение рекомендациями, если есть vibe-вектор пользователя.
	if userID != nil && s.vibeRepo != nil {
		points = s.enrichWithRecommendations(ctx, points, *userID)
	}

	return &models.MapLocationsResponse{
		Points: points,
		Total:  len(points),
	}, nil
}

// enrichWithRecommendations обогащает точки карты скорами рекомендаций из Qdrant.
// Если у пользователя нет vibe-вектора или Qdrant недоступен, возвращает точки без изменений.
func (s *MapService) enrichWithRecommendations(ctx context.Context, points []models.MapPoint, userID uuid.UUID) []models.MapPoint {
	userVector, err := s.vibeRepo.GetVibeVector(ctx, database.CollectionUserVibes, userID)
	if err != nil {
		s.logger.Debug("вектор пользователя не найден, пропускаем обогащение рекомендациями",
			zap.String("user_id", userID.String()),
		)
		return points
	}

	// Поиск ближайших локаций к вектору пользователя.
	results, err := s.vibeRepo.SearchNearest(ctx, database.CollectionLocationVibes, userVector, 50)
	if err != nil {
		s.logger.Warn("ошибка поиска рекомендаций в Qdrant, продолжаем без обогащения",
			zap.Error(err),
		)
		return points
	}

	// Построение маппинга location_id -> score.
	scoreMap := make(map[string]float32, len(results))
	for _, r := range results {
		scoreMap[r.LocationID] = r.Score
	}

	// Обогащение точек скорами.
	for i := range points {
		if score, ok := scoreMap[points[i].ID]; ok {
			points[i].IsRecommended = true
			points[i].RecommendationScore = score
		}
	}

	// Сортировка: рекомендованные первыми, затем по скору.
	sort.Slice(points, func(i, j int) bool {
		if points[i].IsRecommended != points[j].IsRecommended {
			return points[i].IsRecommended
		}
		return points[i].RecommendationScore > points[j].RecommendationScore
	})

	return points
}

// getDemoLocations возвращает curated набор точек с предустановленными рекомендациями.
// Все опубликованные локации возвращаются, а для выбранного demo-профиля
// топ-N помечаются как рекомендованные с curated скорами.
func (s *MapService) getDemoLocations(ctx context.Context, filter *models.MapLocationFilter) (*models.MapLocationsResponse, error) {
	// Получаем все опубликованные локации.
	allFilter := &models.MapLocationFilter{Limit: 200}
	points, err := s.locationRepo.SearchForMap(ctx, allFilter)
	if err != nil {
		return nil, err
	}

	if points == nil {
		points = []models.MapPoint{}
	}

	profile := filter.GetProfile()
	profileConfig := getDemoProfile(profile)

	// Применяем curated рекомендации из профиля.
	slugMap := make(map[string]int, len(points))
	for i, p := range points {
		slugMap[p.ID] = i
	}

	// Ищем локации по slug (description_short как fallback для идентификации).
	// Используем name для сопоставления с профилем.
	nameMap := make(map[string]int, len(points))
	for i, p := range points {
		nameMap[p.Name] = i
	}

	for _, rec := range profileConfig.Recommendations {
		if idx, ok := nameMap[rec.Name]; ok {
			points[idx].IsRecommended = true
			points[idx].RecommendationScore = rec.Score
		}
	}

	// Сортировка: рекомендованные первыми.
	sort.Slice(points, func(i, j int) bool {
		if points[i].IsRecommended != points[j].IsRecommended {
			return points[i].IsRecommended
		}
		return points[i].RecommendationScore > points[j].RecommendationScore
	})

	return &models.MapLocationsResponse{
		Points:  points,
		Total:   len(points),
		Profile: profile,
	}, nil
}

// demoRecommendation — curated рекомендация для demo-профиля.
type demoRecommendation struct {
	// Name — название локации для сопоставления с seed-данными.
	Name string
	// Score — curated скор рекомендации (0.0-1.0).
	Score float32
}

// demoProfile — конфигурация demo-профиля с описанием и рекомендациями.
type demoProfile struct {
	// Description — человекочитаемое описание профиля.
	Description string
	// Recommendations — curated список рекомендаций для профиля.
	Recommendations []demoRecommendation
}

// getDemoProfile возвращает конфигурацию demo-профиля по имени.
// Каждый профиль содержит curated набор рекомендаций с релевантными скорами,
// привязанных к seed-локациям из seed.go.
func getDemoProfile(name string) demoProfile {
	profiles := map[string]demoProfile{
		// Профиль 1: Тишина, горы, вино — спокойный интроверт.
		"calm_wine_mountains": {
			Description: "Спокойный отдых: горы, вино, тишина",
			Recommendations: []demoRecommendation{
				{Name: "Винодельня Лефкадия", Score: 0.96},
				{Name: "Козья ферма дяди Вани", Score: 0.93},
				{Name: "Винодельня Абрау-Дюрсо", Score: 0.91},
				{Name: "Хребет Грачёв и дольмены", Score: 0.88},
				{Name: "Озеро Кардывач", Score: 0.85},
				{Name: "Сыроварня Марии Коваленко", Score: 0.82},
				{Name: "Тихая заводь реки Пшеха", Score: 0.78},
			},
		},
		// Профиль 2: Адреналин, экстрим, природа — активный путешественник.
		"active_adventure": {
			Description: "Активный отдых: каньоны, рафтинг, горы",
			Recommendations: []demoRecommendation{
				{Name: "Каньон реки Белой", Score: 0.97},
				{Name: "Водопады Руфабго", Score: 0.94},
				{Name: "Хребет Грачёв и дольмены", Score: 0.92},
				{Name: "Озеро Кардывач", Score: 0.89},
				{Name: "Подводное погружение в Чёрном море", Score: 0.86},
				{Name: "Конная база Псебай", Score: 0.83},
				{Name: "Ущелье Гуамка", Score: 0.80},
			},
		},
		// Профиль 3: Семейный отдых, дети, фермы — семья с детьми.
		"family_kids": {
			Description: "Семейный отдых: фермы, дети, природа",
			Recommendations: []demoRecommendation{
				{Name: "Козья ферма дяди Вани", Score: 0.95},
				{Name: "Сыроварня Марии Коваленко", Score: 0.93},
				{Name: "Парк Галицкого", Score: 0.91},
				{Name: "Водопады Руфабго", Score: 0.87},
				{Name: "Старый парк Кабардинки", Score: 0.84},
				{Name: "Конная база Псебай", Score: 0.81},
				{Name: "Винодельня Лефкадия", Score: 0.76},
			},
		},
		// Профиль 4: Гастрономия, культура, вино — гурман и ценитель.
		"gastro_cultural": {
			Description: "Гастрономия и культура: вино, сыр, история",
			Recommendations: []demoRecommendation{
				{Name: "Сыроварня Марии Коваленко", Score: 0.97},
				{Name: "Винодельня Абрау-Дюрсо", Score: 0.95},
				{Name: "Винодельня Лефкадия", Score: 0.93},
				{Name: "Козья ферма дяди Вани", Score: 0.90},
				{Name: "Старый парк Кабардинки", Score: 0.87},
				{Name: "Парк Галицкого", Score: 0.83},
				{Name: "Ущелье Гуамка", Score: 0.79},
			},
		},
	}

	if p, ok := profiles[name]; ok {
		return p
	}
	return profiles["calm_wine_mountains"]
}

// ValidationError — ошибка валидации входных данных.
type ValidationError struct {
	Message string
}

// Error реализует интерфейс error.
func (e *ValidationError) Error() string {
	return e.Message
}
