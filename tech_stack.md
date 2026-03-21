# 🛠 Deep Krai — Полный стек технологий

---

## 1. Backend (Golang)

| Компонент | Технология | Версия | Зачем |
|-----------|-----------|--------|-------|
| **Язык** | Go | 1.25.0 | Основной язык бэкенда, высокая производительность, горутины для конкурентности |
| **HTTP-фреймворк** | [GoFiber](https://gofiber.io) (`gofiber/fiber/v3`) | v3.1.0 | REST API, роутинг, middleware (CORS, Recovery, Logger) |
| **JWT-аутентификация** | `golang-jwt/jwt/v5` | v5.3.1 | Генерация и валидация Access/Refresh JWT-токенов |
| **UUID** | `google/uuid` | v1.6.0 | Генерация уникальных идентификаторов для сущностей |
| **WebSocket** | `gorilla/websocket` | v1.5.3 | Real-time уведомления (бронирования, погода, прогресс онбординга) |
| **Конфигурация** | `spf13/viper` | v1.21.0 | Чтение .env и YAML-конфигурации, маппинг в структуры |
| **Логирование** | `uber-go/zap` | v1.27.1 | Структурированное логирование (JSON), уровни debug/info/error |
| **Хэширование паролей** | `golang.org/x/crypto` (bcrypt) | v0.48.0 | bcrypt-хэширование и верификация паролей |
| **Тестирование** | `stretchr/testify` | v1.11.1 | Ассерты, моки, сьюты для unit-тестов |
| **gRPC** | `google.golang.org/grpc` | v1.78.0 | Связь с Qdrant через gRPC-протокол |
| **Protobuf** | `google.golang.org/protobuf` | v1.36.11 | Сериализация для gRPC (Qdrant client) |

### Структура бэкенда (monorepo)

```
api/
├── cmd/api/main.go          — Точка входа
├── internal/
│   ├── config/              — Viper-конфигурация, логгер
│   ├── middleware/           — JWT auth, CORS, rate limit, logging
│   ├── models/              — Структуры данных (User, Location, Trip, Booking...)
│   ├── database/            — Коннекторы ко всем БД + репозитории
│   │   ├── postgres.go      — pgx connection pool
│   │   ├── qdrant.go        — gRPC-клиент Qdrant
│   │   ├── neo4j.go         — Neo4j driver
│   │   ├── clickhouse.go    — ClickHouse driver
│   │   ├── redis.go         — go-redis клиент
│   │   ├── minio.go         — MinIO S3 клиент
│   │   ├── manager.go       — Центральный менеджер подключений
│   │   ├── *_repository.go  — Data access слой (location, user, trip, vibe)
│   │   └── qdrant_collections.go — Автосоздание коллекций
│   ├── services/            — Бизнес-логика
│   │   ├── auth.go          — Регистрация, логин
│   │   ├── jwt.go           — JWT-генерация и валидация
│   │   ├── location.go      — CRUD локаций, PostGIS-запросы
│   │   ├── trip.go          — Trip Details, групповое планирование
│   │   ├── vibe.go          — AI-профилирование, свайпы, рекомендации
│   │   └── storage.go       — Работа с MinIO (загрузка/скачивание файлов)
│   ├── handlers/            — HTTP-хендлеры (тонкий слой над сервисами)
│   ├── ai/                  — Клиенты к AI-сервисам
│   │   ├── client.go        — Базовый HTTP-клиент для AI API
│   │   ├── whisper.go       — OpenAI Whisper STT
│   │   ├── llm.go           — GPT/Claude LLM (vibe-анализ, storytelling)
│   │   ├── embeddings.go    — Embedding-генерация (local MiniLM)
│   │   ├── vosk_stt.go      — Vosk WebSocket STT (локальный)
│   │   ├── models.go        — Структуры запросов/ответов AI
│   │   └── mock.go          — Моки для тестов
│   └── websocket/           — WebSocket hub для real-time
├── migrations/              — SQL-миграции
├── scripts/
│   ├── seed.go              — Seed 20 локаций Кубани с реальными координатами
│   └── setup_ci_envs.sh     — Настройка переменных для CI/CD
├── Dockerfile               — Multi-stage сборка (golang:1.25-alpine → alpine)
├── go.mod
└── go.sum
```

---

## 2. Базы данных (6 штук, Polyglot Persistence)

| БД | Docker-образ | Версия | Драйвер Go | Зачем |
|----|-------------|--------|-------------|-------|
| **PostgreSQL + PostGIS** | `postgis/postgis` | 16-3.4 | `jackc/pgx/v5` v5.8.0 | Основная ACID-БД: пользователи, локации (геоданные), бронирования, поездки, слоты. PostGIS для `ST_DWithin`, `ST_Contains`, bbox-запросов |
| **Qdrant** | `qdrant/qdrant` | latest | `qdrant/go-client` v1.17.1 (gRPC) | Векторная БД: vibe-эмбеддинги пользователей и локаций. Cosine similarity поиск, фильтрация по метаданным. Коллекции: `user_vibes`, `location_vibes` |
| **Neo4j** | `neo4j` | 5 | `neo4j/neo4j-go-driver/v5` v5.28.4 | Графовая БД: граф дорог/маршрутов (Dijkstra/A*), граф доверия (карма для Hidden Gems). Плагины: APOC, Graph Data Science |
| **ClickHouse** | `clickhouse/clickhouse-server` | 24 | `ClickHouse/clickhouse-go/v2` v2.43.0 | Колоночная аналитическая БД: телеметрия (свайпы, клики, GPS-треки), агрегаты для B2G-дашборда, тепловые карты |
| **Redis** | `redis` | 7-alpine | `redis/go-redis/v9` v9.18.0 | Кэш (погода, сессии), PubSub (WebSocket-уведомления), очереди задач (Streams), таймеры бронирований |
| **MinIO** | `minio/minio` | latest | `minio/minio-go/v7` v7.0.99 | S3-совместимое хранилище: .splat файлы (3D-сцены), аудио (mp3), видео от фермеров, изображения |

---

## 3. AI / ML сервисы

### 3.1 Внешние API

| Сервис | Провайдер | Модель | Зачем |
|--------|-----------|--------|-------|
| **LLM** | OnlySQ API (OpenAI-совместимый) | `gpt-5.1` | Vibe-анализ голоса (JSON с осями), генерация описаний локаций, storytelling, уведомления |
| **STT (Whisper)** | OnlySQ API | `whisper-1` | Транскрибация голосовых сообщений (fallback при недоступности Vosk) |

### 3.2 Локальные AI-сервисы (Self-hosted)

| Сервис | Docker-образ | Технология | Зачем |
|--------|-------------|------------|-------|
| **Vosk STT** | `alphacep/kaldi-ru` | Kaldi + Vosk (русская модель) | Бесплатное локальное распознавание речи. WebSocket API на порту 2700. Принимает PCM 16kHz mono |
| **Embeddings Server** | Custom ([Dockerfile.embeddings](file:///mnt/Work/HACKS/VORONKA/20032026/Dockerfile.embeddings)) | Python 3.11 + `sentence-transformers` | Локальная генерация вектора. Модель: `paraphrase-multilingual-MiniLM-L12-v2` (384d). OpenAI-совместимый API на порту 7997 |

### 3.3 Python-стек (Embeddings Server)

| Библиотека | Зачем |
|------------|-------|
| `FastAPI` | HTTP API-сервер, совместимый с OpenAI `/v1/embeddings` форматом |
| `uvicorn` | ASGI-сервер для FastAPI |
| `sentence-transformers` | Загрузка и инференс моделей (MiniLM) |
| `pydantic` | Валидация запросов/ответов |
| `torch` (CPU) | Backend для sentence-transformers |

### 3.4 Python-стек (Vosk Compat Server)

| Библиотека | Зачем |
|------------|-------|
| `vosk` | Python-биндинги к Kaldi для STT |
| `websockets` (v8.x) | WebSocket-сервер для приёма аудио |

---

## 4. Инфраструктура и DevOps

### 4.1 Контейнеризация

| Компонент | Технология | Детали |
|-----------|-----------|--------|
| **Контейнеризация** | Docker | Multi-stage сборка для Go API (builder → alpine runtime) |
| **Оркестрация** | Docker Compose | 9 сервисов: `api`, `postgres`, `redis`, `qdrant`, `neo4j`, `clickhouse`, `minio`, `vosk`, `embeddings-server` |
| **Сеть** | Docker bridge network (`deepkrai-network`) | Все сервисы общаются по внутренним DNS-именам |
| **Volumes** | Docker named volumes (8 шт.) | `postgres_data`, `redis_data`, `qdrant_data`, `neo4j_data`, `neo4j_logs`, `clickhouse_data`, `clickhouse_logs`, `minio_data` |

### 4.2 CI/CD

| Компонент | Технология | Детали |
|-----------|-----------|--------|
| **CI/CD** | GitHub Actions | Workflow [deploy.yml](file:///mnt/Work/HACKS/VORONKA/20032026/.github/workflows/deploy.yml) |
| **SSH Deployment** | `appleboy/ssh-action@v1.0.3` | Деплой по SSH на сервер |
| **Ветки** | `prod`, `dev`, `test-catalyst`, `test-lowcoware` | Автодеплой на push |
| **Стратегия** | Git pull + Docker Compose up | `docker compose --project-name deepkrai-$ENV_NAME up -d --build` |
| **Конфигурация окружений** | `/opt/deepkrai/configs/.env.$ENV_NAME` | Per-environment .env файлы на сервере, автоподстановка Docker-хостов |

### 4.3 Базовые образы

| Образ | Использование |
|-------|--------------|
| `golang:1.25.0-alpine` | Сборочный этап Go API |
| `alpine:latest` | Runtime для Go API |
| `python:3.11-slim` | Embeddings Server |
| `postgis/postgis:16-3.4` | PostgreSQL + PostGIS |
| `redis:7-alpine` | Redis |
| `qdrant/qdrant:latest` | Qdrant |
| `neo4j:5` | Neo4j + APOC + GDS |
| `clickhouse/clickhouse-server:24` | ClickHouse |
| `minio/minio:latest` | MinIO |
| `alphacep/kaldi-ru:latest` | Vosk STT |

---

## 5. Фронтенд (по GDD, в разработке)

| Компонент | Технология | Зачем |
|-----------|-----------|-------|
| **Фреймворк** | Nuxt 3 (SSR + PWA) | Server-side rendering, PWA для офлайна |
| **UI** | Tailwind CSS | Стилизация |
| **3D Engine** | TresJS (Three.js для Vue) | 3D-карта, визуализация маршрутов |
| **Карты** | Mapbox GL JS | Интерактивная карта, тайлы |
| **3D Splatting** | gsplat.js / Luma WebGL Library | Фотореалистичные 3D-экскурсии (Gaussian Splatting) |
| **Офлайн** | Service Worker + IndexedDB + Cache API | Кэширование маршрутов, аудио, тайлов |

---

## 6. Внешние интеграции (по GDD)

| Сервис | Зачем |
|--------|-------|
| **OpenWeatherMap / Яндекс.Погода** | Live-погода для 3D-карты и динамической перестройки маршрутов |
| **ElevenLabs API** | TTS: синтез аудио-историй для AI-аудиогида (голос на русском) |
| **Luma AI API** | Генерация Gaussian Splatting (.splat) из видео для 3D-экскурсий |
| **Voximplant / Zadarma** | Робо-звонки для каскадного подтверждения бронирований (DTMF) |
| **OnlySQ API** | OpenAI-совместимый API-провайдер для LLM и Whisper |

---

## 7. Протоколы и форматы

| Протокол | Где используется |
|----------|-----------------|
| **REST/JSON** | Основное API (Fiber), AI-клиенты |
| **gRPC + Protobuf** | Связь Go → Qdrant |
| **WebSocket** | Real-time уведомления, Vosk STT, погодные обновления |
| **Bolt** | Связь Go → Neo4j |
| **S3 API** | Связь Go → MinIO |
| **GeoJSON** | Маршруты (polyline), карта |
| **JWT (HS256)** | Аутентификация (Access + Refresh tokens) |
| **Multipart/form-data** | Загрузка аудио/видео |
| **PCM 16kHz mono** | Аудио-формат для Vosk STT |

---

## 8. Сводка: 9 Docker-контейнеров в Compose

```
┌─────────────────────────────────────────────────────────────┐
│                    deepkrai-network (bridge)                 │
├─────────────────────────────────────────────────────────────┤
│  api              :8080   — Go Fiber REST + WebSocket       │
│  postgres         :5432   — PostgreSQL 16 + PostGIS 3.4     │
│  redis            :6379   — Redis 7 (cache + pubsub)        │
│  qdrant           :6333/6334 — Vector DB (HTTP + gRPC)      │
│  neo4j            :7474/7687 — Graph DB (HTTP + Bolt)       │
│  clickhouse       :8123/9000 — Analytics DB (HTTP + Native) │
│  minio            :9000/9001 — Object Storage (API + UI)    │
│  vosk             :2700   — Kaldi STT (WebSocket)           │
│  embeddings-server:7997   — MiniLM Embeddings (HTTP)        │
└─────────────────────────────────────────────────────────────┘
```
