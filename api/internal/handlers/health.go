// Файл health.go реализует обработчик GET /api/v1/health.
// Выполняет асинхронный пинг всех 6 баз данных и возвращает JSON
// со статусом каждого сервиса, временем отклика и общим состоянием системы.
package handlers

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
)

// HealthResponse — структура ответа health-эндпоинта.
// Содержит общий статус системы и детальную информацию по каждому сервису БД.
type HealthResponse struct {
	// Status — общий статус системы: "healthy" если все сервисы доступны,
	// "degraded" если хотя бы один сервис недоступен.
	Status string `json:"status"`

	// Timestamp — временная метка выполнения проверки в формате RFC3339.
	Timestamp string `json:"timestamp"`

	// Services — карта статусов каждого сервиса БД.
	Services map[string]database.ServiceHealth `json:"services"`
}

// HealthHandler — обработчик запросов проверки здоровья системы.
type HealthHandler struct {
	dbManager *database.Manager
	logger    *zap.Logger
}

// NewHealthHandler создаёт новый обработчик health-проверки.
// Принимает менеджер БД для выполнения пингов и логгер для записи результатов.
func NewHealthHandler(dbManager *database.Manager, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		dbManager: dbManager,
		logger:    logger,
	}
}

// Live обрабатывает GET /api/v1/health/live.
// Мгновенный ответ без пинга внешних сервисов.
// Используется как Kubernetes liveness probe — проверяет только что процесс жив.
func (h *HealthHandler) Live(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready обрабатывает GET /api/v1/health/ready.
// Выполняет асинхронный пинг всех баз данных через менеджер подключений.
// Возвращает JSON с общим статусом и деталями по каждому сервису.
// Статус-код 200 — все сервисы доступны, 503 — есть недоступные сервисы.
func (h *HealthHandler) Ready(c fiber.Ctx) error {
	h.logger.Debug("выполнение проверки готовности сервисов")

	// Асинхронный пинг всех БД через менеджер подключений.
	services := h.dbManager.PingAll(c.Context())

	// Определение общего статуса: "healthy" только если все сервисы "up".
	overallStatus := "healthy"
	for name, health := range services {
		if health.Status != "up" {
			overallStatus = "degraded"
			h.logger.Warn("сервис недоступен",
				zap.String("service", name),
				zap.String("error", health.Error),
			)
		}
	}

	// Формирование ответа.
	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  services,
	}

	// Выбор статус-кода: 200 если всё здорово, 503 если есть проблемы.
	statusCode := fiber.StatusOK
	if overallStatus != "healthy" {
		statusCode = fiber.StatusServiceUnavailable
	}

	h.logger.Debug("проверка готовности завершена",
		zap.String("status", overallStatus),
		zap.Int("services_count", len(services)),
	)

	return c.Status(statusCode).JSON(response)
}

// Check обрабатывает GET /api/v1/health (обратная совместимость).
// Делегирует в Ready() — полный пинг всех БД.
func (h *HealthHandler) Check(c fiber.Ctx) error {
	return h.Ready(c)
}
