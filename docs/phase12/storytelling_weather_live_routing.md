# Phase 12: Storytelling + Weather + Live Routing

## Goal

Собрать следующий слой после route MVP:

1. `GET /api/v1/weather/region`
2. storytelling route assets
3. `POST /api/v1/route/{id}/rebuild`

## Implemented in this slice

### Weather

- `WeatherService` с региональной погодной сводкой
- `GET /api/v1/weather/region`
- weather advisory для точек маршрута

### Storytelling

- `POST /api/v1/route/{id}/generate-stories`
- `GET /api/v1/route/{id}/stories`
- story metadata хранится в `route_points`
- поля:
  - `audio_story_url`
  - `story_text`
  - `weather_condition`
  - `weather_temp_c`

### Live routing

- `POST /api/v1/route/{id}/rebuild`
- rebuild учитывает weather advisory
- risky outdoor категории (`trail`, `extreme`) при warning-сценарии отодвигаются в конец
- summary и route points обновляются в БД

## Current limitations

- пока нет реального WebSocket transport layer
- погодный источник mock-friendly, не внешний API
- story asset сейчас сохраняется как файл-артефакт через storage, без реального TTS

## Main files

- `api/migrations/012_route_weather_and_storytelling.sql`
- `api/internal/models/weather.go`
- `api/internal/services/weather.go`
- `api/internal/services/storytelling.go`
- `api/internal/services/phase12_route.go`
- `api/internal/handlers/weather.go`
- `api/internal/handlers/route.go`
