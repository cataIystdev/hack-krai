// Тесты для обработчика загрузки медиафайлов.
// Проверяют валидацию размера, MIME-типа и обработку отсутствующего файла.
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

// TestMediaUploadNoFile проверяет ответ при отсутствии файла в запросе.
func TestMediaUploadNoFile(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Создание handler без реального StorageService (nil),
	// так как валидация файла происходит до обращения к хранилищу.
	handler := NewMediaHandler(nil, logger)

	app := fiber.New()
	app.Post("/api/v1/media/upload", handler.Upload)

	// Запрос без multipart body.
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/media/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=test")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var response UploadResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "файл не найден")
}

// TestMediaUploadResponseStructure проверяет, что ответ на запрос
// без файла содержит корректную структуру JSON.
func TestMediaUploadResponseStructure(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	handler := NewMediaHandler(nil, logger)

	app := fiber.New()
	app.Post("/api/v1/media/upload", handler.Upload)

	// Создание multipart-запроса с пустым файлом (без поля "file").
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("name", "test")
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/media/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

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
