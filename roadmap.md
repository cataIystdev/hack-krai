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
6. reviews/karma/hidden gems
7. analytics/B2G
8. sync/offline support endpoints

### Что это значит для планирования

Backend roadmap покрывает полный demo-first контур:

- Фазы 0-5 завершены.
- Фаза 5.5 (техническое усиление) -- следующий приоритет.
- Demo-first flow полностью функционален: voice -> vibe -> swipe -> finalize -> map -> location detail.
- Основные оставшиеся gap-ы: weather, live routing, storytelling, booking, B2G.

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

2. Trip join race / overbooking risk:
   - `CountMembers()` и `AddMember()` выполняются раздельно;
   - при параллельных join возможно превысить `group_size`.
   - Почему это плохо:
     - бизнес-инварианта "в поездке не больше N участников" не гарантируется атомарно;
     - под нагрузкой система может вести себя неконсистентно;
     - на review это классический признак отсутствия транзакционной защиты важного ограничения.

3. Swipe logic mismatch:
   - scene vectors читаются из `user_vibes`;
   - это не совпадает с моделью swipe scenes и делает swipe-flow хрупким/ложным.
   - Почему это плохо:
     - доменная модель "сцена" смешивается с доменной моделью "пользователь";
     - текущий flow может работать только за счет специальных seed-данных или случайно;
     - на review это выглядит как несоответствие реализации продуктовой логике GDD.

4. Voice profiling memory pressure:
   - аудио читается в память;
   - потом снова буферизуется/конвертируется через ffmpeg;
   - при больших файлах и нескольких одновременных запросах это опасно.
   - Почему это плохо:
     - самая тяжелая фича сервиса дополнительно расходует RAM и CPU сверх необходимости;
     - при нескольких одновременных запросах можно быстро упереться в ресурсы;
     - на review это не "микрооптимизация", а реальный operational risk.

5. Health endpoint cost:
   - каждый запрос к health синхронно пингует все внешние сервисы;
   - для production probes и нагрузочного сценария это слишком тяжело.
   - Почему это плохо:
     - мониторинг сам начинает создавать заметную нагрузку;
     - доступность API становится слишком сильно связана с состоянием всех интеграций сразу;
     - на review это обычно просят разделять на cheap liveness и более дорогой readiness/diagnostics.

### Результаты фазы

1. Safe media upload behavior when MinIO unavailable
2. Atomic trip join strategy
3. Correct source of swipe vectors
4. Reduced memory footprint in voice pipeline
5. Cheaper health/readiness split if needed

### Критерии готовности

1. No obvious panic-level traps remain in demo-touched code.
2. Core endpoints stop relying on brittle demo assumptions.
3. MVP work can continue on top of safer contracts and safer runtime behavior.

---

## Phase 6. Trip Details MVP

### Goal

После demo assets закрыть честный обязательный core step из GDD: trip details.

### GDD alignment

Прямое покрытие:

- GDD Feature 1
- раздел Trip Details

### Current status

Частично реализовано.

### Already implemented

1. `POST /api/v1/trips`
2. `GET /api/v1/trips/{id}`
3. Trip validation/model/repository foundation

### Gap to complete

1. Ensure request/response fully match the desired frontend form.
2. Expand payload only where it helps downstream route/group flows.

### Next handoff

Frontend can build the real trip form against mostly real APIs instead of mocking it from zero.

### Deliverables

1. Нормализованный `POST /api/v1/trips`
2. `GET /api/v1/trips/{id}`
3. Поля:
   - dates
   - budget
   - transport
   - group_size
   - group_composition
   - format
   - optional `vibe_vector_id`

4. Validation и error handling.

### Acceptance criteria

1. Trip object можно использовать дальше в route planning.
2. Frontend может честно сохранять форму, а не держать ее локально.

### Frontend handoff

После этой фазы фронт собирает production-like Trip Details screen.

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

Не реализовано.

### Already implemented

1. Building blocks exist:
   - trips;
   - locations;
   - vectors;
   - Neo4j connector.

### Gap to complete

1. First route preview contract.
2. Deterministic/demo-safe route summary.
3. Future-proof endpoint shape for later real routing.

### Next handoff

Frontend route and final CTA flow should wait for at least preview-grade route payload.

### Deliverables

1. `POST /api/v1/route/build`
2. Later `POST /api/v1/trips/{id}/build-route`
3. Response:
   - ordered locations
   - estimated time
   - route summary
   - reasons/fit
   - optional polyline placeholder

### Iteration plan

#### 7A. Demo-safe route preview

- curated route from selected locations;
- deterministic time estimates.

#### 7B. MVP real route

- budget/date/group-aware filtering;
- heuristic ordering;
- Neo4j integration later.

#### 7C. Full GDD route

- merged vibe;
- graph routing;
- weather penalties;
- AI compromise planning.

### Acceptance criteria

1. Уже на этапе 7A endpoint usable из UI.
2. 7B и 7C не ломают контракт.

---

## Phase 8. Group Trip MVP

### Goal

Собрать честный invite/group flow как следующий core layer.

### GDD alignment

Прямое покрытие:

- GDD Feature 1
- раздел Group Planning

### Current status

Частично реализовано.

### Already implemented

1. invite regeneration
2. join endpoint
3. members endpoint

### Gap to complete

1. merged vibe
2. participant preference enrichment
3. route integration

### Next handoff

Frontend later can build group UI mostly on top of existing trip APIs.

### Deliverables

1. `POST /api/v1/trips/{id}/invite`
2. `POST /api/v1/trips/{id}/join`
3. `GET /api/v1/trips/{id}/members`
4. mini-profile payload for participant

### Temporary simplification

Допустимо сначала:

- ручные tags вместо полного mini-vibe;
- merged vibe как deterministic placeholder.

### Full target

- merge participant vectors;
- compromise-aware route.

---

## Phase 9. Auth Hardening

### Goal

Довести auth до уверенного MVP качества.

### GDD alignment

Поддерживающая секция для всего пользовательского контура GDD.

### Current status

Базово уже реализовано.

### Already implemented

1. register
2. login
3. refresh
4. JWT middleware
5. RBAC middleware

### Gap to complete

1. Add only those auth-support helpers that frontend truly needs.
2. Avoid overbuilding auth while map/recommendation demo blockers remain higher priority.

### Deliverables

1. Register/login/refresh polish
2. Role checks
3. demo user bootstrap
4. session UX helper endpoints if needed

### Why now

Auth не должен блокировать demo-first сборку UI, но должен быть готов рано, чтобы не врастать в mock auth.

---

## Phase 10. Host Onboarding

### Goal

Следующий большой product slice после tourist core.

### GDD alignment

Прямое покрытие:

- GDD Feature 5
- Zero-UI onboarding pipeline

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
