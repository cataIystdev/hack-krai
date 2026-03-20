// Файл redis.go реализует подключение к Redis 7.
// Используется библиотека go-redis/v9 для работы с кэшем, PubSub и очередями.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"deep-krai-api/internal/config"
)

// RedisClient — обёртка над клиентом Redis.
// Предоставляет доступ к клиенту и метод проверки связи.
type RedisClient struct {
	// Client — клиент Redis для выполнения команд.
	Client *redis.Client
	logger *zap.Logger
}

// NewRedisClient создаёт новый клиент Redis с заданной конфигурацией.
// Настраивает параметры подключения, пул соединений и таймауты.
// Возвращает инициализированный клиент или ошибку при невозможности подключения.
func NewRedisClient(ctx context.Context, cfg config.RedisConfig, logger *zap.Logger) (*RedisClient, error) {
	logger.Info("подключение к Redis",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB),
	)

	// Создание клиента Redis с настройками пула и таймаутов.
	client := redis.NewClient(&redis.Options{
		Addr:            cfg.Addr(),
		Password:        cfg.Password,
		DB:              cfg.DB,
		DialTimeout:     10 * time.Second,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		PoolSize:        20,
		MinIdleConns:    5,
		ConnMaxIdleTime: 30 * time.Minute,
		ConnMaxLifetime: 1 * time.Hour,
	})

	// Проверка подключения при инициализации.
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(connectCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ошибка пинга Redis при инициализации: %w", err)
	}

	logger.Info("подключение к Redis установлено")

	return &RedisClient{Client: client, logger: logger}, nil
}

// Ping выполняет проверку соединения с Redis.
// Возвращает nil при успешном пинге или ошибку при недоступности сервера.
func (c *RedisClient) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}

// Close закрывает подключение к Redis.
// Должен вызываться при завершении работы приложения.
func (c *RedisClient) Close() error {
	c.logger.Info("закрытие подключения к Redis")
	return c.Client.Close()
}
