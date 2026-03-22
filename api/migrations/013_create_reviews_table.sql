-- Миграция 013: таблица отзывов (reviews) для Phase 13 Социалка и Геймификация.

CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    location_id UUID REFERENCES locations(id) ON DELETE CASCADE,
    booking_id UUID REFERENCES bookings(id) ON DELETE SET NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    text TEXT NOT NULL DEFAULT '',
    is_from_host BOOLEAN NOT NULL DEFAULT FALSE,
    karma_delta INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reviews_author_id
    ON reviews(author_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_reviews_location_id
    ON reviews(location_id, created_at DESC);
