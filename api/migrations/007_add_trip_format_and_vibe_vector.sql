-- Миграция 007: Добавление полей format и vibe_vector_id в таблицу trips.
-- format — формат поездки (day_trip / weekend / multi_day).
-- vibe_vector_id — ссылка на vibe-вектор создателя в Qdrant.

-- =============================================
-- Добавление столбца format
-- =============================================
ALTER TABLE trips
ADD COLUMN IF NOT EXISTS format TEXT NOT NULL DEFAULT 'multi_day';

-- Ограничение: допустимые значения формата.
ALTER TABLE trips
ADD CONSTRAINT trips_format_check CHECK (
    format IN (
        'day_trip',
        'weekend',
        'multi_day'
    )
);

-- =============================================
-- Добавление столбца vibe_vector_id
-- =============================================
-- Ссылка на персональный vibe-вектор создателя в Qdrant.
-- Nullable: поездку можно создать без прохождения профилирования.
ALTER TABLE trips ADD COLUMN IF NOT EXISTS vibe_vector_id UUID;