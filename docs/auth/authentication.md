# Документация: Аутентификация Deep Krai API

## Общие сведения

Система аутентификации Deep Krai API построена на JWT (JSON Web Tokens) с HMAC-SHA256 подписью и ролевой моделью доступа (RBAC).

## Роли пользователей

| Роль      | Описание                                         |
| --------- | ------------------------------------------------ |
| tourist   | Обычный пользователь (турист). Роль по умолчанию |
| host      | Хозяин локации (фермер, владелец агроусадьбы)    |
| b2g_admin | Администратор B2G-платформы                      |

## Токены

- **Access-токен** -- короткоживущий (15 минут), используется для авторизации API-запросов
- **Refresh-токен** -- долгоживущий (7 дней), используется для обновления access-токена

### Payload токена (claims)

| Поле       | Описание                         |
| ---------- | -------------------------------- |
| user_id    | UUID пользователя                |
| role       | Роль (tourist/host/b2g_admin)    |
| token_type | Тип: "access" или "refresh"      |
| iss        | Issuer (deep-krai-api)           |
| sub        | Subject (UUID пользователя)      |
| exp        | Время истечения (Unix timestamp) |
| iat        | Время выпуска (Unix timestamp)   |

## Эндпоинты

### POST /api/v1/auth/register

Регистрация нового пользователя. Пароль хэшируется bcrypt (cost=12).

### POST /api/v1/auth/login

Авторизация по email и паролю. Возвращает пару access+refresh токенов.

### POST /api/v1/auth/refresh

Обновление токенов по refresh-токену.

### GET /api/v1/profile/me

Возвращает профиль текущего пользователя. Требует заголовок `Authorization: Bearer <access_token>`.

## Middleware

### JWT Middleware

Извлекает токен из `Authorization: Bearer <token>`, валидирует, инжектит `user_id` и `role` в контекст запроса.

### RBAC Middleware

Проверяет роль из контекста. Используется через `RequireRole("b2g_admin", "host")`.

## Связанные файлы

- [user.go](file:///mnt/Work/HACKS/VORONKA/20032026/api/internal/models/user.go) -- модель пользователя
- [jwt.go](file:///mnt/Work/HACKS/VORONKA/20032026/api/internal/services/jwt.go) -- JWT-сервис
- [auth.go](file:///mnt/Work/HACKS/VORONKA/20032026/api/internal/services/auth.go) -- сервис аутентификации
- [middleware/jwt.go](file:///mnt/Work/HACKS/VORONKA/20032026/api/internal/middleware/jwt.go) -- JWT middleware
- [middleware/rbac.go](file:///mnt/Work/HACKS/VORONKA/20032026/api/internal/middleware/rbac.go) -- RBAC middleware
- [handlers/auth.go](file:///mnt/Work/HACKS/VORONKA/20032026/api/internal/handlers/auth.go) -- auth handlers
- [handlers/profile.go](file:///mnt/Work/HACKS/VORONKA/20032026/api/internal/handlers/profile.go) -- profile handler
- [001_create_users_table.sql](file:///mnt/Work/HACKS/VORONKA/20032026/api/migrations/001_create_users_table.sql) -- SQL-миграция
