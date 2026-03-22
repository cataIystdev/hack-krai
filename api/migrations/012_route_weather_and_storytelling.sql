-- Миграция 012: weather/storytelling поля для Phase 12.

ALTER TABLE route_points
    ADD COLUMN IF NOT EXISTS audio_story_url TEXT,
    ADD COLUMN IF NOT EXISTS story_text TEXT,
    ADD COLUMN IF NOT EXISTS weather_condition VARCHAR(50),
    ADD COLUMN IF NOT EXISTS weather_temp_c INTEGER;
