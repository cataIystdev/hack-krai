// Файл postgres.go реализует подключение к PostgreSQL + PostGIS.
// Используется пул подключений pgxpool для эффективного управления соединениями.
// Паттерн Singleton реализован через sync.Once в менеджере подключений.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"kudytudy-api/internal/config"
)

// PostgresClient — обёртка над пулом подключений pgxpool.
// Предоставляет доступ к пулу соединений PostgreSQL и метод проверки связи.
type PostgresClient struct {
	// Pool — пул подключений к PostgreSQL.
	// Потокобезопасен, поддерживает автоматическое восстановление соединений.
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewPostgresClient создаёт новый клиент PostgreSQL с пулом подключений.
// Конфигурирует параметры пула: максимальное/минимальное количество соединений,
// период проверки здоровья соединений, таймаут на подключение.
// Возвращает инициализированный клиент или ошибку при невозможности подключения.
func NewPostgresClient(ctx context.Context, cfg config.PostgresConfig, logger *zap.Logger) (*PostgresClient, error) {
	logger.Info("подключение к PostgreSQL",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.DB),
	)

	// Парсинг DSN-строки и настройка параметров пула подключений.
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга DSN PostgreSQL: %w", err)
	}

	// Настройка параметров пула.
	poolCfg.MaxConns = int32(cfg.MaxConns)
	poolCfg.MinConns = int32(cfg.MinConns)
	poolCfg.HealthCheckPeriod = 30 * time.Second
	poolCfg.MaxConnLifetime = 1 * time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute

	// Создание пула с контекстом и таймаутом.
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пула PostgreSQL: %w", err)
	}

	// Проверка подключения при инициализации.
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ошибка пинга PostgreSQL при инициализации: %w", err)
	}

	logger.Info("подключение к PostgreSQL установлено",
		zap.Int32("max_conns", int32(cfg.MaxConns)),
		zap.Int32("min_conns", int32(cfg.MinConns)),
	)

	return &PostgresClient{Pool: pool, logger: logger}, nil
}

// Ping выполняет проверку соединения с PostgreSQL.
// Использует контекст для ограничения времени ожидания ответа.
// Возвращает nil при успешном пинге или ошибку при недоступности сервера.
func (c *PostgresClient) Ping(ctx context.Context) error {
	return c.Pool.Ping(ctx)
}

// Close закрывает пул подключений к PostgreSQL.
// Должен вызываться при завершении работы приложения для освобождения ресурсов.
func (c *PostgresClient) Close() {
	c.logger.Info("закрытие пула подключений PostgreSQL")
	c.Pool.Close()
}
