// Файл auth.go реализует сервис аутентификации и авторизации.
// Предоставляет методы регистрации, авторизации и обновления токенов.
// Использует bcrypt для хэширования паролей и JWTService для работы с токенами.
package services

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"deep-krai-api/internal/database"
)

// BcryptCost — стоимость хэширования bcrypt.
// Значение 12 обеспечивает баланс между безопасностью и производительностью.
const BcryptCost = 12

// Ошибки сервиса аутентификации.
var (
	// ErrInvalidCredentials — неверный email или пароль.
	ErrInvalidCredentials = errors.New("неверный email или пароль")

	// ErrEmailAlreadyRegistered — email уже зарегистрирован.
	ErrEmailAlreadyRegistered = errors.New("email уже зарегистрирован")

	// ErrPasswordTooShort — пароль слишком короткий.
	ErrPasswordTooShort = errors.New("пароль должен содержать не менее 8 символов")

	// ErrInvalidEmail — некорректный формат email.
	ErrInvalidEmail = errors.New("некорректный формат email")
)

// AuthService — сервис аутентификации, управляющий регистрацией,
// авторизацией и обновлением JWT-токенов.
type AuthService struct {
	userRepo   *database.UserRepository
	jwtService *JWTService
	logger     *zap.Logger
}

// NewAuthService создаёт экземпляр сервиса аутентификации.
// Принимает репозиторий пользователей, JWT-сервис и логгер.
func NewAuthService(userRepo *database.UserRepository, jwtService *JWTService, logger *zap.Logger) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
		logger:     logger,
	}
}

// AuthResult — результат операции авторизации.
// Содержит пару токенов и публичное представление пользователя.
type AuthResult struct {
	// Tokens — пара JWT-токенов (access + refresh).
	Tokens *TokenPair `json:"tokens"`

	// User — публичные данные пользователя (без password_hash).
	User interface{} `json:"user"`
}

// Register выполняет регистрацию нового пользователя.
// Процесс: валидация входных данных, хэширование пароля (bcrypt, cost=12),
// создание записи в БД, генерация пары JWT-токенов.
func (s *AuthService) Register(ctx context.Context, email, password, displayName, role string) (*AuthResult, error) {
	// Валидация email.
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	// Валидация пароля.
	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}

	// Хэширование пароля через bcrypt.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		s.logger.Error("ошибка хэширования пароля", zap.Error(err))
		return nil, fmt.Errorf("ошибка хэширования пароля: %w", err)
	}

	// Если displayName пуст, используем часть email до @.
	if displayName == "" {
		displayName = extractNameFromEmail(email)
	}

	// Создание пользователя в БД.
	user, err := s.userRepo.Create(ctx, email, string(hashedPassword), displayName, role)
	if err != nil {
		if errors.Is(err, database.ErrUserAlreadyExists) {
			return nil, ErrEmailAlreadyRegistered
		}
		return nil, fmt.Errorf("ошибка регистрации: %w", err)
	}

	// Генерация пары JWT-токенов.
	tokens, err := s.jwtService.GenerateTokenPair(user.ID.String(), string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации токенов: %w", err)
	}

	s.logger.Info("пользователь зарегистрирован",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)

	return &AuthResult{
		Tokens: tokens,
		User:   user.ToPublicResponse(),
	}, nil
}

// Login выполняет авторизацию пользователя по email и паролю.
// Процесс: поиск пользователя по email, проверка пароля (bcrypt compare),
// генерация пары JWT-токенов.
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	// Поиск пользователя по email.
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			s.logger.Debug("попытка входа с несуществующим email",
				zap.String("email", email),
			)
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	// Сверка пароля с хэшем из БД.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Debug("неверный пароль при входе",
			zap.String("email", email),
		)
		return nil, ErrInvalidCredentials
	}

	// Генерация пары JWT-токенов.
	tokens, err := s.jwtService.GenerateTokenPair(user.ID.String(), string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации токенов: %w", err)
	}

	s.logger.Info("пользователь авторизован",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)

	return &AuthResult{
		Tokens: tokens,
		User:   user.ToPublicResponse(),
	}, nil
}

// RefreshTokens обновляет пару JWT-токенов по refresh-токену.
// Процесс: валидация refresh-токена, извлечение user_id и role,
// генерация новой пары access + refresh.
func (s *AuthService) RefreshTokens(refreshToken string) (*TokenPair, error) {
	// Валидация refresh-токена.
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logger.Debug("невалидный refresh-токен",
			zap.Error(err),
		)
		return nil, fmt.Errorf("невалидный refresh-токен: %w", err)
	}

	// Генерация новой пары токенов.
	tokens, err := s.jwtService.GenerateTokenPair(claims.UserID, claims.Role)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации токенов: %w", err)
	}

	s.logger.Debug("токены обновлены",
		zap.String("user_id", claims.UserID),
	)

	return tokens, nil
}

// isValidEmail выполняет базовую валидацию формата email.
// Проверяет наличие символа @, текста до и после @, наличие точки в домене.
func isValidEmail(email string) bool {
	if len(email) < 5 || len(email) > 255 {
		return false
	}
	atIdx := -1
	for i, ch := range email {
		if ch == '@' {
			if atIdx != -1 {
				return false // Больше одного @.
			}
			atIdx = i
		}
	}
	if atIdx < 1 || atIdx >= len(email)-1 {
		return false
	}
	domain := email[atIdx+1:]
	hasDot := false
	for _, ch := range domain {
		if ch == '.' {
			hasDot = true
		}
	}
	return hasDot
}

// extractNameFromEmail извлекает часть email до символа @ для использования
// в качестве display_name по умолчанию.
func extractNameFromEmail(email string) string {
	for i, ch := range email {
		if ch == '@' {
			return email[:i]
		}
	}
	return email
}
