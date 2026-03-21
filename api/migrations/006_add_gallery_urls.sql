-- Миграция 006: Добавление gallery_urls для карточки локации.
-- Массив URL галерейных изображений используется на экране детали локации
-- для показа нескольких фотографий места (галерея/карусель).
-- preview_image_url остаётся основным hero-изображением,
-- gallery_urls содержит дополнительные фотографии.

ALTER TABLE locations ADD COLUMN IF NOT EXISTS gallery_urls TEXT[] NOT NULL DEFAULT '{}';