// Файл neo4j.go реализует подключение к графовой БД Neo4j 5.
// Neo4j используется для хранения графа маршрутов, связей пользователь-локация
// и системы кармы (Hidden Gems). Подключение через Bolt-протокол.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"

	"deep-krai-api/internal/config"
)

// Neo4jClient — обёртка над драйвером Neo4j.
// Предоставляет доступ к драйверу для создания сессий и выполнения Cypher-запросов.
type Neo4jClient struct {
	// Driver — драйвер подключения к Neo4j.
	// Потокобезопасен, управляет пулом соединений автоматически.
	Driver neo4j.DriverWithContext
	logger *zap.Logger
}

// NewNeo4jClient создаёт новый клиент для подключения к Neo4j через Bolt-протокол.
// Настраивает аутентификацию, параметры пула соединений и проверяет доступность.
// Возвращает инициализированный клиент или ошибку при невозможности подключения.
func NewNeo4jClient(ctx context.Context, cfg config.Neo4jConfig, logger *zap.Logger) (*Neo4jClient, error) {
	logger.Info("подключение к Neo4j",
		zap.String("host", cfg.Host),
		zap.Int("bolt_port", cfg.BoltPort),
	)

	// Создание драйвера с аутентификацией через логин/пароль.
	driver, err := neo4j.NewDriverWithContext(
		cfg.BoltURI(),
		neo4j.BasicAuth(cfg.User, cfg.Password, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания драйвера Neo4j: %w", err)
	}

	// Проверка подключения при инициализации.
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(connectCtx); err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("ошибка проверки подключения к Neo4j: %w", err)
	}

	logger.Info("подключение к Neo4j установлено")

	return &Neo4jClient{Driver: driver, logger: logger}, nil
}

// Ping выполняет проверку соединения с Neo4j через VerifyConnectivity.
// Возвращает nil при успешном пинге или ошибку при недоступности сервера.
func (c *Neo4jClient) Ping(ctx context.Context) error {
	return c.Driver.VerifyConnectivity(ctx)
}

// Close закрывает драйвер подключения к Neo4j и освобождает все ресурсы.
func (c *Neo4jClient) Close(ctx context.Context) error {
	c.logger.Info("закрытие подключения к Neo4j")
	return c.Driver.Close(ctx)
}
