// Файл openapi.go содержит полную спецификацию OpenAPI 3.1 для КудыТуды API.
// Спецификация описывает все реализованные эндпоинты с подробными описаниями
// параметров, тел запросов, ответов, примеров и схем данных.
// Генерируется программно для обеспечения синхронизации с кодом.
package handlers

// OpenAPISpec возвращает полную спецификацию OpenAPI 3.1 в виде Go-структуры.
// Структура сериализуется в JSON при обработке запроса GET /api/v1/docs/openapi.json.
// Параметр baseURL определяет адрес сервера, отображаемый в Scalar UI.
func OpenAPISpec(baseURL string) map[string]any {
	return map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       "КудыТуды API",
			"version":     "1.0.0",
			"description": "REST API платформы пространственного туризма КудыТуды.\n\nКудыТуды -- PWA-платформа, соединяющая путешественников со скрытыми местами Краснодарского края через мультимодальный ИИ, 3D-визуализацию (Gaussian Splatting) и интеллектуальное построение маршрутов.\n\n## Аутентификация\n\nAPI использует JWT (JSON Web Tokens) для аутентификации.\nДля доступа к защищённым эндпоинтам передавайте access-токен в заголовке `Authorization: Bearer <token>`.\nТокены получаются через POST /api/v1/auth/login или POST /api/v1/auth/register.",
			"contact": map[string]any{
				"name": "КудыТуды Team",
			},
			"license": map[string]any{
				"name": "Proprietary",
			},
		},
		"servers": []map[string]any{
			{
				"url":         baseURL,
				"description": "Текущий сервер",
			},
		},
		"tags": []map[string]any{
			{
				"name":        "Мониторинг",
				"description": "Эндпоинты для проверки состояния системы и здоровья сервисов. Используются для мониторинга, алертинга и проверки работоспособности инфраструктуры (liveness/readiness probes).",
			},
			{
				"name":        "Медиа",
				"description": "Эндпоинты для работы с медиафайлами: загрузка изображений, видео, аудио и 3D-файлов (.splat) в S3-совместимое хранилище MinIO. Поддерживается multipart/form-data загрузка с валидацией MIME-типов и ограничением размера.",
			},
			{
				"name":        "Информация",
				"description": "Общая информация о сервере и API.",
			},
			{
				"name":        "Аутентификация",
				"description": "Регистрация, авторизация и обновление JWT-токенов. Публичные эндпоинты, не требующие авторизации.",
			},
			{
				"name":        "Профиль",
				"description": "Управление профилем текущего авторизованного пользователя. Все эндпоинты требуют JWT-аутентификации.",
			},
			{
				"name":        "Локации",
				"description": "CRUD-операции для туристических локаций с пространственными запросами PostGIS. Поддерживает поиск по радиусу (ST_DWithin), по bounding box (ST_Within), фильтрацию по категории, child_friendly и density_level.",
			},
			{
				"name":        "Поездки",
				"description": "Модуль планирования поездок с механикой invite-ссылок для групп. Создание поездки, присоединение участников, генерация invite-токенов, Group Vibe Merge.",
			},
			{
				"name":        "Профилирование",
				"description": "Мультимодальный пайплайн профилирования туриста. Голосовой ввод (Vosk STT), извлечение осей предпочтений (LLM), генерация vibe-вектора (Embeddings), свайп-анкета, vibe passport и поиск рекомендаций.",
			},
			{
				"name":        "Карта",
				"description": "Эндпоинты для работы с интерактивной картой Краснодарского края. Пространственный поиск локаций по bounding box через PostGIS.",
			},
			{
				"name":        "Маршруты",
				"description": "Построение маршрутов между локациями с расчётом расстояния (Haversine) и времени в пути. Demo-режим без внешних API.",
			},
		},
		"paths": map[string]any{
			"/": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Информация"},
					"summary":     "Информация о сервере",
					"description": "Возвращает базовую информацию о сервере КудыТуды API: имя сервиса, версию и текущий статус работы. Используется для быстрой проверки доступности сервера.",
					"operationId": "getServerInfo",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Сервер работает и доступен для обработки запросов.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"$ref": "#/components/schemas/ServerInfo",
									},
									"example": map[string]any{
										"service": "КудыТуды API",
										"version": "1.0.0",
										"status":  "running",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/health": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Мониторинг"},
					"summary":     "Проверка здоровья всех сервисов",
					"description": "Выполняет асинхронный пинг всех 6 баз данных (PostgreSQL, Redis, Qdrant, Neo4j, ClickHouse, MinIO) и возвращает детальный отчёт о состоянии каждого сервиса.\n\nКаждая проверка выполняется в отдельной горутине с таймаутом 5 секунд. Результат включает статус сервиса (\"up\" или \"down\"), время отклика в миллисекундах и описание ошибки при недоступности.\n\n**Общий статус системы:**\n- `healthy` -- все 6 сервисов доступны и отвечают в пределах таймаута.\n- `degraded` -- хотя бы один сервис недоступен или не отвечает.\n\n**HTTP статус-коды:**\n- `200 OK` -- все сервисы здоровы.\n- `503 Service Unavailable` -- есть недоступные сервисы.\n\nПредназначен для использования в системах мониторинга (Prometheus, Grafana), балансировщиках нагрузки (health check) и Kubernetes liveness/readiness probes.",
					"operationId": "healthCheck",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Все сервисы работают корректно. Каждый из 6 сервисов баз данных доступен и отвечает на пинг-запросы.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"$ref": "#/components/schemas/HealthResponse",
									},
									"example": map[string]any{
										"status":    "healthy",
										"timestamp": "2026-03-19T20:37:07Z",
										"services": map[string]any{
											"postgres": map[string]any{
												"status":     "up",
												"latency_ms": 2,
											},
											"redis": map[string]any{
												"status":     "up",
												"latency_ms": 1,
											},
											"qdrant": map[string]any{
												"status":     "up",
												"latency_ms": 3,
											},
											"neo4j": map[string]any{
												"status":     "up",
												"latency_ms": 5,
											},
											"clickhouse": map[string]any{
												"status":     "up",
												"latency_ms": 2,
											},
											"minio": map[string]any{
												"status":     "up",
												"latency_ms": 1,
											},
										},
									},
								},
							},
						},
						"503": map[string]any{
							"description": "Один или несколько сервисов недоступны. Система работает в деградированном режиме -- часть функциональности может быть ограничена.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"$ref": "#/components/schemas/HealthResponse",
									},
									"example": map[string]any{
										"status":    "degraded",
										"timestamp": "2026-03-19T20:37:07Z",
										"services": map[string]any{
											"postgres": map[string]any{
												"status":     "up",
												"latency_ms": 2,
											},
											"redis": map[string]any{
												"status":     "down",
												"latency_ms": 0,
												"error":      "dial tcp 127.0.0.1:6379: connect: connection refused",
											},
											"qdrant": map[string]any{
												"status":     "up",
												"latency_ms": 3,
											},
											"neo4j": map[string]any{
												"status":     "down",
												"latency_ms": 0,
												"error":      "connection reset by peer",
											},
											"clickhouse": map[string]any{
												"status":     "up",
												"latency_ms": 2,
											},
											"minio": map[string]any{
												"status":     "up",
												"latency_ms": 1,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/health/live": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Мониторинг"},
					"summary":     "Liveness probe",
					"description": "Мгновенный ответ без пинга внешних сервисов. Проверяет только что процесс API жив.\n\nИспользуется как Kubernetes liveness probe. Не создаёт нагрузку на внешние сервисы.",
					"operationId": "healthLive",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Процесс жив и обрабатывает запросы.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"status":    map[string]any{"type": "string", "example": "alive"},
											"timestamp": map[string]any{"type": "string", "format": "date-time"},
										},
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/health/ready": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Мониторинг"},
					"summary":     "Readiness probe -- полная проверка всех БД",
					"description": "Выполняет асинхронный пинг всех 6 баз данных. Аналогичен GET /api/v1/health.\n\nИспользуется как Kubernetes readiness probe.",
					"operationId": "healthReady",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Все сервисы доступны.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"$ref": "#/components/schemas/HealthResponse",
									},
								},
							},
						},
						"503": map[string]any{
							"description": "Один или несколько сервисов недоступны.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"$ref": "#/components/schemas/HealthResponse",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/media/upload": map[string]any{
				"post": map[string]any{
					"tags":        []string{"Медиа"},
					"summary":     "Загрузка медиафайла",
					"description": "Загружает медиафайл в S3-совместимое хранилище MinIO.\n\nФайл передаётся через multipart/form-data в поле `file`. Перед загрузкой выполняется валидация:\n\n1. **Наличие файла** -- поле `file` обязательно.\n2. **Размер** -- максимум 100 МБ (104 857 600 байт).\n3. **MIME-тип** -- разрешены только определённые типы файлов.\n\n**Разрешённые MIME-типы:**\n- Изображения: `image/jpeg`, `image/png`, `image/gif`, `image/webp`\n- Видео: `video/mp4`, `video/webm`, `video/quicktime`\n- Аудио: `audio/mpeg`, `audio/mp3`, `audio/ogg`, `audio/webm`, `audio/wav`\n- Бинарные: `application/octet-stream` (для .splat и других 3D-файлов)\n\nПосле успешной загрузки генерируется уникальное имя объекта в формате `uploads/{unix_nano}_{original_name}` и создаётся presigned URL с временем жизни 24 часа для доступа к файлу.\n\n**Примеры использования:**\n- Загрузка фото локации фермером при онбординге.\n- Загрузка видео для генерации 3D-слепка (Gaussian Splatting).\n- Загрузка готового .splat файла.\n- Загрузка аудио-истории для POI.",
					"operationId": "uploadMedia",
					"requestBody": map[string]any{
						"required":    true,
						"description": "Multipart-форма с загружаемым файлом. Поле `file` является обязательным.",
						"content": map[string]any{
							"multipart/form-data": map[string]any{
								"schema": map[string]any{
									"type":     "object",
									"required": []string{"file"},
									"properties": map[string]any{
										"file": map[string]any{
											"type":        "string",
											"format":      "binary",
											"description": "Загружаемый файл. Максимальный размер -- 100 МБ. Допустимые форматы: JPEG, PNG, GIF, WebP, MP4, WebM, MOV, MP3, OGG, WAV, бинарные файлы (.splat).",
										},
									},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Файл успешно загружен в хранилище. Ответ содержит имя объекта, бакет, размер и временный публичный URL для доступа (действителен 24 часа).",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"$ref": "#/components/schemas/UploadResponse",
									},
									"example": map[string]any{
										"success": true,
										"message": "файл успешно загружен",
										"data": map[string]any{
											"object_name": "uploads/1710873427000000_vineyard_photo.jpg",
											"bucket":      "deepkrai-media",
											"size":        245760,
											"url":         "http://localhost:9000/deepkrai-media/uploads/1710873427000000_vineyard_photo.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&...",
										},
									},
								},
							},
						},
						"400": map[string]any{
							"description": "Ошибка валидации запроса. Возможные причины:\n- Файл отсутствует в запросе (поле `file` не найдено).\n- Размер файла превышает 100 МБ.\n- MIME-тип файла не входит в список разрешённых.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"$ref": "#/components/schemas/UploadResponse",
									},
									"examples": map[string]any{
										"file_missing": map[string]any{
											"summary":     "Файл не найден",
											"description": "В запросе отсутствует поле `file` с загружаемым файлом.",
											"value": map[string]any{
												"success": false,
												"message": "файл не найден в запросе, используйте поле 'file'",
											},
										},
										"file_too_large": map[string]any{
											"summary":     "Файл слишком большой",
											"description": "Размер файла превышает максимально допустимый размер в 100 МБ.",
											"value": map[string]any{
												"success": false,
												"message": "размер файла (157286400 байт) превышает максимально допустимый (104857600 байт)",
											},
										},
										"invalid_mime": map[string]any{
											"summary":     "Неподдерживаемый тип",
											"description": "MIME-тип файла не входит в список разрешённых типов.",
											"value": map[string]any{
												"success": false,
												"message": "неподдерживаемый тип файла: text/html",
											},
										},
									},
								},
							},
						},
						"500": map[string]any{
							"description": "Внутренняя ошибка сервера при загрузке файла. Возможные причины: MinIO недоступен, ошибка записи, нехватка места.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"$ref": "#/components/schemas/UploadResponse",
									},
									"example": map[string]any{
										"success": false,
										"message": "ошибка загрузки файла в хранилище",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/auth/register": map[string]any{
				"post": map[string]any{
					"tags":        []string{"Аутентификация"},
					"summary":     "Регистрация нового пользователя",
					"description": "Создаёт нового пользователя с ролью tourist. Пароль хэшируется через bcrypt (cost=12). Возвращает пару JWT-токенов (access + refresh) и данные профиля.",
					"operationId": "registerUser",
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/RegisterRequest"},
							},
						},
					},
					"responses": map[string]any{
						"201": map[string]any{"description": "Пользователь зарегистрирован.", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/AuthResponse"}}}},
						"400": map[string]any{"description": "Ошибка валидации (некорректный email, короткий пароль)."},
						"409": map[string]any{"description": "Пользователь с таким email уже существует."},
					},
				},
			},
			"/api/v1/auth/login": map[string]any{
				"post": map[string]any{
					"tags":        []string{"Аутентификация"},
					"summary":     "Авторизация пользователя",
					"description": "Проверяет email и пароль. При успехе возвращает пару JWT-токенов и данные профиля.",
					"operationId": "loginUser",
					"requestBody": map[string]any{
						"required": true,
						"content":  map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/LoginRequest"}}},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Авторизация успешна.", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/AuthResponse"}}}},
						"401": map[string]any{"description": "Неверный email или пароль."},
					},
				},
			},
			"/api/v1/auth/refresh": map[string]any{
				"post": map[string]any{
					"tags":        []string{"Аутентификация"},
					"summary":     "Обновление токенов",
					"description": "Принимает refresh-токен и возвращает новую пару access + refresh токенов.",
					"operationId": "refreshTokens",
					"requestBody": map[string]any{
						"required": true,
						"content":  map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/RefreshRequest"}}},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Токены обновлены."},
						"401": map[string]any{"description": "Невалидный или истекший refresh-токен."},
					},
				},
			},
			"/api/v1/profile/me": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Профиль"},
					"summary":     "Профиль текущего пользователя",
					"description": "Возвращает данные авторизованного пользователя (без хэша пароля). Требует JWT access-токен.",
					"operationId": "getProfile",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"responses": map[string]any{
						"200": map[string]any{"description": "Данные профиля.", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/UserProfile"}}}},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
					},
				},
				"put": map[string]any{
					"tags":        []string{"Профиль"},
					"summary":     "Обновление профиля",
					"description": "Обновляет display_name текущего авторизованного пользователя. Требует JWT access-токен.",
					"operationId": "updateProfile",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"display_name": map[string]any{
											"type":        "string",
											"description": "Новое отображаемое имя пользователя.",
											"example":     "Иван Путешественник",
										},
									},
									"required": []string{"display_name"},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Профиль обновлён.", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/UserProfile"}}}},
						"400": map[string]any{"description": "Некорректный запрос."},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
					},
				},
			},
			"/api/v1/locations": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Локации"},
					"summary":     "Поиск локаций с пространственными фильтрами",
					"description": "Выполняет пространственный поиск локаций с поддержкой PostGIS. Поддерживает поиск по радиусу (ST_DWithin), по bounding box (ST_Within + ST_MakeEnvelope), фильтрацию по категории, child_friendly, density_level. Результаты пагинируются.",
					"operationId": "searchLocations",
					"parameters": []map[string]any{
						{"name": "lat", "in": "query", "schema": map[string]any{"type": "number", "format": "double"}, "description": "Широта центра поиска (для радиуса). Используется с lon и radius_km.", "example": 45.03},
						{"name": "lon", "in": "query", "schema": map[string]any{"type": "number", "format": "double"}, "description": "Долгота центра поиска (для радиуса). Используется с lat и radius_km.", "example": 38.97},
						{"name": "radius_km", "in": "query", "schema": map[string]any{"type": "number", "format": "double"}, "description": "Радиус поиска в километрах. Требует lat и lon.", "example": 50},
						{"name": "min_lat", "in": "query", "schema": map[string]any{"type": "number", "format": "double"}, "description": "Минимальная широта bbox. Используется с max_lat, min_lon, max_lon.", "example": 43.5},
						{"name": "max_lat", "in": "query", "schema": map[string]any{"type": "number", "format": "double"}, "description": "Максимальная широта bbox.", "example": 46.0},
						{"name": "min_lon", "in": "query", "schema": map[string]any{"type": "number", "format": "double"}, "description": "Минимальная долгота bbox.", "example": 36.5},
						{"name": "max_lon", "in": "query", "schema": map[string]any{"type": "number", "format": "double"}, "description": "Максимальная долгота bbox.", "example": 41.0},
						{"name": "category", "in": "query", "schema": map[string]any{"type": "string"}, "description": "Фильтр по категории (свободный текст: winery, farm, trail, gastro, nature, camping, resort, extreme, cultural, beach и др.).", "example": "winery"},
						{"name": "child_friendly", "in": "query", "schema": map[string]any{"type": "boolean"}, "description": "Фильтр по пригодности для детей.", "example": true},
						{"name": "density_level", "in": "query", "schema": map[string]any{"type": "string", "enum": []string{"red", "yellow", "green"}}, "description": "Фильтр по уровню туристической плотности. red — высокая, yellow — сезонная, green — Hidden Gem.", "example": "green"},
						{"name": "page", "in": "query", "schema": map[string]any{"type": "integer", "default": 1}, "description": "Номер страницы (начиная с 1).", "example": 1},
						{"name": "per_page", "in": "query", "schema": map[string]any{"type": "integer", "default": 20, "maximum": 100}, "description": "Количество записей на странице (макс. 100).", "example": 20},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Список локаций с пагинацией.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"success": map[string]any{"type": "boolean", "example": true},
											"data":    map[string]any{"$ref": "#/components/schemas/LocationListResponse"},
										},
									},
								},
							},
						},
						"500": map[string]any{"description": "Внутренняя ошибка сервера."},
					},
				},
				"post": map[string]any{
					"tags":        []string{"Локации"},
					"summary":     "Создание новой локации",
					"description": "Создаёт новую туристическую локацию. Доступно только пользователям с ролью host или b2g_admin. Автоматически генерирует slug из названия (транслитерация кириллицы).",
					"operationId": "createLocation",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/CreateLocationRequest"},
							},
						},
					},
					"responses": map[string]any{
						"201": map[string]any{
							"description": "Локация создана.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"success": map[string]any{"type": "boolean", "example": true},
											"message": map[string]any{"type": "string", "example": "локация успешно создана"},
											"data":    map[string]any{"$ref": "#/components/schemas/Location"},
										},
									},
								},
							},
						},
						"400": map[string]any{"description": "Ошибка валидации (пустое имя, координаты вне диапазона)."},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
						"403": map[string]any{"description": "Недостаточно прав (требуется роль host или b2g_admin)."},
						"409": map[string]any{"description": "Локация с таким slug уже существует."},
					},
				},
			},
			"/api/v1/locations/{id}": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Локации"},
					"summary":     "Получение локации по ID",
					"description": "Возвращает полные данные локации по UUID. Публичный эндпоинт, не требует авторизации.\n\nОтвет содержит все поля для экрана детали локации: title, category, описания, теги, цена, координаты, hero image, галерея, 3D preview ссылка.",
					"operationId": "getLocationByID",
					"parameters": []map[string]any{
						{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "string", "format": "uuid"}, "description": "UUID локации."},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Данные локации.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"success": map[string]any{"type": "boolean", "example": true},
											"data":    map[string]any{"$ref": "#/components/schemas/Location"},
										},
									},
								},
							},
						},
						"404": map[string]any{"description": "Локация не найдена."},
					},
				},
				"put": map[string]any{
					"tags":        []string{"Локации"},
					"summary":     "Обновление локации",
					"description": "Обновляет данные локации. Доступно только владельцу (owner_id). Все поля опциональны — обновляются только переданные.",
					"operationId": "updateLocation",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"parameters": []map[string]any{
						{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "string", "format": "uuid"}, "description": "UUID локации."},
					},
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/UpdateLocationRequest"},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Локация обновлена.", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{"success": map[string]any{"type": "boolean"}, "data": map[string]any{"$ref": "#/components/schemas/Location"}}}}}},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
						"403": map[string]any{"description": "Нет прав на изменение этой локации (не владелец)."},
						"404": map[string]any{"description": "Локация не найдена."},
					},
				},
				"delete": map[string]any{
					"tags":        []string{"Локации"},
					"summary":     "Удаление локации",
					"description": "Удаляет локацию. Доступно только владельцу (owner_id).",
					"operationId": "deleteLocation",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"parameters": []map[string]any{
						{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "string", "format": "uuid"}, "description": "UUID локации."},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Локация удалена."},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
						"403": map[string]any{"description": "Нет прав на удаление этой локации (не владелец)."},
						"404": map[string]any{"description": "Локация не найдена."},
					},
				},
			},
			"/api/v1/locations/{id}/splat": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Локации"},
					"summary":     "3D-сцена локации (Gaussian Splatting)",
					"description": "Возвращает URL 3D-сцены (.splat) для конкретной локации. Отдельный эндпоинт позволяет фронтенду загружать тяжёлый 3D-контент лениво, не включая его в основной payload location detail.",
					"operationId": "getLocationSplat",
					"parameters": []map[string]any{
						{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "string", "format": "uuid"}, "description": "UUID локации."},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Данные 3D-сцены.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"success": map[string]any{"type": "boolean"},
											"data":    map[string]any{"$ref": "#/components/schemas/SplatResponse"},
										},
									},
								},
							},
						},
						"404": map[string]any{"description": "Локация не найдена."},
					},
				},
			},
			"/api/v1/trips": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Поездки"},
					"summary":     "Список поездок текущего пользователя",
					"description": "Возвращает все поездки, в которых текущий пользователь является участником (creator или member). Для каждой поездки включается полный список участников.",
					"operationId": "listUserTrips",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Список поездок пользователя.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"data": map[string]any{
												"type":  "array",
												"items": map[string]any{"$ref": "#/components/schemas/TripWithMembers"},
											},
											"count": map[string]any{"type": "integer", "example": 2},
										},
									},
								},
							},
						},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
					},
				},
				"post": map[string]any{
					"tags":        []string{"Поездки"},
					"summary":     "Создание поездки",
					"description": "Создаёт новую поездку с датами, бюджетом, транспортом и составом группы. Создатель автоматически добавляется как первый участник с ролью creator. Генерируется invite_token для приглашения участников.",
					"operationId": "createTrip",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/CreateTripRequest"},
							},
						},
					},
					"responses": map[string]any{
						"201": map[string]any{"description": "Поездка создана. Возвращает объект поездки с invite_token и списком участников."},
						"400": map[string]any{"description": "Ошибка валидации (некорректные даты, бюджет, транспорт)."},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
					},
				},
			},
			"/api/v1/trips/{id}": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Поездки"},
					"summary":     "Детали поездки",
					"description": "Возвращает полную информацию о поездке со списком всех участников.",
					"operationId": "getTripByID",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"parameters": []map[string]any{
						{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "string", "format": "uuid"}, "description": "UUID поездки."},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Объект поездки с участниками."},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
						"404": map[string]any{"description": "Поездка не найдена."},
					},
				},
			},
			"/api/v1/trips/{id}/invite": map[string]any{
				"post": map[string]any{
					"tags":        []string{"Поездки"},
					"summary":     "Генерация invite-токена",
					"description": "Генерирует новый invite-токен для поездки. Старый токен становится недействительным. Доступно только создателю поездки.",
					"operationId": "generateTripInvite",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"parameters": []map[string]any{
						{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "string", "format": "uuid"}, "description": "UUID поездки."},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Новый invite-токен и ссылка для приглашения."},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
						"403": map[string]any{"description": "Нет прав (не создатель поездки)."},
						"404": map[string]any{"description": "Поездка не найдена."},
					},
				},
			},
			"/api/v1/trips/{id}/join": map[string]any{
				"post": map[string]any{
					"tags":        []string{"Поездки"},
					"summary":     "Присоединение к поездке",
					"description": "Присоединяет участника к поездке по invite-токену. Эндпоинт доступен БЕЗ авторизации — неавторизованные пользователи передают display_name и теги напрямую. Проверяет валидность токена и ограничение group_size.",
					"operationId": "joinTrip",
					"parameters": []map[string]any{
						{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "string", "format": "uuid"}, "description": "UUID поездки."},
					},
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/JoinTripRequest"},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Участник присоединился. Возвращает объект TripMember."},
						"400": map[string]any{"description": "Ошибка валидации или неверный invite-токен."},
						"404": map[string]any{"description": "Поездка не найдена."},
						"409": map[string]any{"description": "Превышен лимит участников или пользователь уже участвует."},
					},
				},
			},
			"/api/v1/trips/{id}/members": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Поездки"},
					"summary":     "Список участников поездки",
					"description": "Возвращает всех участников поездки, отсортированных по дате присоединения.",
					"operationId": "listTripMembers",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"parameters": []map[string]any{
						{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "string", "format": "uuid"}, "description": "UUID поездки."},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Массив участников поездки."},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
						"404": map[string]any{"description": "Поездка не найдена."},
					},
				},
			},
			"/api/v1/profile/voice": map[string]any{
				"post": map[string]any{
					"tags":        []string{"Профилирование"},
					"summary":     "Голосовое профилирование",
					"description": "Загрузка аудиозаписи для построения vibe-профиля. Пайплайн: Vosk STT -> LLM (оси) -> Embeddings (3072d) -> Qdrant upsert.\n\nВозвращает screenshot-ready vibe passport: вложенные оси, заголовок паспорта, теги и время обработки.\n\nВ mock-режиме (AI_API_KEY пуст) ответ детерминированный с controlled latency 2 секунды.",
					"operationId": "voiceProfile",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"multipart/form-data": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"audio": map[string]any{"type": "string", "format": "binary", "description": "Аудиофайл (mp3, wav, webm, ogg)."},
									},
									"required": []string{"audio"},
								},
							},
						},
					},
					"responses": map[string]any{
						"201": map[string]any{
							"description": "Профиль создан. Возвращает vibe passport с осями, тегами, summary, заголовком и vector_id.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"success": map[string]any{"type": "boolean", "example": true},
											"data":    map[string]any{"$ref": "#/components/schemas/VoiceProfileResponse"},
										},
									},
								},
							},
						},
						"400": map[string]any{"description": "Аудиофайл не передан."},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
						"500": map[string]any{"description": "Аудио не распознано, ошибка STT/LLM/Embeddings."},
					},
				},
			},
			"/api/v1/profile/swipe": map[string]any{
				"post": map[string]any{
					"tags":        []string{"Профилирование"},
					"summary":     "Свайп сцены",
					"description": "Сдвигает vibe-вектор пользователя к (right) или от (left) вектора сцены. Формула: new = normalize(0.85*user ± 0.15*scene).",
					"operationId": "swipeScene",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/SwipeRequest"},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Вектор обновлён."},
						"400": map[string]any{"description": "Невалидный scene_id или direction."},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
					},
				},
			},
			"/api/v1/profile/finalize": map[string]any{
				"post": map[string]any{
					"tags":        []string{"Профилирование"},
					"summary":     "Финализация профиля и получение рекомендаций",
					"description": "Берёт vibe-вектор пользователя из Qdrant и ищет Top-N ближайших локаций по cosine similarity. При отсутствии вектора возвращает curated demo набор (is_curated=true). Тело запроса опционально.",
					"operationId": "finalizeProfile",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"requestBody": map[string]any{
						"required":    false,
						"description": "Опциональные параметры финализации.",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/FinalizeRequest"},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Top-N рекомендаций с обогащёнными данными.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"success": map[string]any{"type": "boolean"},
											"data":    map[string]any{"$ref": "#/components/schemas/FinalizeResponse"},
										},
									},
								},
							},
						},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
					},
				},
			},
			"/api/v1/profile/scenes": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Профилирование"},
					"summary":     "Список сцен для свайпа",
					"description": "Возвращает все сцены свайп-анкеты, отсортированные по display_order.",
					"operationId": "getSwipeScenes",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"responses": map[string]any{
						"200": map[string]any{"description": "Массив сцен с id, title, description, image_url, display_order."},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
					},
				},
			},
			"/api/v1/map/locations": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Карта"},
					"summary":     "Локации для карты",
					"description": "Возвращает локации для отображения на интерактивной карте. Поддерживает три режима:\n\n1. **bbox**: пространственный поиск по bounding box через PostGIS.\n2. **demo**: curated набор с предустановленными рекомендациями для 4 профилей.\n3. **hybrid**: обогащение точек скорами рекомендаций из Qdrant (при наличии JWT).\n\nВсе координаты в формате WGS84 (EPSG:4326). По умолчанию возвращает до 50 локаций.",
					"operationId": "getMapLocations",
					"parameters": []map[string]any{
						{"name": "min_lat", "in": "query", "schema": map[string]any{"type": "number"}, "description": "Минимальная широта bounding box.", "example": 43.5},
						{"name": "max_lat", "in": "query", "schema": map[string]any{"type": "number"}, "description": "Максимальная широта bounding box.", "example": 45.5},
						{"name": "min_lon", "in": "query", "schema": map[string]any{"type": "number"}, "description": "Минимальная долгота bounding box.", "example": 36.5},
						{"name": "max_lon", "in": "query", "schema": map[string]any{"type": "number"}, "description": "Максимальная долгота bounding box.", "example": 41.0},
						{"name": "category", "in": "query", "schema": map[string]any{"type": "string"}, "description": "Фильтр по категории.", "example": "winery"},
						{"name": "density_level", "in": "query", "schema": map[string]any{"type": "string", "enum": []string{"red", "yellow", "green"}}, "description": "Фильтр по плотности."},
						{"name": "is_recommended", "in": "query", "schema": map[string]any{"type": "boolean"}, "description": "Только рекомендованные точки."},
						{"name": "limit", "in": "query", "schema": map[string]any{"type": "integer", "default": 50}, "description": "Максимальное количество точек (макс. 200)."},
						{"name": "demo", "in": "query", "schema": map[string]any{"type": "boolean"}, "description": "Включить demo-режим с curated рекомендациями."},
						{"name": "profile", "in": "query", "schema": map[string]any{"type": "string", "enum": []string{"calm_wine_mountains", "active_adventure", "family_kids", "gastro_cultural"}}, "description": "Demo-профиль. По умолчанию calm_wine_mountains."},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Массив точек для карты.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"success": map[string]any{"type": "boolean"},
											"data": map[string]any{
												"type": "object",
												"properties": map[string]any{
													"points":  map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/MapPoint"}},
													"total":   map[string]any{"type": "integer"},
													"profile": map[string]any{"type": "string", "description": "Название demo-профиля (только в demo-режиме)."},
												},
											},
										},
									},
								},
							},
						},
						"400": map[string]any{"description": "Некорректные параметры (невалидный bbox или профиль)."},
					},
				},
			},
			"/api/v1/route/build": map[string]any{
				"post": map[string]any{
					"tags":        []string{"Маршруты"},
					"summary":     "Построение маршрута",
					"description": "Строит маршрут через указанные локации с расчётом расстояния (Haversine) и предполагаемого времени в пути.\n\nDemo-режим: используются прямые расстояния и настраиваемые скорости по типу транспорта. Полная маршрутизация через Neo4j запланирована в будущих фазах.",
					"operationId": "buildRoute",
					"security":    []map[string]any{{"BearerAuth": []string{}}},
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/BuildRouteRequest"},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Маршрут построен.",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"success": map[string]any{"type": "boolean"},
											"data":    map[string]any{"$ref": "#/components/schemas/RoutePreview"},
										},
									},
								},
							},
						},
						"400": map[string]any{"description": "Менее 2 location_ids или невалидные UUID."},
						"401": map[string]any{"description": "Отсутствует или невалидный токен."},
					},
				},
			},
		},
		"components": map[string]any{
			"schemas": map[string]any{
				"ServerInfo": map[string]any{
					"type":        "object",
					"description": "Базовая информация о сервере КудыТуды API.",
					"properties": map[string]any{
						"service": map[string]any{
							"type":        "string",
							"description": "Название сервиса.",
							"example":     "КудыТуды API",
						},
						"version": map[string]any{
							"type":        "string",
							"description": "Текущая версия API.",
							"example":     "1.0.0",
						},
						"status": map[string]any{
							"type":        "string",
							"description": "Текущий статус сервера.",
							"enum":        []string{"running"},
							"example":     "running",
						},
					},
					"required": []string{"service", "version", "status"},
				},
				"HealthResponse": map[string]any{
					"type":        "object",
					"description": "Результат проверки здоровья системы. Содержит общий статус, временную метку и детальную информацию по каждому из 6 сервисов баз данных.",
					"properties": map[string]any{
						"status": map[string]any{
							"type":        "string",
							"description": "Общий статус системы. `healthy` -- все сервисы работают, `degraded` -- есть недоступные сервисы.",
							"enum":        []string{"healthy", "degraded"},
							"example":     "healthy",
						},
						"timestamp": map[string]any{
							"type":        "string",
							"format":      "date-time",
							"description": "Временная метка выполнения проверки в формате RFC 3339 (UTC).",
							"example":     "2026-03-19T20:37:07Z",
						},
						"services": map[string]any{
							"type":        "object",
							"description": "Карта статусов каждого сервиса баз данных. Ключ -- имя сервиса (postgres, redis, qdrant, neo4j, clickhouse, minio), значение -- объект с деталями проверки.",
							"additionalProperties": map[string]any{
								"$ref": "#/components/schemas/ServiceHealth",
							},
						},
					},
					"required": []string{"status", "timestamp", "services"},
				},
				"ServiceHealth": map[string]any{
					"type":        "object",
					"description": "Результат проверки здоровья одного сервиса базы данных.",
					"properties": map[string]any{
						"status": map[string]any{
							"type":        "string",
							"description": "Текущий статус сервиса. `up` -- сервис доступен и отвечает, `down` -- сервис недоступен.",
							"enum":        []string{"up", "down"},
							"example":     "up",
						},
						"latency_ms": map[string]any{
							"type":        "integer",
							"format":      "int64",
							"description": "Время отклика сервиса в миллисекундах. При статусе `down` может быть 0 (таймаут).",
							"example":     2,
							"minimum":     0,
						},
						"error": map[string]any{
							"type":        "string",
							"description": "Описание ошибки при недоступности сервиса. Пустая строка при статусе `up`.",
							"example":     "",
						},
					},
					"required": []string{"status", "latency_ms"},
				},
				"UploadResponse": map[string]any{
					"type":        "object",
					"description": "Ответ на запрос загрузки медиафайла в хранилище MinIO.",
					"properties": map[string]any{
						"success": map[string]any{
							"type":        "boolean",
							"description": "Флаг успешности операции. `true` -- файл загружен, `false` -- произошла ошибка.",
							"example":     true,
						},
						"message": map[string]any{
							"type":        "string",
							"description": "Человекочитаемое описание результата операции.",
							"example":     "файл успешно загружен",
						},
						"data": map[string]any{
							"description": "Данные о загруженном файле. Присутствует только при успешной загрузке (success=true).",
							"$ref":        "#/components/schemas/UploadResult",
						},
					},
					"required": []string{"success", "message"},
				},
				"UploadResult": map[string]any{
					"type":        "object",
					"description": "Данные о загруженном файле в S3-хранилище.",
					"properties": map[string]any{
						"object_name": map[string]any{
							"type":        "string",
							"description": "Полное имя объекта в бакете MinIO, включая путь. Формат: `uploads/{unix_nano}_{original_name}`.",
							"example":     "uploads/1710873427000000_vineyard_photo.jpg",
						},
						"bucket": map[string]any{
							"type":        "string",
							"description": "Имя бакета MinIO, в который загружен файл.",
							"example":     "deepkrai-media",
						},
						"size": map[string]any{
							"type":        "integer",
							"format":      "int64",
							"description": "Размер загруженного файла в байтах.",
							"example":     245760,
							"minimum":     0,
						},
						"url": map[string]any{
							"type":        "string",
							"format":      "uri",
							"description": "Presigned URL для доступа к загруженному файлу. URL действителен 24 часа с момента генерации.",
							"example":     "http://localhost:9000/deepkrai-media/uploads/1710873427000000_vineyard_photo.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
						},
					},
					"required": []string{"object_name", "bucket", "size", "url"},
				},
				"ErrorResponse": map[string]any{
					"type":        "object",
					"description": "Стандартный ответ при внутренней ошибке сервера. Возвращается middleware Recovery при перехвате паники.",
					"properties": map[string]any{
						"error": map[string]any{
							"type":        "string",
							"description": "Тип ошибки.",
							"example":     "Internal Server Error",
						},
						"message": map[string]any{
							"type":        "string",
							"description": "Описание произошедшей ошибки.",
							"example":     "непредвиденная ошибка: runtime error",
						},
					},
					"required": []string{"error", "message"},
				},
				"RegisterRequest": map[string]any{
					"type":        "object",
					"description": "Запрос регистрации нового пользователя.",
					"properties": map[string]any{
						"email":        map[string]any{"type": "string", "format": "email", "description": "Email пользователя.", "example": "tourist@deepkrai.ru"},
						"password":     map[string]any{"type": "string", "minLength": 8, "description": "Пароль (минимум 8 символов).", "example": "securePass123"},
						"display_name": map[string]any{"type": "string", "description": "Отображаемое имя (необязательно).", "example": "Иван"},
						"role":         map[string]any{"type": "string", "enum": []string{"tourist", "host"}, "default": "tourist", "description": "Роль пользователя. tourist — турист (по умолчанию), host — хост (владелец локаций).", "example": "tourist"},
					},
					"required": []string{"email", "password"},
				},
				"LoginRequest": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"email":    map[string]any{"type": "string", "format": "email", "example": "tourist@deepkrai.ru"},
						"password": map[string]any{"type": "string", "example": "securePass123"},
					},
					"required": []string{"email", "password"},
				},
				"RefreshRequest": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"refresh_token": map[string]any{"type": "string", "description": "Действующий refresh-токен."},
					},
					"required": []string{"refresh_token"},
				},
				"AuthResponse": map[string]any{
					"type":        "object",
					"description": "Ответ аутентификации с парой JWT-токенов и данными пользователя.",
					"properties": map[string]any{
						"success": map[string]any{"type": "boolean", "example": true},
						"message": map[string]any{"type": "string", "example": "авторизация успешна"},
						"data": map[string]any{"type": "object", "properties": map[string]any{
							"tokens": map[string]any{"$ref": "#/components/schemas/TokenPair"},
							"user":   map[string]any{"$ref": "#/components/schemas/UserProfile"},
						}},
					},
				},
				"TokenPair": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"access_token":             map[string]any{"type": "string", "description": "JWT access-токен (15 мин)."},
						"refresh_token":            map[string]any{"type": "string", "description": "JWT refresh-токен (7 дней)."},
						"access_token_expires_at":  map[string]any{"type": "integer", "format": "int64", "description": "Unix timestamp истечения access."},
						"refresh_token_expires_at": map[string]any{"type": "integer", "format": "int64", "description": "Unix timestamp истечения refresh."},
					},
				},
				"UserProfile": map[string]any{
					"type":        "object",
					"description": "Публичный профиль пользователя (без password_hash).",
					"properties": map[string]any{
						"id":             map[string]any{"type": "string", "format": "uuid"},
						"email":          map[string]any{"type": "string", "format": "email"},
						"role":           map[string]any{"type": "string", "enum": []string{"tourist", "host", "b2g_admin"}},
						"display_name":   map[string]any{"type": "string"},
						"karma":          map[string]any{"type": "integer"},
						"vibe_vector_id": map[string]any{"type": "string", "format": "uuid", "nullable": true},
						"created_at":     map[string]any{"type": "string", "format": "date-time"},
						"updated_at":     map[string]any{"type": "string", "format": "date-time"},
					},
				},
				"Location": map[string]any{
					"type":        "object",
					"description": "Туристическая локация Краснодарского края с пространственными координатами (PostGIS).",
					"properties": map[string]any{
						"id":                map[string]any{"type": "string", "format": "uuid", "description": "UUID локации."},
						"owner_id":          map[string]any{"type": "string", "format": "uuid", "description": "UUID владельца (хоста)."},
						"slug":              map[string]any{"type": "string", "description": "URL-дружественный идентификатор.", "example": "vinodelnya-abrau-dyurso"},
						"name":              map[string]any{"type": "string", "description": "Название локации.", "example": "Винодельня Абрау-Дюрсо"},
						"description_short": map[string]any{"type": "string", "description": "Краткое описание (для карточек)."},
						"description_full":  map[string]any{"type": "string", "description": "Полное описание (литературный текст)."},
						"category":          map[string]any{"type": "string", "description": "Категория (свободный текст: winery, farm, trail, gastro, nature, camping, resort, extreme, cultural, beach и др.).", "example": "winery"},
						"tags":              map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Массив тегов для фильтрации.", "example": []string{"вино", "дегустация", "озеро"}},
						"price_per_night":   map[string]any{"type": "integer", "description": "Стоимость за ночь (₽). 0 для бесплатных.", "example": 8000},
						"capacity":          map[string]any{"type": "integer", "description": "Максимальная вместимость.", "example": 20},
						"access_level":      map[string]any{"type": "string", "enum": []string{"open", "semi_open", "hidden"}, "description": "Уровень доступа (Hidden Gems). open — все, semi_open — зарегистрированные, hidden — по карме."},
						"density_level":     map[string]any{"type": "string", "enum": []string{"red", "yellow", "green"}, "description": "Плотность туристов. red — высокая, yellow — сезонная, green — Hidden Gem."},
						"child_friendly":    map[string]any{"type": "boolean", "description": "Подходит ли для детей."},
						"splat_url":         map[string]any{"type": "string", "nullable": true, "description": "URL на .splat файл (3D Gaussian Splatting)."},
						"preview_image_url": map[string]any{"type": "string", "description": "URL hero-изображения локации."},
						"gallery_urls":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Массив URL дополнительных фотографий для галереи."},
						"vibe_vector_id":    map[string]any{"type": "string", "format": "uuid", "nullable": true, "description": "ID вектора vibe-профиля в Qdrant."},
						"address":           map[string]any{"type": "string", "description": "Адрес (населённый пункт, район).", "example": "Краснодарский край, пос. Абрау-Дюрсо"},
						"is_published":      map[string]any{"type": "boolean", "description": "Опубликована ли локация."},
						"latitude":          map[string]any{"type": "number", "format": "double", "description": "Широта (WGS 84).", "example": 44.6979},
						"longitude":         map[string]any{"type": "number", "format": "double", "description": "Долгота (WGS 84).", "example": 37.5949},
						"created_at":        map[string]any{"type": "string", "format": "date-time"},
						"updated_at":        map[string]any{"type": "string", "format": "date-time"},
					},
				},
				"SplatResponse": map[string]any{
					"type":        "object",
					"description": "Ответ с данными 3D-сцены (Gaussian Splatting) для локации.",
					"properties": map[string]any{
						"location_id":   map[string]any{"type": "string", "format": "uuid", "description": "UUID локации."},
						"location_name": map[string]any{"type": "string", "description": "Название локации."},
						"has_splat":     map[string]any{"type": "boolean", "description": "Есть ли 3D-сцена."},
						"splat_url":     map[string]any{"type": "string", "nullable": true, "description": "URL на .splat файл. Присутствует только если has_splat=true."},
					},
				},
				"CreateLocationRequest": map[string]any{
					"type":        "object",
					"description": "Запрос на создание новой локации. Категория — свободный текст (не enum).",
					"properties": map[string]any{
						"name":              map[string]any{"type": "string", "description": "Название локации.", "example": "Козья ферма дяди Вани"},
						"description_short": map[string]any{"type": "string", "description": "Краткое описание."},
						"description_full":  map[string]any{"type": "string", "description": "Полное описание."},
						"category":          map[string]any{"type": "string", "description": "Категория (свободный текст).", "example": "farm"},
						"tags":              map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "example": []string{"сыр", "козы", "тишина"}},
						"price_per_night":   map[string]any{"type": "integer", "description": "Стоимость за ночь (₽).", "example": 5000},
						"capacity":          map[string]any{"type": "integer", "description": "Вместимость.", "example": 4},
						"access_level":      map[string]any{"type": "string", "enum": []string{"open", "semi_open", "hidden"}, "default": "open"},
						"density_level":     map[string]any{"type": "string", "enum": []string{"red", "yellow", "green"}, "default": "green"},
						"child_friendly":    map[string]any{"type": "boolean", "default": false},
						"address":           map[string]any{"type": "string", "example": "Краснодарский край, Хаджох"},
						"is_published":      map[string]any{"type": "boolean", "default": false},
						"latitude":          map[string]any{"type": "number", "format": "double", "description": "Широта [-90, 90].", "example": 44.2878},
						"longitude":         map[string]any{"type": "number", "format": "double", "description": "Долгота [-180, 180].", "example": 40.1763},
					},
					"required": []string{"name", "latitude", "longitude"},
				},
				"UpdateLocationRequest": map[string]any{
					"type":        "object",
					"description": "Запрос на обновление локации. Все поля опциональны — обновляются только переданные.",
					"properties": map[string]any{
						"name":              map[string]any{"type": "string"},
						"description_short": map[string]any{"type": "string"},
						"description_full":  map[string]any{"type": "string"},
						"category":          map[string]any{"type": "string"},
						"tags":              map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"price_per_night":   map[string]any{"type": "integer"},
						"capacity":          map[string]any{"type": "integer"},
						"access_level":      map[string]any{"type": "string", "enum": []string{"open", "semi_open", "hidden"}},
						"density_level":     map[string]any{"type": "string", "enum": []string{"red", "yellow", "green"}},
						"child_friendly":    map[string]any{"type": "boolean"},
						"address":           map[string]any{"type": "string"},
						"is_published":      map[string]any{"type": "boolean"},
						"latitude":          map[string]any{"type": "number", "format": "double"},
						"longitude":         map[string]any{"type": "number", "format": "double"},
					},
				},
				"LocationListResponse": map[string]any{
					"type":        "object",
					"description": "Результат поиска локаций с пагинацией.",
					"properties": map[string]any{
						"locations": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/Location"}, "description": "Массив найденных локаций."},
						"total":     map[string]any{"type": "integer", "description": "Общее количество записей, удовлетворяющих фильтрам.", "example": 20},
						"page":      map[string]any{"type": "integer", "description": "Текущая страница.", "example": 1},
						"per_page":  map[string]any{"type": "integer", "description": "Количество записей на странице.", "example": 20},
					},
				},
				"Trip": map[string]any{
					"type":        "object",
					"description": "Объект поездки с датами, бюджетом, транспортом и составом группы.",
					"properties": map[string]any{
						"id":                    map[string]any{"type": "string", "format": "uuid", "description": "Уникальный идентификатор поездки."},
						"creator_id":            map[string]any{"type": "string", "format": "uuid", "description": "UUID создателя поездки."},
						"date_from":             map[string]any{"type": "string", "format": "date", "description": "Дата начала.", "example": "2026-04-10"},
						"date_to":               map[string]any{"type": "string", "format": "date", "description": "Дата окончания.", "example": "2026-04-13"},
						"budget_rub":            map[string]any{"type": "integer", "description": "Бюджет в рублях.", "example": 50000},
						"budget_tier":           map[string]any{"type": "string", "enum": []string{"economy", "comfort", "premium"}, "description": "Уровень бюджета."},
						"transport":             map[string]any{"type": "string", "enum": []string{"car", "public", "walk", "bike"}, "description": "Вид транспорта."},
						"group_size":            map[string]any{"type": "integer", "description": "Планируемое количество участников.", "example": 4},
						"group_composition":     map[string]any{"type": "object", "description": "Состав группы (JSONB)."},
						"invite_token":          map[string]any{"type": "string", "format": "uuid", "description": "Токен для приглашения участников."},
						"merged_vibe_vector_id": map[string]any{"type": "string", "format": "uuid", "description": "ID средневзвешенного vibe-вектора группы.", "nullable": true},
						"status":                map[string]any{"type": "string", "enum": []string{"planning", "active", "completed", "cancelled"}, "description": "Статус поездки."},
						"created_at":            map[string]any{"type": "string", "format": "date-time"},
						"updated_at":            map[string]any{"type": "string", "format": "date-time"},
					},
				},
				"TripMember": map[string]any{
					"type":        "object",
					"description": "Участник поездки.",
					"properties": map[string]any{
						"id":             map[string]any{"type": "string", "format": "uuid"},
						"trip_id":        map[string]any{"type": "string", "format": "uuid"},
						"user_id":        map[string]any{"type": "string", "format": "uuid", "nullable": true, "description": "UUID пользователя (null для неавторизованных)."},
						"display_name":   map[string]any{"type": "string", "description": "Отображаемое имя.", "example": "Маша"},
						"role":           map[string]any{"type": "string", "enum": []string{"creator", "member"}, "description": "Роль в поездке."},
						"vibe_vector_id": map[string]any{"type": "string", "format": "uuid", "nullable": true},
						"tags":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Теги предпочтений.", "example": []string{"вино", "горы"}},
						"is_child":       map[string]any{"type": "boolean", "description": "Ребёнок."},
						"joined_at":      map[string]any{"type": "string", "format": "date-time"},
					},
				},
				"CreateTripRequest": map[string]any{
					"type":        "object",
					"description": "Запрос создания поездки.",
					"properties": map[string]any{
						"date_from":         map[string]any{"type": "string", "format": "date", "description": "Дата начала (YYYY-MM-DD).", "example": "2026-04-10"},
						"date_to":           map[string]any{"type": "string", "format": "date", "description": "Дата окончания (YYYY-MM-DD).", "example": "2026-04-13"},
						"budget_rub":        map[string]any{"type": "integer", "description": "Бюджет в рублях.", "example": 50000},
						"budget_tier":       map[string]any{"type": "string", "enum": []string{"economy", "comfort", "premium"}, "default": "comfort"},
						"transport":         map[string]any{"type": "string", "enum": []string{"car", "public", "walk", "bike"}, "default": "car"},
						"group_size":        map[string]any{"type": "integer", "default": 1, "description": "Количество участников.", "example": 4},
						"group_composition": map[string]any{"type": "object", "description": "Состав группы.", "example": map[string]any{"adults": 2, "children": []map[string]any{{"age": 8}, {"age": 12}}}},
					},
					"required": []string{"date_from", "date_to"},
				},
				"JoinTripRequest": map[string]any{
					"type":        "object",
					"description": "Запрос присоединения к поездке по invite-ссылке.",
					"properties": map[string]any{
						"invite_token": map[string]any{"type": "string", "format": "uuid", "description": "Токен приглашения."},
						"display_name": map[string]any{"type": "string", "description": "Имя участника.", "example": "Жена Маша"},
						"tags":         map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Теги предпочтений.", "example": []string{"сыр", "ферма", "тишина"}},
						"is_child":     map[string]any{"type": "boolean", "default": false, "description": "Является ли участник ребёнком."},
					},
					"required": []string{"invite_token", "display_name"},
				},
				"SwipeRequest": map[string]any{
					"type":        "object",
					"description": "Запрос свайпа сцены.",
					"properties": map[string]any{
						"scene_id":  map[string]any{"type": "string", "format": "uuid", "description": "UUID сцены."},
						"direction": map[string]any{"type": "string", "enum": []string{"right", "left"}, "description": "Направление свайпа."},
					},
					"required": []string{"scene_id", "direction"},
				},
				"VoiceProfileResponse": map[string]any{
					"type":        "object",
					"description": "Результат голосового профилирования. Screenshot-ready формат для vibe passport.",
					"properties": map[string]any{
						"transcription":       map[string]any{"type": "string", "description": "Распознанный текст из аудио.", "example": "Мне нравится отдыхать на природе, подальше от города."},
						"axes":                map[string]any{"$ref": "#/components/schemas/VibeAxes"},
						"extracted_tags":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "3-8 семантических тегов предпочтений.", "example": []string{"горы", "лес", "ферма", "виноградник", "каякинг"}},
						"vibe_summary":        map[string]any{"type": "string", "description": "Краткое резюме настроения на русском.", "example": "Семейный турист, предпочитающий природный отдых с элементами приключений."},
						"vibe_passport_title": map[string]any{"type": "string", "description": "Заголовок vibe-паспорта для экрана.", "example": "Исследователь Кубани"},
						"vector_id":           map[string]any{"type": "string", "format": "uuid", "description": "ID вектора в Qdrant."},
						"processing_time_ms":  map[string]any{"type": "integer", "description": "Время обработки пайплайна в миллисекундах.", "example": 2012},
					},
				},
				"VibeAxes": map[string]any{
					"type":        "object",
					"description": "Оси vibe-профиля пользователя. Каждая ось описывает предпочтение по шкале.",
					"properties": map[string]any{
						"stress_level":         map[string]any{"type": "number", "description": "Уровень стресса (0.0 = спокоен, 1.0 = устал).", "example": 0.7},
						"solitude_vs_social":   map[string]any{"type": "number", "description": "Уединение vs компания (-1.0 to 1.0).", "example": -0.5},
						"relax_vs_adrenaline":  map[string]any{"type": "number", "description": "Релакс vs адреналин (-1.0 to 1.0).", "example": -0.3},
						"gastro_vs_nature":     map[string]any{"type": "number", "description": "Гастрономия vs природа (-1.0 to 1.0).", "example": 0.4},
						"culture_vs_adventure": map[string]any{"type": "number", "description": "Культура vs приключения (-1.0 to 1.0).", "example": 0.3},
					},
				},
				"FinalizeRequest": map[string]any{
					"type":        "object",
					"description": "Опциональные параметры запроса финализации профиля.",
					"properties": map[string]any{
						"limit":               map[string]any{"type": "integer", "default": 10, "maximum": 50, "description": "Максимальное количество рекомендаций."},
						"child_friendly_only": map[string]any{"type": "boolean", "default": false, "description": "Фильтровать только детские локации."},
					},
				},
				"FinalizeResponse": map[string]any{
					"type":        "object",
					"description": "Результат финализации профиля с обогащёнными рекомендациями.",
					"properties": map[string]any{
						"recommendations": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/LocationRecommendation"}, "description": "Top-N рекомендованных локаций."},
						"total_found":     map[string]any{"type": "integer", "description": "Количество найденных совпадений."},
						"is_curated":      map[string]any{"type": "boolean", "description": "true, если рекомендации из curated demo набора (вектор не найден)."},
					},
				},
				"LocationRecommendation": map[string]any{
					"type":        "object",
					"description": "Обогащённая рекомендация локации для карточки на фронтенде.",
					"properties": map[string]any{
						"location_id":       map[string]any{"type": "string", "format": "uuid", "description": "UUID локации."},
						"score":             map[string]any{"type": "number", "description": "Cosine similarity (0-1).", "example": 0.92},
						"name":              map[string]any{"type": "string", "description": "Название локации.", "example": "Винодельня Лефкадия"},
						"category":          map[string]any{"type": "string", "description": "Категория.", "example": "winery"},
						"description_short": map[string]any{"type": "string", "description": "Краткое описание для карточки."},
						"tags":              map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Теги локации."},
						"preview_image_url": map[string]any{"type": "string", "description": "URL hero-изображения."},
						"splat_url":         map[string]any{"type": "string", "description": "URL 3D-сцены (.splat)."},
						"latitude":          map[string]any{"type": "number", "description": "Широта (WGS84)."},
						"longitude":         map[string]any{"type": "number", "description": "Долгота (WGS84)."},
						"density_level":     map[string]any{"type": "string", "enum": []string{"red", "yellow", "green"}, "description": "Уровень туристической плотности."},
						"reason_short":      map[string]any{"type": "string", "description": "Краткая причина рекомендации.", "example": "Высокое совпадение: природа, виноградник"},
						"tags_match":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Совпавшие теги между профилем и локацией."},
						"child_friendly":    map[string]any{"type": "boolean", "description": "Подходит для детей."},
					},
				},
				"MapPoint": map[string]any{
					"type":        "object",
					"description": "Точка для отображения на интерактивной карте.",
					"properties": map[string]any{
						"id":                   map[string]any{"type": "string", "format": "uuid", "description": "UUID локации."},
						"name":                 map[string]any{"type": "string", "description": "Название."},
						"category":             map[string]any{"type": "string", "description": "Категория."},
						"latitude":             map[string]any{"type": "number", "description": "Широта."},
						"longitude":            map[string]any{"type": "number", "description": "Долгота."},
						"density_level":        map[string]any{"type": "string", "enum": []string{"red", "yellow", "green"}, "description": "Цвет маркера плотности."},
						"preview_image_url":    map[string]any{"type": "string", "description": "URL превью."},
						"description_short":    map[string]any{"type": "string", "description": "Краткое описание для tooltip."},
						"is_recommended":       map[string]any{"type": "boolean", "description": "Рекомендована ли точка для текущего профиля."},
						"recommendation_score": map[string]any{"type": "number", "format": "float", "description": "Скор рекомендации (0.0-1.0). Заполняется только для рекомендованных."},
					},
				},
				"BuildRouteRequest": map[string]any{
					"type":        "object",
					"description": "Запрос на построение маршрута.",
					"properties": map[string]any{
						"location_ids": map[string]any{"type": "array", "items": map[string]any{"type": "string", "format": "uuid"}, "description": "UUID локаций маршрута (минимум 2).", "minItems": 2},
						"transport":    map[string]any{"type": "string", "enum": []string{"car", "walk", "bike", "public"}, "default": "car", "description": "Тип транспорта."},
						"optimize":     map[string]any{"type": "boolean", "default": false, "description": "Оптимизировать порядок точек."},
					},
					"required": []string{"location_ids"},
				},
				"RoutePreview": map[string]any{
					"type":        "object",
					"description": "Предварительный просмотр маршрута.",
					"properties": map[string]any{
						"points":            map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/RoutePoint"}, "description": "Точки маршрута в порядке следования."},
						"total_distance_km": map[string]any{"type": "number", "description": "Общая дистанция маршрута в км."},
						"total_time_min":    map[string]any{"type": "number", "description": "Общее время в пути в минутах."},
						"transport":         map[string]any{"type": "string", "description": "Тип транспорта."},
					},
				},
				"RoutePoint": map[string]any{
					"type":        "object",
					"description": "Точка маршрута.",
					"properties": map[string]any{
						"location_id":     map[string]any{"type": "string", "format": "uuid"},
						"name":            map[string]any{"type": "string"},
						"latitude":        map[string]any{"type": "number"},
						"longitude":       map[string]any{"type": "number"},
						"order":           map[string]any{"type": "integer", "description": "Порядковый номер в маршруте."},
						"distance_km":     map[string]any{"type": "number", "description": "Расстояние до следующей точки в км."},
						"travel_time_min": map[string]any{"type": "number", "description": "Время до следующей точки в минутах."},
					},
				},
			},
			"securitySchemes": map[string]any{
				"BearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
					"description":  "JWT access-токен. Получите через POST /api/v1/auth/login.",
				},
			},
		},
	}
}
