-- Миграция 010: создание таблицы ai_tasks для отслеживания задач AI-обработки.
-- Используется модулем онбординга хостов (Phase 10) для track прогресса
-- pipeline: STT -> LLM -> Embedding -> Location Draft.
-- Типы задач: onboarding, splatting, story_generation, voice_profile.
-- Статусы: queued, processing, completed, failed.

CREATE TABLE IF NOT EXISTS ai_tasks (
    -- Уникальный идентификатор задачи (UUID v4).
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Идентификатор пользователя, создавшего задачу.
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Тип задачи: onboarding, splatting, story_generation, voice_profile.
    type VARCHAR(50) NOT NULL,

    -- Статус выполнения: queued, processing, completed, failed.
    status VARCHAR(20) NOT NULL DEFAULT 'queued',

    -- Прогресс выполнения (0-100) для отображения прогресс-бара.
    progress INTEGER NOT NULL DEFAULT 0,

    -- Входные данные задачи (JSONB): audio_url, media_urls, coordinates и др.
    input_data JSONB,

    -- Результат выполнения (JSONB): location_id, extracted_data и др.
    output_data JSONB,

    -- Текст ошибки (заполняется при status = 'failed').
    error TEXT,

    -- Временная метка создания задачи.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Временная метка последнего обновления.
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Временная метка завершения (заполняется при status = 'completed' или 'failed').
    completed_at TIMESTAMPTZ
);

-- Индекс для быстрого поиска задач по пользователю.
CREATE INDEX IF NOT EXISTS idx_ai_tasks_user_id ON ai_tasks(user_id);

-- Индекс для фильтрации задач по статусу (очередь, в процессе).
CREATE INDEX IF NOT EXISTS idx_ai_tasks_status ON ai_tasks(status);

-- Индекс для фильтрации задач по типу.
CREATE INDEX IF NOT EXISTS idx_ai_tasks_type ON ai_tasks(type);
