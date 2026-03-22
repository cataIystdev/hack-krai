-- Миграция 011: booking slots и bookings для Phase 11 Booking MVP.

CREATE TABLE IF NOT EXISTS booking_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    slot_date DATE NOT NULL,
    total_capacity INTEGER NOT NULL CHECK (total_capacity > 0),
    available_capacity INTEGER NOT NULL CHECK (available_capacity >= 0 AND available_capacity <= total_capacity),
    is_closed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(location_id, slot_date)
);

CREATE INDEX IF NOT EXISTS idx_booking_slots_location_date
    ON booking_slots(location_id, slot_date);

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date_from DATE NOT NULL,
    date_to DATE NOT NULL,
    guests_count INTEGER NOT NULL CHECK (guests_count > 0),
    total_price INTEGER NOT NULL CHECK (total_price >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    contact_name VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(50) NOT NULL,
    comment TEXT NOT NULL DEFAULT '',
    host_comment TEXT NOT NULL DEFAULT '',
    cancelled_by VARCHAR(20),
    confirmed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_bookings_dates CHECK (date_to > date_from),
    CONSTRAINT chk_bookings_status CHECK (status IN ('pending', 'confirmed', 'rejected', 'cancelled')),
    CONSTRAINT chk_bookings_cancelled_by CHECK (cancelled_by IS NULL OR cancelled_by IN ('tourist', 'host'))
);

CREATE INDEX IF NOT EXISTS idx_bookings_user_id
    ON bookings(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_bookings_location_id
    ON bookings(location_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_bookings_status
    ON bookings(status);
