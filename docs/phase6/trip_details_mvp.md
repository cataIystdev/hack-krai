# Phase 6: Trip Details MVP — Документация

## Обзор

Phase 6 полностью закрывает core step **Trip Details** из GDD.

## API Endpoints

### POST /api/v1/trips — Создание поездки

**Требует JWT.**

```json
{
  "date_from": "2026-04-10",
  "date_to": "2026-04-13",
  "budget_rub": 50000,
  "budget_tier": "comfort",
  "transport": "car",
  "group_size": 4,
  "group_composition": { "adults": 2, "children": [{ "age": 8 }] },
  "format": "multi_day",
  "vibe_vector_id": "qdrant-uuid-or-null"
}
```

| Поле              | Тип    | Обязательное | Default   | Описание                     |
| ----------------- | ------ | :----------: | --------- | ---------------------------- |
| date_from         | date   |      ✅      | —         | Дата начала YYYY-MM-DD       |
| date_to           | date   |      ✅      | —         | Дата окончания               |
| budget_rub        | int    |      ❌      | 0         | Бюджет ₽                     |
| budget_tier       | enum   |      ❌      | comfort   | economy/comfort/premium      |
| transport         | enum   |      ❌      | car       | car/public/walk/bike         |
| group_size        | int    |      ❌      | 1         | Кол-во участников            |
| group_composition | object |      ❌      | {}        | JSONB состава                |
| format            | enum   |      ❌      | multi_day | day_trip/weekend/multi_day   |
| vibe_vector_id    | uuid   |      ❌      | null      | Qdrant vibe-вектор создателя |

### PUT /api/v1/trips/{id} — Обновление поездки

**Требует JWT.** Доступно только создателю. Partial update — обновляются только переданные поля.

```json
{
  "budget_rub": 75000,
  "format": "weekend"
}
```

### GET /api/v1/trips/{id} — Детали поездки

Возвращает `TripWithMembers` — объект поездки со всеми полями + массив `members`.

### GET /api/v1/trips — Список поездок пользователя

Возвращает все поездки, в которых пользователь является участником.

## Миграция

```sql
-- 007_add_trip_format_and_vibe_vector.sql
ALTER TABLE trips ADD COLUMN format TEXT NOT NULL DEFAULT 'multi_day';
ALTER TABLE trips ADD COLUMN vibe_vector_id UUID;
ALTER TABLE trips ADD CONSTRAINT trips_format_check CHECK (format IN ('day_trip', 'weekend', 'multi_day'));
```

## Для фронтенда

- Форма Trip Details отправляет все поля в `POST /api/v1/trips`
- `format` определяет UX: day_trip = однодневная, weekend = выходные, multi_day = многодневная
- `vibe_vector_id` — передавать `vector_id` из `POST /api/v1/profile/voice` если пользователь прошёл профилирование
- Обновление через `PUT /api/v1/trips/{id}` — отправлять только изменённые поля
