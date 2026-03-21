# Фаза 5: Demo-срез Location Detail

## Описание

Экран детали локации — мост между рекомендацией AI и конкретным местом.
Покрывает GDD Feature 3, частично GDD Feature 1 result flow.

## Эндпоинты

### GET /api/v1/locations/{id}

Полные данные локации для экрана детали. Публичный.

**Payload:**

- title (name) + slug
- category, tags
- description_short / description_full
- price_per_night, capacity
- coordinates (latitude/longitude)
- preview_image_url (hero image)
- gallery_urls (массив фотографий)
- splat_url (3D-сцена)
- density_level (цветовая индикация плотности)
- access_level (Hidden Gems)
- child_friendly
- address

### GET /api/v1/locations/{id}/splat

Отдельный fetch 3D-сцены (Gaussian Splatting) для ленивой загрузки.

**Payload:**

```json
{
  "success": true,
  "data": {
    "location_id": "uuid",
    "location_name": "Винодельня Абрау-Дюрсо",
    "has_splat": true,
    "splat_url": "https://storage.kudytudy.ru/splats/abrau-dyurso.splat"
  }
}
```

### POST /api/v1/trips

Создание trip details из экрана location flow (уже реализовано в фазе 3).

## Миграция

`006_add_gallery_urls.sql` — добавляет `gallery_urls TEXT[]` для массива фотографий галереи.

## Связанные файлы

| Файл                                  | Назначение                       |
| ------------------------------------- | -------------------------------- |
| `models/location.go`                  | Location модель с gallery_urls   |
| `database/location_repository.go`     | SQL SELECT + Scan с gallery_urls |
| `handlers/location.go`                | GetByID + GetSplat handlers      |
| `handlers/router.go`                  | /locations/:id/splat маршрут     |
| `handlers/openapi.go`                 | Location, SplatResponse schema   |
| `migrations/006_add_gallery_urls.sql` | Миграция                         |
