// Пакет config содержит логику загрузки конфигурации приложения.
// Загрузка выполняется через Viper: чтение .env файла, маппинг переменных
// окружения на Go-структуры, валидация обязательных полей.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Load загружает конфигурацию приложения из .env файла и переменных окружения.
// Параметр envPath — путь к .env файлу (например, "../.env" или "/app/.env").
// Переменные окружения имеют приоритет над значениями из файла.
// Возвращает заполненную структуру AppConfig или ошибку при загрузке/валидации.
func Load(envPath string) (*AppConfig, error) {
	v := viper.New()

	// Установка значений по умолчанию для всех параметров.
	// Значения по умолчанию используются, если переменная окружения
	// не задана ни в .env файле, ни в системных переменных окружения.
	setDefaults(v)

	// Настройка чтения .env файла.
	// Viper ищет файл по указанному пути и загружает его содержимое.
	v.SetConfigFile(envPath)
	v.SetConfigType("env")

	// Чтение .env файла. Если файл не найден — используются значения по умолчанию
	// и системные переменные окружения. Ошибка чтения файла не является фатальной.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("ошибка чтения файла конфигурации: %w", err)
		}
	}

	// Включение автоматического чтения переменных окружения.
	// Все ключи конфигурации автоматически маппятся на переменные окружения.
	v.AutomaticEnv()

	// Замена точки на подчёркивание для вложенных ключей.
	// Например, ключ "postgres.host" маппится на переменную "POSTGRES_HOST".
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Привязка переменных окружения к ключам конфигурации.
	// Каждый ключ явно привязывается для корректного маппинга.
	bindEnvVariables(v)

	// Сборка конфигурации из загруженных значений.
	cfg := buildConfig(v)

	// Валидация обязательных параметров конфигурации.
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("ошибка валидации конфигурации: %w", err)
	}

	return cfg, nil
}

// setDefaults устанавливает значения по умолчанию для всех параметров конфигурации.
// Значения выбраны для режима разработки (development).
func setDefaults(v *viper.Viper) {
	// Приложение
	v.SetDefault("APP_NAME", "kudytudy-api")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_PORT", 8080)
	v.SetDefault("APP_LOG_LEVEL", "debug")

	// PostgreSQL
	v.SetDefault("POSTGRES_HOST", "localhost")
	v.SetDefault("POSTGRES_PORT", 5432)
	v.SetDefault("POSTGRES_USER", "deepkrai")
	v.SetDefault("POSTGRES_PASSWORD", "")
	v.SetDefault("POSTGRES_DB", "deepkrai")
	v.SetDefault("POSTGRES_MAX_CONNS", 20)
	v.SetDefault("POSTGRES_MIN_CONNS", 5)

	// Redis
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)

	// Qdrant
	v.SetDefault("QDRANT_HOST", "localhost")
	v.SetDefault("QDRANT_HTTP_PORT", 6333)
	v.SetDefault("QDRANT_GRPC_PORT", 6334)

	// Neo4j
	v.SetDefault("NEO4J_HOST", "localhost")
	v.SetDefault("NEO4J_BOLT_PORT", 7687)
	v.SetDefault("NEO4J_HTTP_PORT", 7474)
	v.SetDefault("NEO4J_USER", "neo4j")
	v.SetDefault("NEO4J_PASSWORD", "")

	// ClickHouse
	v.SetDefault("CLICKHOUSE_HOST", "localhost")
	v.SetDefault("CLICKHOUSE_PORT", 9000)
	v.SetDefault("CLICKHOUSE_HTTP_PORT", 8123)
	v.SetDefault("CLICKHOUSE_USER", "deepkrai")
	v.SetDefault("CLICKHOUSE_PASSWORD", "")
	v.SetDefault("CLICKHOUSE_DB", "deepkrai_analytics")

	// MinIO
	v.SetDefault("MINIO_HOST", "localhost")
	v.SetDefault("MINIO_API_PORT", 9000)
	v.SetDefault("MINIO_CONSOLE_PORT", 9001)
	v.SetDefault("MINIO_ROOT_USER", "")
	v.SetDefault("MINIO_ROOT_PASSWORD", "")
	v.SetDefault("MINIO_BUCKET", "deepkrai-media")
	v.SetDefault("MINIO_USE_SSL", false)

	// JWT
	v.SetDefault("JWT_SECRET", "")
	v.SetDefault("JWT_ACCESS_TTL_MINUTES", 15)
	v.SetDefault("JWT_REFRESH_TTL_HOURS", 168)
	v.SetDefault("JWT_ISSUER", "kudytudy-api")

	// AI (OnlySQ API)
	v.SetDefault("AI_BASE_URL", "https://api.onlysq.ru/ai/openai/")
	v.SetDefault("AI_API_KEY", "")
	v.SetDefault("AI_WHISPER_MODEL", "whisper-1")
	v.SetDefault("AI_LLM_MODEL", "gpt-4o-mini")
	v.SetDefault("AI_EMBEDDINGS_MODEL", "text-embedding-3-large")
	v.SetDefault("VOSK_HOST", "localhost")
	v.SetDefault("VOSK_PORT", 2700)
}

// bindEnvVariables привязывает каждый ключ конфигурации к соответствующей
// переменной окружения. Для плоской структуры .env файла переменные
// маппятся напрямую (POSTGRES_HOST -> POSTGRES_HOST).
func bindEnvVariables(v *viper.Viper) {
	envBindings := []string{
		"APP_NAME", "APP_ENV", "APP_PORT", "APP_LOG_LEVEL",
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD",
		"POSTGRES_DB", "POSTGRES_MAX_CONNS", "POSTGRES_MIN_CONNS",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
		"QDRANT_HOST", "QDRANT_HTTP_PORT", "QDRANT_GRPC_PORT",
		"NEO4J_HOST", "NEO4J_BOLT_PORT", "NEO4J_HTTP_PORT", "NEO4J_USER", "NEO4J_PASSWORD",
		"CLICKHOUSE_HOST", "CLICKHOUSE_PORT", "CLICKHOUSE_HTTP_PORT",
		"CLICKHOUSE_USER", "CLICKHOUSE_PASSWORD", "CLICKHOUSE_DB",
		"MINIO_HOST", "MINIO_API_PORT", "MINIO_CONSOLE_PORT",
		"MINIO_ROOT_USER", "MINIO_ROOT_PASSWORD", "MINIO_BUCKET", "MINIO_USE_SSL",
		"JWT_SECRET", "JWT_ACCESS_TTL_MINUTES", "JWT_REFRESH_TTL_HOURS", "JWT_ISSUER",
		"AI_BASE_URL", "AI_API_KEY", "AI_WHISPER_MODEL", "AI_LLM_MODEL", "AI_EMBEDDINGS_MODEL", "AI_EMBEDDINGS_BASE_URL",
		"VOSK_HOST", "VOSK_PORT",
	}
	for _, key := range envBindings {
		_ = v.BindEnv(key)
	}
}

// buildConfig собирает структуру AppConfig из значений Viper.
// Каждое поле структуры заполняется соответствующим значением конфигурации.
func buildConfig(v *viper.Viper) *AppConfig {
	return &AppConfig{
		App: AppSettings{
			Name:     v.GetString("APP_NAME"),
			Env:      v.GetString("APP_ENV"),
			Port:     v.GetInt("APP_PORT"),
			LogLevel: v.GetString("APP_LOG_LEVEL"),
		},
		Postgres: PostgresConfig{
			Host:     v.GetString("POSTGRES_HOST"),
			Port:     v.GetInt("POSTGRES_PORT"),
			User:     v.GetString("POSTGRES_USER"),
			Password: v.GetString("POSTGRES_PASSWORD"),
			DB:       v.GetString("POSTGRES_DB"),
			MaxConns: v.GetInt("POSTGRES_MAX_CONNS"),
			MinConns: v.GetInt("POSTGRES_MIN_CONNS"),
		},
		Redis: RedisConfig{
			Host:     v.GetString("REDIS_HOST"),
			Port:     v.GetInt("REDIS_PORT"),
			Password: v.GetString("REDIS_PASSWORD"),
			DB:       v.GetInt("REDIS_DB"),
		},
		Qdrant: QdrantConfig{
			Host:     v.GetString("QDRANT_HOST"),
			HTTPPort: v.GetInt("QDRANT_HTTP_PORT"),
			GRPCPort: v.GetInt("QDRANT_GRPC_PORT"),
		},
		Neo4j: Neo4jConfig{
			Host:     v.GetString("NEO4J_HOST"),
			BoltPort: v.GetInt("NEO4J_BOLT_PORT"),
			HTTPPort: v.GetInt("NEO4J_HTTP_PORT"),
			User:     v.GetString("NEO4J_USER"),
			Password: v.GetString("NEO4J_PASSWORD"),
		},
		ClickHouse: ClickHouseConfig{
			Host:     v.GetString("CLICKHOUSE_HOST"),
			Port:     v.GetInt("CLICKHOUSE_PORT"),
			HTTPPort: v.GetInt("CLICKHOUSE_HTTP_PORT"),
			User:     v.GetString("CLICKHOUSE_USER"),
			Password: v.GetString("CLICKHOUSE_PASSWORD"),
			DB:       v.GetString("CLICKHOUSE_DB"),
		},
		MinIO: MinIOConfig{
			Host:         v.GetString("MINIO_HOST"),
			APIPort:      v.GetInt("MINIO_API_PORT"),
			ConsolePort:  v.GetInt("MINIO_CONSOLE_PORT"),
			RootUser:     v.GetString("MINIO_ROOT_USER"),
			RootPassword: v.GetString("MINIO_ROOT_PASSWORD"),
			Bucket:       v.GetString("MINIO_BUCKET"),
			UseSSL:       v.GetBool("MINIO_USE_SSL"),
		},
		JWT: JWTConfig{
			SecretKey:             v.GetString("JWT_SECRET"),
			AccessTokenTTLMinutes: v.GetInt("JWT_ACCESS_TTL_MINUTES"),
			RefreshTokenTTLHours:  v.GetInt("JWT_REFRESH_TTL_HOURS"),
			Issuer:                v.GetString("JWT_ISSUER"),
		},
		AI: AIConfig{
			BaseURL:           v.GetString("AI_BASE_URL"),
			APIKey:            v.GetString("AI_API_KEY"),
			WhisperModel:      v.GetString("AI_WHISPER_MODEL"),
			LLMModel:          v.GetString("AI_LLM_MODEL"),
			EmbeddingsModel:   v.GetString("AI_EMBEDDINGS_MODEL"),
			EmbeddingsBaseURL: v.GetString("AI_EMBEDDINGS_BASE_URL"),
			VoskHost:          v.GetString("VOSK_HOST"),
			VoskPort:          v.GetInt("VOSK_PORT"),
		},
	}
}

// validate проверяет обязательные параметры конфигурации.
// Возвращает ошибку, если какой-либо критический параметр не задан.
func validate(cfg *AppConfig) error {
	if cfg.Postgres.Password == "" {
		return fmt.Errorf("POSTGRES_PASSWORD обязателен")
	}
	if cfg.Neo4j.Password == "" {
		return fmt.Errorf("NEO4J_PASSWORD обязателен")
	}
	if cfg.MinIO.RootUser == "" {
		return fmt.Errorf("MINIO_ROOT_USER обязателен")
	}
	if cfg.MinIO.RootPassword == "" {
		return fmt.Errorf("MINIO_ROOT_PASSWORD обязателен")
	}
	if cfg.App.Port <= 0 || cfg.App.Port > 65535 {
		return fmt.Errorf("APP_PORT должен быть в диапазоне 1-65535, получено: %d", cfg.App.Port)
	}
	if cfg.JWT.SecretKey == "" {
		return fmt.Errorf("JWT_SECRET обязателен")
	}
	if len(cfg.JWT.SecretKey) < 32 {
		return fmt.Errorf("JWT_SECRET должен содержать не менее 32 символов")
	}
	return nil
}
