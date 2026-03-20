# Запуск инфраструктуры через Docker Compose

## Предварительные требования

- Docker 24+
- Docker Compose v2+

## Запуск

1. Убедиться, что файл `.env` находится в корне проекта
   (рядом с `docker-compose.yml`).

2. Запустить все сервисы:

   ```bash
   docker compose up -d
   ```

3. Проверить статус контейнеров:

   ```bash
   docker compose ps
   ```

   Все контейнеры должны иметь статус "healthy".

## Сервисы и порты

| Сервис     | Контейнер           | Порт(ы)                    |
| ---------- | ------------------- | -------------------------- |
| PostgreSQL | deepkrai-postgres   | 5432                       |
| Redis      | deepkrai-redis      | 6379                       |
| Qdrant     | deepkrai-qdrant     | 6333 (HTTP), 6334 (gRPC)   |
| Neo4j      | deepkrai-neo4j      | 7474 (HTTP), 7687 (Bolt)   |
| ClickHouse | deepkrai-clickhouse | 8123 (HTTP), 9000 (native) |
| MinIO      | deepkrai-minio      | 9000 (API), 9001 (Console) |

## Управление

```bash
# Остановить все сервисы
docker compose down

# Остановить с удалением данных (volumes)
docker compose down -v

# Просмотр логов конкретного сервиса
docker compose logs -f postgres

# Перезапуск одного сервиса
docker compose restart redis
```

## Проверка доступности

```bash
# PostgreSQL
docker compose exec postgres pg_isready -U deepkrai

# Redis
docker compose exec redis redis-cli -a redis_secret_2026 ping

# Qdrant
curl -s http://localhost:6333/healthz

# Neo4j
curl -s http://localhost:7474

# ClickHouse
curl -s http://localhost:8123/ping

# MinIO Console
# Открыть в браузере: http://localhost:9001
```
