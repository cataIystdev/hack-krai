# Модуль планирования поездок (Trip Details)

Фаза 2.5: планирование поездок с invite-ссылками для групп.

## Схема данных

### Таблица trips

| Поле                  | Тип              | Описание                                |
| --------------------- | ---------------- | --------------------------------------- |
| id                    | UUID PK          | Идентификатор поездки                   |
| creator_id            | UUID FK -> users | Создатель                               |
| date_from             | DATE             | Дата начала                             |
| date_to               | DATE             | Дата окончания                          |
| budget_rub            | INTEGER          | Бюджет (рубли)                          |
| budget_tier           | TEXT             | economy/comfort/premium                 |
| transport             | TEXT             | car/public/walk/bike                    |
| group_size            | INTEGER          | Макс. участников                        |
| group_composition     | JSONB            | {"adults": N, "children": [{"age": N}]} |
| invite_token          | UUID UNIQUE      | Токен приглашения                       |
| merged_vibe_vector_id | UUID             | Средневзвешенный вibe-вектор            |
| status                | TEXT             | planning/active/completed/cancelled     |

### Таблица trip_members

| Поле           | Тип              | Описание                |
| -------------- | ---------------- | ----------------------- |
| id             | UUID PK          | Идентификатор записи    |
| trip_id        | UUID FK -> trips | Поездка                 |
| user_id        | UUID FK -> users | Пользователь (nullable) |
| display_name   | TEXT             | Имя участника           |
| role           | TEXT             | creator/member          |
| vibe_vector_id | UUID             | vibe-вектор (nullable)  |
| tags           | TEXT[]           | Теги предпочтений       |
| is_child       | BOOLEAN          | Ребёнок                 |

## API-эндпоинты

### POST /api/v1/trips (JWT)

Создание поездки. Создатель автоматически добавляется как участник.

```bash
curl -X POST http://localhost:8080/api/v1/trips \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{
    "date_from": "2026-04-10",
    "date_to": "2026-04-13",
    "budget_rub": 50000,
    "transport": "car",
    "group_size": 4,
    "group_composition": {"adults": 2, "children": [{"age": 8}, {"age": 12}]}
  }'
```

### GET /api/v1/trips/{id} (JWT)

Детали поездки со списком участников.

```bash
curl http://localhost:8080/api/v1/trips/<trip_id> \
  -H "Authorization: Bearer <jwt>"
```

### POST /api/v1/trips/{id}/invite (JWT, только создатель)

Генерация нового invite-токена (старый аннулируется).

```bash
curl -X POST http://localhost:8080/api/v1/trips/<trip_id>/invite \
  -H "Authorization: Bearer <jwt>"
```

### POST /api/v1/trips/{id}/join (без авторизации)

Присоединение участника по invite-токену.

```bash
curl -X POST http://localhost:8080/api/v1/trips/<trip_id>/join \
  -H "Content-Type: application/json" \
  -d '{"invite_token":"<uuid>","display_name":"Маша","tags":["сыр","ферма"]}'
```

### GET /api/v1/trips/{id}/members (JWT)

Список участников поездки.

```bash
curl http://localhost:8080/api/v1/trips/<trip_id>/members \
  -H "Authorization: Bearer <jwt>"
```

## Group Vibe Merge

Заглушка, собирающая vibe_vector_id всех участников. В будущем:

1. Загрузка векторов из Qdrant
2. Вычисление средневзвешенного (дети получают доп. вес на child_friendly)
3. Upsert merged вектора в Qdrant
4. Обновление trips.merged_vibe_vector_id

## Связи

- trips.creator_id -> users.id (ON DELETE CASCADE)
- trip_members.trip_id -> trips.id (ON DELETE CASCADE)
- trip_members.user_id -> users.id (ON DELETE SET NULL)
- UNIQUE(trip_id, user_id) WHERE user_id IS NOT NULL
