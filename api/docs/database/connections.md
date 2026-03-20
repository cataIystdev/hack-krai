# Подключения к базам данных

## Паттерн Singleton

Все подключения реализованы через паттерн Singleton с использованием `sync.Once`.
Менеджер `database.Manager` хранит клиенты всех 6 БД и предоставляет единый
интерфейс для подключения, проверки здоровья и закрытия.

## PostgreSQL + PostGIS

- Библиотека: `github.com/jackc/pgx/v5/pgxpool`
- Протокол: PostgreSQL wire protocol
- Пул: настраиваемый (MaxConns, MinConns, HealthCheckPeriod)
- Проверка: `Pool.Ping(ctx)`
- DSN: `postgres://user:pass@host:port/db?sslmode=disable`

## Redis 7

- Библиотека: `github.com/redis/go-redis/v9`
- Протокол: RESP3
- Пул: автоматический (PoolSize, MinIdleConns)
- Проверка: `Client.Ping(ctx)`
- Адрес: `host:port`

## Qdrant

- Библиотека: `github.com/qdrant/go-client` (gRPC)
- Протокол: gRPC (порт 6334)
- Проверка: `Collections.List(ctx)` -- запрос списка коллекций
- Адрес: `host:grpc_port`

## Neo4j 5

- Библиотека: `github.com/neo4j/neo4j-go-driver/v5`
- Протокол: Bolt (порт 7687)
- Аутентификация: BasicAuth (логин/пароль)
- Проверка: `Driver.VerifyConnectivity(ctx)`
- URI: `bolt://host:port`

## ClickHouse

- Библиотека: `github.com/ClickHouse/clickhouse-go/v2`
- Протокол: нативный (порт 9000)
- Сжатие: LZ4
- Проверка: `Conn.Ping(ctx)`
- Адрес: `host:port`

## MinIO (S3)

- Библиотека: `github.com/minio/minio-go/v7`
- Протокол: HTTP (S3 API)
- Аутентификация: Static V4 (Access Key / Secret Key)
- Проверка: `Client.ListBuckets(ctx)`
- Автосоздание бакета при инициализации

## Конфигурация

Все параметры подключений загружаются из `.env` файла через Viper.
Каждая структура конфигурации содержит методы для формирования
строк подключения (DSN, Addr, URI, Endpoint).

## Асинхронный Ping

Метод `Manager.PingAll()` выполняет пинг всех 6 БД параллельно
в отдельных горутинах. Каждый пинг имеет таймаут 5 секунд.
Результат содержит статус ("up"/"down"), латентность в мс и ошибку.
