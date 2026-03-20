-- Миграция 002: Создание таблицы локаций с пространственными данными PostGIS.
-- Определяет типы перечислений для уровней доступа и плотности,
-- создаёт таблицу locations с полем GEOMETRY(Point, 4326) и GiST-индексом
-- для эффективных пространственных запросов (ST_DWithin, ST_Within).

-- Подключение расширения PostGIS для пространственных типов данных.
CREATE EXTENSION IF NOT EXISTS postgis;

-- Перечисление уровней доступа к локации (Hidden Gems).
-- open      -- видна всем пользователям.
-- semi_open -- видна зарегистрированным с профилем.
-- hidden    -- только для пользователей с karma >= threshold.
CREATE TYPE access_level_type AS ENUM ('open', 'semi_open', 'hidden');

-- Перечисление уровней туристической плотности.
-- red    -- высокая загруженность (много туристов круглый год).
-- yellow -- сезонная загруженность (в сезон много, вне -- мало).
-- green  -- скрытое место (редко посещается, Hidden Gem).
CREATE TYPE density_level_type AS ENUM ('red', 'yellow', 'green');

-- Таблица локаций.
-- Хранит данные о туристических объектах Краснодарского края:
-- координаты (PostGIS), описания, категорию, теги, параметры бронирования,
-- уровень доступа (Hidden Gems) и плотность.
CREATE TABLE IF NOT EXISTS locations (
    -- Уникальный идентификатор локации (UUID v4).
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

-- Идентификатор владельца (хост). Внешний ключ на таблицу users.
owner_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,

-- URL-дружественный идентификатор (slug).
-- Уникальный, используется в маршрутах /l/{slug}.
slug VARCHAR(255) NOT NULL UNIQUE,

-- Название локации (отображается на карте и в каталоге).
name VARCHAR(255) NOT NULL,

-- Краткое описание (до 500 символов) для карточек и tooltips.
description_short TEXT NOT NULL DEFAULT '',

-- Полное описание (литературный текст, сгенерированный ИИ или введённый вручную).
description_full TEXT NOT NULL DEFAULT '',

-- Категория локации (свободный текст).
-- Примеры: winery, farm, trail, guesthouse, gastro, camping, extreme, cultural.
-- Не ограничена перечислением, допускает произвольные значения.
category VARCHAR(100) NOT NULL DEFAULT '',

-- Теги для фильтрации и поиска (массив строк).
-- Примеры: {"вино", "дегустация", "горы", "тишина"}.
tags TEXT[] NOT NULL DEFAULT '{}',

-- Стоимость проживания за ночь в рублях. 0 для бесплатных объектов.
price_per_night INTEGER NOT NULL DEFAULT 0,

-- Максимальная вместимость (количество гостей).
capacity INTEGER NOT NULL DEFAULT 1,

-- Уровень доступа к локации (Hidden Gems).
access_level access_level_type NOT NULL DEFAULT 'open',

-- Уровень туристической плотности (для цветовой маркировки на карте).
density_level density_level_type NOT NULL DEFAULT 'green',

-- Подходит ли для посещения с детьми.
child_friendly BOOLEAN NOT NULL DEFAULT false,

-- URL на .splat файл (3D Gaussian Splatting) в MinIO.
-- NULL до генерации 3D-сцены.
splat_url TEXT,

-- Идентификатор вектора vibe-профиля локации в Qdrant.
-- NULL до генерации эмбеддинга описания.
vibe_vector_id UUID,

-- Полный адрес локации (населённый пункт, район).
address TEXT NOT NULL DEFAULT '',

-- Опубликована ли локация (видна ли на карте и в поиске).
is_published BOOLEAN NOT NULL DEFAULT false,

-- Пространственные координаты (долгота, широта) в системе WGS 84 (SRID 4326).
-- Хранится как GEOMETRY(Point, 4326) для поддержки GiST-индекса.
geo GEOMETRY (Point, 4326),

-- Временная метка создания записи (UTC).
created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

-- Временная метка последнего обновления записи (UTC).
updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() );

-- GiST-индекс для пространственного поиска.
-- Обеспечивает эффективные запросы ST_DWithin (радиус) и ST_Within (bbox).
CREATE INDEX IF NOT EXISTS idx_locations_geo ON locations USING GIST (geo);

-- Индекс для поиска локаций конкретного владельца.
CREATE INDEX IF NOT EXISTS idx_locations_owner_id ON locations (owner_id);

-- Индекс для поиска по slug (быстрый lookup для URL-маршрутов).
CREATE INDEX IF NOT EXISTS idx_locations_slug ON locations (slug);

-- Индекс для фильтрации по категории.
CREATE INDEX IF NOT EXISTS idx_locations_category ON locations (category);

-- Индекс для фильтрации по статусу публикации.
CREATE INDEX IF NOT EXISTS idx_locations_published ON locations (is_published);

-- Комментарии к таблице и ключевым колонкам.
COMMENT ON
TABLE locations IS 'Туристические локации КудыТуды с пространственными координатами (PostGIS)';

COMMENT ON COLUMN locations.geo IS 'Координаты GEOMETRY(Point, 4326) для пространственных запросов (GiST)';

COMMENT ON COLUMN locations.category IS 'Свободная строка категории (winery, farm, trail, guesthouse, gastro и др.)';

COMMENT ON COLUMN locations.access_level IS 'Уровень доступа: open/semi_open/hidden (Hidden Gems)';

COMMENT ON COLUMN locations.density_level IS 'Плотность туристов: red/yellow/green (цветовая маркировка на карте)';

COMMENT ON COLUMN locations.tags IS 'Массив тегов для фильтрации (PostgreSQL TEXT[])';