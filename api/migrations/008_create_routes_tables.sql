-- Миграция 008: создание таблиц routes и route_points.
-- Маршруты привязаны к поездке (trip_id) и пользователю (user_id).
-- Точки маршрута упорядочены по position, распределены по дням и тайм-слотам.
--
-- Идемпотентная миграция: IF NOT EXISTS, ON CONFLICT DO NOTHING.

-- Таблица маршрутов.
CREATE TABLE IF NOT EXISTS routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(200) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    total_distance_km FLOAT NOT NULL DEFAULT 0,
    estimated_duration_min INTEGER NOT NULL DEFAULT 0,
    estimated_cost_rub INTEGER NOT NULL DEFAULT 0,
    transport VARCHAR(20) NOT NULL DEFAULT 'car',
    points_count INTEGER NOT NULL DEFAULT 0,
    summary TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индекс для быстрого поиска маршрутов по поездке.
CREATE INDEX IF NOT EXISTS idx_routes_trip_id ON routes(trip_id);

-- Индекс для быстрого поиска маршрутов по пользователю.
CREATE INDEX IF NOT EXISTS idx_routes_user_id ON routes(user_id);

-- Таблица точек маршрута.
CREATE TABLE IF NOT EXISTS route_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    location_id UUID NOT NULL REFERENCES locations(id),
    position INTEGER NOT NULL DEFAULT 1,
    day_number INTEGER NOT NULL DEFAULT 1,
    time_slot VARCHAR(20) NOT NULL DEFAULT 'morning',
    target_audience VARCHAR(20) NOT NULL DEFAULT 'all',
    stay_duration_min INTEGER NOT NULL DEFAULT 60,
    distance_from_prev_km FLOAT NOT NULL DEFAULT 0,
    duration_from_prev_min INTEGER NOT NULL DEFAULT 0
);

-- Индекс для получения точек маршрута в правильном порядке.
CREATE INDEX IF NOT EXISTS idx_route_points_route_id_position ON route_points(route_id, position);

-- CHECK-ограничения для валидации данных.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_routes_status'
    ) THEN
        ALTER TABLE routes ADD CONSTRAINT chk_routes_status
            CHECK (status IN ('draft', 'active', 'completed'));
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_route_points_time_slot'
    ) THEN
        ALTER TABLE route_points ADD CONSTRAINT chk_route_points_time_slot
            CHECK (time_slot IN ('morning', 'afternoon', 'evening'));
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_route_points_target_audience'
    ) THEN
        ALTER TABLE route_points ADD CONSTRAINT chk_route_points_target_audience
            CHECK (target_audience IN ('all', 'adults', 'children'));
    END IF;
END $$;
