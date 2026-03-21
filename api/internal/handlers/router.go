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

	// Обработчик загрузки медиафайлов.
	mediaHandler := NewMediaHandler(storage, logger)
	v1.Post("/media/upload", mediaHandler.Upload)

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

	// Поездки — публичный эндпоинт (присоединение по invite-ссылке без авторизации).
	tripHandler := NewTripHandler(tripService, logger)
	v1.Post("/trips/:id/join", tripHandler.Join)

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

	// Локации — защищённые эндпоинты (создание, обновление, удаление).
	// POST доступен только хостам и администраторам (RBAC).
	locationsProtected := v1.Group("/locations", jwtMiddleware)
	locationsProtected.Post("/", jwtMiddleware, middleware.RequireRole(logger, string(models.RoleHost), string(models.RoleB2GAdmin)), locationHandler.Create)
	locationsProtected.Put("/:id", locationHandler.Update)
	locationsProtected.Delete("/:id", locationHandler.Delete)

	// Поездки — защищённые эндпоинты (создание, детали, invite, участники).
	tripsProtected := v1.Group("/trips", jwtMiddleware)
	tripsProtected.Get("/", tripHandler.ListTrips)
	tripsProtected.Post("/", tripHandler.Create)
	tripsProtected.Get("/:id", tripHandler.GetByID)
	tripsProtected.Post("/:id/invite", tripHandler.GenerateInvite)
	tripsProtected.Get("/:id/members", tripHandler.ListMembers)

	// Маршруты — защищённый эндпоинт (построение маршрута).
	if routeService != nil {
		routeHandler := NewRouteHandler(routeService, logger)
		routeGroup := v1.Group("/route", jwtMiddleware)
		routeGroup.Post("/build", routeHandler.BuildRoute)
	}
}
