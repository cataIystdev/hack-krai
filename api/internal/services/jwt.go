// Файл jwt.go реализует сервис генерации и валидации JSON Web Tokens.
// Использует HMAC-SHA256 для подписи токенов. Поддерживает два типа токенов:
// access (короткоживущий, для авторизации запросов) и refresh (долгоживущий,
// для обновления access-токена без повторного ввода пароля).
package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"kudytudy-api/internal/config"
)

// Типы токенов для различения access и refresh при валидации.
const (
	// TokenTypeAccess — тип access-токена. Используется для авторизации запросов.
	TokenTypeAccess = "access"

	// TokenTypeRefresh — тип refresh-токена. Используется для обновления access-токена.
	TokenTypeRefresh = "refresh"
)

// Ошибки JWT-сервиса.
var (
	// ErrInvalidToken — токен имеет некорректный формат или невалидную подпись.
	ErrInvalidToken = errors.New("невалидный токен")

	// ErrExpiredToken — срок действия токена истёк.
	ErrExpiredToken = errors.New("срок действия токена истёк")

	// ErrInvalidTokenType — тип токена не соответствует ожидаемому.
	ErrInvalidTokenType = errors.New("некорректный тип токена")
)

// TokenClaims — структура claims для JWT-токена КудыТуды.
// Содержит идентификатор пользователя, роль и тип токена (access/refresh).
// Встраивает стандартные RegisteredClaims (iss, sub, exp, iat).
type TokenClaims struct {
	// UserID — уникальный идентификатор пользователя (UUID строкой).
	UserID string `json:"user_id"`

	// Role — роль пользователя (tourist, host, b2g_admin).
	Role string `json:"role"`

	// TokenType — тип токена: "access" или "refresh".
	TokenType string `json:"token_type"`

	// RegisteredClaims — стандартные JWT claims (iss, sub, exp, iat, nbf).
	jwt.RegisteredClaims
}

// TokenPair — пара токенов (access + refresh), возвращаемая при авторизации.
type TokenPair struct {
	// AccessToken — JWT для авторизации запросов. Короткоживущий.
	AccessToken string `json:"access_token"`

	// RefreshToken — JWT для обновления access-токена. Долгоживущий.
	RefreshToken string `json:"refresh_token"`

	// AccessTokenExpiresAt — время истечения access-токена (Unix timestamp).
	AccessTokenExpiresAt int64 `json:"access_token_expires_at"`

	// RefreshTokenExpiresAt — время истечения refresh-токена (Unix timestamp).
	RefreshTokenExpiresAt int64 `json:"refresh_token_expires_at"`
}

// JWTService — сервис генерации и валидации JWT-токенов.
// Инкапсулирует секретный ключ, настройки TTL и логирование.
type JWTService struct {
	secretKey  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	issuer     string
	logger     *zap.Logger
}

// NewJWTService создаёт экземпляр JWT-сервиса из конфигурации.
// Секретный ключ преобразуется в байтовый массив для HMAC-SHA256.
func NewJWTService(cfg config.JWTConfig, logger *zap.Logger) *JWTService {
	return &JWTService{
		secretKey:  []byte(cfg.SecretKey),
		accessTTL:  time.Duration(cfg.AccessTokenTTLMinutes) * time.Minute,
		refreshTTL: time.Duration(cfg.RefreshTokenTTLHours) * time.Hour,
		issuer:     cfg.Issuer,
		logger:     logger,
	}
}

// GenerateAccessToken создаёт access JWT-токен для указанного пользователя.
// Access-токен имеет короткий TTL и используется для авторизации API-запросов.
func (s *JWTService) GenerateAccessToken(userID, role string) (string, int64, error) {
	expiresAt := time.Now().Add(s.accessTTL)
	claims := TokenClaims{
		UserID:    userID,
		Role:      role,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		s.logger.Error("ошибка подписи access-токена",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return "", 0, fmt.Errorf("ошибка генерации access-токена: %w", err)
	}

	return tokenString, expiresAt.Unix(), nil
}

// GenerateRefreshToken создаёт refresh JWT-токен для указанного пользователя.
// Refresh-токен имеет длинный TTL и используется для получения нового access-токена.
func (s *JWTService) GenerateRefreshToken(userID, role string) (string, int64, error) {
	expiresAt := time.Now().Add(s.refreshTTL)
	claims := TokenClaims{
		UserID:    userID,
		Role:      role,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		s.logger.Error("ошибка подписи refresh-токена",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return "", 0, fmt.Errorf("ошибка генерации refresh-токена: %w", err)
	}

	return tokenString, expiresAt.Unix(), nil
}

// GenerateTokenPair создаёт пару токенов (access + refresh) для пользователя.
// Используется при регистрации, авторизации и обновлении токенов.
func (s *JWTService) GenerateTokenPair(userID, role string) (*TokenPair, error) {
	accessToken, accessExp, err := s.GenerateAccessToken(userID, role)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExp, err := s.GenerateRefreshToken(userID, role)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("сгенерирована пара токенов",
		zap.String("user_id", userID),
		zap.String("role", role),
	)

	return &TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
	}, nil
}

// ValidateToken проверяет подпись и claims JWT-токена.
// Возвращает распарсенные claims при успешной валидации.
// Проверяет: подпись (HMAC-SHA256), срок действия, issuer, наличие exp.
func (s *JWTService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&TokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Проверка алгоритма подписи — допускается только HMAC.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный алгоритм подписи: %v", token.Header["alg"])
			}
			return s.secretKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateAccessToken проверяет, что токен является валидным access-токеном.
// Дополнительно проверяет поле token_type.
func (s *JWTService) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeAccess {
		return nil, ErrInvalidTokenType
	}
	return claims, nil
}

// ValidateRefreshToken проверяет, что токен является валидным refresh-токеном.
// Дополнительно проверяет поле token_type.
func (s *JWTService) ValidateRefreshToken(tokenString string) (*TokenClaims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidTokenType
	}
	return claims, nil
}
