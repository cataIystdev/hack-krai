// Тесты для пакета config.
// Покрывают загрузку конфигурации, значения по умолчанию, валидацию
// обязательных полей, парсинг уровня логирования и методы формирования строк.
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadFromEnvFile проверяет загрузку конфигурации из .env файла.
// Создаёт временный .env файл с минимальным набором параметров
// и убеждается, что все значения корректно загружены.
func TestLoadFromEnvFile(t *testing.T) {
	// Создание временного .env файла с обязательными параметрами.
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := `APP_NAME=test-app
APP_ENV=production
APP_PORT=9090
APP_LOG_LEVEL=info
POSTGRES_HOST=pg-host
POSTGRES_PORT=5433
POSTGRES_USER=testuser
POSTGRES_PASSWORD=testpass
POSTGRES_DB=testdb
POSTGRES_MAX_CONNS=10
POSTGRES_MIN_CONNS=2
REDIS_HOST=redis-host
REDIS_PORT=6380
REDIS_PASSWORD=redispass
REDIS_DB=1
QDRANT_HOST=qdrant-host
QDRANT_HTTP_PORT=6335
QDRANT_GRPC_PORT=6336
NEO4J_HOST=neo4j-host
NEO4J_BOLT_PORT=7688
NEO4J_HTTP_PORT=7475
NEO4J_USER=testneo
NEO4J_PASSWORD=neopass
CLICKHOUSE_HOST=ch-host
CLICKHOUSE_PORT=9001
CLICKHOUSE_HTTP_PORT=8124
CLICKHOUSE_USER=chuser
CLICKHOUSE_PASSWORD=chpass
CLICKHOUSE_DB=chdb
MINIO_HOST=minio-host
MINIO_API_PORT=9002
MINIO_CONSOLE_PORT=9003
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=miniopass
MINIO_BUCKET=test-bucket
MINIO_USE_SSL=true
JWT_SECRET=test-secret-key-for-testing-must-be-32-characters-long
JWT_ACCESS_TTL_MINUTES=30
JWT_REFRESH_TTL_HOURS=72
JWT_ISSUER=test-issuer
`
	err := os.WriteFile(envFile, []byte(content), 0644)
	require.NoError(t, err, "запись .env файла должна пройти без ошибок")

	cfg, err := Load(envFile)
	require.NoError(t, err, "загрузка конфигурации должна пройти без ошибок")

	// Проверка параметров приложения.
	assert.Equal(t, "test-app", cfg.App.Name)
	assert.Equal(t, "production", cfg.App.Env)
	assert.Equal(t, 9090, cfg.App.Port)
	assert.Equal(t, "info", cfg.App.LogLevel)

	// Проверка PostgreSQL.
	assert.Equal(t, "pg-host", cfg.Postgres.Host)
	assert.Equal(t, 5433, cfg.Postgres.Port)
	assert.Equal(t, "testuser", cfg.Postgres.User)
	assert.Equal(t, "testpass", cfg.Postgres.Password)
	assert.Equal(t, "testdb", cfg.Postgres.DB)
	assert.Equal(t, 10, cfg.Postgres.MaxConns)
	assert.Equal(t, 2, cfg.Postgres.MinConns)

	// Проверка Redis.
	assert.Equal(t, "redis-host", cfg.Redis.Host)
	assert.Equal(t, 6380, cfg.Redis.Port)
	assert.Equal(t, "redispass", cfg.Redis.Password)
	assert.Equal(t, 1, cfg.Redis.DB)

	// Проверка Qdrant.
	assert.Equal(t, "qdrant-host", cfg.Qdrant.Host)
	assert.Equal(t, 6335, cfg.Qdrant.HTTPPort)
	assert.Equal(t, 6336, cfg.Qdrant.GRPCPort)

	// Проверка Neo4j.
	assert.Equal(t, "neo4j-host", cfg.Neo4j.Host)
	assert.Equal(t, 7688, cfg.Neo4j.BoltPort)
	assert.Equal(t, 7475, cfg.Neo4j.HTTPPort)
	assert.Equal(t, "testneo", cfg.Neo4j.User)
	assert.Equal(t, "neopass", cfg.Neo4j.Password)

	// Проверка ClickHouse.
	assert.Equal(t, "ch-host", cfg.ClickHouse.Host)
	assert.Equal(t, 9001, cfg.ClickHouse.Port)
	assert.Equal(t, 8124, cfg.ClickHouse.HTTPPort)
	assert.Equal(t, "chuser", cfg.ClickHouse.User)
	assert.Equal(t, "chpass", cfg.ClickHouse.Password)
	assert.Equal(t, "chdb", cfg.ClickHouse.DB)

	// Проверка MinIO.
	assert.Equal(t, "minio-host", cfg.MinIO.Host)
	assert.Equal(t, 9002, cfg.MinIO.APIPort)
	assert.Equal(t, 9003, cfg.MinIO.ConsolePort)
	assert.Equal(t, "minioadmin", cfg.MinIO.RootUser)
	assert.Equal(t, "miniopass", cfg.MinIO.RootPassword)
	assert.Equal(t, "test-bucket", cfg.MinIO.Bucket)
	assert.True(t, cfg.MinIO.UseSSL)
}

// TestLoadValidationFailsOnMissingPassword проверяет, что загрузка
// конфигурации завершается ошибкой, если не задан обязательный параметр.
func TestLoadValidationFailsOnMissingPassword(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	// Файл без POSTGRES_PASSWORD — валидация должна вернуть ошибку.
	content := `APP_NAME=test
`
	err := os.WriteFile(envFile, []byte(content), 0644)
	require.NoError(t, err)

	_, err = Load(envFile)
	assert.Error(t, err, "загрузка без POSTGRES_PASSWORD должна вернуть ошибку")
	assert.Contains(t, err.Error(), "POSTGRES_PASSWORD")
}

// TestLoadValidationFailsOnInvalidPort проверяет валидацию порта приложения.
func TestLoadValidationFailsOnInvalidPort(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := `APP_PORT=0
POSTGRES_PASSWORD=pass
NEO4J_PASSWORD=pass
MINIO_ROOT_USER=user
MINIO_ROOT_PASSWORD=pass
JWT_SECRET=test-secret-key-for-testing-must-be-32-characters-long
`
	err := os.WriteFile(envFile, []byte(content), 0644)
	require.NoError(t, err)

	_, err = Load(envFile)
	assert.Error(t, err, "загрузка с APP_PORT=0 должна вернуть ошибку")
	assert.Contains(t, err.Error(), "APP_PORT")
}

// TestPostgresDSN проверяет формирование DSN-строки для PostgreSQL.
func TestPostgresDSN(t *testing.T) {
	cfg := PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "pass",
		DB:       "testdb",
	}
	expected := "postgres://user:pass@localhost:5432/testdb?sslmode=disable"
	assert.Equal(t, expected, cfg.DSN())
}

// TestRedisAddr проверяет формирование адреса для Redis.
func TestRedisAddr(t *testing.T) {
	cfg := RedisConfig{Host: "redis-host", Port: 6379}
	assert.Equal(t, "redis-host:6379", cfg.Addr())
}

// TestQdrantGRPCAddr проверяет формирование gRPC-адреса для Qdrant.
func TestQdrantGRPCAddr(t *testing.T) {
	cfg := QdrantConfig{Host: "qdrant-host", GRPCPort: 6334}
	assert.Equal(t, "qdrant-host:6334", cfg.GRPCAddr())
}

// TestNeo4jBoltURI проверяет формирование Bolt URI для Neo4j.
func TestNeo4jBoltURI(t *testing.T) {
	cfg := Neo4jConfig{Host: "neo4j-host", BoltPort: 7687}
	assert.Equal(t, "bolt://neo4j-host:7687", cfg.BoltURI())
}

// TestMinIOEndpoint проверяет формирование адреса для MinIO API.
func TestMinIOEndpoint(t *testing.T) {
	cfg := MinIOConfig{Host: "minio-host", APIPort: 9000}
	assert.Equal(t, "minio-host:9000", cfg.Endpoint())
}

// TestParseLogLevel проверяет корректный парсинг уровней логирования.
func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"debug", false},
		{"info", false},
		{"warn", false},
		{"error", false},
		{"unknown", true},
	}
	for _, tt := range tests {
		_, err := parseLogLevel(tt.input)
		if tt.wantErr {
			assert.Error(t, err, "уровень %q должен вернуть ошибку", tt.input)
		} else {
			assert.NoError(t, err, "уровень %q не должен вернуть ошибку", tt.input)
		}
	}
}

// TestNewLoggerDevelopment проверяет создание логгера в режиме development.
func TestNewLoggerDevelopment(t *testing.T) {
	settings := AppSettings{
		Name:     "test-app",
		Env:      "development",
		LogLevel: "debug",
	}
	logger, err := NewLogger(settings)
	require.NoError(t, err, "создание логгера не должно вернуть ошибку")
	require.NotNil(t, logger, "логгер не должен быть nil")
	logger.Sync()
}

// TestNewLoggerProduction проверяет создание логгера в режиме production.
func TestNewLoggerProduction(t *testing.T) {
	settings := AppSettings{
		Name:     "test-app",
		Env:      "production",
		LogLevel: "info",
	}
	logger, err := NewLogger(settings)
	require.NoError(t, err, "создание логгера не должно вернуть ошибку")
	require.NotNil(t, logger, "логгер не должен быть nil")
	logger.Sync()
}

// TestNewLoggerInvalidLevel проверяет, что невалидный уровень возвращает ошибку.
func TestNewLoggerInvalidLevel(t *testing.T) {
	settings := AppSettings{
		Name:     "test-app",
		Env:      "development",
		LogLevel: "invalid",
	}
	_, err := NewLogger(settings)
	assert.Error(t, err, "невалидный уровень логирования должен вернуть ошибку")
}

// TestItoa проверяет вспомогательную функцию преобразования int в строку.
func TestItoa(t *testing.T) {
	assert.Equal(t, "0", itoa(0))
	assert.Equal(t, "5432", itoa(5432))
	assert.Equal(t, "1", itoa(1))
	assert.Equal(t, "65535", itoa(65535))
}
