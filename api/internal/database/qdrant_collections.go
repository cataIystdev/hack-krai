// Файл qdrant_collections.go реализует инициализацию коллекций Qdrant при старте приложения.
// Создаёт коллекции user_vibes и location_vibes с размером вектора 384 и метрикой Cosine Similarity,
// если они ещё не существуют. Используется text-embedding-3-large (OpenAI / OnlySQ API).
package database

import (
	"context"
	"fmt"

	pb "github.com/qdrant/go-client/qdrant"
	"go.uber.org/zap"
)

// QdrantCollectionConfig — конфигурация коллекции Qdrant.
type QdrantCollectionConfig struct {
	// Name — имя коллекции.
	Name string

	// VectorSize — размерность вектора (384 для paraphrase-multilingual-MiniLM-L12-v2).
	VectorSize uint64

	// Distance — метрика расстояния (Cosine Similarity).
	Distance pb.Distance
}

// DefaultCollections — конфигурации коллекций, создаваемых при старте приложения.
var DefaultCollections = []QdrantCollectionConfig{
	{
		Name:       "user_vibes",
		VectorSize: 384,
		Distance:   pb.Distance_Cosine,
	},
	{
		Name:       "location_vibes",
		VectorSize: 384,
		Distance:   pb.Distance_Cosine,
	},
	{
		Name:       "scene_vibes",
		VectorSize: 384,
		Distance:   pb.Distance_Cosine,
	},
}

// EnsureCollections проверяет наличие и при необходимости создаёт коллекции Qdrant.
// Вызывается из main.go после успешного подключения к Qdrant.
// Для каждой коллекции проверяет её существование через CollectionInfo.
// Если коллекция не найдена — создаёт новую с указанными параметрами.
func EnsureCollections(ctx context.Context, client *QdrantClient, logger *zap.Logger) error {
	if client == nil {
		logger.Warn("Qdrant клиент не инициализирован, пропуск создания коллекций")
		return nil
	}

	for _, col := range DefaultCollections {
		if err := ensureCollection(ctx, client, col, logger); err != nil {
			return fmt.Errorf("ошибка создания коллекции %s: %w", col.Name, err)
		}
	}

	logger.Info("все коллекции Qdrant инициализированы",
		zap.Int("count", len(DefaultCollections)),
	)
	return nil
}

// ensureCollection проверяет существование одной коллекции и создаёт её при отсутствии.
func ensureCollection(ctx context.Context, client *QdrantClient, col QdrantCollectionConfig, logger *zap.Logger) error {
	// Проверка существования коллекции через CollectionInfo.
	_, err := client.Collections.Get(ctx, &pb.GetCollectionInfoRequest{
		CollectionName: col.Name,
	})

	if err == nil {
		// Коллекция существует.
		logger.Debug("коллекция Qdrant уже существует",
			zap.String("collection", col.Name),
		)
		return nil
	}

	// Коллекция не найдена — создаём новую.
	logger.Info("создание коллекции Qdrant",
		zap.String("collection", col.Name),
		zap.Uint64("vector_size", col.VectorSize),
		zap.String("distance", col.Distance.String()),
	)

	_, err = client.Collections.Create(ctx, &pb.CreateCollection{
		CollectionName: col.Name,
		VectorsConfig: &pb.VectorsConfig{
			Config: &pb.VectorsConfig_Params{
				Params: &pb.VectorParams{
					Size:     col.VectorSize,
					Distance: col.Distance,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("ошибка создания коллекции Qdrant %s: %w", col.Name, err)
	}

	logger.Info("коллекция Qdrant создана",
		zap.String("collection", col.Name),
	)
	return nil
}
