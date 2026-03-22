// Файл route_repository.go реализует слой доступа к данным для таблиц routes и route_points.
// Предоставляет методы: транзакционное создание маршрута с точками,
// поиск маршрута по ID с загрузкой точек, поиск маршрутов по trip_id.
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

// Ошибки репозитория маршрутов.
var (
	// ErrRouteNotFound — маршрут не найден по указанному ID.
	ErrRouteNotFound = errors.New("маршрут не найден")
)

// RouteRepository — репозиторий для работы с таблицами routes и route_points.
type RouteRepository struct {
	pg     *PostgresClient
	logger *zap.Logger
}

// NewRouteRepository создаёт экземпляр репозитория маршрутов.
func NewRouteRepository(pg *PostgresClient, logger *zap.Logger) *RouteRepository {
	return &RouteRepository{
		pg:     pg,
		logger: logger.Named("route_repo"),
	}
}

// Create транзакционно создаёт маршрут и его точки в БД.
// Сначала вставляет запись в routes, затем все точки в route_points.
// При ошибке вся транзакция откатывается.
func (r *RouteRepository) Create(ctx context.Context, route *models.Route, points []models.RoutePointDB) (*models.Route, error) {
	tx, err := r.pg.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	// Вставка маршрута.
	routeQuery := `
		INSERT INTO routes (trip_id, user_id, name, status, total_distance_km,
			estimated_duration_min, estimated_cost_rub, transport, points_count, summary)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, trip_id, user_id, name, status, total_distance_km,
			estimated_duration_min, estimated_cost_rub, transport, points_count, summary, created_at
	`

	var created models.Route
	err = tx.QueryRow(ctx, routeQuery,
		route.TripID, route.UserID, route.Name, route.Status,
		route.TotalDistanceKm, route.EstimatedDurationMin, route.EstimatedCostRub,
		route.Transport, route.PointsCount, route.Summary,
	).Scan(
		&created.ID, &created.TripID, &created.UserID, &created.Name,
		&created.Status, &created.TotalDistanceKm, &created.EstimatedDurationMin,
		&created.EstimatedCostRub, &created.Transport, &created.PointsCount,
		&created.Summary, &created.CreatedAt,
	)
	if err != nil {
		r.logger.Error("ошибка создания маршрута",
			zap.String("trip_id", route.TripID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка создания маршрута: %w", err)
	}

	// Вставка точек маршрута.
	for _, point := range points {
		pointQuery := `
			INSERT INTO route_points (route_id, location_id, position, day_number,
				time_slot, target_audience, stay_duration_min,
				distance_from_prev_km, duration_from_prev_min, audio_story_url,
				story_text, story_debug_info, weather_condition, weather_temp_c)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`
		_, err = tx.Exec(ctx, pointQuery,
			created.ID, point.LocationID, point.Position, point.DayNumber,
			point.TimeSlot, point.TargetAudience, point.StayDurationMin,
			point.DistanceFromPrevKm, point.DurationFromPrevMin,
			point.AudioStoryURL, point.StoryText, point.StoryDebugInfo, point.WeatherCondition, point.WeatherTempC,
		)
		if err != nil {
			r.logger.Error("ошибка вставки точки маршрута",
				zap.String("route_id", created.ID.String()),
				zap.String("location_id", point.LocationID.String()),
				zap.Error(err),
			)
			return nil, fmt.Errorf("ошибка вставки точки маршрута: %w", err)
		}
	}

	// Фиксация транзакции.
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	r.logger.Info("маршрут создан",
		zap.String("route_id", created.ID.String()),
		zap.String("trip_id", created.TripID.String()),
		zap.Int("points_count", len(points)),
	)

	return &created, nil
}

// FindByID находит маршрут по UUID.
func (r *RouteRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Route, error) {
	query := `
		SELECT id, trip_id, user_id, name, status, total_distance_km,
			estimated_duration_min, estimated_cost_rub, transport, points_count, summary, created_at
		FROM routes WHERE id = $1
	`

	var route models.Route
	err := r.pg.Pool.QueryRow(ctx, query, id).Scan(
		&route.ID, &route.TripID, &route.UserID, &route.Name,
		&route.Status, &route.TotalDistanceKm, &route.EstimatedDurationMin,
		&route.EstimatedCostRub, &route.Transport, &route.PointsCount,
		&route.Summary, &route.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRouteNotFound
		}
		return nil, fmt.Errorf("ошибка поиска маршрута: %w", err)
	}

	return &route, nil
}

// FindPointsByRouteID возвращает точки маршрута, отсортированные по position.
func (r *RouteRepository) FindPointsByRouteID(ctx context.Context, routeID uuid.UUID) ([]models.RoutePointDB, error) {
	query := `
		SELECT id, route_id, location_id, position, day_number, time_slot,
			target_audience, stay_duration_min, distance_from_prev_km, duration_from_prev_min,
			COALESCE(audio_story_url, ''), COALESCE(story_text, ''), COALESCE(story_debug_info, ''), COALESCE(weather_condition, ''), weather_temp_c
		FROM route_points
		WHERE route_id = $1
		ORDER BY position ASC
	`

	rows, err := r.pg.Pool.Query(ctx, query, routeID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения точек маршрута: %w", err)
	}
	defer rows.Close()

	var points []models.RoutePointDB
	for rows.Next() {
		var p models.RoutePointDB
		if err := rows.Scan(
			&p.ID, &p.RouteID, &p.LocationID, &p.Position,
			&p.DayNumber, &p.TimeSlot, &p.TargetAudience,
			&p.StayDurationMin, &p.DistanceFromPrevKm, &p.DurationFromPrevMin,
			&p.AudioStoryURL, &p.StoryText, &p.StoryDebugInfo, &p.WeatherCondition, &p.WeatherTempC,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования точки маршрута: %w", err)
		}
		points = append(points, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации точек маршрута: %w", err)
	}

	return points, nil
}

// ReplacePoints полностью заменяет точки маршрута новым набором.
func (r *RouteRepository) ReplacePoints(ctx context.Context, routeID uuid.UUID, points []models.RoutePointDB) error {
	tx, err := r.pg.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции замены точек маршрута: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM route_points WHERE route_id = $1`, routeID); err != nil {
		return fmt.Errorf("ошибка удаления старых точек маршрута: %w", err)
	}

	for _, point := range points {
		_, err := tx.Exec(ctx, `
			INSERT INTO route_points (
				route_id, location_id, position, day_number, time_slot, target_audience,
				stay_duration_min, distance_from_prev_km, duration_from_prev_min,
				audio_story_url, story_text, story_debug_info, weather_condition, weather_temp_c
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`,
			routeID, point.LocationID, point.Position, point.DayNumber, point.TimeSlot,
			point.TargetAudience, point.StayDurationMin, point.DistanceFromPrevKm,
			point.DurationFromPrevMin, point.AudioStoryURL, point.StoryText,
			point.StoryDebugInfo, point.WeatherCondition, point.WeatherTempC,
		)
		if err != nil {
			return fmt.Errorf("ошибка вставки новой точки маршрута: %w", err)
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE routes SET points_count = $1 WHERE id = $2`, len(points), routeID); err != nil {
		return fmt.Errorf("ошибка обновления points_count: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка commit замены route_points: %w", err)
	}

	return nil
}

// UpdatePointStory обновляет storytelling-артефакт конкретной точки маршрута.
func (r *RouteRepository) UpdatePointStory(ctx context.Context, routeID, pointID uuid.UUID, audioURL, storyText, debugInfo string) error {
	tag, err := r.pg.Pool.Exec(ctx, `
		UPDATE route_points
		SET audio_story_url = $1, story_text = $2, story_debug_info = $3
		WHERE route_id = $4 AND id = $5
	`, audioURL, storyText, debugInfo, routeID, pointID)
	if err != nil {
		return fmt.Errorf("ошибка обновления story asset: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrRouteNotFound
	}
	return nil
}

// UpdateRouteSummaryAndMetrics обновляет итоговые метрики маршрута после rebuild.
func (r *RouteRepository) UpdateRouteSummaryAndMetrics(
	ctx context.Context,
	routeID uuid.UUID,
	totalDistance float64,
	totalDuration int,
	estimatedCost int,
	summary string,
) error {
	tag, err := r.pg.Pool.Exec(ctx, `
		UPDATE routes
		SET total_distance_km = $1,
		    estimated_duration_min = $2,
		    estimated_cost_rub = $3,
		    summary = $4
		WHERE id = $5
	`, totalDistance, totalDuration, estimatedCost, summary, routeID)
	if err != nil {
		return fmt.Errorf("ошибка обновления route summary: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrRouteNotFound
	}
	return nil
}

// FindByTripID возвращает все маршруты поездки, отсортированные по дате создания.
func (r *RouteRepository) FindByTripID(ctx context.Context, tripID uuid.UUID) ([]models.Route, error) {
	query := `
		SELECT id, trip_id, user_id, name, status, total_distance_km,
			estimated_duration_min, estimated_cost_rub, transport, points_count, summary, created_at
		FROM routes
		WHERE trip_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pg.Pool.Query(ctx, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения маршрутов поездки: %w", err)
	}
	defer rows.Close()

	var routes []models.Route
	for rows.Next() {
		var route models.Route
		if err := rows.Scan(
			&route.ID, &route.TripID, &route.UserID, &route.Name,
			&route.Status, &route.TotalDistanceKm, &route.EstimatedDurationMin,
			&route.EstimatedCostRub, &route.Transport, &route.PointsCount,
			&route.Summary, &route.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования маршрута: %w", err)
		}
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации маршрутов: %w", err)
	}

	return routes, nil
}
