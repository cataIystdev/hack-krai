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

| Фича GDD | Фазы roadmap бэкенда |
| --- | --- |
| Feature 1. Smart trip planning | 0, 1, 2, 3, 5, 6, 7, 8, 9 |
| Feature 2. 3D map | 4, 7, 12, 15 |
| Feature 3. 3D tours / splat | 5, 10, 15 |
| Feature 4. Live routing & storytelling | 7, 12, 15 |
| Feature 5. Host onboarding | 10, 15 |
| Feature 6. Booking | 11, 15 |
| Feature 7. Hidden Gems / karma | 13, 15 |
| Feature 8. Analytics / B2G | 14, 15 |
| Feature 9. Offline mode backend support | 12, 14, 15 |

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
   - migrations;
   - public list/detail;
   - protected create/update/delete;
   - PostGIS search foundation.

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

7. Media/docs:
   - media upload;
   - health endpoint;
   - OpenAPI;
   - Scalar UI.

8. Schema/migrations:
   - users;
   - locations;
   - trips/trip_members;
   - swipe_scenes.

### Частично реализовано

1. Recommendations exist, but payload is still thinner than the frontend demo needs.
2. Group trip exists, but merged vibe and route-build are still missing.
3. `.splat` support exists in schema/model/media storage, but no dedicated delivery flow for product UI.
4. Polyglot infra is wired, but many storages are not yet used by shipped product slices.

### Пока не реализовано

1. `GET /api/v1/map/locations`
2. `POST /api/v1/route/build`
3. `POST /api/v1/trips/{id}/build-route`
4. weather/live routing
5. storytelling
6. host onboarding
7. booking
8. reviews/karma/hidden gems
9. analytics/B2G
10. sync/offline support endpoints

### Что это значит для планирования

Backend roadmap starts from a substantial core, not from zero:

- Phase 0 is mainly contract freeze and alignment work.
- Phases 2, 3, 6, 8, 9 are already partially de-risked by existing code.
- The biggest near-term backend gaps for demo-first are:
  - map payloads;
  - enriched recommendation payloads;
  - route preview payloads;
  - demo content curation.

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

### Цель

Заморозить минимальный набор API-контрактов для первого скринкаста и скриншотов.

### GDD alignment

Закрывает подготовительный слой для:

- GDD Feature 1
- GDD Feature 2
- GDD Feature 3

### Текущее состояние

Частично уже реализовано.

### Уже реализовано

1. Большая часть core endpoints уже существует.
2. OpenAPI/Scalar уже генерируются из кода.
3. Основные сущности и их response shapes уже можно опереть на реальный код.

### Что осталось доделать

1. Зафиксировать demo-specific DTOs для map/location/route.
2. Согласовать JSON examples под конкретные screenshot/screencast screens.
3. Убрать неявность между "что уже есть" и "что фронт реально должен ждать".

### Следующий handoff

Frontend должен получить frozen payloads для:

- voice result;
- vibe passport;
- map points;
- location detail;
- route preview.

### Результаты фазы

1. Список demo-сцен и их payload.
2. OpenAPI/markdown контракты для:
   - `POST /api/v1/profile/voice`
   - `GET /api/v1/profile/scenes`
   - `POST /api/v1/profile/swipe`
   - `POST /api/v1/profile/finalize`
   - `POST /api/v1/trips`
   - `GET /api/v1/locations`
   - `GET /api/v1/locations/{id}`
   - `GET /api/v1/map/locations`
   - `POST /api/v1/route/build`
3. Единый JSON shape для:
   - vibe passport;
   - recommendation cards;
   - map points;
   - location detail;
   - route preview.

### Критерии готовности

1. Frontend может работать без догадок.
2. Ни один ключевой payload не придется радикально переписывать после demo.

### Handoff во фронтенд

После этой фазы фронт может начинать:

- hero flow;
- voice UI;
- vibe passport UI;
- map screen UI;
- location detail UI.

---

## Фаза 1. Основа demo-данных

### Цель

Подготовить данные, на которых можно быстро собрать красивое и правдоподобное демо.

### GDD alignment

Подготавливает данные для:

- Feature 1 recommendations
- Feature 2 map points
- Feature 3 location previews

### Текущее состояние

Частично реализовано.

### Уже реализовано

1. Есть migration/seed основа для swipe scenes.
2. Есть общая location schema и seed script foundation.
3. Есть mock AI outputs для детерминированных demo-ответов.

### Что осталось доделать

1. Curated demo locations set для презентации.
2. Preview images and optional fake splat assets.
3. Stable recommendation bundles для 1-2 demo profiles.
4. Demo users/demo trip baseline.

### Следующий handoff

Frontend получает стабильные demo fixtures, совпадающие с backend IDs и DTOs.

### Результаты фазы

1. Seed users:
   - tourist demo user;
   - host demo user.

2. Seed locations:
   - 8-12 красивых точек;
   - 3-5 приоритетных demo locations;
   - нормальные категории, теги, description_short, description_full;
   - координаты, пригодные для карты.

3. Seed swipe scenes:
   - 6-8 сцен;
   - изображения/обложки;
   - display order.

4. Seed recommendation set:
   - заранее подготовленные top recommendations для 1-2 demo профилей.

5. Demo media references:
   - preview images;
   - optional fake `splat_url`;
   - location hero assets.

### Политика моков

Допустимо:

- заранее зафиксировать 1-2 демо-профиля;
- заранее зафиксировать топ локаций под эти профили;
- временно подставить route preview вместо реального graph build.

### Критерии готовности

1. Один и тот же demo-user стабильно получает красивый и ожидаемый результат.
2. Все demo assets согласованы с будущими сущностями в БД.

### Handoff во фронтенд

После этой фазы фронт может:

- рендерить реальные карточки локаций;
- рендерить карту с осмысленными точками;
- рендерить экраны без вымышленных полей.

---

## Фаза 2. Demo-срез Voice-To-Vibe

### Цель

Дать полностью рабочий backend flow для сцены:
voice input -> AI thinking -> vibe passport.

### GDD alignment

Прямое покрытие:

- GDD Feature 1
- раздел Voice-to-Vibe
- vibe axes / summary / tags / vector pipeline

### Текущее состояние

В значительной степени уже реализовано.

### Уже реализовано

1. `POST /api/v1/profile/voice`
2. STT abstraction with mock mode
3. LLM extraction layer
4. Embeddings layer
5. Qdrant upsert
6. user vibe vector persistence

### Что осталось доделать

1. Зафиксировать screenshot-ready response shape.
2. Подготовить deterministic demo profile.
3. При необходимости добавить controlled demo latency.
4. Полировать error/loading semantics for frontend.

### Следующий handoff

Frontend может начинать voice/passport screens почти сразу после freeze примеров ответа.

### Результаты фазы

1. `POST /api/v1/profile/voice`
   - принимает audio;
   - возвращает:
     - transcription;
     - axes;
     - extracted_tags;
     - vibe_summary;
     - vector_id.

2. Demo mode:
   - deterministic response for selected demo clip and/or missing API keys;
   - configurable mock mode, не требующий переписывать frontend.

3. Persist profile result:
   - upsert vector;
   - update user vibe reference.

4. Response timing policy:
   - для demo можно держать controlled latency, чтобы "ИИ думает" выглядел естественно, но не тормозил.

### Что желательно сделать по-настоящему

Желательно сделать по-настоящему уже сейчас:

- endpoint shape;
- сохранение результата в user profile;
- mock/real AI abstraction;
- единая модель vibe axes.

### Что можно временно замокать

- transcription;
- LLM extraction;
- embeddings generation.

### Критерии готовности

1. Один запрос дает стабильный vibe passport.
2. Payload полностью пригоден для скриншота без фронтовых костылей.
3. Frontend не знает, mock там или real.

### Handoff во фронтенд

После этой фазы фронт должен закрыть:

- voice record UI;
- loading/AI thinking state;
- vibe passport screen;
- screenshot-ready layout.

---

## Фаза 3. Swipe + finalize recommendations

### Цель

Дать второй кусок core-vibe flow:
scenes -> swipe -> finalize -> recommendations.

### GDD alignment

Прямое покрытие:

- GDD Feature 1
- раздел Emotional 3D Swipe
- переход к recommendation layer

### Текущее состояние

Частично реализовано.

### Уже реализовано

1. `GET /api/v1/profile/scenes`
2. `POST /api/v1/profile/swipe`
3. `POST /api/v1/profile/finalize`
4. Vector search foundation in Qdrant

### Что осталось доделать

1. Enrich recommendation payload for frontend screens.
2. Stabilize curated finalize output for demo capture.
3. Decide whether swipe stays visible in first screencast or optional for later.

### Следующий handoff

Frontend needs enriched finalize DTO to build map and recommendation transitions.

### Результаты фазы

1. `GET /api/v1/profile/scenes`
2. `POST /api/v1/profile/swipe`
3. `POST /api/v1/profile/finalize`
4. Recommendation payload:
   - `location_id`
   - `name`
   - `category`
   - `score`
   - `preview image`
   - optional `splat_url`
   - short reason / tags match

### Цель MVP-реализации

Сделать честно:

- scenes из PostgreSQL;
- finalize через Qdrant или controlled demo mapping;
- выдача top-N.

### Временное упрощение

Если не хватает времени:

- не делать обязательный swipe в первом скринкасте;
- `finalize` может использовать уже полученный voice-profile и возвращать стабильный curated top-N.

### Критерии готовности

1. Recommendation set связан с реальными location IDs.
2. Один и тот же профиль дает предсказуемый набор demo-рекомендаций.

### Handoff во фронтенд

После этой фазы фронт может собирать:

- recommendation list;
- recommendation map;
- переход из vibe passport в карту/локации.

---

## Фаза 4. Map API для демо

### Цель

Отдать фронтенду API для wow-сцены карты и screenshot hero shot.

### GDD alignment

Прямое покрытие:

- GDD Feature 2
- `GET /api/v1/map/locations`
- карта рекомендованных / доступных локаций

### Текущее состояние

Не реализовано.

### Уже реализовано

1. PostGIS location search foundation exists.
2. Recommendation source data can already come from vibe finalize.

### Что осталось доделать

1. Introduce dedicated `GET /api/v1/map/locations`.
2. Shape map-point DTO for frontend.
3. Decide first implementation mode:
   - bbox;
   - recommended points;
   - hybrid.

### Следующий handoff

Это один из главных backend blockers для screenshot-ready map screen.

### Результаты фазы

1. `GET /api/v1/map/locations`
   - bbox or simple recommended mode;
   - отдача точек для карты;
   - support flags:
     - recommended;
     - category;
     - density;
     - score.

2. Map point schema:
   - id
   - name
   - lat/lon
   - category
   - density level
   - preview image
   - is_recommended
   - recommendation_score

3. Optional demo query modes:
   - `?demo=true`
   - `?profile=calm_wine_mountains`

### Политика моков

Допустимо:

- сначала отдать curated set points;
- потом заменить внутреннюю выборку на реальный hybrid search.

### Критерии готовности

1. Карта выглядит наполненной и осмысленной.
2. Payload годится и для screenshot, и для будущей реальной карты.

### Handoff во фронтенд

После этой фазы фронт может финализировать screenshot:

- карта края;
- glowing markers;
- highlighted recommendations.

---

## Фаза 5. Demo-срез Location Detail

### Цель

Собрать экран, который завершает demo-story: рекомендация превращается в конкретное место.

### GDD alignment

Прямое покрытие:

- GDD Feature 3
- частично GDD Feature 1 result flow
- location detail as bridge from recommendation to route

### Текущее состояние

Частично реализовано.

### Уже реализовано

1. `GET /api/v1/locations/:id`
2. Rich location schema with category/tags/descriptions/coordinates
3. `splat_url` field already exists in model/schema

### Что осталось доделать

1. Confirm frontend-ready DTO shape.
2. Curate high-quality demo content for selected locations.
3. Add optional `GET /api/v1/locations/{id}/splat` if product UI needs a separate fetch.

### Следующий handoff

Frontend can start location detail implementation as soon as content and shape are frozen.

### Результаты фазы

1. `GET /api/v1/locations/{id}`
2. Optional `GET /api/v1/locations/{id}/splat`
3. Detail payload:
   - title
   - category
   - short/full description
   - tags
   - price
   - gallery/hero image
   - coordinates
   - optional 3D preview link

4. `POST /api/v1/trips`
   - создать trip details из экрана location flow.

### Политика моков

Допустимо:

- `splat_url` временно вести на заранее подготовленный asset;
- CTA "Собрать маршрут" пока может вести к route preview response.

### Критерии готовности

1. Экран выглядит как будущая продуктовая карточка, а не как временная заглушка.
2. Данные location detail не противоречат seed и recommendation payload.

### Handoff во фронтенд

После этой фазы можно снимать:

- карточку локации;
- финал скринкаста;
- третий скриншот.

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
