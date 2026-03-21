# Phase 7: Route MVP

## Обзор

Trip-aware построение маршрутов с vibe-интеграцией, фильтрацией по параметрам поездки, распределением по дням и сохранением в БД.

## Реализованные эндпоинты

### POST /api/v1/route/build (Phase 7A — ручной режим)
- Маршрут по заданным ID локаций
- Haversine-расчёт расстояний
- Время по типу транспорта (car/public/walk/bike)

### POST /api/v1/trips/{id}/build-route (Phase 7B+7C — trip-aware)
- Автоподбор локаций на основе vibe-скоринга через Qdrant
- Фильтрация по: budget_tier, child_friendly, format
- Merged vibe vector: взвешенное среднее vibe-векторов всех участников
- Распределение по дням и time_slot (morning/afternoon/evening)
- Target audience tagging (all/adults/children)
- Сохранение в БД (routes + route_points)
- Текстовое summary

## БД: таблицы

### routes
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID PK | ID маршрута |
| trip_id | UUID FK→trips | Поездка |
| user_id | UUID FK→users | Создатель |
| name | VARCHAR(200) | Название |
| status | VARCHAR(20) | draft/active/completed |
| total_distance_km | FLOAT | Дистанция |
| estimated_duration_min | INT | Время (мин) |
| estimated_cost_rub | INT | Стоимость (₽) |
| transport | VARCHAR(20) | Транспорт |
| points_count | INT | Кол-во точек |
| summary | TEXT | Описание |

### route_points
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID PK | ID точки |
| route_id | UUID FK→routes | Маршрут |
| location_id | UUID FK→locations | Локация |
| position | INT | Порядок |
| day_number | INT | День |
| time_slot | VARCHAR(20) | morning/afternoon/evening |
| target_audience | VARCHAR(20) | all/adults/children |
| stay_duration_min | INT | Время пребывания |

## Алгоритм trip-aware маршрута

1. Загрузка Trip + members
2. Определение наличия детей (is_child, group_composition)
3. Сбор vibe-векторов участников из Qdrant
4. Вычисление merged vector (среднее + L2-нормализация)
5. SearchNearest по location_vibes → vibe_score
6. Фильтрация: budget_tier → max price_per_night, child_friendly
7. Ранжирование: vibe_score + child бонус +0.1
8. Ограничение: max_points по формату (day_trip=5, weekend=8, multi_day=15)
9. Haversine routing + stay durations
10. Day planning: 3 точки/день, morning→afternoon→evening
11. Target audience: winery/gastro/extreme → adults, farm/beach/nature → children
12. Summary generation
13. Persist в routes + route_points
