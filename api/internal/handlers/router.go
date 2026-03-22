// Файл router.go настраивает маршрутизацию HTTP-сервера.
// Регистрирует публичные маршруты (health, auth, docs, locations read, trip join) и защищённые
// маршруты (profile, locations write, media, trips), применяя JWT и RBAC middleware.
// Все маршруты API версионируются через префикс /api/v1.
package handlers

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
	"kudytudy-api/internal/middleware"
	"kudytudy-api/internal/models"
	"kudytudy-api/internal/services"
)

// SetupRoutes настраивает все маршруты API.
// Разделяет маршруты на публичные (без авторизации) и защищённые (JWT middleware).
func SetupRoutes(
	app *fiber.App,
	dbManager *database.Manager,
	storage *services.StorageService,
	authService *services.AuthService,
	jwtService *services.JWTService,
	userRepo *database.UserRepository,
	locationService *services.LocationService,
	tripService *services.TripService,
	vibeService *services.VibeService,
	mapService *services.MapService,
	routeService *services.RouteService,
	onboardingService *services.OnboardingService,
	logger *zap.Logger,
) {
	// Корневой маршрут — базовая информация о сервере.
	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "КудыТуды API",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// Группа маршрутов API версии 1.
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// --- Публичные маршруты (без авторизации) ---

	// Обработчик проверки здоровья сервисов.
	healthHandler := NewHealthHandler(dbManager, logger)
	v1.Get("/health", healthHandler.Check)       // обратная совместимость (полный пинг)
	v1.Get("/health/live", healthHandler.Live)   // liveness probe (мгновенный ответ)
	v1.Get("/health/ready", healthHandler.Ready) // readiness probe (пинг всех БД)

	// Обработчик загрузки медиафайлов (перемещён в защищённую зону — см. ниже).
	mediaHandler := NewMediaHandler(storage, logger)

	// Scalar API документация.
	scalarHandler := NewScalarHandler(logger)
	v1.Get("/docs", scalarHandler.ServeUI)
	v1.Get("/docs/openapi.json", scalarHandler.ServeSpec)

	// Аутентификация — публичные эндпоинты (регистрация, вход, обновление токена).
	authHandler := NewAuthHandler(authService, logger)
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)

	// Локации — публичные эндпоинты (чтение).
	locationHandler := NewLocationHandler(locationService, logger)
	v1.Get("/locations", locationHandler.List)
	v1.Get("/locations/:id", locationHandler.GetByID)
	v1.Get("/locations/:id/splat", locationHandler.GetSplat)

	// Карта — публичный эндпоинт (точки для маркеров).
	if mapService != nil {
		mapHandler := NewMapHandler(mapService, logger)
		v1.Get("/map/locations", mapHandler.GetMapLocations)
	}

	// Поездки — публичный эндпоинт с необязательной JWT-аутентификацией.
	// OptionalJWTAuth парсит JWT если передан, но не блокирует запрос при отсутствии.
	// Это позволяет auth-flex join: авторизованные получают user_id из профиля,
	// неавторизованные — проходят анонимно.
	optionalJWT := middleware.NewOptionalJWTAuth(jwtService, logger)
	tripHandler := NewTripHandler(tripService, logger)
	v1.Post("/trips/:id/join", optionalJWT, tripHandler.Join)

	// --- Защищённые маршруты (требуют JWT) ---

	// JWT middleware применяется ко всем маршрутам в этой группе.
	jwtMiddleware := middleware.NewJWTAuth(jwtService, logger)

	// Профиль текущего пользователя.
	profileHandler := NewProfileHandler(userRepo, logger)
	profile := v1.Group("/profile", jwtMiddleware)
	profile.Get("/me", profileHandler.GetMe)
	profile.Put("/me", profileHandler.UpdateMe)

	// Профилирование — голосовой ввод, свайп сцен, финализация.
	vibeHandler := NewVibeHandler(vibeService, logger)
	profile.Post("/voice", vibeHandler.VoiceProfile)
	profile.Post("/swipe", vibeHandler.Swipe)
	profile.Post("/finalize", vibeHandler.Finalize)
	profile.Get("/scenes", vibeHandler.GetScenes)

	// Загрузка медиафайлов — защищённый эндпоинт (JWT обязателен).
	v1.Post("/media/upload", jwtMiddleware, mediaHandler.Upload)

	// Локации — защищённые эндпоинты (создание, обновление, удаление).
	// Все операции записи доступны только хостам и администраторам (RBAC).
	locationsProtected := v1.Group("/locations", jwtMiddleware)
	locationsRBAC := middleware.RequireRole(logger, string(models.RoleHost), string(models.RoleB2GAdmin))
	locationsProtected.Post("/", locationsRBAC, locationHandler.Create)
	locationsProtected.Put("/:id", locationsRBAC, locationHandler.Update)
	locationsProtected.Delete("/:id", locationsRBAC, locationHandler.Delete)

	// Поездки — защищённые эндпоинты (создание, детали, invite, участники, маршрут).
	tripsProtected := v1.Group("/trips", jwtMiddleware)
	tripsProtected.Get("/", tripHandler.ListTrips)
	tripsProtected.Post("/", tripHandler.Create)
	tripsProtected.Get("/:id", tripHandler.GetByID)
	tripsProtected.Put("/:id", tripHandler.Update)
	tripsProtected.Post("/:id/invite", tripHandler.GenerateInvite)
	tripsProtected.Get("/:id/members", tripHandler.ListMembers)

	// Маршруты — защищённые эндпоинты (ручной и trip-aware режимы).
	if routeService != nil {
		routeHandler := NewRouteHandler(routeService, logger)
		routeGroup := v1.Group("/route", jwtMiddleware)
		routeGroup.Post("/build", routeHandler.BuildRoute)

		// Trip-aware построение маршрута.
		tripsProtected.Post("/:id/build-route", routeHandler.BuildTripRoute)
	}

	// Онбординг хостов — Zero-UI pipeline (GDD Feature 5).
	// Все эндпоинты требуют JWT и роль host или b2g_admin.
	if onboardingService != nil {
		onboardingHandler := NewOnboardingHandler(onboardingService, logger)
		hostGroup := v1.Group("/host", jwtMiddleware, locationsRBAC)
		hostGroup.Post("/onboard", onboardingHandler.Onboard)
		hostGroup.Post("/splat", onboardingHandler.StartSplatting)
		hostGroup.Get("/tasks/:id", onboardingHandler.GetTaskStatus)
		hostGroup.Get("/locations", onboardingHandler.GetHostLocations)
	}
}
