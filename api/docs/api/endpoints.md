# API Endpoints

## GET /api/v1/health

Проверка здоровья всех сервисов.

Выполняет асинхронный пинг 6 баз данных через горутины.
Возвращает статус каждого сервиса и общее состояние системы.

### Ответ 200 (все сервисы доступны)

```json
{
  "status": "healthy",
  "timestamp": "2026-03-19T20:37:07Z",
  "services": {
    "postgres": { "status": "up", "latency_ms": 2 },
    "redis": { "status": "up", "latency_ms": 1 },
    "qdrant": { "status": "up", "latency_ms": 3 },
    "neo4j": { "status": "up", "latency_ms": 5 },
    "clickhouse": { "status": "up", "latency_ms": 2 },
    "minio": { "status": "up", "latency_ms": 1 }
  }
}
```

### Ответ 503 (есть недоступные сервисы)

```json
{
  "status": "degraded",
  "timestamp": "2026-03-19T20:37:07Z",
  "services": {
    "postgres": { "status": "up", "latency_ms": 2 },
    "redis": {
      "status": "down",
      "latency_ms": 0,
      "error": "connection refused"
    },
    "qdrant": { "status": "up", "latency_ms": 3 },
    "neo4j": { "status": "up", "latency_ms": 5 },
    "clickhouse": { "status": "up", "latency_ms": 2 },
    "minio": { "status": "up", "latency_ms": 1 }
  }
}
```

---

## POST /api/v1/media/upload

Загрузка медиафайлов в S3-хранилище (MinIO).

### Запрос

Тип: `multipart/form-data`
Поле: `file` -- загружаемый файл.

Ограничения:

- Максимальный размер: 100 МБ
- Разрешённые MIME-типы: image/jpeg, image/png, image/gif, image/webp,
  video/mp4, video/webm, video/quicktime, audio/mpeg, audio/mp3,
  audio/ogg, audio/webm, audio/wav, application/octet-stream

### Пример запроса

```bash
curl -X POST http://localhost:8080/api/v1/media/upload \
  -F "file=@photo.jpg"
```

### Ответ 200 (успешная загрузка)

```json
{
  "success": true,
  "message": "файл успешно загружен",
  "data": {
    "object_name": "uploads/1710873427000000_photo.jpg",
    "bucket": "deepkrai-media",
    "size": 245760,
    "url": "http://localhost:9000/deepkrai-media/uploads/..."
  }
}
```

### Ответ 400 (ошибка валидации)

```json
{
  "success": false,
  "message": "неподдерживаемый тип файла: text/html"
}
```
