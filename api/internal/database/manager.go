// Файл manager.go реализует централизованный менеджер подключений ко всем базам данных.
// Паттерн Singleton обеспечивается через sync.Once: каждый клиент создаётся однократно.
// Менеджер предоставляет единый интерфейс для подключения, проверки здоровья
// и корректного завершения работы всех клиентов БД.
package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"deep-krai-api/internal/config"
)

// ServiceHealth — результат проверки здоровья одного сервиса БД.
type ServiceHealth struct {
	// Status — текущий статус сервиса: "up" или "down".
	Status string `json:"status"`

	// LatencyMs — время отклика в миллисекундах.
	LatencyMs int64 `json:"latency_ms"`

	// Error — описание ошибки, если сервис недоступен (пустая строка при статусе "up").
	Error string `json:"error,omitempty"`
}

// Manager — централизованный менеджер подключений ко всем базам данных.
// Реализует паттерн Singleton через sync.Once для каждого клиента.
// Предоставляет методы подключения, проверки здоровья (PingAll) и закрытия.
type Manager struct {
	// Postgres — клиент PostgreSQL + PostGIS.
	Postgres *PostgresClient

	// Redis — клиент Redis.
	Redis *RedisClient

	// Qdrant — клиент векторной БД Qdrant.
	Qdrant *QdrantClient

	// Neo4j — клиент графовой БД Neo4j.
	Neo4j *Neo4jClient

	// ClickHouse — клиент аналитической БД ClickHouse.
	ClickHouse *ClickHouseClient

	// MinIO — клиент объектного хранилища MinIO.
	MinIO *MinIOClient

	logger *zap.Logger
	once   sync.Once
}

// singleton — глобальный экземпляр менеджера подключений.
// Инициализируется единожды при вызове GetManager().
var (
	singleton     *Manager
	singletonOnce sync.Once
)

// GetManager возвращает глобальный экземпляр менеджера подключений.
// При первом вызове создаёт новый менеджер с указанным логгером.
// Все последующие вызовы возвращают тот же экземпляр (Singleton).
func GetManager(logger *zap.Logger) *Manager {
	singletonOnce.Do(func() {
		singleton = &Manager{
			logger: logger,
		}
	})
	return singleton
}

// ConnectAll устанавливает подключения ко всем 6 базам данных параллельно.
// Каждое подключение выполняется в отдельной горутине для ускорения запуска.
// Подключение выполняется ровно один раз (sync.Once).
// Возвращает ошибку, если хотя бы одно подключение не удалось.
func (m *Manager) ConnectAll(ctx context.Context, cfg *config.AppConfig) error {
	var connectErr error

	m.once.Do(func() {
		m.logger.Info("инициализация подключений ко всем базам данных")

		type result struct {
			name string
			err  error
		}

		results := make(chan result, 6)
		var wg sync.WaitGroup

		// Подключение к PostgreSQL.
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := NewPostgresClient(ctx, cfg.Postgres, m.logger)
			if err == nil {
				m.Postgres = client
			}
			results <- result{name: "PostgreSQL", err: err}
		}()

		// Подключение к Redis.
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := NewRedisClient(ctx, cfg.Redis, m.logger)
			if err == nil {
				m.Redis = client
			}
			results <- result{name: "Redis", err: err}
		}()

		// Подключение к Qdrant.
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := NewQdrantClient(ctx, cfg.Qdrant, m.logger)
			if err == nil {
				m.Qdrant = client
			}
			results <- result{name: "Qdrant", err: err}
		}()

		// Подключение к Neo4j.
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := NewNeo4jClient(ctx, cfg.Neo4j, m.logger)
			if err == nil {
				m.Neo4j = client
			}
			results <- result{name: "Neo4j", err: err}
		}()

		// Подключение к ClickHouse.
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := NewClickHouseClient(ctx, cfg.ClickHouse, m.logger)
			if err == nil {
				m.ClickHouse = client
			}
			results <- result{name: "ClickHouse", err: err}
		}()

		// Подключение к MinIO.
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := NewMinIOClient(ctx, cfg.MinIO, m.logger)
			if err == nil {
				m.MinIO = client
			}
			results <- result{name: "MinIO", err: err}
		}()

		// Ожидание завершения всех подключений.
		wg.Wait()
		close(results)

		// Сбор ошибок.
		var errors []string
		for r := range results {
			if r.err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", r.name, r.err))
				m.logger.Error("ошибка подключения к БД",
					zap.String("service", r.name),
					zap.Error(r.err),
				)
			}
		}

		if len(errors) > 0 {
			connectErr = fmt.Errorf("ошибки подключения к БД: %v", errors)
		} else {
			m.logger.Info("все подключения к базам данных установлены")
		}
	})

	return connectErr
}

// PingAll выполняет асинхронный пинг всех 6 баз данных.
// Каждый пинг выполняется в отдельной горутине с измерением времени отклика.
// Возвращает карту результатов проверки здоровья для каждого сервиса.
func (m *Manager) PingAll(ctx context.Context) map[string]ServiceHealth {
	healthMap := make(map[string]ServiceHealth)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// pingService — вспомогательная функция для асинхронного пинга одного сервиса.
	// Измеряет время отклика и записывает результат в общую карту.
	pingService := func(name string, pingFn func(ctx context.Context) error) {
		defer wg.Done()

		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		start := time.Now()
		err := pingFn(pingCtx)
		latency := time.Since(start).Milliseconds()

		health := ServiceHealth{
			Status:    "up",
			LatencyMs: latency,
		}
		if err != nil {
			health.Status = "down"
			health.Error = err.Error()
		}

		mu.Lock()
		healthMap[name] = health
		mu.Unlock()
	}

	// Запуск пинга всех сервисов параллельно.
	if m.Postgres != nil {
		wg.Add(1)
		go pingService("postgres", m.Postgres.Ping)
	} else {
		healthMap["postgres"] = ServiceHealth{Status: "down", Error: "клиент не инициализирован"}
	}

	if m.Redis != nil {
		wg.Add(1)
		go pingService("redis", m.Redis.Ping)
	} else {
		healthMap["redis"] = ServiceHealth{Status: "down", Error: "клиент не инициализирован"}
	}

	if m.Qdrant != nil {
		wg.Add(1)
		go pingService("qdrant", m.Qdrant.Ping)
	} else {
		healthMap["qdrant"] = ServiceHealth{Status: "down", Error: "клиент не инициализирован"}
	}

	if m.Neo4j != nil {
		wg.Add(1)
		go pingService("neo4j", m.Neo4j.Ping)
	} else {
		healthMap["neo4j"] = ServiceHealth{Status: "down", Error: "клиент не инициализирован"}
	}

	if m.ClickHouse != nil {
		wg.Add(1)
		go pingService("clickhouse", m.ClickHouse.Ping)
	} else {
		healthMap["clickhouse"] = ServiceHealth{Status: "down", Error: "клиент не инициализирован"}
	}

	if m.MinIO != nil {
		wg.Add(1)
		go pingService("minio", m.MinIO.Ping)
	} else {
		healthMap["minio"] = ServiceHealth{Status: "down", Error: "клиент не инициализирован"}
	}

	wg.Wait()
	return healthMap
}

// Close корректно закрывает все подключения к базам данных.
// Вызывается при завершении работы приложения для освобождения ресурсов.
// Ошибки закрытия логируются, но не прерывают процесс завершения.
func (m *Manager) Close(ctx context.Context) {
	m.logger.Info("закрытие всех подключений к базам данных")

	if m.Postgres != nil {
		m.Postgres.Close()
	}
	if m.Redis != nil {
		if err := m.Redis.Close(); err != nil {
			m.logger.Error("ошибка закрытия Redis", zap.Error(err))
		}
	}
	if m.Qdrant != nil {
		if err := m.Qdrant.Close(); err != nil {
			m.logger.Error("ошибка закрытия Qdrant", zap.Error(err))
		}
	}
	if m.Neo4j != nil {
		if err := m.Neo4j.Close(ctx); err != nil {
			m.logger.Error("ошибка закрытия Neo4j", zap.Error(err))
		}
	}
	if m.ClickHouse != nil {
		if err := m.ClickHouse.Close(); err != nil {
			m.logger.Error("ошибка закрытия ClickHouse", zap.Error(err))
		}
	}
	if m.MinIO != nil {
		m.MinIO.Close()
	}

	m.logger.Info("все подключения к базам данных закрыты")
}

// ResetSingleton сбрасывает глобальный экземпляр менеджера.
// Используется только в тестах для обеспечения изоляции между тестовыми случаями.
func ResetSingleton() {
	singleton = nil
	singletonOnce = sync.Once{}
}
