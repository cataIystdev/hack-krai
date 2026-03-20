// Файл location_repository.go реализует репозиторий для работы с таблицей locations.
// Предоставляет методы CRUD и пространственного поиска PostGIS:
// Create (ST_SetSRID + ST_MakePoint), FindByID/FindBySlug (ST_X/ST_Y),
// Update (динамические SET), Delete, Search (ST_DWithin, ST_Within + ST_MakeEnvelope).
package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
)

// ErrLocationNotFound — ошибка при отсутствии локации в БД.
var ErrLocationNotFound = errors.New("локация не найдена")

// ErrLocationSlugExists — ошибка при дублировании slug.
var ErrLocationSlugExists = errors.New("локация с таким slug уже существует")

// ErrLocationForbidden — ошибка при попытке изменения чужой локации.
var ErrLocationForbidden = errors.New("нет прав на изменение этой локации")

// locationColumns — список колонок для SELECT-запросов.
// Координаты извлекаются через ST_X(geo) и ST_Y(geo).
const locationColumns = `
	id, owner_id, slug, name, description_short, description_full,
	category, tags, price_per_night, capacity, access_level, density_level,
	child_friendly, splat_url, vibe_vector_id, address, is_published,
	ST_X(geo) AS longitude, ST_Y(geo) AS latitude,
	created_at, updated_at
`

// LocationRepository — репозиторий для работы с таблицей locations.
// Инкапсулирует SQL-запросы с пространственными функциями PostGIS.
type LocationRepository struct {
	pg     *PostgresClient
	logger *zap.Logger
}

// NewLocationRepository создаёт экземпляр репозитория локаций.
func NewLocationRepository(pg *PostgresClient, logger *zap.Logger) *LocationRepository {
	return &LocationRepository{
		pg:     pg,
		logger: logger,
	}
}

// scanLocation выполняет маппинг строки результата на структуру Location.
func scanLocation(row pgx.Row) (*models.Location, error) {
	var loc models.Location
	err := row.Scan(
		&loc.ID, &loc.OwnerID, &loc.Slug, &loc.Name,
		&loc.DescriptionShort, &loc.DescriptionFull,
		&loc.Category, &loc.Tags, &loc.PricePerNight, &loc.Capacity,
		&loc.AccessLevel, &loc.DensityLevel, &loc.ChildFriendly,
		&loc.SplatURL, &loc.VibeVectorID, &loc.Address, &loc.IsPublished,
		&loc.Longitude, &loc.Latitude,
		&loc.CreatedAt, &loc.UpdatedAt,
	)
	return &loc, err
}

// Create создаёт новую локацию в БД.
// Координаты вставляются через ST_SetSRID(ST_MakePoint(lon, lat), 4326).
func (r *LocationRepository) Create(ctx context.Context, loc *models.Location) (*models.Location, error) {
	query := `
		INSERT INTO locations (
			owner_id, slug, name, description_short, description_full,
			category, tags, price_per_night, capacity, access_level, density_level,
			child_friendly, address, is_published,
			geo
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11,
			$12, $13, $14,
			ST_SetSRID(ST_MakePoint($15, $16), 4326)
		)
		RETURNING ` + locationColumns

	row := r.pg.Pool.QueryRow(ctx, query,
		loc.OwnerID, loc.Slug, loc.Name,
		loc.DescriptionShort, loc.DescriptionFull,
		loc.Category, loc.Tags, loc.PricePerNight, loc.Capacity,
		loc.AccessLevel, loc.DensityLevel, loc.ChildFriendly,
		loc.Address, loc.IsPublished,
		loc.Longitude, loc.Latitude,
	)

	created, err := scanLocation(row)
	if err != nil {
		if contains(err.Error(), "duplicate key") && contains(err.Error(), "slug") {
			return nil, ErrLocationSlugExists
		}
		r.logger.Error("ошибка создания локации",
			zap.String("name", loc.Name),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка создания локации: %w", err)
	}

	r.logger.Info("локация создана",
		zap.String("location_id", created.ID.String()),
		zap.String("slug", created.Slug),
	)

	return created, nil
}

// FindByID находит локацию по UUID.
func (r *LocationRepository) FindByID(ctx context.Context, id string) (*models.Location, error) {
	query := `SELECT ` + locationColumns + ` FROM locations WHERE id = $1`
	row := r.pg.Pool.QueryRow(ctx, query, id)
	loc, err := scanLocation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLocationNotFound
		}
		r.logger.Error("ошибка поиска локации по ID",
			zap.String("id", id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка поиска локации: %w", err)
	}
	return loc, nil
}

// FindBySlug находит локацию по slug.
func (r *LocationRepository) FindBySlug(ctx context.Context, slug string) (*models.Location, error) {
	query := `SELECT ` + locationColumns + ` FROM locations WHERE slug = $1`
	row := r.pg.Pool.QueryRow(ctx, query, slug)
	loc, err := scanLocation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLocationNotFound
		}
		r.logger.Error("ошибка поиска локации по slug",
			zap.String("slug", slug),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка поиска локации: %w", err)
	}
	return loc, nil
}

// Update обновляет локацию. Только владелец (owner_id) может обновить запись.
// Строит динамический SET на основе переданных полей.
func (r *LocationRepository) Update(ctx context.Context, id, ownerID string, req *models.UpdateLocationRequest) (*models.Location, error) {
	// Проверка владельца.
	var actualOwnerID string
	err := r.pg.Pool.QueryRow(ctx, `SELECT owner_id FROM locations WHERE id = $1`, id).Scan(&actualOwnerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLocationNotFound
		}
		return nil, fmt.Errorf("ошибка проверки владельца: %w", err)
	}
	if actualOwnerID != ownerID {
		return nil, ErrLocationForbidden
	}

	// Построение динамического SET.
	setClauses := []string{"updated_at = NOW()"}
	args := []any{}
	argIdx := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.DescriptionShort != nil {
		setClauses = append(setClauses, fmt.Sprintf("description_short = $%d", argIdx))
		args = append(args, *req.DescriptionShort)
		argIdx++
	}
	if req.DescriptionFull != nil {
		setClauses = append(setClauses, fmt.Sprintf("description_full = $%d", argIdx))
		args = append(args, *req.DescriptionFull)
		argIdx++
	}
	if req.Category != nil {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *req.Category)
		argIdx++
	}
	if req.Tags != nil {
		setClauses = append(setClauses, fmt.Sprintf("tags = $%d", argIdx))
		args = append(args, req.Tags)
		argIdx++
	}
	if req.PricePerNight != nil {
		setClauses = append(setClauses, fmt.Sprintf("price_per_night = $%d", argIdx))
		args = append(args, *req.PricePerNight)
		argIdx++
	}
	if req.Capacity != nil {
		setClauses = append(setClauses, fmt.Sprintf("capacity = $%d", argIdx))
		args = append(args, *req.Capacity)
		argIdx++
	}
	if req.AccessLevel != nil {
		setClauses = append(setClauses, fmt.Sprintf("access_level = $%d", argIdx))
		args = append(args, *req.AccessLevel)
		argIdx++
	}
	if req.DensityLevel != nil {
		setClauses = append(setClauses, fmt.Sprintf("density_level = $%d", argIdx))
		args = append(args, *req.DensityLevel)
		argIdx++
	}
	if req.ChildFriendly != nil {
		setClauses = append(setClauses, fmt.Sprintf("child_friendly = $%d", argIdx))
		args = append(args, *req.ChildFriendly)
		argIdx++
	}
	if req.Address != nil {
		setClauses = append(setClauses, fmt.Sprintf("address = $%d", argIdx))
		args = append(args, *req.Address)
		argIdx++
	}
	if req.IsPublished != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_published = $%d", argIdx))
		args = append(args, *req.IsPublished)
		argIdx++
	}
	if req.Latitude != nil && req.Longitude != nil {
		setClauses = append(setClauses, fmt.Sprintf("geo = ST_SetSRID(ST_MakePoint($%d, $%d), 4326)", argIdx, argIdx+1))
		args = append(args, *req.Longitude, *req.Latitude)
		argIdx += 2
	}

	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE locations SET %s WHERE id = $%d RETURNING %s`,
		strings.Join(setClauses, ", "), argIdx, locationColumns,
	)

	row := r.pg.Pool.QueryRow(ctx, query, args...)
	updated, err := scanLocation(row)
	if err != nil {
		r.logger.Error("ошибка обновления локации",
			zap.String("id", id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("ошибка обновления локации: %w", err)
	}

	r.logger.Info("локация обновлена",
		zap.String("location_id", updated.ID.String()),
		zap.String("slug", updated.Slug),
	)

	return updated, nil
}

// Delete удаляет локацию. Только владелец (owner_id) может удалить запись.
func (r *LocationRepository) Delete(ctx context.Context, id, ownerID string) error {
	// Проверка владельца.
	var actualOwnerID string
	err := r.pg.Pool.QueryRow(ctx, `SELECT owner_id FROM locations WHERE id = $1`, id).Scan(&actualOwnerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLocationNotFound
		}
		return fmt.Errorf("ошибка проверки владельца: %w", err)
	}
	if actualOwnerID != ownerID {
		return ErrLocationForbidden
	}

	tag, err := r.pg.Pool.Exec(ctx, `DELETE FROM locations WHERE id = $1 AND owner_id = $2`, id, ownerID)
	if err != nil {
		r.logger.Error("ошибка удаления локации",
			zap.String("id", id),
			zap.Error(err),
		)
		return fmt.Errorf("ошибка удаления локации: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrLocationNotFound
	}

	r.logger.Info("локация удалена", zap.String("location_id", id))
	return nil
}

// Search выполняет пространственный поиск локаций с фильтрами.
// Поддерживает: поиск по радиусу (ST_DWithin), по bbox (ST_Within + ST_MakeEnvelope),
// фильтрацию по category, child_friendly, density_level.
// Возвращает список локаций и общее количество записей.
func (r *LocationRepository) Search(ctx context.Context, filter *models.LocationFilter) ([]models.Location, int, error) {
	filter.NormalizePagination()

	whereClauses := []string{"is_published = true"}
	args := []any{}
	argIdx := 1

	// Пространственный поиск: радиус (ST_DWithin).
	if filter.HasRadiusSearch() {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"ST_DWithin(geo::geography, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography, $%d)",
			argIdx, argIdx+1, argIdx+2,
		))
		args = append(args, *filter.Lon, *filter.Lat, *filter.RadiusKm*1000) // км -> метры
		argIdx += 3
	}

	// Пространственный поиск: bbox (ST_Within + ST_MakeEnvelope).
	if filter.HasBBoxSearch() {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"ST_Within(geo, ST_MakeEnvelope($%d, $%d, $%d, $%d, 4326))",
			argIdx, argIdx+1, argIdx+2, argIdx+3,
		))
		args = append(args, *filter.MinLon, *filter.MinLat, *filter.MaxLon, *filter.MaxLat)
		argIdx += 4
	}

	// Фильтр по категории.
	if filter.Category != nil && *filter.Category != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *filter.Category)
		argIdx++
	}

	// Фильтр по child_friendly.
	if filter.ChildFriendly != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("child_friendly = $%d", argIdx))
		args = append(args, *filter.ChildFriendly)
		argIdx++
	}

	// Фильтр по density_level.
	if filter.DensityLevel != nil && *filter.DensityLevel != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("density_level = $%d", argIdx))
		args = append(args, *filter.DensityLevel)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Подсчёт общего количества записей.
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM locations WHERE %s`, whereSQL)
	var total int
	err := r.pg.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("ошибка подсчёта локаций", zap.Error(err))
		return nil, 0, fmt.Errorf("ошибка подсчёта локаций: %w", err)
	}

	// Сортировка: по расстоянию если радиус, иначе по дате создания.
	orderBy := "created_at DESC"
	if filter.HasRadiusSearch() {
		orderBy = fmt.Sprintf(
			"geo <-> ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geometry",
			argIdx, argIdx+1,
		)
		args = append(args, *filter.Lon, *filter.Lat)
		argIdx += 2
	}

	// Основной запрос с пагинацией.
	args = append(args, filter.PerPage, filter.Offset())
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM locations WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		locationColumns, whereSQL, orderBy, argIdx, argIdx+1,
	)

	rows, err := r.pg.Pool.Query(ctx, dataQuery, args...)
	if err != nil {
		r.logger.Error("ошибка поиска локаций", zap.Error(err))
		return nil, 0, fmt.Errorf("ошибка поиска локаций: %w", err)
	}
	defer rows.Close()

	var locations []models.Location
	for rows.Next() {
		var loc models.Location
		err := rows.Scan(
			&loc.ID, &loc.OwnerID, &loc.Slug, &loc.Name,
			&loc.DescriptionShort, &loc.DescriptionFull,
			&loc.Category, &loc.Tags, &loc.PricePerNight, &loc.Capacity,
			&loc.AccessLevel, &loc.DensityLevel, &loc.ChildFriendly,
			&loc.SplatURL, &loc.VibeVectorID, &loc.Address, &loc.IsPublished,
			&loc.Longitude, &loc.Latitude,
			&loc.CreatedAt, &loc.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("ошибка сканирования строки локации", zap.Error(err))
			return nil, 0, fmt.Errorf("ошибка сканирования: %w", err)
		}
		locations = append(locations, loc)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("ошибка итерации: %w", err)
	}

	return locations, total, nil
}
