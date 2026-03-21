# Phase 10: Host Onboarding — Документация

## Цель

Реализация Zero-UI онбординга для владельцев локаций (GDD Feature 5).
Фермер / хозяин усадьбы записывает голосовое описание + загружает медиа → AI создаёт черновик локации.

## Pipeline

```
Голос (audio) ──► Vosk STT ──► Транскрипция
                                    │
                                    ▼
                              LLM (GPT-4o-mini)
                              onboardingSystemPrompt
                                    │
                                    ▼
                            Структурированные данные
                           (name, description, tags,
                            price, category, amenities,
                            capacity)
                                    │
                    ┌───────────────┼───────────────┐
                    ▼                               ▼
              Embedding API                   PostgreSQL
              → Qdrant upsert              Location (draft)
              (location_vibes)            is_published=false
```

## API Endpoints

### POST /api/v1/host/onboard
- **Auth**: JWT + RBAC (host, b2g_admin)
- **Content-Type**: multipart/form-data
- **Fields**:
  - `audio` (обязательно) — голосовое описание (webm/wav/mp3/ogg)
  - `media_urls[]` — URL предзагруженных медиафайлов
  - `latitude`, `longitude` — координаты
  - `address` — адрес
- **Response**: `OnboardResponse` с `task_id`, `status`, `progress`, `location_id`, `extracted_data`

### GET /api/v1/host/tasks/{id}
- **Auth**: JWT + RBAC
- **Response**: `OnboardResponse` (для polling прогресса)

### GET /api/v1/host/locations
- **Auth**: JWT + RBAC
- **Response**: Массив `Location` текущего хоста

## Модели данных

### AITask (таблица ai_tasks)
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| type | VARCHAR(50) | onboarding / splatting / story_generation |
| status | VARCHAR(20) | queued / processing / completed / failed |
| progress | INTEGER | 0-100 |
| input_data | JSONB | audio_url, media_urls, coordinates |
| output_data | JSONB | location_id, transcription, extracted_data |
| error | TEXT | ошибка (при failed) |

### OnboardingResult (из LLM)
| Поле | Тип | Описание |
|------|-----|----------|
| name_suggestion | string | Предложенное название |
| description_short | string | Краткое описание |
| description_literary | string | Литературное описание |
| tags | []string | 3-8 тегов |
| category | string | farm/winery/trail/guesthouse/... |
| price_per_night | int | Цена за ночь (руб) |
| amenities | []string | Удобства |
| capacity | int | Вместимость |

## Demo Policy

- **Mock-режим** (AI_API_KEY=""): fake STT возвращает «ну я дядя ваня делаю сыр…», fake LLM возвращает «Козья ферма дяди Вани»
- **Real AI**: Vosk STT → LLM (onboardingSystemPrompt) → реальные данные

## Затронутые файлы

| Файл | Действие | Описание |
|------|----------|----------|
| migrations/010_create_ai_tasks.sql | NEW | Таблица ai_tasks |
| models/onboarding.go | NEW | AITask, OnboardingResult, OnboardResponse |
| models/onboarding_test.go | NEW | 4 unit-теста |
| ai/llm.go | MODIFY | ExtractLocationData + onboardingSystemPrompt |
| ai/mock.go | MODIFY | mockLocationData, mockOnboardingTranscription |
| database/ai_task_repository.go | NEW | CRUD для ai_tasks |
| database/location_repository.go | MODIFY | FindByOwnerID |
| services/onboarding.go | NEW | OnboardingService (pipeline) |
| handlers/onboarding.go | NEW | 3 HTTP-обработчика |
| handlers/router.go | MODIFY | Host group + RBAC |
| handlers/openapi.go | MODIFY | 3 новых endpoint |
| cmd/api/main.go | MODIFY | Injection OnboardingService |

## Acceptance Criteria

- [x] `go build ./...` — OK
- [x] `go test ./internal/models/ -v` — 4 новых теста PASS
- [x] `go test ./internal/middleware/ -v` — 8 тестов PASS
- [x] `go test ./internal/services/ -v` — все тесты PASS
- [x] POST /host/onboard требует JWT + role host/b2g_admin
- [x] Pipeline: STT → LLM → Embedding → Location draft
- [x] Mock-режим: детерминированный ответ «Козья ферма дяди Вани»
- [x] Миграция 010 автоматически выполняется при запуске
