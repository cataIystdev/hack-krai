// Файл storage.go реализует сервис хранилища файлов через MinIO (S3).
// Предоставляет методы для загрузки файлов в бакет, генерации публичных URL
// (presigned URL) и удаления объектов.
package services

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"

	"deep-krai-api/internal/database"
)

// StorageService — сервис для работы с объектным хранилищем MinIO.
// Обеспечивает загрузку файлов, генерацию публичных URL и удаление объектов.
type StorageService struct {
	minioClient *database.MinIOClient
	logger      *zap.Logger
}

// UploadResult — результат загрузки файла в хранилище.
type UploadResult struct {
	// ObjectName — имя объекта в бакете (включая путь).
	ObjectName string `json:"object_name"`

	// Bucket — имя бакета, в который загружен файл.
	Bucket string `json:"bucket"`

	// Size — размер загруженного файла в байтах.
	Size int64 `json:"size"`

	// URL — публичный URL для доступа к файлу (presigned URL, TTL 24 часа).
	URL string `json:"url"`
}

// NewStorageService создаёт новый сервис хранилища.
// Принимает клиент MinIO и логгер для записи операций.
func NewStorageService(minioClient *database.MinIOClient, logger *zap.Logger) *StorageService {
	return &StorageService{
		minioClient: minioClient,
		logger:      logger,
	}
}

// Upload загружает файл в бакет MinIO.
// Параметры:
//   - ctx: контекст для отмены операции.
//   - objectName: имя объекта в бакете (может включать путь, например "uploads/image.jpg").
//   - reader: источник данных для загрузки.
//   - fileSize: размер файла в байтах (-1 для неизвестного размера).
//   - contentType: MIME-тип файла (например, "image/jpeg").
//
// Возвращает результат загрузки с публичным URL или ошибку.
func (s *StorageService) Upload(ctx context.Context, objectName string, reader io.Reader, fileSize int64, contentType string) (*UploadResult, error) {
	s.logger.Info("загрузка файла в MinIO",
		zap.String("object", objectName),
		zap.Int64("size", fileSize),
		zap.String("content_type", contentType),
	)

	// Загрузка объекта в бакет.
	info, err := s.minioClient.Client.PutObject(
		ctx,
		s.minioClient.Bucket,
		objectName,
		reader,
		fileSize,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки файла в MinIO: %w", err)
	}

	// Генерация presigned URL для доступа к загруженному файлу.
	presignedURL, err := s.GeneratePresignedURL(ctx, objectName, 24*time.Hour)
	if err != nil {
		s.logger.Warn("не удалось сгенерировать presigned URL, используется прямой путь",
			zap.Error(err),
		)
		presignedURL = fmt.Sprintf("/%s/%s", s.minioClient.Bucket, objectName)
	}

	s.logger.Info("файл успешно загружен",
		zap.String("object", objectName),
		zap.Int64("size", info.Size),
	)

	return &UploadResult{
		ObjectName: objectName,
		Bucket:     s.minioClient.Bucket,
		Size:       info.Size,
		URL:        presignedURL,
	}, nil
}

// GeneratePresignedURL генерирует временный публичный URL для доступа к объекту.
// URL имеет ограниченное время жизни (TTL), после чего становится недействительным.
// Параметры:
//   - ctx: контекст для отмены операции.
//   - objectName: имя объекта в бакете.
//   - ttl: время жизни URL.
//
// Возвращает строку URL или ошибку при генерации.
func (s *StorageService) GeneratePresignedURL(ctx context.Context, objectName string, ttl time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := s.minioClient.Client.PresignedGetObject(
		ctx,
		s.minioClient.Bucket,
		objectName,
		ttl,
		reqParams,
	)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// Delete удаляет объект из бакета MinIO.
// Параметры:
//   - ctx: контекст для отмены операции.
//   - objectName: имя удаляемого объекта.
//
// Возвращает nil при успешном удалении или ошибку.
func (s *StorageService) Delete(ctx context.Context, objectName string) error {
	s.logger.Info("удаление файла из MinIO", zap.String("object", objectName))

	err := s.minioClient.Client.RemoveObject(
		ctx,
		s.minioClient.Bucket,
		objectName,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("ошибка удаления файла из MinIO: %w", err)
	}

	s.logger.Info("файл удалён", zap.String("object", objectName))
	return nil
}

// GenerateObjectName создаёт уникальное имя объекта для загрузки.
// Формат: uploads/{timestamp}_{originalName}
// Использование временной метки предотвращает конфликты имён.
func GenerateObjectName(originalName string) string {
	timestamp := time.Now().UnixNano()
	ext := path.Ext(originalName)
	baseName := originalName[:len(originalName)-len(ext)]
	return fmt.Sprintf("uploads/%d_%s%s", timestamp, baseName, ext)
}
