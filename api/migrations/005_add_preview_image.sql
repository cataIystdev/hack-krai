-- Миграция 005: добавление поля preview_image_url в таблицу locations.
-- Поле хранит URL hero-изображения для карточек и маркеров на карте.

ALTER TABLE locations
ADD COLUMN IF NOT EXISTS preview_image_url TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN locations.preview_image_url IS 'URL hero-изображения локации для карточек, карты и рекомендаций';