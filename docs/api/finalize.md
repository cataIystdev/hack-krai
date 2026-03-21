# Финализация профиля: POST /api/v1/profile/finalize

## Статус

mock-safe, shape frozen

## Используется экранами

- Recommendation list
- Recommendation map
- Переход из vibe passport в карту/локации

## Запрос

Метод: POST
Тело запроса (опционально):

```json
{
  "limit": 10,
  "child_friendly_only": false
}
```

- `limit` -- максимальное количество рекомендаций (по умолчанию 10, максимум 50)
- `child_friendly_only` -- фильтровать только child-friendly локации

При пустом теле запроса используются значения по умолчанию.

## Пример ответа (200)

```json
{
  "success": true,
  "data": {
    "recommendations": [
      {
        "location_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "score": 0.92,
        "name": "Винодельня Лефкадия",
        "category": "winery",
        "description_short": "Премиальная винодельня с дегустацией и экскурсиями по виноградникам.",
        "tags": ["вино", "природа", "гастрономия"],
        "preview_image_url": "https://cdn.kudytudy.ru/locations/lefkadia.jpg",
        "splat_url": "",
        "latitude": 44.5123,
        "longitude": 38.1456,
        "density_level": "green",
        "reason_short": "Идеальное совпадение: вино, природа",
        "tags_match": ["вино", "природа"],
        "child_friendly": true
      }
    ],
    "total_found": 1,
    "is_curated": false
  }
}
```

## Новые поля (Фаза 3)

- `reason_short` -- краткая причина рекомендации на русском (генерируется из score + tags_match)
- `tags_match` -- теги, совпавшие между профилем пользователя и локацией (пересечение)
- `child_friendly` -- подходит ли локация для детей
- `is_curated` -- true, если рекомендации из curated demo набора (вектор пользователя не найден)

## Поведение curated fallback

Если вектор пользователя ещё не создан (голосовое профилирование не пройдено),
endpoint возвращает curated набор опубликованных локаций с синтетическими score (0.95 -> 0.40).
Поле `is_curated` будет `true`.

## Известные ограничения

- `tags_match` зависит от наличия тегов в Qdrant payload профиля
- `reason_short` генерируется алгоритмически, не через LLM

## Безопасно для фронтенда

Да. Shape обратно совместим: новые поля добавлены, старые не удалены.
