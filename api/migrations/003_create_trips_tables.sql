-- Миграция 003: Создание таблиц trips и trip_members.
-- Таблица trips хранит объект поездки с датами, бюджетом, транспортом,
-- составом группы и invite-токеном для приглашения участников.
-- Таблица trip_members хранит участников поездки с их ролями,
-- тегами предпочтений и ссылками на vibe-векторы.

-- =============================================
-- Таблица поездок
-- =============================================
CREATE TABLE IF NOT EXISTS trips (
    -- Уникальный идентификатор поездки.
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

-- Создатель поездки (ссылка на users.id).
creator_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,

-- Дата начала поездки.
date_from DATE NOT NULL,

-- Дата окончания поездки.
date_to DATE NOT NULL,

-- Общий бюджет поездки в рублях.
budget_rub INTEGER NOT NULL DEFAULT 0,

-- Уровень бюджета: economy, comfort, premium.
budget_tier TEXT NOT NULL DEFAULT 'comfort',

-- Вид транспорта: car, public, walk, bike.
transport TEXT NOT NULL DEFAULT 'car',

-- Планируемое количество участников.
group_size INTEGER NOT NULL DEFAULT 1,

-- Состав группы в формате JSONB: {"adults": 2, "children": [{"age": 8}]}.
group_composition JSONB NOT NULL DEFAULT '{}',

-- Уникальный токен для приглашения участников.
invite_token UUID NOT NULL UNIQUE DEFAULT gen_random_uuid (),

-- Идентификатор средневзвешенного vibe-вектора группы в Qdrant.
merged_vibe_vector_id UUID,

-- Статус поездки: planning, active, completed, cancelled.
status TEXT NOT NULL DEFAULT 'planning',

-- Временные метки.
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

-- Ограничение: дата окончания не раньше даты начала.
CONSTRAINT trips_dates_check CHECK (date_to >= date_from),

-- Ограничение: размер группы положительный.
CONSTRAINT trips_group_size_check CHECK (group_size > 0),

-- Ограничение: бюджет неотрицательный.
CONSTRAINT trips_budget_check CHECK (budget_rub >= 0) );

-- Индекс для быстрого поиска поездок по создателю.
CREATE INDEX IF NOT EXISTS idx_trips_creator_id ON trips (creator_id);

-- Индекс для поиска поездки по invite-токену (уникальный, уже покрыт UNIQUE constraint).
-- Дополнительный индекс по статусу для фильтрации активных поездок.
CREATE INDEX IF NOT EXISTS idx_trips_status ON trips (status);

-- =============================================
-- Таблица участников поездки
-- =============================================
CREATE TABLE IF NOT EXISTS trip_members (
    -- Уникальный идентификатор записи участника.
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

-- Ссылка на поездку. При удалении поездки — каскадное удаление участников.
trip_id UUID NOT NULL REFERENCES trips (id) ON DELETE CASCADE,

-- Ссылка на пользователя (nullable: неавторизованные участники
-- присоединяются через invite-ссылку без аккаунта).
user_id UUID REFERENCES users (id) ON DELETE SET NULL,

-- Отображаемое имя участника в контексте поездки.
display_name TEXT NOT NULL,

-- Роль участника: creator (создатель), member (обычный участник).
role TEXT NOT NULL DEFAULT 'member',

-- Идентификатор vibe-вектора участника в Qdrant (nullable).
vibe_vector_id UUID,

-- Теги предпочтений участника (вино, горы, тишина и т.д.).
tags TEXT[] NOT NULL DEFAULT '{}',

-- Является ли участник ребёнком.
is_child BOOLEAN NOT NULL DEFAULT false,

-- Временная метка присоединения.
joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW() );

-- Индекс для быстрого получения всех участников поездки.
CREATE INDEX IF NOT EXISTS idx_trip_members_trip_id ON trip_members (trip_id);

-- Индекс для поиска поездок пользователя.
CREATE INDEX IF NOT EXISTS idx_trip_members_user_id ON trip_members (user_id);

-- Ограничение уникальности: авторизованный пользователь не может
-- присоединиться к одной поездке дважды.
CREATE UNIQUE INDEX IF NOT EXISTS idx_trip_members_unique_user ON trip_members (trip_id, user_id)
WHERE
    user_id IS NOT NULL;

-- =============================================
-- Триггер автообновления updated_at для trips
-- =============================================
CREATE OR REPLACE FUNCTION update_trips_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_trips_updated_at
    BEFORE UPDATE ON trips
    FOR EACH ROW
    EXECUTE FUNCTION update_trips_updated_at();