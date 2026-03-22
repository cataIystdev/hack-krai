-- Файл: 014_create_events.sql
-- Создание таблицы для аналитических событий (телеметрии) B2G дашборда.
-- Таблица использует движок MergeTree для эффективного хранения временных рядов.

CREATE TABLE IF NOT EXISTS events (
    event_id UUID,
    timestamp DateTime,
    event_type String,
    user_id UUID,
    location_id UUID,
    lat Float64,
    lon Float64,
    metadata String
) ENGINE = MergeTree()
ORDER BY (event_type, timestamp)
PARTITION BY toYYYYMM(timestamp);
