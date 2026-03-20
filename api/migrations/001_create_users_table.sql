-- Миграция 001: Создание таблицы пользователей.
-- Определяет тип перечисления user_role и создаёт таблицу users
-- со всеми полями, необходимыми для аутентификации, авторизации и профиля.

-- Подключение расширения для генерации UUID v4.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Перечисление допустимых ролей пользователя.
-- tourist  — обычный пользователь (турист), роль по умолчанию.
-- host     — хозяин локации (фермер, владелец агроусадьбы).
-- b2g_admin — администратор B2G-платформы (муниципальные органы).
CREATE TYPE user_role AS ENUM ('tourist', 'host', 'b2g_admin');

-- Таблица пользователей.
-- Хранит учётные данные, профиль и метаданные каждого пользователя системы.
CREATE TABLE IF NOT EXISTS users (
    -- Уникальный идентификатор пользователя (UUID v4).
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

-- Адрес электронной почты. Используется для аутентификации.
-- Уникальный и обязательный.
email VARCHAR(255) NOT NULL UNIQUE,

-- Хэш пароля (bcrypt). Никогда не возвращается клиенту.
password_hash VARCHAR(255) NOT NULL,

-- Роль пользователя в системе. По умолчанию — tourist.
role user_role NOT NULL DEFAULT 'tourist',

-- Отображаемое имя пользователя в интерфейсе.
display_name VARCHAR(255) NOT NULL DEFAULT '',

-- Очки кармы пользователя. Начальное значение — 0.
-- Карма увеличивается при полезных действиях (отзывы, рекомендации).
karma INTEGER NOT NULL DEFAULT 0,

-- Идентификатор вектора vibe-профиля в Qdrant.
-- NULL до прохождения onboarding-опроса.
vibe_vector_id UUID,

-- Временная метка создания записи (UTC).
created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

-- Временная метка последнего обновления записи (UTC).
updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() );

-- Индекс для ускорения поиска пользователя по email при авторизации.
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Комментарии к таблице и колонкам для документации схемы.
COMMENT ON TABLE users IS 'Пользователи платформы Deep Krai';

COMMENT ON COLUMN users.id IS 'Уникальный идентификатор (UUID v4)';

COMMENT ON COLUMN users.email IS 'Email для аутентификации';

COMMENT ON COLUMN users.password_hash IS 'Хэш пароля (bcrypt, cost=12)';

COMMENT ON COLUMN users.role IS 'Роль: tourist, host, b2g_admin';

COMMENT ON COLUMN users.display_name IS 'Отображаемое имя';

COMMENT ON COLUMN users.karma IS 'Очки кармы';

COMMENT ON COLUMN users.vibe_vector_id IS 'ID вектора vibe-профиля в Qdrant';

COMMENT ON COLUMN users.created_at IS 'Дата создания (UTC)';

COMMENT ON COLUMN users.updated_at IS 'Дата обновления (UTC)';