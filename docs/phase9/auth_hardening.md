# Phase 9: Auth Hardening

## Обзор

Укрепление модуля аутентификации и авторизации до MVP-качества.
Полный аудит эндпоинтов, RBAC-консистентность, OptionalJWTAuth для mixed endpoints,
демо-пользователи для интеграции фронтенда.

## Выявленные и исправленные проблемы

### 1. Media Upload без авторизации

**Проблема:** `POST /media/upload` был публичным — любой пользователь мог загружать файлы.

**Решение:** Эндпоинт перемещён под JWT middleware. Загрузка файлов требует авторизации.

### 2. Locations PUT/DELETE без RBAC

**Проблема:** `PUT/DELETE /locations/:id` имели JWT middleware, но не имели RBAC.
Любой авторизованный пользователь (включая tourist) мог редактировать/удалять локации.

**Решение:** Добавлен `RequireRole(host, b2g_admin)` для PUT и DELETE операций.
POST уже имел RBAC, но с двойным JWT middleware (убрано дублирование).

### 3. Join без OptionalJWTAuth

**Проблема:** `POST /trips/:id/join` был полностью публичным. `c.Locals("user_id")` в handler
всегда возвращал nil, т.к. JWT middleware не применялся. Auth-flex не работал.

**Решение:** Создан `OptionalJWTAuth` middleware — парсит JWT если передан, но не блокирует
запрос при отсутствии. При наличии JWT user_id инжектится в Locals.

### 4. Отсутствие демо-пользователей

**Проблема:** Для тестирования и интеграции фронтенда требовалось ручное создание аккаунтов.

**Решение:** Идемпотентный bootstrap при запуске API:
- `demo@deepkrai.ru` / `demo1234` / tourist
- `host@deepkrai.ru` / `host1234` / host

## Authorization Matrix (итоговая)

| Endpoint | Method | Auth | RBAC | Resource |
|----------|--------|------|------|----------|
| `/auth/register` | POST | - | - | - |
| `/auth/login` | POST | - | - | - |
| `/auth/refresh` | POST | - | - | - |
| `/health/*` | GET | - | - | - |
| `/docs/*` | GET | - | - | - |
| `/locations` | GET | - | - | - |
| `/locations/:id` | GET | - | - | - |
| `/map/locations` | GET | - | - | - |
| `/media/upload` | POST | JWT | - | - |
| `/profile/me` | GET/PUT | JWT | - | - |
| `/profile/voice` | POST | JWT | - | - |
| `/profile/swipe` | POST | JWT | - | - |
| `/profile/finalize` | POST | JWT | - | - |
| `/profile/scenes` | GET | JWT | - | - |
| `/locations/` | POST | JWT | host,b2g_admin | - |
| `/locations/:id` | PUT | JWT | host,b2g_admin | - |
| `/locations/:id` | DELETE | JWT | host,b2g_admin | - |
| `/trips/` | GET/POST | JWT | - | - |
| `/trips/:id` | GET | JWT | - | membership |
| `/trips/:id` | PUT | JWT | - | creator |
| `/trips/:id/invite` | POST | JWT | - | creator |
| `/trips/:id/members` | GET | JWT | - | membership |
| `/trips/:id/join` | POST | OptionalJWT | - | invite_token |
| `/trips/:id/build-route` | POST | JWT | - | - |
| `/route/build` | POST | JWT | - | - |

## Файлы

| Файл | Действие | Описание |
|------|----------|----------|
| `middleware/optional_jwt.go` | NEW | OptionalJWTAuth middleware |
| `middleware/optional_jwt_test.go` | NEW | 4 тестовых сценария |
| `services/bootstrap.go` | NEW | Демо-пользователи при запуске |
| `handlers/router.go` | MODIFY | Authorization matrix fix |
| `cmd/api/main.go` | MODIFY | Bootstrap вызов |
| `handlers/openapi.go` | MODIFY | Security и RBAC descriptions |
