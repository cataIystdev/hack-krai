// Файл route.go реализует бизнес-логику построения маршрутов.
// RouteService принимает список ID локаций, загружает данные из PostgreSQL,
// рассчитывает расстояния между точками по формуле Haversine
// и формирует превью маршрута.
//
// Текущая реализация — demo-версия. Реальный routing через Neo4j
// будет реализован в фазе 7 (Live Routing).
package services

import (
	"context"
	"fmt"
	"math"

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

// RouteService — сервис построения маршрутов.
// Рассчитывает расстояния по формуле Haversine и формирует превью маршрута.
type RouteService struct {
	locationRepo *database.LocationRepository
	logger       *zap.Logger
}

// NewRouteService создаёт экземпляр сервиса маршрутов.
func NewRouteService(locationRepo *database.LocationRepository, logger *zap.Logger) *RouteService {
	return &RouteService{
		locationRepo: locationRepo,
		logger:       logger.Named("route_service"),
	}
}

// BuildRoute строит маршрут по списку ID локаций.
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
			Order:           len(points) + 1,
			StayDurationMin: getStayDuration(loc.Category),
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
