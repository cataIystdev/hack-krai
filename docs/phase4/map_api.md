# Фаза 4: Map API для демо

## Описание

API для отображения локаций на интерактивной карте Краснодарского края.
Покрывает GDD Feature 2 (интерактивная 3D-карта).

## Эндпоинт

```
GET /api/v1/map/locations
```

Публичный эндпоинт. JWT опционален — при наличии авторизации точки обогащаются персональными рекомендациями.

## Три режима работы

### 1. Bbox (по умолчанию)

Пространственный поиск через PostGIS ST_Within.

```
GET /api/v1/map/locations?min_lat=43.5&max_lat=45.5&min_lon=36.5&max_lon=41.0
```

### 2. Demo

Curated набор с предустановленными рекомендациями для выбранного профиля.

```
GET /api/v1/map/locations?demo=true&profile=calm_wine_mountains
```

Доступные профили:

| Профиль               | Описание                                   | Топ-рекомендации                   |
| --------------------- | ------------------------------------------ | ---------------------------------- |
| `calm_wine_mountains` | Спокойный отдых: горы, вино, тишина        | Лефкадия, Козья ферма, Абрау-Дюрсо |
| `active_adventure`    | Активный отдых: каньоны, рафтинг, горы     | Каньон Белой, Руфабго, Грачёв      |
| `family_kids`         | Семейный отдых: фермы, дети, природа       | Козья ферма, Сыроварня, Галицкого  |
| `gastro_cultural`     | Гастрономия и культура: вино, сыр, история | Сыроварня, Абрау-Дюрсо, Лефкадия   |

### 3. Hybrid

При наличии JWT и vibe-вектора пользователя, точки обогащаются скорами из Qdrant.

## Параметры запроса

| Параметр         | Тип     | Описание                               |
| ---------------- | ------- | -------------------------------------- |
| `min_lat`        | number  | Минимальная широта bbox                |
| `max_lat`        | number  | Максимальная широта bbox               |
| `min_lon`        | number  | Минимальная долгота bbox               |
| `max_lon`        | number  | Максимальная долгота bbox              |
| `category`       | string  | Фильтр по категории                    |
| `density_level`  | string  | Фильтр по плотности (red/yellow/green) |
| `is_recommended` | boolean | Только рекомендованные                 |
| `limit`          | integer | Лимит (по умолчанию 50, макс. 200)     |
| `demo`           | boolean | Demo-режим                             |
| `profile`        | string  | Demo-профиль                           |

## Формат ответа

```json
{
  "success": true,
  "data": {
    "points": [
      {
        "id": "uuid",
        "name": "Винодельня Лефкадия",
        "latitude": 44.723,
        "longitude": 37.781,
        "category": "winery",
        "density_level": "yellow",
        "preview_image_url": "https://...",
        "description_short": "Премиальная винодельня в долине Лефкадия",
        "is_recommended": true,
        "recommendation_score": 0.96
      }
    ],
    "total": 20,
    "profile": "calm_wine_mountains"
  }
}
```

## Связанные файлы

| Файл                              | Назначение                                                    |
| --------------------------------- | ------------------------------------------------------------- |
| `models/map.go`                   | DTO: MapPoint, MapLocationFilter, MapLocationsResponse        |
| `services/map.go`                 | Бизнес-логика: 3 режима, 4 demo-профиля, обогащение из Qdrant |
| `handlers/map.go`                 | HTTP-обработчик с опциональным JWT                            |
| `database/location_repository.go` | SearchForMap (PostGIS)                                        |
| `handlers/openapi.go`             | OpenAPI спецификация                                          |
| `models/map_test.go`              | Unit-тесты моделей                                            |
| `services/map_service_test.go`    | Unit-тесты сервиса                                            |
