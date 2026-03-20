// Файл clickhouse.go реализует подключение к аналитической БД ClickHouse.
// ClickHouse используется для хранения телеметрии, тепловых карт
// и агрегатов для B2G-дашборда. Подключение через нативный протокол.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"

	"kudytudy-api/internal/config"
)

// ClickHouseClient — обёртка над подключением к ClickHouse.
// Предоставляет доступ к соединению для выполнения SQL-запросов.
type ClickHouseClient struct {
	// Conn — соединение с ClickHouse через нативный протокол.
	Conn   clickhouse.Conn
	logger *zap.Logger
}

// NewClickHouseClient создаёт новый клиент для подключения к ClickHouse.
// Настраивает параметры подключения, аутентификацию, сжатие и таймауты.
// Возвращает инициализированный клиент или ошибку при невозможности подключения.
func NewClickHouseClient(ctx context.Context, cfg config.ClickHouseConfig, logger *zap.Logger) (*ClickHouseClient, error) {
	logger.Info("подключение к ClickHouse",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.DB),
	)

	// Создание подключения с настройками аутентификации и сжатия.
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.DB,
			Username: cfg.User,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:     10 * time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 1 * time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания подключения к ClickHouse: %w", err)
	}

	// Проверка подключения при инициализации.
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := conn.Ping(connectCtx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ошибка пинга ClickHouse при инициализации: %w", err)
	}

	logger.Info("подключение к ClickHouse установлено")

	return &ClickHouseClient{Conn: conn, logger: logger}, nil
}

// Ping выполняет проверку соединения с ClickHouse.
// Возвращает nil при успешном пинге или ошибку при недоступности сервера.
func (c *ClickHouseClient) Ping(ctx context.Context) error {
	return c.Conn.Ping(ctx)
}

// Close закрывает подключение к ClickHouse.
func (c *ClickHouseClient) Close() error {
	c.logger.Info("закрытие подключения к ClickHouse")
	return c.Conn.Close()
}
