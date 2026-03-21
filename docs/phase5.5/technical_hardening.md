# Фаза 5.5: Техническое усиление после демо

## Дата: 2026-03-21

## Обзор

Закрытие 5 технических рисков, обнаруженных при code review фаз 0-5.
Все исправления затрагивают demo-touched код и устраняют panic-level traps,
race conditions и operational risks.

---

## 1. Safe Media Upload

**Проблема:** `POST /api/v1/media/upload` — при `storage == nil` (MinIO недоступен) вызов
`h.storage.Upload(...)` приводил к nil pointer dereference (panic).

**Решение:** Nil-guard в начале `Upload()` — возвращает 503 Service Unavailable.

**Файлы:**

- `handlers/media.go` — nil-guard на `h.storage`
- `handlers/media_test.go` — `TestMediaUploadStorageUnavailable`

---

## 2. Atomic Trip Join

**Проблема:** `CountMembers()` и `AddMember()` — два отдельных SQL-запроса.
При параллельных JOIN оба могут пройти проверку `count < group_size`.

**Решение:** Атомарный `INSERT ... SELECT WHERE (SELECT count(*) ...) < maxGroupSize`.

**Файлы:**

- `database/trip_repository.go` — `AddMemberAtomic()`, `ErrTripFull`
- `services/trip.go` — `Join()` использует `AddMemberAtomic()`

---

## 3. Scene Vectors Isolation

**Проблема:** `SeedSceneVectors()` записывал эмбеддинги сцен в коллекцию `user_vibes`.
`Swipe()` читал вектор сцены из той же коллекции. Доменная модель "сцена" смешивалась
с "пользователем" — при совпадении UUID происходила перезапись.

**Решение:** Выделена отдельная коллекция `scene_vibes` в Qdrant.

**Файлы:**

- `database/vibe_repository.go` — `CollectionSceneVibes`
- `database/qdrant_collections.go` — конфигурация `scene_vibes`
- `services/vibe.go` — `SeedSceneVectors()` и `Swipe()` используют `CollectionSceneVibes`

---

## 4. Voice Pipeline — ограничение размера

**Проблема:** `io.ReadAll(file)` буферизовал весь аудиофайл в RAM.
При 100 МБ файлах и нескольких параллельных запросах — memory pressure.

**Решение:** `io.LimitReader(file, 25MB+1)` + проверка размера → 413 Payload Too Large.

**Файлы:**

- `handlers/vibe.go` — `maxAudioSize`, `io.LimitReader`

---

## 5. Health — Liveness/Readiness Split

**Проблема:** Каждый запрос к `/health` пинговал все 6 БД синхронно.
Kubernetes probes создавали нагрузку.

**Решение:** Разделение на три endpoint-а:

- `GET /health/live` — мгновенный ответ (liveness)
- `GET /health/ready` — полный пинг (readiness)
- `GET /health` — alias на `Ready` (обратная совместимость)

**Файлы:**

- `handlers/health.go` — `Live()`, `Ready()`, `Check()` как alias
- `handlers/router.go` — маршруты `/health/live` и `/health/ready`
- `handlers/openapi.go` — схемы для новых endpoint-ов

---

## Связанные файлы

| Файл                             | Изменения                                                   |
| -------------------------------- | ----------------------------------------------------------- |
| `handlers/media.go`              | nil-guard на storage                                        |
| `handlers/media_test.go`         | `TestMediaUploadStorageUnavailable`, обновлены существующие |
| `database/trip_repository.go`    | `AddMemberAtomic()`, `ErrTripFull`                          |
| `services/trip.go`               | atomic `Join()`                                             |
| `database/vibe_repository.go`    | `CollectionSceneVibes`                                      |
| `database/qdrant_collections.go` | конфигурация `scene_vibes`                                  |
| `services/vibe.go`               | `SeedSceneVectors` / `Swipe` → `CollectionSceneVibes`       |
| `handlers/vibe.go`               | `maxAudioSize`, `io.LimitReader`                            |
| `handlers/health.go`             | `Live()`, `Ready()`                                         |
| `handlers/router.go`             | маршруты health                                             |
| `handlers/openapi.go`            | health/live, health/ready                                   |
