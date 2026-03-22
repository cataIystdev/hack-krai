// Пакет config предоставляет структуры конфигурации приложения КудыТуды.
// Все параметры загружаются из .env файла и переменных окружения через Viper.
package config

// AppConfig — корневая структура конфигурации приложения.
// Содержит вложенные конфигурации для HTTP-сервера и каждого сервиса БД.
type AppConfig struct {
	// App — общие параметры приложения (имя, окружение, порт, уровень логирования).
	App AppSettings `mapstructure:"app"`

	// Postgres — конфигурация подключения к PostgreSQL + PostGIS.
	Postgres PostgresConfig `mapstructure:"postgres"`

	// Redis — конфигурация подключения к Redis.
	Redis RedisConfig `mapstructure:"redis"`

	// Qdrant — конфигурация подключения к векторной БД Qdrant.
	Qdrant QdrantConfig `mapstructure:"qdrant"`

	// Neo4j — конфигурация подключения к графовой БД Neo4j.
	Neo4j Neo4jConfig `mapstructure:"neo4j"`

	// ClickHouse — конфигурация подключения к аналитической БД ClickHouse.
	ClickHouse ClickHouseConfig `mapstructure:"clickhouse"`

	// MinIO — конфигурация подключения к S3-совместимому хранилищу MinIO.
	MinIO MinIOConfig `mapstructure:"minio"`

	// JWT — конфигурация JSON Web Token аутентификации.
	JWT JWTConfig `mapstructure:"jwt"`

	// AI — конфигурация подключения к AI API (OnlySQ / OpenAI-совместимый).
	AI AIConfig `mapstructure:"ai"`
}

// AppSettings — общие параметры приложения.
type AppSettings struct {
	// Name — имя приложения, используется в логах и метриках.
	Name string `mapstructure:"name"`

	// Env — окружение запуска: development, staging, production.
	Env string `mapstructure:"env"`

	// Port — порт HTTP-сервера.
	Port int `mapstructure:"port"`

	// LogLevel — уровень логирования: debug, info, warn, error.
	LogLevel string `mapstructure:"log_level"`
}

// PostgresConfig — параметры подключения к PostgreSQL.
type PostgresConfig struct {
	// Host — адрес сервера PostgreSQL.
	Host string `mapstructure:"host"`

	// Port — порт подключения.
	Port int `mapstructure:"port"`

	// User — имя пользователя БД.
	User string `mapstructure:"user"`

	// Password — пароль пользователя БД.
	Password string `mapstructure:"password"`

	// DB — имя базы данных.
	DB string `mapstructure:"db"`

	// MaxConns — максимальное количество подключений в пуле.
	MaxConns int `mapstructure:"max_conns"`

	// MinConns — минимальное количество подключений в пуле.
	MinConns int `mapstructure:"min_conns"`
}

// DSN формирует строку подключения к PostgreSQL в формате URI.
func (c PostgresConfig) DSN() string {
	return "postgres://" + c.User + ":" + c.Password +
		"@" + c.Host + ":" + itoa(c.Port) +
		"/" + c.DB + "?sslmode=disable"
}

// RedisConfig — параметры подключения к Redis.
type RedisConfig struct {
	// Host — адрес сервера Redis.
	Host string `mapstructure:"host"`

	// Port — порт подключения.
	Port int `mapstructure:"port"`

	// Password — пароль авторизации.
	Password string `mapstructure:"password"`

	// DB — номер базы данных Redis (0-15).
	DB int `mapstructure:"db"`
}

// Addr формирует строку адреса для подключения к Redis.
func (c RedisConfig) Addr() string {
	return c.Host + ":" + itoa(c.Port)
}

// QdrantConfig — параметры подключения к векторной БД Qdrant.
type QdrantConfig struct {
	// Host — адрес сервера Qdrant.
	Host string `mapstructure:"host"`

	// HTTPPort — порт REST API Qdrant.
	HTTPPort int `mapstructure:"http_port"`

	// GRPCPort — порт gRPC API Qdrant.
	GRPCPort int `mapstructure:"grpc_port"`
}

// GRPCAddr формирует строку адреса gRPC для подключения к Qdrant.
func (c QdrantConfig) GRPCAddr() string {
	return c.Host + ":" + itoa(c.GRPCPort)
}

// Neo4jConfig — параметры подключения к графовой БД Neo4j.
type Neo4jConfig struct {
	// Host — адрес сервера Neo4j.
	Host string `mapstructure:"host"`

	// BoltPort — порт Bolt-протокола для запросов.
	BoltPort int `mapstructure:"bolt_port"`

	// HTTPPort — порт веб-интерфейса Neo4j Browser.
	HTTPPort int `mapstructure:"http_port"`

	// User — имя пользователя.
	User string `mapstructure:"user"`

	// Password — пароль.
	Password string `mapstructure:"password"`
}

// BoltURI формирует URI для подключения через Bolt-протокол.
func (c Neo4jConfig) BoltURI() string {
	return "bolt://" + c.Host + ":" + itoa(c.BoltPort)
}

// ClickHouseConfig — параметры подключения к аналитической БД ClickHouse.
type ClickHouseConfig struct {
	// Host — адрес сервера ClickHouse.
	Host string `mapstructure:"host"`
	// Port — порт нативного протокола ClickHouse.
	Port int `mapstructure:"port"`

	// HTTPPort — порт HTTP API.
	HTTPPort int `mapstructure:"http_port"`

	// User — имя пользователя.
	User string `mapstructure:"user"`

	// Password — пароль.
	Password string `mapstructure:"password"`

	// DB — имя базы данных.
	DB string `mapstructure:"db"`
}

// MinIOConfig — параметры подключения к объектному хранилищу MinIO.
type MinIOConfig struct {
	// Host — адрес сервера MinIO.
	Host string `mapstructure:"host"`

	// APIPort — порт S3 API.
	APIPort int `mapstructure:"api_port"`

	// ConsolePort — порт веб-консоли управления.
	ConsolePort int `mapstructure:"console_port"`

	// RootUser — имя корневого пользователя (аналог AWS Access Key).
	RootUser string `mapstructure:"root_user"`

	// RootPassword — пароль корневого пользователя (аналог AWS Secret Key).
	RootPassword string `mapstructure:"root_password"`

	// Bucket — имя бакета для хранения медиафайлов.
	Bucket string `mapstructure:"bucket"`

	// UseSSL — использовать ли SSL для подключения.
	UseSSL bool `mapstructure:"use_ssl"`

	// PublicURL — публичный базовый URL для доступа к файлам (например http://141.98.7.225:9102).
	// Если задан, Upload() формирует прямой URL: {PublicURL}/{bucket}/{key}.
	// Если пуст, используется presigned URL через внутренний MinIO endpoint.
	PublicURL string `mapstructure:"public_url"`
}

// Endpoint формирует строку адреса для подключения к MinIO API.
func (c MinIOConfig) Endpoint() string {
	return c.Host + ":" + itoa(c.APIPort)
}

// JWTConfig — параметры JWT-аутентификации.
type JWTConfig struct {
	// SecretKey — секретный ключ для подписи JWT (HMAC-SHA256).
	// Должен быть криптографически стойкой строкой длиной не менее 32 символов.
	SecretKey string `mapstructure:"secret_key"`

	// AccessTokenTTLMinutes — время жизни access-токена в минутах.
	// Рекомендуемое значение: 15 минут.
	AccessTokenTTLMinutes int `mapstructure:"access_ttl_minutes"`

	// RefreshTokenTTLHours — время жизни refresh-токена в часах.
	// Рекомендуемое значение: 168 часов (7 дней).
	RefreshTokenTTLHours int `mapstructure:"refresh_ttl_hours"`

	// Issuer — идентификатор выпускающей стороны токена.
	// Используется для валидации поля "iss" в JWT claims.
	Issuer string `mapstructure:"issuer"`
}

// AIConfig — параметры подключения к AI API (OnlySQ, Vosk).
type AIConfig struct {
	// BaseURL — базовый URL AI API (по умолчанию https://api.onlysq.ru/ai/openai/).
	BaseURL string `mapstructure:"base_url"`

	// APIKey — ключ авторизации AI API.
	APIKey string `mapstructure:"api_key"`

	// WhisperModel — модель для распознавания речи (по умолчанию whisper-1). Deprecated: используй Vosk.
	WhisperModel string `mapstructure:"whisper_model"`

	// LLMModel — модель для структурированного извлечения осей (по умолчанию gpt-4o-mini).
	LLMModel string `mapstructure:"llm_model"`

	// EmbeddingsModel — модель для генерации эмбеддингов (по умолчанию text-embedding-3-large).
	EmbeddingsModel string `mapstructure:"embeddings_model"`

	// EmbeddingsBaseURL — прямой хост для локальной Embeddings API (без v1/).
	EmbeddingsBaseURL string `mapstructure:"embeddings_base_url"`

	// VoskHost — хост Vosk-server (по умолчанию localhost).
	VoskHost string `mapstructure:"vosk_host"`

	// VoskPort — порт Vosk-server (по умолчанию 2700).
	VoskPort int `mapstructure:"vosk_port"`

	// DemoLatencyMs — искусственная задержка в миллисекундах для mock-режима.
	// Создаёт естественную UX-анимацию "ИИ думает" при демонстрации.
	// По умолчанию 2000ms. Диапазон: 800-5000ms.
	DemoLatencyMs int `mapstructure:"demo_latency_ms"`
}

// IsMockMode возвращает true, если API-ключ не задан и нужно использовать mock-клиенты.
func (c AIConfig) IsMockMode() bool {
	return c.APIKey == ""
}

// itoa — вспомогательная функция для преобразования int в строку.
// Используется в методах формирования строк подключения для избежания
// зависимости от пакета strconv в методах-форматтерах.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	negative := false
	if i < 0 {
		negative = true
		i = -i
	}
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	if negative {
		result = "-" + result
	}
	return result
}
