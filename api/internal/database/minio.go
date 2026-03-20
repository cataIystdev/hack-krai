// Файл minio.go реализует подключение к S3-совместимому хранилищу MinIO.
// MinIO используется для хранения .splat файлов (3D-сцены), аудиофайлов,
// видео от фермеров и изображений.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"

	"deep-krai-api/internal/config"
)

// MinIOClient — обёртка над клиентом MinIO (S3).
// Предоставляет доступ к клиенту для операций с объектами и бакетами.
type MinIOClient struct {
	// Client — клиент MinIO для работы с S3 API.
	Client *minio.Client

	// Bucket — имя бакета по умолчанию для хранения медиафайлов.
	Bucket string

	logger *zap.Logger
}

// NewMinIOClient создаёт новый клиент для подключения к MinIO.
// Настраивает аутентификацию, SSL и автоматически создаёт бакет,
// если он ещё не существует.
// Возвращает инициализированный клиент или ошибку при невозможности подключения.
func NewMinIOClient(ctx context.Context, cfg config.MinIOConfig, logger *zap.Logger) (*MinIOClient, error) {
	logger.Info("подключение к MinIO",
		zap.String("endpoint", cfg.Endpoint()),
		zap.String("bucket", cfg.Bucket),
		zap.Bool("ssl", cfg.UseSSL),
	)

	// Создание клиента MinIO с указанными учётными данными.
	client, err := minio.New(cfg.Endpoint(), &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.RootUser, cfg.RootPassword, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента MinIO: %w", err)
	}

	// Проверка подключения через запрос информации о сервере.
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Проверка доступности сервера через ListBuckets.
	_, err = client.ListBuckets(connectCtx)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки подключения к MinIO: %w", err)
	}

	// Автоматическое создание бакета, если он не существует.
	exists, err := client.BucketExists(connectCtx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существования бакета MinIO: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(connectCtx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("ошибка создания бакета MinIO %q: %w", cfg.Bucket, err)
		}
		logger.Info("бакет MinIO создан", zap.String("bucket", cfg.Bucket))
	}

	logger.Info("подключение к MinIO установлено", zap.String("bucket", cfg.Bucket))

	return &MinIOClient{
		Client: client,
		Bucket: cfg.Bucket,
		logger: logger,
	}, nil
}

// Ping выполняет проверку соединения с MinIO через запрос списка бакетов.
// Возвращает nil при успешном пинге или ошибку при недоступности сервера.
func (c *MinIOClient) Ping(ctx context.Context) error {
	_, err := c.Client.ListBuckets(ctx)
	return err
}

// Close выполняет корректное завершение работы клиента MinIO.
// MinIO SDK не требует явного закрытия подключения, так как использует HTTP.
// Метод предоставлен для единообразия интерфейса.
func (c *MinIOClient) Close() {
	c.logger.Info("клиент MinIO остановлен")
}
