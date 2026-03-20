# Архитектура проекта Deep Krai API

## Общее описание

Deep Krai API -- серверная часть PWA-платформы пространственного туризма.
Реализована на Go (Fiber v3) с polyglot persistence подходом (6 баз данных).

## Структура проекта

```
api/
  cmd/api/             -- Точка входа приложения (main.go)
  internal/
    config/            -- Конфигурация (Viper) и логирование (Zap)
    database/          -- Подключения к 6 БД (Singleton, Manager)
    handlers/          -- HTTP-обработчики (health, media, router)
    middleware/        -- CORS, Logger, Recovery
    models/            -- Модели данных (расширяется)
    services/          -- Бизнес-логика (storage)
    websocket/         -- WebSocket-обработчики (расширяется)
  migrations/          -- SQL-миграции
  scripts/             -- Вспомогательные скрипты
  docs/                -- Документация
```

## Слой баз данных

| БД                   | Пакет         | Назначение                          |
| -------------------- | ------------- | ----------------------------------- |
| PostgreSQL + PostGIS | postgres.go   | Пользователи, локации, бронирования |
| Redis 7              | redis.go      | Кэш, PubSub, очереди задач          |
| Qdrant               | qdrant.go     | Векторные эмбеддинги (vibe-профили) |
| Neo4j 5              | neo4j.go      | Граф маршрутов, карма, Hidden Gems  |
| ClickHouse           | clickhouse.go | Телеметрия, аналитика, B2G-дашборд  |
| MinIO                | minio.go      | .splat, аудио, видео, изображения   |

## Паттерн Singleton (Manager)

Все подключения к БД управляются через `database.Manager` (Singleton).
Глобальный экземпляр создается через `GetManager()` с `sync.Once`.

Методы менеджера:

- `ConnectAll()` -- параллельное подключение ко всем 6 БД.
- `PingAll()` -- асинхронный пинг всех БД с измерением латентности.
- `Close()` -- корректное закрытие всех подключений.

## Связи между компонентами

```
main.go
  |-- config.Load()         --> AppConfig
  |-- config.NewLogger()    --> zap.Logger
  |-- database.GetManager() --> Manager
  |     |-- ConnectAll()    --> 6 клиентов БД
  |-- services.NewStorageService() --> StorageService
  |-- fiber.New()           --> App
  |     |-- middleware.NewRecovery()
  |     |-- middleware.NewRequestLogger()
  |     |-- middleware.NewCORS()
  |-- handlers.SetupRoutes()
  |     |-- HealthHandler   --> GET /api/v1/health
  |     |-- MediaHandler    --> POST /api/v1/media/upload
  |-- Graceful Shutdown     --> SIGINT/SIGTERM
```
