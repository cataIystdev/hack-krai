// Файл bootstrap.go реализует создание демонстрационных пользователей при запуске API.
// Обеспечивает идемпотентное создание предустановленных аккаунтов для тестирования
// и интеграции фронтенда. При повторном запуске существующие аккаунты не затрагиваются.
package services

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"kudytudy-api/internal/database"
)

// DemoUser — конфигурация демонстрационного пользователя для bootstrap.
type DemoUser struct {
	// Email — адрес электронной почты.
	Email string

	// Password — пароль в открытом виде (хэшируется при создании).
	Password string

	// DisplayName — отображаемое имя.
	DisplayName string

	// Role — роль пользователя (tourist, host, b2g_admin).
	Role string
}

// defaultDemoUsers — список предустановленных демонстрационных пользователей.
// Создаются при первом запуске API для снижения friction при интеграции фронтенда.
var defaultDemoUsers = []DemoUser{
	{
		Email:       "demo@deepkrai.ru",
		Password:    "demo1234",
		DisplayName: "Демо Турист",
		Role:        "tourist",
	},
	{
		Email:       "host@deepkrai.ru",
		Password:    "host1234",
		DisplayName: "Демо Хост",
		Role:        "host",
	},
}

// BootstrapDemoUsers создаёт демонстрационных пользователей, если они ещё не существуют.
// Операция идемпотентна: при повторном вызове уже существующие аккаунты пропускаются.
// Возвращает количество созданных пользователей.
func BootstrapDemoUsers(ctx context.Context, userRepo *database.UserRepository, logger *zap.Logger) (int, error) {
	created := 0

	for _, demo := range defaultDemoUsers {
		// Проверка существования пользователя по email.
		_, err := userRepo.FindByEmail(ctx, demo.Email)
		if err == nil {
			// Пользователь уже существует — пропускаем.
			logger.Debug("демо-пользователь уже существует, пропуск",
				zap.String("email", demo.Email),
			)
			continue
		}
		if !errors.Is(err, database.ErrUserNotFound) {
			// Ошибка при проверке — пропускаем, но логируем.
			logger.Warn("ошибка проверки демо-пользователя",
				zap.String("email", demo.Email),
				zap.Error(err),
			)
			continue
		}

		// Хэширование пароля.
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(demo.Password), BcryptCost)
		if err != nil {
			return created, fmt.Errorf("ошибка хэширования пароля демо-пользователя %s: %w", demo.Email, err)
		}

		// Создание пользователя в БД.
		_, err = userRepo.Create(ctx, demo.Email, string(hashedPassword), demo.DisplayName, demo.Role)
		if err != nil {
			if errors.Is(err, database.ErrUserAlreadyExists) {
				// Конкурентное создание — уже существует.
				logger.Debug("демо-пользователь создан конкурентно",
					zap.String("email", demo.Email),
				)
				continue
			}
			logger.Warn("ошибка создания демо-пользователя",
				zap.String("email", demo.Email),
				zap.Error(err),
			)
			continue
		}

		created++
		logger.Info("демо-пользователь создан",
			zap.String("email", demo.Email),
			zap.String("role", demo.Role),
		)
	}

	if created > 0 {
		logger.Info("bootstrap демо-пользователей завершён",
			zap.Int("created", created),
			zap.Int("total", len(defaultDemoUsers)),
		)
	}

	return created, nil
}
