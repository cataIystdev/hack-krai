# Модуль Vibe-профилирования (Voice-to-Vibe)

## Описание

Мультимодальный пайплайн профилирования туриста:

1. **Голосовой ввод** → Whisper STT (транскрипция)
2. **Извлечение осей** → LLM (структурированный JSON)
3. **Векторизация** → Embeddings (3072d)
4. **Сохранение** → Qdrant (user_vibes) + PostgreSQL (vibe_vector_id)

## Оси профиля

| Ось                | 0.0                 | 1.0              |
| ------------------ | ------------------- | ---------------- |
| stress_level       | Полное расслабление | Экстрим          |
| solitude_vs_social | Уединение           | Тусовка          |
| budget_sensitivity | Бюджет не важен     | Строгая экономия |
| nature_vs_urban    | Природа             | Город            |
| adventure_level    | Пляжный отдых       | Экстрим-туризм   |

## API Endpoints

### POST /api/v1/profile/voice

Голосовое профилирование. Multipart form с полем `audio`.

```bash
curl -X POST http://localhost:8080/api/v1/profile/voice \
  -H "Authorization: Bearer <jwt>" \
  -F "audio=@recording.mp3"
```

### POST /api/v1/profile/swipe

Свайп сцены. Сдвигает vibe-вектор к/от сцены.

```bash
curl -X POST http://localhost:8080/api/v1/profile/swipe \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"scene_id":"a1b2c3d4-1111-4000-8000-000000000001","direction":"right"}'
```

### POST /api/v1/profile/finalize

Top-10 рекомендаций по cosine similarity.

```bash
curl -X POST http://localhost:8080/api/v1/profile/finalize \
  -H "Authorization: Bearer <jwt>"
```

### GET /api/v1/profile/scenes

Список сцен для свайп-анкеты.

```bash
curl http://localhost:8080/api/v1/profile/scenes \
  -H "Authorization: Bearer <jwt>"
```

## Формула свайпа

```
new = normalize(α * user_vector ± (1-α) * scene_vector)
α = 0.85 (85% исходного вектора сохраняется)
+ для right (нравится), - для left (не нравится)
```

## Конфигурация

```env
AI_BASE_URL=https://api.onlysq.ru/ai/openai/
AI_API_KEY=<key>
AI_WHISPER_MODEL=whisper-1
AI_LLM_MODEL=gpt-4o-mini
AI_EMBEDDINGS_MODEL=text-embedding-3-large
```

При пустом `AI_API_KEY` используется mock-режим.

## Схема данных

### swipe_scenes

| Поле          | Тип  | Описание        |
| ------------- | ---- | --------------- |
| id            | UUID | PK              |
| title         | TEXT | Заголовок       |
| description   | TEXT | Описание        |
| image_url     | TEXT | URL изображения |
| display_order | INT  | Порядок         |

### Qdrant: user_vibes / location_vibes

- Размерность: 3072 (text-embedding-3-large)
- Метрика: Cosine Similarity
- Point ID = User/Location UUID
