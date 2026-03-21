// Файл media.go реализует обработчик POST /api/v1/media/upload.
// Принимает multipart/form-data с файлом, валидирует размер и MIME-тип,
// загружает в MinIO и возвращает публичный URL.
package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"kudytudy-api/internal/services"
)

// maxFileSize — максимальный размер загружаемого файла (100 МБ).
// Ограничение защищает от перегрузки сервера большими файлами.
const maxFileSize = 100 * 1024 * 1024 // 100 МБ

// allowedMIMETypes — карта разрешённых MIME-типов для загрузки.
// Включает изображения, видео, аудио и 3D-файлы, используемые в платформе.
var allowedMIMETypes = map[string]bool{
	"image/jpeg":               true,
	"image/png":                true,
	"image/gif":                true,
	"image/webp":               true,
	"video/mp4":                true,
	"video/webm":               true,
	"video/quicktime":          true,
	"audio/mpeg":               true,
	"audio/mp3":                true,
	"audio/ogg":                true,
	"audio/webm":               true,
	"audio/wav":                true,
	"application/octet-stream": true, // .splat и другие бинарные файлы
}

// UploadResponse — структура JSON-ответа на запрос загрузки файла.
type UploadResponse struct {
	// Success — флаг успешной загрузки.
	Success bool `json:"success"`

	// Message — описание результата операции.
	Message string `json:"message"`

	// Data — данные о загруженном файле (имя объекта, URL, размер).
	Data *services.UploadResult `json:"data,omitempty"`
}

// MediaHandler — обработчик запросов загрузки медиафайлов.
type MediaHandler struct {
	storage *services.StorageService
	logger  *zap.Logger
}

// NewMediaHandler создаёт новый обработчик загрузки медиафайлов.
// Принимает сервис хранилища для работы с MinIO и логгер.
func NewMediaHandler(storage *services.StorageService, logger *zap.Logger) *MediaHandler {
	return &MediaHandler{
		storage: storage,
		logger:  logger,
	}
}

// Upload обрабатывает POST /api/v1/media/upload.
// Принимает multipart/form-data с полем "file".
// Валидирует: наличие файла, размер (до 100 МБ), MIME-тип.
// Загружает файл в MinIO и возвращает JSON с публичным URL.
//
// Возможные ответы:
//   - 200: файл успешно загружен, возвращается URL.
//   - 400: файл не найден, превышен размер или неподдерживаемый тип.
//   - 500: ошибка при загрузке в хранилище.
func (h *MediaHandler) Upload(c fiber.Ctx) error {
	// Проверка доступности хранилища.
	// При недоступности MinIO возвращаем 503 вместо panic на nil pointer.
	if h.storage == nil {
		h.logger.Error("хранилище медиафайлов недоступно: StorageService не инициализирован")
		return c.Status(fiber.StatusServiceUnavailable).JSON(UploadResponse{
			Success: false,
			Message: "хранилище медиафайлов временно недоступно",
		})
	}

	h.logger.Debug("получен запрос на загрузку файла")

	// Извлечение файла из multipart-формы.
	fileHeader, err := c.FormFile("file")
	if err != nil {
		h.logger.Warn("файл не найден в запросе", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(UploadResponse{
			Success: false,
			Message: "файл не найден в запросе, используйте поле 'file'",
		})
	}

	// Проверка размера файла.
	if fileHeader.Size > maxFileSize {
		h.logger.Warn("превышен максимальный размер файла",
			zap.Int64("size", fileHeader.Size),
			zap.Int64("max_size", maxFileSize),
		)
		return c.Status(fiber.StatusBadRequest).JSON(UploadResponse{
			Success: false,
			Message: fmt.Sprintf("размер файла (%d байт) превышает максимально допустимый (%d байт)", fileHeader.Size, maxFileSize),
		})
	}

	// Определение и проверка MIME-типа.
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if !allowedMIMETypes[contentType] {
		h.logger.Warn("неподдерживаемый MIME-тип",
			zap.String("content_type", contentType),
			zap.String("filename", fileHeader.Filename),
		)
		return c.Status(fiber.StatusBadRequest).JSON(UploadResponse{
			Success: false,
			Message: fmt.Sprintf("неподдерживаемый тип файла: %s", contentType),
		})
	}

	// Открытие файла для чтения.
	file, err := fileHeader.Open()
	if err != nil {
		h.logger.Error("ошибка открытия загруженного файла", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(UploadResponse{
			Success: false,
			Message: "ошибка обработки загруженного файла",
		})
	}
	defer file.Close()

	// Генерация уникального имени объекта.
	objectName := services.GenerateObjectName(fileHeader.Filename)

	// Загрузка файла в MinIO через сервис хранилища.
	result, err := h.storage.Upload(c.Context(), objectName, file, fileHeader.Size, contentType)
	if err != nil {
		h.logger.Error("ошибка загрузки файла в хранилище",
			zap.String("filename", fileHeader.Filename),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(UploadResponse{
			Success: false,
			Message: "ошибка загрузки файла в хранилище",
		})
	}

	h.logger.Info("файл успешно загружен",
		zap.String("object", result.ObjectName),
		zap.Int64("size", result.Size),
	)

	return c.Status(fiber.StatusOK).JSON(UploadResponse{
		Success: true,
		Message: "файл успешно загружен",
		Data:    result,
	})
}
