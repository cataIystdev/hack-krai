// Файл vibe_repository.go реализует репозиторий для работы с Qdrant-векторами
// и таблицей swipe_scenes в PostgreSQL. Предоставляет методы для CRUD-операций
// со сценами и операций с векторами (upsert, get, search) в Qdrant.
package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	pb "github.com/qdrant/go-client/qdrant"
	"go.uber.org/zap"

	"deep-krai-api/internal/models"
)

// Константы коллекций Qdrant.
const (
	// CollectionUserVibes — коллекция vibe-профилей пользователей.
	CollectionUserVibes = "user_vibes"

	// CollectionLocationVibes — коллекция vibe-профилей локаций.
	CollectionLocationVibes = "location_vibes"
)

// ErrSceneNotFound — ошибка, возвращаемая при отсутствии сцены свайпа.
var ErrSceneNotFound = errors.New("сцена свайпа не найдена")

// ErrVectorNotFound — ошибка, возвращаемая при отсутствии вектора в Qdrant.
var ErrVectorNotFound = errors.New("вектор не найден в Qdrant")

// VibeRepository — репозиторий для работы с vibe-векторами и сценами свайпа.
type VibeRepository struct {
	pg     *PostgresClient
	qdrant *QdrantClient
	logger *zap.Logger
}

// NewVibeRepository создаёт репозиторий vibe-профилирования.
func NewVibeRepository(pg *PostgresClient, qdrant *QdrantClient, logger *zap.Logger) *VibeRepository {
	return &VibeRepository{
		pg:     pg,
		qdrant: qdrant,
		logger: logger.Named("vibe_repo"),
	}
}

// GetAllScenes возвращает все сцены свайпа, отсортированные по display_order.
func (r *VibeRepository) GetAllScenes(ctx context.Context) ([]models.SwipeScene, error) {
	query := `
		SELECT id, title, description, image_url, display_order, created_at
		FROM swipe_scenes
		ORDER BY display_order ASC
	`

	rows, err := r.pg.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сцен свайпа: %w", err)
	}
	defer rows.Close()

	var scenes []models.SwipeScene
	for rows.Next() {
		var scene models.SwipeScene
		if err := rows.Scan(
			&scene.ID, &scene.Title, &scene.Description,
			&scene.ImageURL, &scene.DisplayOrder, &scene.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования сцены: %w", err)
		}
		scenes = append(scenes, scene)
	}

	return scenes, nil
}

// GetSceneByID возвращает сцену свайпа по UUID.
func (r *VibeRepository) GetSceneByID(ctx context.Context, id string) (*models.SwipeScene, error) {
	query := `
		SELECT id, title, description, image_url, display_order, created_at
		FROM swipe_scenes
		WHERE id = $1
	`

	var scene models.SwipeScene
	err := r.pg.Pool.QueryRow(ctx, query, id).Scan(
		&scene.ID, &scene.Title, &scene.Description,
		&scene.ImageURL, &scene.DisplayOrder, &scene.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSceneNotFound
		}
		return nil, fmt.Errorf("ошибка получения сцены: %w", err)
	}

	return &scene, nil
}

// UpsertVibeVector выполняет upsert вектора в коллекцию Qdrant.
// Если точка с данным ID существует — обновляет вектор и payload.
// Если нет — создаёт новую точку.
// Параметры:
//   - collection: имя коллекции (user_vibes / location_vibes).
//   - pointID: UUID точки (user_id или location_id).
//   - vector: вектор 384 измерений.
//   - payload: дополнительные данные (имя, теги и т.д.).
func (r *VibeRepository) UpsertVibeVector(ctx context.Context, collection string, pointID uuid.UUID, vector []float32, payload map[string]any) error {
	r.logger.Info("upsert вектора в Qdrant",
		zap.String("collection", collection),
		zap.String("point_id", pointID.String()),
		zap.Int("vector_size", len(vector)),
	)

	// Конвертация UUID в строку для Qdrant.
	idStr := pointID.String()

	// Конвертация payload в формат Qdrant.
	var qdrantPayload map[string]*pb.Value
	if payload != nil {
		qdrantPayload = make(map[string]*pb.Value)
		for k, v := range payload {
			switch val := v.(type) {
			case string:
				qdrantPayload[k] = &pb.Value{
					Kind: &pb.Value_StringValue{StringValue: val},
				}
			case float64:
				qdrantPayload[k] = &pb.Value{
					Kind: &pb.Value_DoubleValue{DoubleValue: val},
				}
			case bool:
				qdrantPayload[k] = &pb.Value{
					Kind: &pb.Value_BoolValue{BoolValue: val},
				}
			case int:
				qdrantPayload[k] = &pb.Value{
					Kind: &pb.Value_IntegerValue{IntegerValue: int64(val)},
				}
			}
		}
	}

	// Корректировка для Qdrant: float64 -> float32.
	vectors := make([]float32, len(vector))
	copy(vectors, vector)

	_, err := r.qdrant.Points.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: collection,
		Points: []*pb.PointStruct{
			{
				Id: &pb.PointId{
					PointIdOptions: &pb.PointId_Uuid{Uuid: idStr},
				},
				Vectors: &pb.Vectors{
					VectorsOptions: &pb.Vectors_Vector{
						Vector: &pb.Vector{Data: vectors},
					},
				},
				Payload: qdrantPayload,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("ошибка upsert в Qdrant: %w", err)
	}

	r.logger.Info("вектор успешно upserted в Qdrant",
		zap.String("collection", collection),
		zap.String("point_id", pointID.String()),
	)
	return nil
}

// GetVibeVector получает вектор из коллекции Qdrant по UUID точки.
// Возвращает ErrVectorNotFound, если точка не найдена.
func (r *VibeRepository) GetVibeVector(ctx context.Context, collection string, pointID uuid.UUID) ([]float32, error) {
	idStr := pointID.String()
	withVector := true

	resp, err := r.qdrant.Points.Get(ctx, &pb.GetPoints{
		CollectionName: collection,
		Ids: []*pb.PointId{
			{PointIdOptions: &pb.PointId_Uuid{Uuid: idStr}},
		},
		WithVectors: &pb.WithVectorsSelector{
			SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: withVector},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка получения вектора из Qdrant: %w", err)
	}

	if len(resp.GetResult()) == 0 {
		return nil, ErrVectorNotFound
	}

	point := resp.GetResult()[0]
	vectors := point.GetVectors()
	if vectors == nil {
		return nil, ErrVectorNotFound
	}

	vec := vectors.GetVector()
	if vec == nil {
		return nil, ErrVectorNotFound
	}

	return vec.GetData(), nil
}

// SearchNearest ищет Top-N ближайших точек к заданному вектору в коллекции Qdrant.
// Возвращает массив результатов с ID, score и payload.
func (r *VibeRepository) SearchNearest(ctx context.Context, collection string, vector []float32, limit uint64) ([]models.LocationRecommendation, error) {
	r.logger.Info("поиск ближайших векторов",
		zap.String("collection", collection),
		zap.Uint64("limit", limit),
	)

	withPayload := true
	resp, err := r.qdrant.Points.Search(ctx, &pb.SearchPoints{
		CollectionName: collection,
		Vector:         vector,
		Limit:          limit,
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: withPayload},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска в Qdrant: %w", err)
	}

	results := make([]models.LocationRecommendation, 0, len(resp.GetResult()))
	for _, scored := range resp.GetResult() {
		rec := models.LocationRecommendation{
			Score: scored.GetScore(),
		}

		// Извлечение ID из PointId.
		if scored.GetId() != nil {
			if uuidVal := scored.GetId().GetUuid(); uuidVal != "" {
				rec.LocationID = uuidVal
			}
		}

		// Извлечение payload (name, category).
		if payload := scored.GetPayload(); payload != nil {
			if nameVal, ok := payload["name"]; ok {
				rec.Name = nameVal.GetStringValue()
			}
			if catVal, ok := payload["category"]; ok {
				rec.Category = catVal.GetStringValue()
			}
		}

		results = append(results, rec)
	}

	r.logger.Info("поиск завершён",
		zap.Int("found", len(results)),
	)

	return results, nil
}

// UpdateVibeVectorID обновляет vibe_vector_id пользователя в PostgreSQL.
// Вызывается после успешного upsert вектора в Qdrant.
func (r *VibeRepository) UpdateVibeVectorID(ctx context.Context, userID uuid.UUID, vibeVectorID uuid.UUID) error {
	query := `
		UPDATE users SET vibe_vector_id = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.pg.Pool.Exec(ctx, query, vibeVectorID, userID)
	if err != nil {
		return fmt.Errorf("ошибка обновления vibe_vector_id: %w", err)
	}

	r.logger.Info("vibe_vector_id обновлён",
		zap.String("user_id", userID.String()),
		zap.String("vibe_vector_id", vibeVectorID.String()),
	)
	return nil
}
