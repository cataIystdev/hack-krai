// Файл main.go — точка входа приложения Deep Krai API.
// Выполняет инициализацию конфигурации, логгера, подключений к БД,
// сервисов аутентификации, HTTP-сервера (Fiber) с middleware и graceful shutdown.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"deep-krai-api/internal/ai"
	"deep-krai-api/internal/config"
	"deep-krai-api/internal/database"
	"deep-krai-api/internal/handlers"
	"deep-krai-api/internal/middleware"
	"deep-krai-api/internal/services"
)

func main() {
	// --- 1. Загрузка конфигурации ---
	// Путь к .env файлу относительно директории запуска (api/).
	cfg, err := config.Load("../.env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	// --- 2. Инициализация логгера ---
	logger, err := config.NewLogger(cfg.App)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка инициализации логгера: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("запуск приложения",
		zap.String("name", cfg.App.Name),
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.Port),
	)

	// --- 3. Подключение к базам данных ---
	ctx := context.Background()
	dbManager := database.GetManager(logger)

	if err := dbManager.ConnectAll(ctx, cfg); err != nil {
		logger.Warn("не все базы данных доступны при запуске, сервер продолжит работу",
			zap.Error(err),
		)
	}

	// --- 4. Инициализация сервисов ---

	// Сервис хранилища (MinIO).
	var storageService *services.StorageService
	if dbManager.MinIO != nil {
		storageService = services.NewStorageService(dbManager.MinIO, logger)
	} else {
		logger.Warn("MinIO недоступен, загрузка файлов будет отключена")
	}

	// JWT-сервис для генерации и валидации токенов.
	jwtService := services.NewJWTService(cfg.JWT, logger)

	// Репозиторий пользователей (PostgreSQL).
	var userRepo *database.UserRepository
	if dbManager.Postgres != nil {
		userRepo = database.NewUserRepository(dbManager.Postgres, logger)
	} else {
		logger.Warn("PostgreSQL недоступен, аутентификация будет отключена")
	}

	// Сервис аутентификации.
	var authService *services.AuthService
	if userRepo != nil {
		authService = services.NewAuthService(userRepo, jwtService, logger)
	}

	// Репозиторий и сервис локаций (PostGIS).
	var locationService *services.LocationService
	if dbManager.Postgres != nil {
		locationRepo := database.NewLocationRepository(dbManager.Postgres, logger)
		locationService = services.NewLocationService(locationRepo, logger)
	}

	// Репозиторий и сервис поездок.
	var tripService *services.TripService
	if dbManager.Postgres != nil && userRepo != nil {
		tripRepo := database.NewTripRepository(dbManager.Postgres, logger)
		tripService = services.NewTripService(tripRepo, userRepo, logger)
	}

	// --- 4.5 Инициализация AI-клиентов и vibe-сервиса ---

	// Инициализация коллекций Qdrant (user_vibes, location_vibes).
	if dbManager.Qdrant != nil {
		if err := database.EnsureCollections(ctx, dbManager.Qdrant, logger); err != nil {
			logger.Warn("ошибка инициализации коллекций Qdrant", zap.Error(err))
		}
	}

	// AI-клиенты (OnlySQ API / mock).
	aiClient := ai.NewClient(cfg.AI, logger)
	llmClient := ai.NewLLMClient(aiClient, cfg.AI, logger)
	embeddingsClient := ai.NewEmbeddingsClient(aiClient, cfg.AI.EmbeddingsModel, cfg.AI.EmbeddingsBaseURL, logger)

	// STT-клиент: Vosk-server (локальный, бесплатный).
	voskClient := ai.NewVoskClient(cfg.AI.VoskHost, cfg.AI.VoskPort, cfg.AI.IsMockMode(), logger)

	if cfg.AI.IsMockMode() {
		logger.Warn("AI-клиенты работают в mock-режиме (AI_API_KEY не задан)")
	} else {
		logger.Info("AI-клиенты инициализированы",
			zap.String("base_url", cfg.AI.BaseURL),
			zap.String("llm_model", cfg.AI.LLMModel),
			zap.String("vosk_url", fmt.Sprintf("ws://%s:%d", cfg.AI.VoskHost, cfg.AI.VoskPort)),
		)
	}

	// Vibe-сервис профилирования.
	var vibeService *services.VibeService
	if dbManager.Postgres != nil && dbManager.Qdrant != nil {
		vibeRepo := database.NewVibeRepository(dbManager.Postgres, dbManager.Qdrant, logger)
		vibeService = services.NewVibeService(voskClient, llmClient, embeddingsClient, vibeRepo, storageService, logger)
	} else {
		logger.Warn("Vibe-сервис недоступен: требуется PostgreSQL и Qdrant")
	}

	// --- 5. Создание HTTP-сервера Fiber ---
	app := fiber.New(fiber.Config{
		// ServerHeader — заголовок Server в HTTP-ответах.
		ServerHeader: cfg.App.Name,

		// ReadTimeout — максимальное время чтения запроса.
		ReadTimeout: 15 * time.Second,

		// WriteTimeout — максимальное время записи ответа.
		WriteTimeout: 15 * time.Second,

		// IdleTimeout — максимальное время ожидания следующего запроса (keep-alive).
		IdleTimeout: 60 * time.Second,

		// BodyLimit — максимальный размер тела запроса (100 МБ для загрузки файлов).
		BodyLimit: 100 * 1024 * 1024,

		// AppName — имя приложения для отладки.
		AppName: cfg.App.Name,
	})

	// --- 6. Подключение middleware ---
	// Порядок подключения middleware важен:
	// Recovery первым перехватывает паники, Logger логирует все запросы,
	// CORS добавляет заголовки для кросс-доменных запросов.
	app.Use(middleware.NewRecovery(logger))
	app.Use(middleware.NewRequestLogger(logger))
	app.Use(middleware.NewCORS())

	// --- 7. Регистрация маршрутов ---
	handlers.SetupRoutes(app, dbManager, storageService, authService, jwtService, userRepo, locationService, tripService, vibeService, logger)

	// --- 8. Graceful Shutdown ---
	// Создание канала для перехвата сигналов завершения (SIGINT, SIGTERM).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP-сервера в отдельной горутине.
	go func() {
		addr := fmt.Sprintf(":%d", cfg.App.Port)
		logger.Info("HTTP-сервер запущен", zap.String("addr", addr))

		if err := app.Listen(addr); err != nil {
			logger.Fatal("ошибка HTTP-сервера", zap.Error(err))
		}
	}()

	// Ожидание сигнала завершения.
	sig := <-quit
	logger.Info("получен сигнал завершения, начинается graceful shutdown",
		zap.String("signal", sig.String()),
	)

	// Таймаут на graceful shutdown — 30 секунд.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Остановка HTTP-сервера: дожидается завершения текущих запросов.
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("ошибка при graceful shutdown HTTP-сервера", zap.Error(err))
	}

	// Закрытие подключений к базам данных.
	dbManager.Close(shutdownCtx)

	logger.Info("приложение корректно остановлено")
}
