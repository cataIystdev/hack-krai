# Phase 11: Booking MVP

## Goal

Закрыть базовую коммерческую механику Phase 11 из roadmap:

1. booking slots schema and endpoints
2. create booking
3. confirm/reject booking
4. host bookings / my bookings

## Implemented

### Schema

- `011_create_bookings_tables.sql`
- `booking_slots`:
  - `location_id`
  - `slot_date`
  - `total_capacity`
  - `available_capacity`
  - `is_closed`
  - unique `(location_id, slot_date)`
- `bookings`:
  - `location_id`
  - `user_id`
  - `date_from`, `date_to`
  - `guests_count`
  - `total_price`
  - `status`
  - `contact_name`, `contact_phone`
  - `comment`, `host_comment`

### API

- `GET /api/v1/locations/:id/slots?month=YYYY-MM`
- `POST /api/v1/bookings`
- `GET /api/v1/bookings/my`
- `POST /api/v1/bookings/:id/confirm`
- `POST /api/v1/bookings/:id/cancel`
- `GET /api/v1/host/bookings`

### Business rules

- Слоты создаются лениво из `location.capacity`, если для месяца их ещё нет.
- Создание брони атомарно резервирует capacity по всем дням диапазона `[date_from, date_to)`.
- `reject` и `cancel` возвращают capacity обратно в `booking_slots`.
- `confirm` доступен только владельцу локации или `b2g_admin`.
- `cancel` доступен только создателю брони.

## Not included in MVP

- телефония / robocall
- WebSocket notifications
- auto-confirm policies
- отдельный inventory UI
- интеграция с ClickHouse events

## Files

- `api/internal/models/booking.go`
- `api/internal/database/booking_repository.go`
- `api/internal/services/booking.go`
- `api/internal/handlers/booking.go`
- `api/internal/handlers/router.go`
- `api/internal/handlers/openapi.go`
- `api/cmd/api/main.go`
