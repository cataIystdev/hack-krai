// Файл qdrant.go реализует подключение к векторной БД Qdrant через gRPC.
// Qdrant используется для хранения и поиска vibe-эмбеддингов пользователей и локаций.
package database

import (
	"context"
	"fmt"
	"time"

	pb "github.com/qdrant/go-client/qdrant"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"kudytudy-api/internal/config"
)

// QdrantClient — обёртка над gRPC-клиентом Qdrant.
// Предоставляет доступ к API коллекций, точек и метод проверки связи.
type QdrantClient struct {
	// Conn — gRPC-соединение с сервером Qdrant.
	Conn *grpc.ClientConn

	// Collections — клиент для управления коллекциями (создание, удаление, информация).
	Collections pb.CollectionsClient

	// Points — клиент для операций с точками (вставка, поиск, обновление).
	Points pb.PointsClient

	logger *zap.Logger
}

// NewQdrantClient создаёт новый gRPC-клиент для подключения к Qdrant.
// Устанавливает соединение через gRPC на указанный адрес (по умолчанию порт 6334).
// Возвращает инициализированный клиент или ошибку при невозможности подключения.
func NewQdrantClient(ctx context.Context, cfg config.QdrantConfig, logger *zap.Logger) (*QdrantClient, error) {
	logger.Info("подключение к Qdrant",
		zap.String("host", cfg.Host),
		zap.Int("grpc_port", cfg.GRPCPort),
	)

	// Установка gRPC-соединения с таймаутом.
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		connectCtx,
		cfg.GRPCAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к Qdrant gRPC: %w", err)
	}

	// Создание клиентов для API коллекций и точек.
	collectionsClient := pb.NewCollectionsClient(conn)
	pointsClient := pb.NewPointsClient(conn)

	// Проверка подключения через запрос списка коллекций.
	if _, err := collectionsClient.List(connectCtx, &pb.ListCollectionsRequest{}); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ошибка проверки подключения к Qdrant: %w", err)
	}

	logger.Info("подключение к Qdrant установлено")

	return &QdrantClient{
		Conn:        conn,
		Collections: collectionsClient,
		Points:      pointsClient,
		logger:      logger,
	}, nil
}

// Ping выполняет проверку соединения с Qdrant через запрос списка коллекций.
// Возвращает nil при успешном пинге или ошибку при недоступности сервера.
func (c *QdrantClient) Ping(ctx context.Context) error {
	_, err := c.Collections.List(ctx, &pb.ListCollectionsRequest{})
	return err
}

// Close закрывает gRPC-соединение с Qdrant.
func (c *QdrantClient) Close() error {
	c.logger.Info("закрытие подключения к Qdrant")
	return c.Conn.Close()
}
