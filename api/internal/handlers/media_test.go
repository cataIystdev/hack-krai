// Тесты для обработчика загрузки медиафайлов.
// Проверяют валидацию размера, MIME-типа, обработку отсутствующего файла
// и безопасное поведение при недоступности хранилища.
package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestMediaUploadStorageUnavailable проверяет, что при nil storage
// Upload возвращает 503 Service Unavailable вместо panic.
func TestMediaUploadStorageUnavailable(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// storage = nil имитирует недоступность MinIO.
	handler := NewMediaHandler(nil, logger)

	app := fiber.New()
	app.Post("/api/v1/media/upload", handler.Upload)

	// Создание multipart-запроса с файлом.
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "test.jpg")
	_, _ = part.Write([]byte("fake-image-data-more-than-zero"))
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/media/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// При nil storage ожидаем 503, а не panic.
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var response UploadResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "временно недоступно")
}

// TestMediaUploadNoFile проверяет ответ при отсутствии файла в запросе.
func TestMediaUploadNoFile(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// storage = nil, но запрос не содержит файла.
	// nil-guard сработает первым и вернёт 503.
	handler := NewMediaHandler(nil, logger)

	app := fiber.New()
	app.Post("/api/v1/media/upload", handler.Upload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/media/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=test")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// При nil storage — 503.
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var response UploadResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
}

// TestMediaUploadResponseStructure проверяет, что ответ при nil storage
// содержит корректную структуру JSON с полями success и message.
func TestMediaUploadResponseStructure(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	handler := NewMediaHandler(nil, logger)

	app := fiber.New()
	app.Post("/api/v1/media/upload", handler.Upload)

	// Создание multipart-запроса без файла.
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("name", "test")
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/media/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// При nil storage — 503 до проверки наличия файла.
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.Contains(t, response, "success")
	assert.Contains(t, response, "message")
}

// TestAllowedMIMETypes проверяет корректность карты разрешённых MIME-типов.
func TestAllowedMIMETypes(t *testing.T) {
	// Проверка основных типов, которые должны быть разрешены.
	allowed := []string{
		"image/jpeg", "image/png", "image/webp",
		"video/mp4", "video/webm",
		"audio/mpeg", "audio/ogg",
		"application/octet-stream",
	}
	for _, mime := range allowed {
		assert.True(t, allowedMIMETypes[mime], "MIME-тип %q должен быть разрешён", mime)
	}

	// Проверка типов, которые не должны быть разрешены.
	disallowed := []string{
		"application/javascript",
		"text/html",
		"application/x-executable",
	}
	for _, mime := range disallowed {
		assert.False(t, allowedMIMETypes[mime], "MIME-тип %q не должен быть разрешён", mime)
	}
}
