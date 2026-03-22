# Phase 10 Update: Location Media & 3D Splat

## 1. Управление фото локации

### Что изменилось

- Добавлены `preview_image_url` и `gallery_urls` в `CreateLocationRequest` и `UpdateLocationRequest`
- `POST /api/v1/locations` — теперь принимает hero-фото и галерею при создании
- `PUT /api/v1/locations/{id}` — хост может обновлять фото через:
  ```json
  {
    "preview_image_url": "http://.../preview.jpg",
    "gallery_urls": ["http://.../photo1.jpg", "http://.../photo2.jpg"]
  }
  ```
- `LocationRepository.Create` — INSERT включает `preview_image_url`, `gallery_urls`
- `LocationRepository.Update` — динамический SET поддерживает обновление фото

### Затронутые файлы

| Файл | Изменения |
|------|-----------|
| models/location.go | +PreviewImageURL, +GalleryURLs в CreateLocationRequest и UpdateLocationRequest |
| services/location.go | +preview/gallery в Location при создании |
| database/location_repository.go | +preview/gallery в INSERT и UPDATE SET |

---

## 2. 3D Gaussian Splatting (Mock Pipeline)

### Endpoint

`POST /api/v1/host/splat` — JWT + RBAC (host / b2g_admin)

### Request

```json
{
  "location_id": "uuid-of-location",
  "video_url": "http://.../my-farm-walkthrough.mp4"
}
```

### Response

```json
{
  "success": true,
  "message": "3D-сцена готова",
  "data": {
    "task_id": "uuid",
    "status": "completed",
    "progress": 100,
    "message": "3D-сцена готова",
    "splat_url": "http://141.98.7.225:9102/deepkrai-media/locations/splat/kitchen.splat"
  }
}
```

### Mock Pipeline

```
Видео → Извлечение кадров → Облако точек → Обучение GS → Экспорт 3D → Оптимизация → Привязка к локации
```

Каждый шаг симулируется с controlled latency и обновлением прогресса в `ai_tasks`.

### Category → Splat Mapping

| Категория | Splat файл | Обоснование |
|-----------|-----------|-------------|
| winery | garden.splat | Виноградники, открытые пространства |
| farm | kitchen.splat | Фермерские постройки |
| nature | stump.splat | Природные объекты, деревья |
| gastro | counter.splat | Прилавки, интерьеры |
| resort/guesthouse | room.splat | Комнаты, помещения |
| extreme | bicycle.splat | Активный отдых |
| cultural | bonsai.splat | Культурные объекты |
| trail | stump.splat | Тропы, природа |
| fallback | garden.splat | Универсальный |

### Затронутые файлы

| Файл | Изменения |
|------|-----------|
| models/onboarding.go | +SplattingInputData, +SplattingOutputData, +SplattingResponse, +SplattingProgressMessage |
| services/onboarding.go | +StartSplatting (mock pipeline), +selectMockSplatURL |
| handlers/onboarding.go | +StartSplatting handler |
| handlers/router.go | +POST /host/splat route |
| handlers/openapi.go | +POST /host/splat endpoint |
| database/location_repository.go | +UpdateSplatURL |
