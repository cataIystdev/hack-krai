# 🗺 KUDYTUDY — Актуальный Roadmap (vs GDD)

> Дата актуализации: 22 марта 2026.
> Этот документ отражает реальный статус кодовой базы по отношению к изначальному бизнес-видению (GDD).

---

## 🏗 Executive Summary

На данный момент полностью реализован **Tourist Core MVP** и значительная часть "магии" из GDD. Бэкенд выдерживает основные юзкейсы туриста: от голосового профилирования до 3D-карты и построения маршрута. Бизнес-часть (бронирование, онбординг) переведена в стадию MVP (код написан, базовые флоу закрыты и протестированы).

**Главная задача сейчас** — довести инфраструктурные и социальные фичи (WebSocket, Offline, B2G, Karma) до продакшен-состояния.

---

## 🎯 Сверка статуса по фичам GDD

### 🧠 Фича 1: Умное планирование поездки
| Функционал | Статус | Комментарий |
|---|---|---|
| Голосовое профилирование | ✅ DONE | Whisper + LLM + Qdrant. Ограничение аудио 25МБ. Работает стабильно. |
| Эмоциональные 3D-свайпы | ✅ DONE | Tinder-механика на `scene_vibes` и Qdrant. |
| Trip Details (бюджет, транспорт и т.д.) | ✅ DONE | Редактирование, валидация форматов (day_trip, weekend, multi_day). |
| Групповое планирование (Join by invite) | ✅ DONE | Приглашения, `AddMemberAtomic`, `MergeGroupVibes`, потребление merged вектора в роутинге. |
| Строгий Privacy & Concurrency | ✅ DONE | Membership-guard в `BuildTripRoute`. `AddMemberAtomic` с `SELECT FOR UPDATE`. |

### 🌍 Фича 2: Интерактивная 3D-карта
| Функционал | Статус | Комментарий |
|---|---|---|
| Map API (bbox, density, recommendations) | ✅ DONE | Эндпоинты возвращают нужный JSON (цветовая загруженность 🔴🟡🟢, координаты). |
| Live-погода по регионам | ✅ DONE | `GET /api/v1/weather/region` возвращает 5 ключевых точек с severity. |
| WebSocket WebSocket push погоды | ✅ DONE | `/ws/v1/weather` + `/ws/v1/notifications` реализованы. WeatherTicker 60с. |

### 🧊 Фича 3: Immersive 3D-экскурсии (Gaussian Splatting)
| Функционал | Статус | Комментарий |
|---|---|---|
| Хранение и раздача `.splat` файлов | ✅ DONE | Поднято в MinIO, работает lazy load через `GET /locations/{id}/splat`. |
| Pipeline генерации | 🔶 PARTIAL | Сделан mock-пайплайн (POST /host/splat) для демо-целей. Реальной интеграции с Luma AI API нет. |

### 🚗 Фича 4: AI Live Routing & Storytelling
| Функционал | Статус | Комментарий |
|---|---|---|
| AI Аудиогид (ElevenLabs TTS) | ✅ DONE | Интеграция с ElevenLabs завершена, генерит mp3 в MinIO и возвращает URL. Работает fallback на txt. |
| Базовый роутинг | ✅ DONE | `POST /api/v1/route/build` работает логически. |
| Live Routing (перестройка от погоды) | ✅ DONE | `POST /api/v1/route/{id}/rebuild` сделан. |

### 🚜 Фича 5: Zero-UI Онбординг для владельцев
| Функционал | Статус | Комментарий |
|---|---|---|
| Pipeline обработки (голос → листинг) | ✅ DONE | `POST /host/onboard`. Whisper → LLM → Qdrant → PostgreSQL работает end-to-end. |
| Управление галереей и превью | ✅ DONE | `preview_image_url`, `gallery_urls` поддерживаются в БД и API. |

### 📞 Фича 6: Бронирование и Инвентаризация
| Функционал | Статус | Комментарий |
|---|---|---|
| Базовый CRUD (слоты, создание, отмена) | ✅ DONE | Слой данных и API реализованы, тесты зелёные. |
| Каскадное подтверждение (Уровень 1 - Web) | ✅ DONE | Владелец может подтвердить / отклонить. |
| Робо-звонилка (Уровень 2 - Телефония) | ❌ TODO | Интеграции с Voximplant / Zadarma API нет. |

### 🛡 Фича 7: Защита «Скрытых Сокровищ» (Karma)
| Функционал | Статус | Комментарий |
|---|---|---|
| Фундамент БД | ✅ DONE | Поля `access_level`, `karma`, `karma_threshold` существуют в схемах. |
| Логика отзывов и начисления кармы | ✅ DONE | Отзывы, начисление кармы и API реализованы. |

### 📊 Фича 8: B2G Аналитический дашборд
| Функционал | Статус | Комментарий |
|---|---|---|
| Инфраструктура БД (ClickHouse) | ✅ DONE | Инфраструктура развернута. |
| Событийная телеметрия (Стрим ивентов) | ✅ DONE | Реализовано: `telemetry_service` через Redis Streams + ClickHouse консьюмер. |
| B2G Endpoints (Heatmap, Predictions) | ✅ DONE | Эндпоинты агрегации (Heatmap, Predictions) реализованы. |

### 📡 Фича 9: PWA Офлайн-режим
| Функционал | Статус | Комментарий |
|---|---|---|
| Offline-bundle сборка | ✅ DONE | `GET /api/v1/route/{id}/offline-bundle` реализован. |
| Sync endpoint | ✅ DONE | `POST /api/v1/sync` реализован. |

---

## 📝 Разбор бэклога: Что делать дальше

Основываясь на чекапе выше, у нас вырисовывается четкий бэклог до 100% покрытия:

### ✅ 1. Архитектурный P0 (Долги Core MVP) — Phase 16
- [x] **Privacy & Concurrency в Trips**: membership-проверка в `BuildTripRoute`, `AddMemberAtomic` с `SELECT FOR UPDATE`.
- [x] **Merged Vibe Vector**: `computeVibeScores` использует предрассчитанный `merged_vibe_vector_id` из Qdrant `group_vibes` с fallback на per-member averaging.

### ✅ 2. Социалка и Геймификация (Phase 13)
- [x] Написать `review-service` (создание отзывов).
- [x] Реализовать механизм обновления `karma` юзера на основе отзывов.
- [x] Опционально: фильтрация `hidden` локаций в `map-service`.

### ✅ 3. Инфраструктура реального времени — Phase 17
- [x] WebSocket Hub для `/ws/v1/weather` и `/ws/v1/notifications` (PubSub через Redis).
- [x] WeatherTicker — пуш обновлений погоды каждые 60с через WS Hub.

### ✅ 4. Data & Analytics (Phase 14)
- [x] Написать воркер-генератор событий (`event-ingestion` middleware + ClickHouse consumer) для стриминга в аналитику.
- [x] Сделать аналитические эндпоинты для B2G (`/analytics/heatmap`, `/analytics/predictions`).

### 5. Offline Sync (Phase 15)
- Собирать ZIP-bundle маршрута на сервере (`/api/v1/route/{id}/offline-bundle`).
- Реализовать endpoint синхронизации офлайн-событий.
