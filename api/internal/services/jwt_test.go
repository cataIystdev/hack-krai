// Файл jwt_test.go содержит unit-тесты для JWT-сервиса.
// Тестируемые сценарии: генерация токенов, валидация, истечение срока,
// неверная подпись, некорректный тип токена.
package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"deep-krai-api/internal/config"
)

// testJWTConfig возвращает конфигурацию JWT для тестов.
func testJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		SecretKey:             "test-secret-key-for-jwt-testing-32chars-minimum",
		AccessTokenTTLMinutes: 15,
		RefreshTokenTTLHours:  168,
		Issuer:                "test-issuer",
	}
}

// testLogger возвращает логгер без вывода для тестов.
func testLogger() *zap.Logger {
	return zap.NewNop()
}

// TestNewJWTService проверяет создание экземпляра JWT-сервиса.
func TestNewJWTService(t *testing.T) {
	cfg := testJWTConfig()
	svc := NewJWTService(cfg, testLogger())

	assert.NotNil(t, svc)
	assert.Equal(t, []byte(cfg.SecretKey), svc.secretKey)
	assert.Equal(t, time.Duration(cfg.AccessTokenTTLMinutes)*time.Minute, svc.accessTTL)
	assert.Equal(t, time.Duration(cfg.RefreshTokenTTLHours)*time.Hour, svc.refreshTTL)
	assert.Equal(t, cfg.Issuer, svc.issuer)
}

// TestGenerateAccessToken проверяет генерацию access-токена.
func TestGenerateAccessToken(t *testing.T) {
	svc := NewJWTService(testJWTConfig(), testLogger())

	token, exp, err := svc.GenerateAccessToken("user-123", "tourist")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, exp, time.Now().Unix())
}

// TestGenerateRefreshToken проверяет генерацию refresh-токена.
func TestGenerateRefreshToken(t *testing.T) {
	svc := NewJWTService(testJWTConfig(), testLogger())

	token, exp, err := svc.GenerateRefreshToken("user-456", "host")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, exp, time.Now().Unix())
}

// TestGenerateTokenPair проверяет генерацию пары токенов.
func TestGenerateTokenPair(t *testing.T) {
	svc := NewJWTService(testJWTConfig(), testLogger())

	pair, err := svc.GenerateTokenPair("user-789", "b2g_admin")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Greater(t, pair.AccessTokenExpiresAt, time.Now().Unix())
	assert.Greater(t, pair.RefreshTokenExpiresAt, time.Now().Unix())
	// Refresh должен истекать позже, чем access.
	assert.Greater(t, pair.RefreshTokenExpiresAt, pair.AccessTokenExpiresAt)
}

// TestValidateAccessToken проверяет успешную валидацию access-токена.
func TestValidateAccessToken(t *testing.T) {
	svc := NewJWTService(testJWTConfig(), testLogger())

	token, _, err := svc.GenerateAccessToken("user-123", "tourist")
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "tourist", claims.Role)
	assert.Equal(t, TokenTypeAccess, claims.TokenType)
	assert.Equal(t, "test-issuer", claims.Issuer)
}

// TestValidateRefreshToken проверяет успешную валидацию refresh-токена.
func TestValidateRefreshToken(t *testing.T) {
	svc := NewJWTService(testJWTConfig(), testLogger())

	token, _, err := svc.GenerateRefreshToken("user-456", "host")
	require.NoError(t, err)

	claims, err := svc.ValidateRefreshToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-456", claims.UserID)
	assert.Equal(t, "host", claims.Role)
	assert.Equal(t, TokenTypeRefresh, claims.TokenType)
}

// TestValidateAccessTokenWithRefreshToken проверяет, что refresh-токен
// не проходит валидацию как access-токен.
func TestValidateAccessTokenWithRefreshToken(t *testing.T) {
	svc := NewJWTService(testJWTConfig(), testLogger())

	refreshToken, _, err := svc.GenerateRefreshToken("user-123", "tourist")
	require.NoError(t, err)

	_, err = svc.ValidateAccessToken(refreshToken)
	assert.ErrorIs(t, err, ErrInvalidTokenType)
}

// TestValidateRefreshTokenWithAccessToken проверяет, что access-токен
// не проходит валидацию как refresh-токен.
func TestValidateRefreshTokenWithAccessToken(t *testing.T) {
	svc := NewJWTService(testJWTConfig(), testLogger())

	accessToken, _, err := svc.GenerateAccessToken("user-123", "tourist")
	require.NoError(t, err)

	_, err = svc.ValidateRefreshToken(accessToken)
	assert.ErrorIs(t, err, ErrInvalidTokenType)
}

// TestValidateInvalidToken проверяет отклонение некорректного токена.
func TestValidateInvalidToken(t *testing.T) {
	svc := NewJWTService(testJWTConfig(), testLogger())

	_, err := svc.ValidateToken("invalid.token.string")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

// TestValidateTokenWithWrongSecret проверяет отклонение токена,
// подписанного другим секретом.
func TestValidateTokenWithWrongSecret(t *testing.T) {
	// Генерация токена с одним секретом.
	cfg1 := testJWTConfig()
	svc1 := NewJWTService(cfg1, testLogger())

	token, _, err := svc1.GenerateAccessToken("user-123", "tourist")
	require.NoError(t, err)

	// Попытка валидации с другим секретом.
	cfg2 := testJWTConfig()
	cfg2.SecretKey = "another-secret-key-that-is-32-chars-long"
	svc2 := NewJWTService(cfg2, testLogger())

	_, err = svc2.ValidateToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

// TestValidateExpiredToken проверяет отклонение истекшего токена.
func TestValidateExpiredToken(t *testing.T) {
	cfg := testJWTConfig()
	svc := NewJWTService(cfg, testLogger())

	// Создание токена с истекшим сроком вручную.
	claims := TokenClaims{
		UserID:    "user-123",
		Role:      "tourist",
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	_, err = svc.ValidateToken(tokenString)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

// TestValidateTokenWithWrongIssuer проверяет отклонение токена
// с некорректным issuer.
func TestValidateTokenWithWrongIssuer(t *testing.T) {
	cfg := testJWTConfig()
	svc := NewJWTService(cfg, testLogger())

	claims := TokenClaims{
		UserID:    "user-123",
		Role:      "tourist",
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wrong-issuer",
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	_, err = svc.ValidateToken(tokenString)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

// TestTokenClaimsContent проверяет содержимое claims сгенерированного токена.
func TestTokenClaimsContent(t *testing.T) {
	svc := NewJWTService(testJWTConfig(), testLogger())

	pair, err := svc.GenerateTokenPair("user-abc-def", "host")
	require.NoError(t, err)

	// Проверка claims access-токена.
	accessClaims, err := svc.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "user-abc-def", accessClaims.UserID)
	assert.Equal(t, "host", accessClaims.Role)
	assert.Equal(t, "user-abc-def", accessClaims.Subject)

	// Проверка claims refresh-токена.
	refreshClaims, err := svc.ValidateRefreshToken(pair.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, "user-abc-def", refreshClaims.UserID)
	assert.Equal(t, "host", refreshClaims.Role)
}
