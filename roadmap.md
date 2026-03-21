# Roadmap бэкенда Deep Krai

## Назначение

Этот roadmap задает порядок разработки backend-части с учетом двух жестких принципов:

1. Сначала собираем demo-first flow для скринкаста и скриншотов.
2. Demo flow должен визуально и поведенчески совпадать с тем, как это будет выглядеть в релизной версии, даже если часть данных временно мокается.

Backend планируется так, чтобы:

- не тратить лишнее время на одноразовый хардкод, если можно быстро сделать честный API;
- временные моки были тонким слоем поверх будущего настоящего flow;
- фронтенд мог начинать сборку экранов как можно раньше на стабильных контрактах;
- после demo-first этапа команда сразу переходила в MVP core без переделки всей архитектуры.

## GDD как источник истины

Главный источник истины для roadmap:

- [deep_krai_gdd.md](/c:/Users/lowcoware/Projects/hack-krai/deep_krai_gdd.md)

Этот roadmap не заменяет GDD, а декомпозирует его в инженерный порядок реализации.

Правило согласования:

1. Все крупные backend-фазы должны явно указывать, какую фичу или секцию GDD они закрывают.
2. Demo-first фазы покрывают только подмножество GDD, но через те же сущности, контракты и сценарии, что и финальный продукт.
3. Phase 15 считается завершенной только когда roadmap покрывает весь продуктовый контур GDD, а не только demo/MVP.

---

## Модель покрытия GDD

### Покрытие на этапе demo-first

Первый съемочный контур покрывает подмножество GDD:

1. GDD Feature 1: Умное планирование поездки
   - точка входа с 2 CTA
   - Voice-to-Vibe
   - Vibe passport
   - базовые рекомендации

2. GDD Feature 2: Интерактивная 3D-карта Краснодарского края
   - только demo map / recommended points layer

3. GDD Feature 3: Immersive 3D-экскурсии
   - только location detail / preview shell

### Покрытие на этапе MVP-core

Следующий обязательный слой:

1. GDD Feature 1 полностью в MVP-ядре:
   - trip details
   - group invite/join
   - recommendation flow

2. GDD Feature 2 в MVP-версии:
   - map endpoints
   - location selection
   - route preview

3. Частичный GDD Feature 4:
   - базовый route build без полной live-перестройки

### Полное покрытие GDD

Полное покрытие достигается только после реализации:

1. Feature 1. Smart Planning
2. Feature 2. 3D Map
3. Feature 3. 3D Tours
4. Feature 4. Live Routing & Storytelling
5. Feature 5. Zero-UI Onboarding
6. Feature 6. Booking & Inventory
7. Feature 7. Hidden Gems
8. Feature 8. Analytics / B2G
9. Feature 9. Offline-related backend endpoints / sync support

---

## Связь roadmap с фичами GDD

| Фича GDD                                | Фазы roadmap бэкенда      |
| --------------------------------------- | ------------------------- |
| Feature 1. Smart trip planning          | 0, 1, 2, 3, 5, 6, 7, 8, 9 |
| Feature 2. 3D map                       | 4, 7, 12, 15              |
| Feature 3. 3D tours / splat             | 5, 10, 15                 |
| Feature 4. Live routing & storytelling  | 7, 12, 15                 |
| Feature 5. Host onboarding              | 10, 15                    |
| Feature 6. Booking                      | 11, 15                    |
| Feature 7. Hidden Gems / karma          | 13, 15                    |
| Feature 8. Analytics / B2G              | 14, 15                    |
| Feature 9. Offline mode backend support | 12, 14, 15                |

---

## Текущее состояние проекта

### Уже реализовано

1. App/server foundation:
   - Fiber app bootstrap;
   - config;
   - logger;
   - middleware;
   - graceful shutdown.

2. Infrastructure connection layer:
   - PostgreSQL;
   - Redis;
   - Qdrant;
   - Neo4j;
   - ClickHouse;
   - MinIO.

3. Auth/profile:
   - register;
   - login;
   - refresh;
   - JWT middleware;
   - RBAC middleware;
   - `GET /api/v1/profile/me`;
   - `PUT /api/v1/profile/me`.

4. Locations:
   - migrations (001-006);
   - public list/detail;
   - protected create/update/delete;
   - PostGIS search foundation;
   - `GET /api/v1/locations/{id}/splat` -- ленивая загрузка 3D;
   - `gallery_urls` для массива фотографий.

5. Trips:
   - create;
   - detail;
   - invite refresh;
   - join;
   - members list.

6. Vibe:
   - voice profiling;
   - scenes;
   - swipe;
   - finalize;
   - AI mock mode;
   - Qdrant integration.

7. Map API:
   - `GET /api/v1/map/locations` -- 3 режима (bbox, demo, hybrid);
   - 4 curated demo-профиля (calm_wine_mountains, active_adventure, family_kids, gastro_cultural);
   - обогащение рекомендациями из Qdrant.

8. Route:
   - `POST /api/v1/route/build` -- построение маршрутов.

9. Media/docs:
   - media upload;
   - health endpoint;
   - OpenAPI;
   - Scalar UI.

10. Schema/migrations:
    - 001: users;
    - 002: locations;
    - 003: trips/trip_members;
    - 004: swipe_scenes;
    - 005: preview_image;
    - 006: gallery_urls.

11. Seed-данные:
    - 20 seed-локаций Краснодарского края;
    - seed-хост и demo-турист;
    - загружено на dev и test-catalyst серверы.

### Ранее частично реализовано, теперь закрыто

1. Recommendations -- закрыто в фазах 3 и 4 (enriched payload, curated demo profiles).
2. `.splat` support -- закрыто в фазе 5 (`GET /api/v1/locations/{id}/splat`).
3. Map API -- закрыто в фазе 4 (3 режима, hybrid enrichment).

### Пока не реализовано

1. `POST /api/v1/trips/{id}/build-route`
2. weather/live routing
3. storytelling
4. host onboarding
5. booking
6. reviews/karma/hidden gems logic
7. analytics/B2G endpoints
8. sync/offline support endpoints

### Что это значит для планирования

Backend roadmap покрывает полный demo-first контур:

- Фазы 0-6 завершены.
- Фаза 5.5 (техническое усиление) -- ЗАВЕРШЕНА.
- Demo-first flow полностью функционален: voice -> vibe -> swipe -> finalize -> map -> location detail.
- Core MVP закрыт частично глубже demo-first: trip details, invite/join, route preview, auth hardening.
- Основные оставшиеся gap-ы: merged group vibe, trip-aware route build, host onboarding, booking, storytelling, weather/live routing, analytics/B2G.
- После code review выявлены отдельные P0 архитектурные долги в trip/auth-контуре:
  - privacy/access control для trip detail и members;
  - неконсистентный auth contract для public/auth join flow;
  - отсутствие по-настоящему безопасной capacity guarantee при конкурентных join.

---

## Приоритеты продукта

### Артефакты demo-first

Первым должны быть готовы:

1. Скринкаст 30 секунд:
   - стартовый экран с 2 CTA;
   - voice input + состояние "ИИ думает";
   - vibe passport;
   - 3D-карта / карта с рекомендованными точками;
   - карточка локации / 3D-preview + CTA "Собрать маршрут".

2. Скриншоты для презентации:
   - vibe passport;
   - 3D-карта с рекомендациями;
   - карточка локации или 3D-экскурсия.

### Приоритеты core MVP после demo-first

Сразу после demo-first:

1. Auth
2. Vibe profiling
3. Trip details
4. Recommendations
5. Location details
6. Route-building MVP
7. Group trip/invite

### Более низкий приоритет после core

Делать позже, если не хватает времени/сил:

1. Storytelling
2. Live weather
3. Booking automation
4. Host onboarding
5. Hidden Gems / karma
6. Analytics / B2G
7. Full offline sync

### Проверка приоритетов относительно GDD

Это согласовано с GDD так:

1. Сначала реализуется самый сильный туристический сценарий из Feature 1.
2. Затем доводится spatial presentation layer из Feature 2 и Feature 3.
3. Потом достраиваются route/group flows как продолжение Feature 1 и Feature 4.
4. Только после consumer core двигаются host, booking, hidden gems, analytics, B2G.

---

## Стратегия бэкенда

### Правило 1: предпочитать настоящие API одноразовому demo-коду

Если фичу можно сделать честно за сопоставимое время, делаем честно.

### Правило 2: если нужны моки, мокать на границах

Временные моки допустимы:

- в AI-ответах;
- в seed-данных;
- в маршрутах;
- в погоде;
- в .splat и media preview;
- в предварительно подготовленных рекомендациях.

Нежелательно мокать:

- форму и контракт API;
- статусную модель сущностей;
- shape JSON-ответов;
- идентификаторы и связи между users, trips, locations.

### Правило 3: demo-эндпоинты должны доживать до MVP

Если для demo делается endpoint, он не должен выбрасываться после съемки. Допустимо только:

- заменить mock-источник на реальный;
- расширить payload;
- усилить валидацию;
- добавить асинхронность и очереди.

---

## Модель поставки

Работа делится на фазы и handoff между backend и frontend.

### Паттерн handoff

1. Backend фиксирует контракт.
2. Backend отдает стабильный mock-or-real endpoint.
3. Frontend собирает финальный UI под этот контракт.
4. Backend идет на следующий блок.
5. Позже backend заменяет mock internals на real internals без ломки UI.

---

## Фаза 0. Заморозка контрактов для демо

> **Статус: ЗАВЕРШЕНА**

### Цель

Заморозить минимальный набор API-контрактов для первого скринкаста и скриншотов.

### GDD alignment

Закрывает подготовительный слой для:

- GDD Feature 1
- GDD Feature 2
- GDD Feature 3

### Реализация

Все контракты зафиксированы и реализованы в коде:

1. OpenAPI спецификация генерируется из `handlers/openapi.go` (1300+ строк).
2. Scalar UI доступен по `/api/v1/docs`.
3. Все demo-specific DTO зафиксированы:
   - `MapPoint` (id, name, lat/lon, category, density_level, preview_image_url, description_short, is_recommended, recommendation_score);
   - `MapLocationsResponse` (points, total, profile);
   - `Location` (20+ полей включая gallery_urls, splat_url);
   - `SplatResponse` (location_id, location_name, has_splat, splat_url);
   - `LocationRecommendation` (location_id, name, category, score, preview_image_url, splat_url, description, tags).
4. Контракты согласованы с фронтендом -- JSON shape стабилен.

### Файлы реализации

- `handlers/openapi.go` -- OpenAPI спецификация;
- `handlers/scalar.go` -- Scalar UI;
- `models/` -- все DTO.

### Критерии готовности -- выполнены

1. Frontend работает на стабильных контрактах.
2. Ни один payload не потребовал радикальной переделки.

---

## Фаза 1. Основа demo-данных

> **Статус: ЗАВЕРШЕНА**

### Цель

Подготовить данные, на которых можно быстро собрать красивое и правдоподобное демо.

### GDD alignment

Подготавливает данные для:

- Feature 1 recommendations
- Feature 2 map points
- Feature 3 location previews

### Реализация

Все seed-данные подготовлены и загружены на серверы (dev + test-catalyst):

1. **Seed users:**
   - `seed-host@deepkrai.ru` (host, owner_id: `a0000000-...0001`);
   - `demo-tourist@deepkrai.ru` (tourist).

2. **Seed locations -- 20 точек Краснодарского края:**
   - Винодельня Абрау-Дюрсо (winery, red density);
   - Винодельня Лефкадия (winery, yellow);
   - Козья ферма дяди Вани (farm, green, Hidden Gem);
   - Каньон реки Белой (extreme, yellow);
   - Водопады Руфабго (nature, red);
   - Хребет Грачёв и дольмены (cultural, green);
   - Озеро Кардывач (nature, green, hidden access);
   - Сыроварня Марии Коваленко (gastro, green);
   - Парк Галицкого (cultural, red);
   - Тихая заводь реки Пшеха (nature, green, hidden);
   - Подводное погружение в Чёрном море (extreme, yellow);
   - Старый парк Кабардинки (cultural, yellow);
   - Конная база Псебай (extreme, green);
   - Ущелье Гуамка (nature, yellow);
   - Глэмпинг Роза Хутор (resort, green);
   - Медовые водопады (nature, red);
   - Кипарисовое озеро Сукко (nature, red);
   - Чайные плантации Дагомыса (gastro, yellow);
   - Грязевой вулкан Шуго (nature, yellow);
   - Лысая гора Горячий Ключ (trail, yellow).
   - Все точки с полными описаниями, тегами, координатами, категориями.

3. **Seed swipe scenes:** 6 сцен (миграция 004).

4. **Seed recommendation bundles:** 4 curated demo-профиля (по 7 рекомендаций каждый), встроенные в MapService.

5. **Загрузка seed:** SQL-скрипт `/tmp/seed_locations.sql` выполнен на обоих серверах.

### Файлы реализации

- `scripts/seed.go` -- Go-скрипт seed;
- `/tmp/seed_locations.sql` -- SQL seed 20 локаций;
- `migrations/004_create_swipe_scenes.sql` -- seed сцен;
- `services/map.go` -- curated demo-профили.

### Критерии готовности -- выполнены

1. Demo-user стабильно получает красивый результат (22 точки, 7 рекомендованных).
2. Все seed-данные согласованы с DTO и demo-профилями.

---

## Фаза 2. Demo-срез Voice-To-Vibe

> **Статус: ЗАВЕРШЕНА**

### Цель

Дать полностью рабочий backend flow для сцены:
voice input -> AI thinking -> vibe passport.

### GDD alignment

Прямое покрытие:

- GDD Feature 1
- раздел Voice-to-Vibe
- vibe axes / summary / tags / vector pipeline

### Реализация

Полный pipeline голосового профилирования реализован:

1. **`POST /api/v1/profile/voice`:**
   - принимает audio (multipart/form-data);
   - STT (Whisper mock/real) -> транскрипция;
   - LLM extraction -> axes (adventure, culture, nature, social, comfort), summary, tags;
   - Embeddings -> 384-мерный вектор;
   - Qdrant upsert + user vibe_vector_id persistence;
   - возвращает: transcription, axes, extracted_tags, vibe_summary, vector_id.

2. **Demo mode:** детерминированный ответ при отсутствии API-ключей или mock-режиме. Frontend не различает mock/real.

3. **Persist profile result:** вектор сохраняется в Qdrant, ссылка привязывается к user profile.

### Файлы реализации

- `services/vibe.go` -- VoiceProfile, AI pipeline;
- `handlers/vibe.go` -- HTTP handler;
- `ai/llm.go` -- LLM abstraction;
- `ai/embeddings.go` -- embeddings layer;
- `database/qdrant.go` -- Qdrant client.

### Критерии готовности -- выполнены

1. Один запрос дает стабильный vibe passport.
2. Payload полностью пригоден для скриншота.
3. Frontend не знает, mock или real.

---

## Фаза 3. Swipe + finalize recommendations

> **Статус: ЗАВЕРШЕНА**

### Цель

Дать второй кусок core-vibe flow:
scenes -> swipe -> finalize -> recommendations.

### GDD alignment

Прямое покрытие:

- GDD Feature 1
- раздел Emotional 3D Swipe
- переход к recommendation layer

### Реализация

Полный swipe + finalize pipeline реализован:

1. **`GET /api/v1/profile/scenes`:** возвращает сцены из PostgreSQL с display_order.
2. **`POST /api/v1/profile/swipe`:** принимает scene_id + direction (like/dislike/skip), корректирует вектор и ищет рекомендации в Qdrant (SearchNearest).
3. **`POST /api/v1/profile/finalize`:** финализирует vibe-профиль, возвращает enriched LocationRecommendation:
   - location_id, name, category, score;
   - preview_image_url, splat_url;
   - description, tags.
4. **Swipe доступен наравне с voice** -- пользователь может выбрать один из двух путей профилирования.
5. Qdrant SearchNearest используется для реального vector similarity search.

### Файлы реализации

- `services/vibe.go` -- Swipe, Finalize, SearchNearest;
- `handlers/vibe.go` -- HTTP handlers (Swipe, Finalize, GetScenes);
- `models/vibe.go` -- LocationRecommendation, SwipeRequest;
- `database/vibe_repository.go` -- Qdrant integration;
- `migrations/004_create_swipe_scenes.sql` -- seed сцен.

### Критерии готовности -- выполнены

1. Recommendation set связан с реальными location IDs из seed.
2. Один vibe-профиль дает предсказуемый curated top-N.

---

## Фаза 4. Map API для демо

> **Статус: ЗАВЕРШЕНА** (commit `36259a7`)

### Цель

Отдать фронтенду API для wow-сцены карты и screenshot hero shot.

### GDD alignment

Прямое покрытие:

- GDD Feature 2
- `GET /api/v1/map/locations`
- карта рекомендованных / доступных локаций

### Реализация

`GET /api/v1/map/locations` реализован с тремя режимами:

1. **Bbox** (по умолчанию) -- пространственный поиск через PostGIS ST_Within.
   - Параметры: `min_lat`, `max_lat`, `min_lon`, `max_lon`.

2. **Demo** (`?demo=true&profile=...`) -- curated набор с предустановленными рекомендациями.
   - 4 профиля (по 7 рекомендаций каждый):
     - `calm_wine_mountains` -- горы, вино, тишина;
     - `active_adventure` -- каньоны, рафтинг, горы;
     - `family_kids` -- фермы, дети, природа;
     - `gastro_cultural` -- гастрономия, культура, история.

3. **Hybrid** -- при наличии JWT обогащает точки скорами из Qdrant.

**MapPoint DTO:** id, name, lat/lon, category, density_level, preview_image_url, description_short, is_recommended, recommendation_score.

### Файлы реализации

- `models/map.go` -- MapLocationFilter (Demo, Profile), MapPoint (DescriptionShort, IsRecommended, RecommendationScore);
- `services/map.go` -- 3 режима, 4 curated профиля, enrichWithRecommendations, getDemoLocations;
- `handlers/map.go` -- опциональный JWT, парсинг demo/profile;
- `database/location_repository.go` -- SearchForMap с description_short;
- `handlers/openapi.go` -- MapPoint schema, demo/profile параметры;
- `models/map_test.go` -- unit-тесты моделей (IsDemo, GetProfile, валидация 4 профилей);
- `services/map_service_test.go` -- unit-тесты сервиса (profiles, scores, uniqueness, fallback);
- `docs/phase4/map_api.md` -- документация.

### Метрики

- 9 файлов, +623/-29 строк.
- 62/62 тестов PASS.
- 22 точки на карте (20 seed + 2 пользовательские), 7 рекомендованных.

### Критерии готовности -- выполнены

1. Карта наполнена (22 точки) и осмысленна (curated рекомендации).
2. Payload пригоден для screenshot и будущей реальной карты.

---

## Фаза 5. Demo-срез Location Detail

> **Статус: ЗАВЕРШЕНА** (commit `8122907`)

### Цель

Собрать экран, который завершает demo-story: рекомендация превращается в конкретное место.

### GDD alignment

Прямое покрытие:

- GDD Feature 3
- частично GDD Feature 1 result flow
- location detail as bridge from recommendation to route

### Реализация

1. **`GET /api/v1/locations/{id}`** -- полный detail payload:
   - title (name), slug;
   - category, tags;
   - description_short, description_full;
   - price_per_night, capacity;
   - coordinates (latitude/longitude);
   - preview_image_url (hero image);
   - gallery_urls (массив фотографий галереи);
   - splat_url (3D Gaussian Splatting);
   - density_level, access_level;
   - child_friendly, address.

2. **`GET /api/v1/locations/{id}/splat`** -- ленивая загрузка 3D-сцены:
   - location_id, location_name, has_splat, splat_url;
   - отдельный endpoint для фронтенда -- тяжёлый 3D-контент не включён в основной payload.

3. **Миграция 006** (`gallery_urls TEXT[]`) -- массив URL дополнительных фотографий.

4. **`POST /api/v1/trips`** -- создание trip (реализовано в фазе 3).

### Файлы реализации

- `models/location.go` -- GalleryURLs поле;
- `database/location_repository.go` -- gallery_urls в SELECT/Scan (4 сайта);
- `handlers/location.go` -- GetSplat handler;
- `handlers/router.go` -- маршрут `/locations/:id/splat`;
- `handlers/openapi.go` -- Location schema (gallery_urls, preview_image_url), SplatResponse;
- `migrations/006_add_gallery_urls.sql`;
- `docs/phase5/location_detail.md`.

### Метрики

- 7 файлов, +169/-6 строк.
- Все тесты PASS.

### Критерии готовности -- выполнены

1. Экран выглядит как продуктовая карточка (20+ полей, галерея, 3D preview).
2. Данные согласованы с seed и recommendation payload.

---

## Фаза 5.5. Техническое усиление после демо

> СТАТУС: ЗАВЕРШЕНА (2026-03-21)

### Цель

Сразу после сборки и съемки demo-first контура закрыть тяжелые технические замечания, которые могут вызвать вопросы на code review или помешать нормальному росту в MVP.

### Почему именно здесь

До demo-first эти исправления не должны блокировать визуальный прогресс, но после demo их уже опасно откладывать, потому что дальше они начнут портить MVP-core.

### Обязательные исправления по текущему code review

1. Media upload panic risk:
   - `POST /api/v1/media/upload` регистрируется всегда;
   - при `storage == nil` обработчик упадет на вызове `h.storage.Upload(...)`.
   - Почему это плохо:
     - недоступность MinIO превращается в runtime crash-path вместо контролируемой деградации;
     - внешняя инфраструктурная проблема начинает ломать само приложение;
     - на review это выглядит как отсутствие defensive handling для опциональной зависимости.
   - **Исправлено:** Nil-guard в начале `Upload()` — возвращает 503 Service Unavailable с сообщением "хранилище медиафайлов временно недоступно".

2. Trip join race / overbooking risk:
   - `CountMembers()` и `AddMember()` выполняются раздельно;
   - при параллельных join возможно превысить `group_size`.
   - Почему это плохо:
     - бизнес-инварианта "в поездке не больше N участников" не гарантируется атомарно;
     - под нагрузкой система может вести себя неконсистентно;
     - на review это классический признак отсутствия транзакционной защиты важного ограничения.
   - **Частично исправлено:** Добавлен `AddMemberAtomic()` — `INSERT ... SELECT WHERE (SELECT count(\*) ...) < maxGroupSize`.
   - **Оставшийся риск:** это лучше, чем два раздельных запроса, но всё ещё не даёт строгой гарантии против overbooking при конкурентных транзакциях. Нужен отдельный post-review fix в trip repository / transaction model.

3. Swipe logic mismatch:
   - scene vectors читаются из `user_vibes`;
   - это не совпадает с моделью swipe scenes и делает swipe-flow хрупким/ложным.
   - Почему это плохо:
     - доменная модель "сцена" смешивается с доменной моделью "пользователь";
     - текущий flow может работать только за счет специальных seed-данных или случайно;
     - на review это выглядит как несоответствие реализации продуктовой логике GDD.
   - **Исправлено:** Новая коллекция `scene_vibes` в Qdrant. `SeedSceneVectors()` пишет в `scene_vibes`, `Swipe()` читает сцену из `scene_vibes`, пользователя из `user_vibes`.

4. Voice profiling memory pressure:
   - аудио читается в память;
   - потом снова буферизуется/конвертируется через ffmpeg;
   - при больших файлах и нескольких одновременных запросах это опасно.
   - Почему это плохо:
     - самая тяжелая фича сервиса дополнительно расходует RAM и CPU сверх необходимости;
     - при нескольких одновременных запросах можно быстро упереться в ресурсы;
     - на review это не "микрооптимизация", а реальный operational risk.
   - **Исправлено:** `io.LimitReader(file, 25MB+1)` — ограничение аудио 25 МБ. При превышении — 413 Payload Too Large.

5. Health endpoint cost:
   - каждый запрос к health синхронно пингует все внешние сервисы;
   - для production probes и нагрузочного сценария это слишком тяжело.
   - Почему это плохо:
     - мониторинг сам начинает создавать заметную нагрузку;
     - доступность API становится слишком сильно связана с состоянием всех интеграций сразу;
     - на review это обычно просят разделять на cheap liveness и более дорогой readiness/diagnostics.
   - **Исправлено:** Разделение на `GET /health/live` (мгновенный, liveness) и `GET /health/ready` (полный пинг, readiness). `GET /health` оставлен как alias.

### Реализация

| Исправление  | Подход                   | Файлы                                                                               |
| ------------ | ------------------------ | ----------------------------------------------------------------------------------- |
| Media panic  | Nil-guard → 503          | `handlers/media.go`, `handlers/media_test.go`                                       |
| Trip race    | `INSERT WHERE count < N` | `database/trip_repository.go`, `services/trip.go`                                   |
| Swipe source | Коллекция `scene_vibes`  | `database/vibe_repository.go`, `database/qdrant_collections.go`, `services/vibe.go` |
| Voice memory | `io.LimitReader` 25 МБ   | `handlers/vibe.go`                                                                  |
| Health cost  | Live/Ready split         | `handlers/health.go`, `handlers/router.go`, `handlers/openapi.go`                   |

### Метрики

- Build: PASS, 0 ошибок компиляции
- Тесты: 7/7 пакетов PASS
- Новые тесты: `TestMediaUploadStorageUnavailable`
- Документация: `docs/phase5.5/technical_hardening.md`
- OpenAPI: обновлён (`/health/live`, `/health/ready`)

### Критерии готовности -- выполнены

1. No obvious panic-level traps remain in demo-touched code -- media nil-guard закрывает единственный panic-path.
2. Core endpoints stop relying on brittle demo assumptions -- scene vectors изолированы, для trip join добавлена первичная защита от очевидной гонки.
3. MVP work can continue on top of safer contracts and safer runtime behavior -- все 5 исправлений применены.

---

## Phase 6. Trip Details MVP

### Goal

После demo assets закрыть честный обязательный core step из GDD: trip details.

### GDD alignment

Прямое покрытие:

- GDD Feature 1
- раздел Trip Details

### Current status

✅ **DONE** (21.03.2026)

### Already implemented

1. `POST /api/v1/trips`
2. `GET /api/v1/trips/{id}`
3. `GET /api/v1/trips` — список поездок пользователя
4. `PUT /api/v1/trips/{id}` — частичное обновление поездки (partial update)
5. Trip validation/model/repository foundation
6. Invite-токен: `POST /api/v1/trips/{id}/invite`
7. Присоединение: `POST /api/v1/trips/{id}/join` (с атомарной проверкой group_size)
8. Список участников: `GET /api/v1/trips/{id}/members`

### Gap analysis (что было выявлено)

При сравнении кода с GDD (строки 217-256) выявлены следующие проблемы:

| Элемент                               | Статус до Phase 6 | Проблема                                                                                       |
| ------------------------------------- | :---------------: | ---------------------------------------------------------------------------------------------- |
| `format` (day_trip/weekend/multi_day) |        ❌         | Отсутствовал в DB, модели, DTO, OpenAPI                                                        |
| `vibe_vector_id` (ссылка на Qdrant)   |        ❌         | Отсутствовал в модели Trip и CreateTripRequest (был только `merged_vibe_vector_id` для группы) |
| `PUT /api/v1/trips/{id}`              |        ❌         | Не было endpoint'а для обновления (фронт не мог редактировать поездку)                         |
| `TripWithMembers` (OpenAPI)           |        ❌         | Схема не была определена в OpenAPI, хотя использовалась в ответах                              |
| `UpdateTripRequest` DTO               |        ❌         | Не существовал                                                                                 |
| Дублирование Scan в репозитории       |        ⚠️         | 4 метода (FindByID, FindByInviteToken, FindByUserID, Create) дублировали одинаковый Scan       |

### Что сделано

#### Миграция БД

- `api/migrations/007_add_trip_format_and_vibe_vector.sql`
  - `format TEXT NOT NULL DEFAULT 'multi_day'` + CHECK constraint
  - `vibe_vector_id UUID` (nullable)
  - Миграция идемпотентна (`IF NOT EXISTS`), применяется автоматически при старте API через `runMigrations()`

#### Модели (`api/internal/models/trip.go`)

- Добавлены поля `Format string` и `VibeVectorID *uuid.UUID` в struct `Trip`
- Добавлены поля `Format string` и `VibeVectorID string` в `CreateTripRequest`
- Добавлена карта `ValidFormats` (day_trip, weekend, multi_day)
- Обновлена `Validate()` — валидация `format` по карте, `vibe_vector_id` как UUID
- Обновлена `NormalizeDefaults()` — дефолт `format = "multi_day"`
- Создан `UpdateTripRequest` — DTO для partial update, все поля указатели (nil = не обновлять), своя `Validate()` проверяет только переданные поля

#### Репозиторий (`api/internal/database/trip_repository.go`)

- Вынесен `scanTrip()` helper — единая функция разбора строки в `models.Trip`, устраняет дублирование в 4 методах
- Вынесена `tripColumns` const — список столбцов для SELECT
- Обновлены все SQL-запросы: `Create`, `FindByID`, `FindByInviteToken`, `FindByUserID` — добавлены `format`, `vibe_vector_id`
- Создан метод `Update()` — динамический SQL (`SET col = $N`) только для переданных полей, с RETURNING

#### Сервис (`api/internal/services/trip.go`)

- `Create()` — передаёт `Format` и `VibeVectorID` из DTO в модель, парсинг UUID из строки
- Создан `Update()` — загрузка поездки, проверка `CreatorID == userID`, формирование map изменённых полей, вызов `repo.Update()`, загрузка участников

#### Handler + Router

- `api/internal/handlers/trip.go` — создан `Update()` handler (парсинг JWT, path ID, bind body, validate, service call)
- `api/internal/handlers/router.go` — зарегистрирован `tripsProtected.Put("/:id", tripHandler.Update)`

#### OpenAPI (`api/internal/handlers/openapi.go`)

- Trip schema: добавлены `format` (enum) и `vibe_vector_id` (uuid, nullable)
- CreateTripRequest schema: добавлены `format` и `vibe_vector_id`
- Добавлена `UpdateTripRequest` schema — все поля опциональны
- Добавлена `TripWithMembers` schema (allOf: Trip + members array)
- Добавлен `PUT /api/v1/trips/{id}` endpoint с описанием ответов

#### Тесты (`api/internal/models/trip_test.go`)

- Расширены тесты `CreateTripRequest` — добавлены кейсы format (корректный/некорректный, day_trip, weekend), vibe_vector_id (UUID/не-UUID)
- Добавлен `TestCreateTripRequestNormalizeDefaults_NoOverwrite` — проверка что NormalizeDefaults не перезаписывает явно установленные значения
- Создан полный `TestUpdateTripRequestValidation` — 15 кейсов: пустой запрос, пустые даты, некорректные форматы, порядок дат, отрицательные значения, некорректные enum'ы, корректные partial updates, очистка vibe_vector_id
- Создан `TestValidFormats` — проверка карты допустимых форматов

#### Документация

- `docs/phase6/trip_details_mvp.md` — хендовер-документация для фронтенда

### Баги и проблемы

Критических багов не выявлено. Основные проблемы были архитектурного характера:

1. **Дублирование Scan** — 4 метода репозитория содержали идентичный Scan 12+ полей. Решено: вынесен `scanTrip()` helper.
2. **Порядок полей в Trip struct** — при добавлении `format` между `group_composition` и `invite_token`, а `vibe_vector_id` после `invite_token`, необходимо было синхронно обновить порядок в `tripColumns`, `scanTrip()` и всех INSERT/RETURNING. Решено полной перезаписью файла репозитория.

### Результаты тестирования

- `go build ./...` — ✅ компиляция без ошибок
- `go test ./internal/models/ -v` — ✅ 39 тестов, все PASS
  - 15 TestCreateTripRequestValidation
  - 1 TestCreateTripRequestParseDates
  - 2 TestCreateTripRequestNormalizeDefaults (включая NoOverwrite)
  - 5 TestJoinTripRequestValidation
  - 15 TestUpdateTripRequestValidation
  - 6 TestConstants (Status, Role, BudgetTiers, Transports, Formats)

### Acceptance criteria

1. ✅ Trip object содержит все поля из GDD: dates, budget, transport, group_size, group_composition, format, vibe_vector_id
2. ✅ Frontend может честно сохранять форму через `POST /api/v1/trips` и редактировать через `PUT /api/v1/trips/{id}`
3. ✅ Trip можно использовать в route planning (все необходимые поля присутствуют)

### Post-review follow-up tasks

1. Ограничить доступ к `GET /api/v1/trips/{id}`:
   - detail должен быть доступен только creator/member/admin согласно явной privacy policy;
   - service layer должен принимать `currentUserID` и делать resource-level access check, а не только JWT-проверку на уровне router.
2. Ограничить доступ к `GET /api/v1/trips/{id}/members`:
   - members endpoint не должен раскрывать состав чужой поездки любому авторизованному пользователю;
   - нужно синхронизировать policy между detail, members и invite flow.
3. Зафиксировать privacy contract в OpenAPI и фазе handoff:
   - какие trip-ресурсы публичны по invite-token;
   - какие доступны только участникам;
   - какие доступны только creator/admin.

### Frontend handoff

После этой фазы фронт собирает production-like Trip Details screen. API полностью готово:

- Создание поездки с полными данными формы
- Редактирование через partial update (только изменённые поля)
- Все поля возвращаются в GET-ответах
- OpenAPI документация актуальна в Scalar UI

---

## Phase 7. Route MVP

### Goal

Дать минимально честный route-building, который можно постепенно углублять.

### GDD alignment

Покрывает:

- GDD Feature 1 downstream route logic
- GDD Feature 2 route presentation needs
- частично GDD Feature 4 routing-service

### Current status

✅ **DONE** (Phase 7A + 7B + 7C реализованы)

### Implementation tracking

#### 7A. Demo-safe route preview — ✅ DONE (ранее)

- `POST /api/v1/route/build` — ручной режим, Haversine, transport-aware duration

#### 7B. Trip-aware route building — ✅ DONE

**Что реализовано:**

1. **Миграция 008**: таблицы `routes` и `route_points` с CHECK-ограничениями (status, time_slot, target_audience)
2. **`POST /api/v1/trips/{id}/build-route`**: trip-aware построение маршрута
3. **Trip-aware фильтрация**:
   - `budget_tier` → `max_price_per_night` (economy: 3000₽, comfort: 7000₽, premium: 50000₽)
   - `has_children` (из trip_members.is_child + group_composition) → child_friendly фильтр
   - `format` → max_points лимит (day_trip: 5, weekend: 8, multi_day: 15)
4. **Day planning**: распределение по дням, 3 слота/день (morning/afternoon/evening)
5. **Target audience**: all/adults/children на основе категории локации
6. **Cost estimation**: средняя price_per_night × дни
7. **Summary generation**: детерминистический текст с форматом, транспортом, дистанцией, категориями
8. **DB persistence**: транзакционное сохранение route + route_points

**Файлы:**

| Файл | Действие | Описание |
|------|----------|----------|
| `migrations/008_create_routes_tables.sql` | NEW | Таблицы routes, route_points |
| `models/route.go` | REWRITE | +BuildTripRouteRequest, +Route, +RoutePointDB, расширены RoutePreview/RoutePoint |
| `models/route_test.go` | REWRITE | 17 тестов: Validate, NormalizeDefaults, Constants |
| `database/route_repository.go` | NEW | Create, FindByID, FindPointsByRouteID, FindByTripID |
| `services/route.go` | REWRITE | +BuildTripRoute, +tripRepo/routeRepo/vibeRepo зависимости |
| `handlers/route.go` | REWRITE | +BuildTripRoute handler |
| `handlers/router.go` | MODIFY | +POST /:id/build-route |
| `cmd/api/main.go` | MODIFY | tripRepo/routeRepo вынесены, переданы в RouteService |
| `handlers/openapi.go` | MODIFY | +BuildTripRouteRequest, расширены RoutePreview/RoutePoint schemas |

#### 7C. Vibe integration — ✅ DONE

**Что реализовано:**

1. **Merged vibe vector**: загрузка vibe-векторов всех trip_members из Qdrant, fallback на trip.vibe_vector_id
2. **Среднее арифметическое + L2-нормализация** merged vector
3. **SearchNearest** по `location_vibes` для vibe-скоринга
4. **Ранжирование**: vibe_score + child_friendly бонус (+0.1)
5. **Target audience tagging**: winery/gastro/extreme → adults, farm/beach/nature → children

**Файлы:** Интегрировано в `services/route.go` (computeVibeScores, mergeVectors, determineAudience)

### Bugs fixed

1. **Отсутствующий `ai.Client`**: восстановлен `ai/client.go` — базовый HTTP-клиент для LLM/Embeddings/Whisper (IsMock, buildURL, setAuthHeaders). Был утерян из codebase.

### Test results

```
go build ./... — OK
go test ./internal/models/ -v — 17/17 PASS (route tests)
  TestBuildRouteRequest_Validate — 6 sub-tests
  TestBuildRouteRequest_NormalizeDefaults — 3 sub-tests
  TestBuildTripRouteRequest_Validate — 6 sub-tests
  TestBuildTripRouteRequest_NormalizeDefaults — 5 sub-tests
  TestMaxRoutePointsPerFormat — OK
  TestRouteConstants — OK
```

### Remaining (future phases)

1. Neo4j-backed graph routing (вместо Haversine)
2. Weather penalties
3. LLM-based compromise planning (замена детерминистического summary)
4. Route optimization (nearest-neighbor / TSP heuristic)
5. Polyline generation

### Acceptance criteria

1. ✅ `POST /api/v1/route/build` — usable из UI
2. ✅ `POST /api/v1/trips/{id}/build-route` — trip-aware с vibe scoring
3. ✅ 7B и 7C не ломают контракт 7A
4. ✅ Маршруты сохраняются в БД (routes + route_points)
5. ✅ Day planning + target audience

---

## Phase 8. Group Trip MVP

### Goal

Собрать честный invite/group flow как следующий core layer.

### GDD alignment

Прямое покрытие:

- GDD Feature 1
- раздел Group Planning

### Current status

 **DONE** (все gap-ы закрыты)

### Implementation tracking

#### Concurrency-Safe Join

**Проблема:** `AddMemberAtomic` использовал `INSERT...SELECT WHERE count < N` без блокировки строки, что допускало race condition при параллельных join.

**Решение:** Переписан на транзакционный подход:
1. `SELECT ... FOR UPDATE` на trips row (exclusive row lock)
2. `COUNT(*)` текущих членов
3. Проверка дубликатов: user_id для авторизованных, `LOWER(display_name)` для анонимных
4. `INSERT` нового участника
5. `COMMIT`

#### Membership Privacy

**Проблема:** `GetByID` и `GetMembers` не проверяли, является ли запрашивающий участником.

**Решение:** Добавлен `checkMembership` — проверяет `CreatorID` или `IsMember`. Неучастники получают 403.

#### Auth-Flex Join

**Проблема:** Authenticated join не связывал `trip_member` с данными из `users`.

**Решение:**
- При auth join: `display_name` и `vibe_vector_id` подтягиваются из профиля users
- `user_id` всегда привязывается к `trip_member`
- Fallback на данные из запроса, если профиль недоступен

#### MergeGroupVibes

**Проблема:** Stub, возвращавший nil.

**Решение:**
1. Загрузка vibe-векторов всех участников из Qdrant (CollectionUserVibes)
2. Взвешенное среднее: детские профили = вес 1.2x
3. L2-нормализация
4. Upsert в Qdrant (CollectionGroupVibes)
5. Обновление `trips.merged_vibe_vector_id`
6. SHA1-based UUID для идемпотентности
7. Запускается асинхронно после каждого Join

#### Duplicate Prevention

| Сценарий | Механизм |
|----------|----------|
| Авторизованный | UNIQUE INDEX `(trip_id, user_id) WHERE user_id IS NOT NULL` (миграция 003) |
| Анонимный | UNIQUE INDEX `(trip_id, LOWER(display_name)) WHERE user_id IS NULL` (миграция 009) |

### Файлы

| Файл | Действие | Описание |
|------|----------|----------|
| `migrations/009_group_trip_hardening.sql` | NEW | Unique index для анонимных участников |
| `database/trip_repository.go` | MODIFY | Транзакционный AddMemberAtomic + IsMember + FindMemberByUserID |
| `services/trip.go` | REWRITE | Privacy, auth-flex Join, MergeGroupVibes, vibeRepo |
| `handlers/trip.go` | MODIFY | GetByID/ListMembers передают userID |
| `cmd/api/main.go` | MODIFY | vibeRepo в NewTripService |
| `handlers/openapi.go` | MODIFY | 403 на GET trip/members, обновлён POST join |
| `docs/phase8/group_trip_mvp.md` | NEW | Документация Phase 8 |

### Test results

```
go build ./... -- OK
go test ./internal/models/ -v -- ALL PASS
```

### Remaining (future phases)

1. Route integration (compromise-aware route на основе merged vibe)
2. Mini-vibe profile для участников (замена ручных tags)
3. Member removal / kick
4. Trip status transitions (planning -> active -> completed)

### Acceptance criteria

1. Concurrency-safe join с FOR UPDATE -- capacity guarantee строго соблюдается
2. Membership privacy -- неучастники получают 403
3. Auth-flex join -- авторизованные получают display_name и vibe_vector из профиля
4. MergeGroupVibes -- полноценный merge с Qdrant upsert
5. Duplicate prevention одинаково работает для auth и anon join

---

## Phase 9. Auth Hardening

### Goal

Довести auth до уверенного MVP качества.

### GDD alignment

Поддерживающая секция для всего пользовательского контура GDD.

### Current status

 **DONE** (все gap-ы закрыты)

### Implementation tracking

#### Endpoint Audit (4 проблемы найдены и исправлены)

| Проблема | Описание | Решение |
|----------|----------|---------|
| media/upload без auth | POST /media/upload был публичным | Перемещён под JWT middleware |
| locations без RBAC | PUT/DELETE имели JWT, но без RequireRole | Добавлен RequireRole(host, b2g_admin) |
| Join без OptionalJWT | c.Locals("user_id") всегда nil | Создан OptionalJWTAuth middleware |
| locations POST двойной JWT | POST имел jwtMiddleware дважды | Убрано дублирование |

#### OptionalJWTAuth Middleware

Новый middleware для auth-flex эндпоинтов:
- Если JWT передан и валиден — user_id/role в Locals
- Если JWT отсутствует — пропускает без ошибки
- Если JWT невалиден — пропускает без ошибки (логирует Debug)
- 4 unit-теста: no token, invalid token, valid token, bad format

#### Demo User Bootstrap

Идемпотентное создание при запуске API:
- `demo@deepkrai.ru` / `demo1234` / tourist
- `host@deepkrai.ru` / `host1234` / host
- При повторном запуске ничего не создаёт

### Файлы

| Файл | Действие | Описание |
|------|----------|----------|
| `middleware/optional_jwt.go` | NEW | OptionalJWTAuth middleware |
| `middleware/optional_jwt_test.go` | NEW | 4 тестовых сценария |
| `services/bootstrap.go` | NEW | Демо-пользователи при запуске |
| `handlers/router.go` | MODIFY | Authorization matrix fix |
| `cmd/api/main.go` | MODIFY | Bootstrap вызов |
| `handlers/openapi.go` | MODIFY | Security и RBAC descriptions |
| `docs/phase9/auth_hardening.md` | NEW | Полная документация + Authorization matrix |

### Test results

```
go build ./... -- OK
go test ./internal/middleware/ -- 15 tests ALL PASS
go test ./internal/services/ -- 37 tests ALL PASS
go test ./internal/models/ -- ALL PASS
```

### Acceptance criteria

1. POST /media/upload требует JWT
2. PUT/DELETE /locations/:id требуют JWT + RBAC(host, b2g_admin)
3. POST /trips/:id/join использует OptionalJWTAuth (auth-flex работает)
4. Демо-пользователи создаются при первом запуске
5. Полная Authorization Matrix на 27 эндпоинтов задокументирована

---

## Phase 10. Host Onboarding

### Goal

Следующий большой product slice после tourist core.

### GDD alignment

Прямое покрытие:

- GDD Feature 5
- Zero-UI onboarding pipeline

### Current status

❌ Не реализовано.

### Deliverables

1. `POST /api/v1/host/onboard`
2. task/status model
3. parsing voice into structured location
4. media upload hookup
5. generated location draft

### Demo policy

Можно сделать staged:

1. demo-sync fake progress;
2. async task real contract;
3. real AI/media internals.

---

## Phase 11. Booking MVP

### Goal

Закрыть базовую коммерческую механику.

### GDD alignment

Прямое покрытие:

- GDD Feature 6

### Current status

❌ Не реализовано.

### Deliverables

1. booking slots schema and endpoints
2. create booking
3. confirm/reject booking
4. host bookings / my bookings

### Can wait until after tourist core

Да, это не входит в первый обязательный visual demo slice.

---

## Phase 12. Storytelling + Weather + Live Routing

### Goal

Достроить wow-features после core.

### GDD alignment

Прямое покрытие:

- GDD Feature 4
- части GDD Feature 2
- части offline/live сценариев

### Current status

❌ Не реализовано.

### Deliverables

1. `GET /api/v1/weather/region`
2. `/ws/v1/weather`
3. storytelling route assets
4. live route rebuild

---

## Phase 13. Hidden Gems + Reviews + Karma

### Goal

Постепенно включить социальную и access-логику из GDD.

### GDD alignment

Прямое покрытие:

- GDD Feature 7

### Current status

❌ Не реализовано как feature slice.

### Already implemented

1. schema/data foundation for `access_level`
2. `karma` field in users
3. hidden/semi-open enums in models and migrations

### Deliverables

1. reviews
2. karma updates
3. access levels
4. hidden location gating

---

## Phase 14. Analytics + B2G

### Goal

Закрыть B2G-ветку GDD после consumer/host core.

### GDD alignment

Прямое покрытие:

- GDD Feature 8

### Current status

❌ Не реализовано как product/API slice.

### Already implemented

1. ClickHouse connection layer
2. health visibility for ClickHouse in infra/ops contour

### Deliverables

1. event ingestion
2. ClickHouse pipeline
3. analytics endpoints
4. predictions/problematic zones

---

## Phase 15. Full GDD Completion

### Goal

Довести backend до полного покрытия GDD.

### GDD alignment

Эта фаза закрывает все неохваченные части GDD и является контрольной точкой "roadmap == GDD".

### Must include

1. Map hybrid search
2. Route graph + Neo4j
3. Weather-aware rerouting
4. Storytelling
5. Booking + notifications
6. Host onboarding full pipeline
7. Hidden Gems + karma
8. Analytics + B2G
9. Offline sync support endpoints
10. WebSocket channels

### Current status

🔶 Частично реализовано как foundation only.

### Already implemented

1. infra connectors for PostgreSQL, Redis, Qdrant, Neo4j, ClickHouse, MinIO
2. tourist core demo-first API
3. auth/profile/trips/map/locations/media/route preview base

### Gap to complete

1. full host/business slice
2. booking/inventory
3. live/weather/storytelling
4. hidden gems logic
5. analytics/B2G APIs
6. offline/sync
7. websocket channels
8. trip privacy / membership authorization model fully enforced
9. consistent authenticated/public join contract
10. strict concurrency-safe trip capacity enforcement

---

## Архитектурный backlog после code review

### P0 -- обязательно до уверенного MVP

1. **Trip privacy and access control**
   - закрыть несанкционированный доступ к `GET /api/v1/trips/{id}` и `GET /api/v1/trips/{id}/members`;
   - внедрить resource-level authorization в service layer;
   - формализовать matrix доступа: creator / member / admin / invite-only.

2. **Join auth contract**
   - разделить и явно описать public invite join и authenticated join;
   - убрать зависимость handler/service от `user_id` в контексте там, где JWT middleware не применяется;
   - гарантировать корректную привязку `trip_member.user_id` для залогиненных участников.

3. **Concurrency-safe trip capacity**
   - заменить текущую count-based вставку на решение, которое реально держит инвариант `members <= group_size` под параллельной нагрузкой;
   - покрыть этот сценарий интеграционным concurrency-тестом.

### P1 -- сразу после P0

1. Синхронизировать OpenAPI, router и service contracts для всех trip/group endpoint-ов.
2. Добавить явные acceptance criteria на privacy/auth behavior для invite/group flows.

### P2 -- до полного закрытия GDD

1. Довести merged group vibe до рабочей интеграции с route planning.
2. Расширить participant mini-profile и compromise logic для group planning.

---

## Backend Sprint Plan

## Sprint B1. Demo Contract + Seed Layer

### Backend work

1. Freeze payloads
2. Prepare seed/demo data
3. Add demo mode switches
4. Finalize API docs for frontend handoff

### Frontend dependency

Frontend starts static assembly immediately after payload freeze.

---

## Sprint B2. Voice + Passport

### Backend work

1. Finish `/profile/voice`
2. deterministic AI fallback
3. user profile persistence

### Frontend handoff

Frontend builds:

1. voice screen
2. AI thinking state
3. vibe passport

---

## Sprint B3. Recommendations + Map Payload

### Backend work

1. `/profile/finalize`
2. `/map/locations`
3. recommendation data join

### Frontend handoff

Frontend builds:

1. recommendation/map screen
2. screenshot-ready map

---

## Sprint B4. Location Detail + Route Preview

### Backend work

1. location detail polish
2. optional splat url
3. route preview endpoint

### Frontend handoff

Frontend finalizes:

1. location card/detail
2. CTA "Собрать маршрут"
3. final shot for screencast

---

## Sprint B5. Trip Details + Invite Core

### Backend work

1. trip details
2. invite
3. join
4. members

### Frontend handoff

Frontend can now extend from promo/demo to true app flow.

---

## Completion Checklist For Demo-First Backend

### Required before shooting

- [ ] Stable voice profiling response
- [ ] Stable vibe passport payload
- [ ] Stable recommendation payload
- [ ] Stable map payload
- [ ] Stable location detail payload
- [ ] Demo seed data loaded
- [ ] No frontend-only fake entities that contradict backend schema

### Nice to have before shooting

- [ ] Real trip creation behind CTA
- [ ] Real location details
- [ ] Optional real swipe scenes

---

## Риски

### Риск 1. Потратить слишком много времени на одноразовые demo-хаки

Как снижать:

- мокать только внутренности;
- сохранять стабильные контракты эндпоинтов;
- переиспользовать demo-код в MVP.

### Риск 2. Фронтенд блокируется нестабильными backend-схемами

Как снижать:

- рано замораживать payload;
- версионировать demo-примеры payload;
- избегать постоянной смены формы ответов.

### Риск 3. Слишком рано переусложнить не-core фичи GDD

Как снижать:

- не идти в погоду, аналитику, телефонию и B2G до закрытия tourist core;
- не делать полный graph routing до route preview уровня MVP.

---

## Definition of done по этапам

### Готово для демо

- бэкенд полностью поддерживает 30-секундный сценарий скринкаста;
- фронтенд умеет рендерить финально выглядящий UI на стабильных контрактах;
- временные моки находятся за future-proof эндпоинтами и не ломают дальнейший рост.

### Готово для MVP

- пользователь может авторизоваться;
- создать и использовать свой профиль;
- пройти полный vibe-flow;
- создать поездку;
- посмотреть рекомендации;
- открыть детальную страницу локации;
- собрать базовый маршрут;
- пригласить участников.

### Полное покрытие GDD

- все основные туристические, хостовые, live-, аналитические и B2G-системы из GDD реализованы на реальной инфраструктуре и стабильных контрактах.
