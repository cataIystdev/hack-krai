# Phase 8: Group Trip MVP

## Обзор

Укрепление модуля групповых поездок: membership privacy, concurrency-safe join,
auth-flex идентификация, полноценный MergeGroupVibes.

## Реализованные улучшения

### 1. Concurrency-safe Add Member

Переписан `AddMemberAtomic` с INSERT...SELECT на транзакционный подход:
1. `SELECT ... FOR UPDATE` на строке trips (блокировка на уровне строки)
2. `COUNT(*)` текущих участников
3. Проверка дубликатов: по `user_id` для авторизованных, по `LOWER(display_name)` для анонимных
4. `INSERT` нового участника
5. `COMMIT`

Гарантирует строгое соблюдение `group_size` даже при параллельных join-запросах.

### 2. Membership Privacy

`GetByID` и `GetMembers` теперь проверяют, что запрашивающий пользователь
является участником или создателем поездки. Неучастники получают 403 Forbidden.

### 3. Auth-Flex Join

При Join с JWT-токеном:
- `display_name` подтягивается из профиля users (если не передан в запросе)
- `vibe_vector_id` подтягивается из профиля users автоматически
- `user_id` корректно связывается с `trip_member`

При Join без JWT:
- Используются `display_name` и `tags` из запроса
- Проверка дубликатов по `LOWER(display_name)`

### 4. MergeGroupVibes (реализация)

Полноценная имплементация (замена stub):
1. Загрузка vibe-векторов всех участников из Qdrant (коллекция `user_vibes`)
2. Взвешенное среднее: детские профили получают вес 1.2x
3. L2-нормализация результата
4. Upsert merged вектора в Qdrant (коллекция `group_vibes`)
5. Обновление `trips.merged_vibe_vector_id`
6. Детерминистический UUID через `uuid.NewSHA1` для идемпотентности

Пересчёт запускается асинхронно после каждого Join.

### 5. Duplicate Prevention

| Сценарий | Механизм |
|----------|----------|
| Авторизованный user | UNIQUE INDEX `(trip_id, user_id) WHERE user_id IS NOT NULL` |
| Анонимный user | UNIQUE INDEX `(trip_id, LOWER(display_name)) WHERE user_id IS NULL` (миграция 009) |

## Файлы

| Файл | Действие | Описание |
|------|----------|----------|
| `migrations/009_group_trip_hardening.sql` | NEW | Индекс уникальности анонимных участников |
| `database/trip_repository.go` | MODIFY | Транзакционный AddMemberAtomic + IsMember + FindMemberByUserID |
| `services/trip.go` | REWRITE | Membership privacy, auth-flex Join, MergeGroupVibes, vibeRepo зависимость |
| `handlers/trip.go` | MODIFY | GetByID/ListMembers передают userID для privacy |
| `cmd/api/main.go` | MODIFY | Передача vibeRepo в NewTripService |
| `handlers/openapi.go` | MODIFY | Добавлены 403 на GET trip/members, обновлён POST join |
