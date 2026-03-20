// Файл scalar.go реализует обработчики для Scalar API документации.
// Предоставляет два эндпоинта:
//   - GET /api/v1/docs -- интерактивная документация Scalar UI (HTML).
//   - GET /api/v1/docs/openapi.json -- спецификация OpenAPI 3.1 (JSON).
//
// Scalar UI загружается через CDN и рендерит спецификацию OpenAPI
// в современном интерактивном интерфейсе с поддержкой тёмной темы.
package handlers

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// scalarHTML — шаблон HTML-страницы с Scalar API Reference.
// Scalar загружается через CDN (@scalar/api-reference), получает спецификацию
// из эндпоинта /api/v1/docs/openapi.json и рендерит интерактивную документацию.
// Настроена тёмная тема (deep-space), боковая панель и современный макет.
const scalarHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Deep Krai API — Документация</title>
  <meta name="description" content="Интерактивная документация REST API платформы пространственного туризма Deep Krai." />
  <style>
    body { margin: 0; padding: 0; }
  </style>
</head>
<body>
  <script
    id="api-reference"
    data-url="/api/v1/docs/openapi.json"
    data-configuration='{
      "theme": "deepSpace",
      "layout": "modern",
      "darkMode": true,
      "showSidebar": true,
      "hideDownloadButton": false,
      "searchHotKey": "k",
      "metaData": {
        "title": "Deep Krai API — Документация"
      }
    }'>
  </script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`

// ScalarHandler — обработчик для Scalar API документации.
type ScalarHandler struct {
	logger *zap.Logger
}

// NewScalarHandler создаёт новый обработчик Scalar документации.
func NewScalarHandler(logger *zap.Logger) *ScalarHandler {
	return &ScalarHandler{logger: logger}
}

// ServeUI обрабатывает GET /api/v1/docs.
// Возвращает HTML-страницу с интерактивной документацией Scalar UI.
// Спецификация OpenAPI загружается клиентом из /api/v1/docs/openapi.json.
func (h *ScalarHandler) ServeUI(c fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(scalarHTML)
}

// ServeSpec обрабатывает GET /api/v1/docs/openapi.json.
// Возвращает полную спецификацию OpenAPI 3.1 в формате JSON.
// URL сервера формируется динамически из заголовка Host входящего запроса.
func (h *ScalarHandler) ServeSpec(c fiber.Ctx) error {
	scheme := "http"
	if c.Secure() {
		scheme = "https"
	}
	host := c.Get("Host")
	baseURL := scheme + "://" + host
	return c.JSON(OpenAPISpec(baseURL))
}
