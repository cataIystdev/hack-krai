# Документация: Локации (CRUD + пространственный поиск)

## Обзор

API локаций предоставляет CRUD-операции для управления туристическими точками Краснодарского края. Данные хранятся в PostgreSQL с расширением PostGIS для пространственных запросов.

## Схема данных

| Поле                | Тип                   | Описание                                                               |
| ------------------- | --------------------- | ---------------------------------------------------------------------- |
| `id`                | UUID                  | Первичный ключ                                                         |
| `owner_id`          | UUID FK               | Владелец (host)                                                        |
| `slug`              | TEXT UNIQUE           | URL-дружественный идентификатор (автогенерация из name)                |
| `name`              | TEXT                  | Название                                                               |
| `description_short` | TEXT                  | Краткое описание (для карточек)                                        |
| `description_full`  | TEXT                  | Полное описание (литературный текст)                                   |
| `category`          | TEXT                  | Свободная категория (winery, farm, trail, gastro, nature, camping...)  |
| `tags`              | TEXT[]                | Массив тегов                                                           |
| `price_per_night`   | INTEGER               | Стоимость за ночь (₽), 0 = бесплатно                                   |
| `capacity`          | INTEGER               | Вместимость                                                            |
| `access_level`      | ENUM                  | `open` / `semi_open` / `hidden` (Hidden Gems)                          |
| `density_level`     | ENUM                  | `red` (высокая) / `yellow` (сезонная) / `green` (пустынно, Hidden Gem) |
| `child_friendly`    | BOOLEAN               | Подходит для детей                                                     |
| `splat_url`         | TEXT                  | URL на .splat файл (3D Gaussian Splatting)                             |
| `vibe_vector_id`    | UUID                  | ID вектора vibe-профиля в Qdrant                                       |
| `address`           | TEXT                  | Адрес                                                                  |
| `is_published`      | BOOLEAN               | Публикация                                                             |
| `geo`               | GEOMETRY(Point, 4326) | PostGIS: координаты (GiST-индекс)                                      |

## Endpoints

### GET /api/v1/locations — Поиск локаций

Публичный. Возвращает опубликованные локации с фильтрами.

**Query параметры:**

| Параметр                                      | Тип    | Описание                                            |
| --------------------------------------------- | ------ | --------------------------------------------------- |
| `lat` + `lon` + `radius_km`                   | float  | Поиск по радиусу (PostGIS ST_DWithin)               |
| `min_lat` + `max_lat` + `min_lon` + `max_lon` | float  | Поиск по bbox (PostGIS ST_Within + ST_MakeEnvelope) |
| `category`                                    | string | Фильтр по категории                                 |
| `child_friendly`                              | bool   | Фильтр по пригодности для детей                     |
| `density_level`                               | string | `red` / `yellow` / `green`                          |
| `page`                                        | int    | Страница (default: 1)                               |
| `per_page`                                    | int    | Записей на странице (default: 20, max: 100)         |

**Примеры:**

```bash
# Все локации
curl http://localhost:8080/api/v1/locations

# Радиус 50 км от Краснодара
curl "http://localhost:8080/api/v1/locations?lat=45.03&lon=38.97&radius_km=50"

# Винодельни
curl "http://localhost:8080/api/v1/locations?category=winery"

# Hidden Gems (зелёные, тихие)
curl "http://localhost:8080/api/v1/locations?density_level=green"

# Для детей
curl "http://localhost:8080/api/v1/locations?child_friendly=true"
```

### POST /api/v1/locations — Создание

**Требует:** JWT + роль `host` или `b2g_admin`.

```bash
curl -X POST http://localhost:8080/api/v1/locations \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Козья ферма дяди Вани",
    "latitude": 44.2878,
    "longitude": 40.1763,
    "category": "farm",
    "tags": ["сыр", "козы"],
    "price_per_night": 5000,
    "capacity": 4,
    "access_level": "semi_open",
    "density_level": "green",
    "child_friendly": true,
    "address": "Краснодарский край, Хаджох",
    "is_published": true
  }'
```

### GET /api/v1/locations/:id — Получение по ID

Публичный. Возвращает все поля локации, включая latitude/longitude.

### PUT /api/v1/locations/:id — Обновление

**Требует:** JWT + владелец (owner_id). Все поля опциональны.

### DELETE /api/v1/locations/:id — Удаление

**Требует:** JWT + владелец (owner_id).

## PostGIS

Координаты хранятся как `GEOMETRY(Point, 4326)` с GiST-индексом.

- **Радиус:** `ST_DWithin(geo::geography, point::geography, meters)` — точный расчёт расстояния на сфере (в метрах).
- **BBox:** `ST_Within(geo, ST_MakeEnvelope(xmin, ymin, xmax, ymax, 4326))` — прямоугольная область.
- **Сортировка:** При радиусном поиске результаты сортируются по расстоянию (`<->` оператор).

## Категории

Категории — **свободный текст** (не enum). В seed-данных используются:
`winery`, `farm`, `trail`, `guesthouse`, `gastro`, `nature`, `camping`, `resort`, `extreme`, `cultural`, `beach`.

Можно добавлять любые новые категории без миграций.
