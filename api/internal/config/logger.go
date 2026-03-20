// Файл logger.go отвечает за инициализацию структурированного логгера Zap.
// Поддерживает два режима: development (console, debug) и production (JSON, info).
// Уровень логирования настраивается через конфигурацию приложения.
package config

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger создаёт и настраивает экземпляр Zap-логгера на основе параметров приложения.
// Параметр settings содержит окружение (development/production) и уровень логирования.
// В режиме development используется человекочитаемый формат вывода (console).
// В режиме production — компактный JSON-формат для парсинга системами мониторинга.
// Возвращает настроенный логгер и функцию для синхронизации буферов при завершении.
func NewLogger(settings AppSettings) (*zap.Logger, error) {
	// Определение уровня логирования из строковой конфигурации.
	level, err := parseLogLevel(settings.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("ошибка разбора уровня логирования: %w", err)
	}

	var cfg zap.Config

	// Выбор конфигурации в зависимости от окружения.
	// Development: цветной вывод, вызовы caller, stacktrace на уровне warn.
	// Production: JSON-вывод, stacktrace на уровне error.
	if settings.Env == "development" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Применение уровня логирования из конфигурации.
	cfg.Level = zap.NewAtomicLevelAt(level)

	// Добавление имени приложения в каждую запись лога.
	cfg.InitialFields = map[string]interface{}{
		"service": settings.Name,
	}

	// Сборка логгера с указанной конфигурацией.
	// AddCallerSkip(0) обеспечивает корректное отображение вызывающего файла/строки.
	logger, err := cfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания логгера: %w", err)
	}

	return logger, nil
}

// parseLogLevel преобразует строковое представление уровня логирования
// в тип zapcore.Level. Поддерживает значения: debug, info, warn, error.
// Возвращает ошибку для неизвестных значений.
func parseLogLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("неизвестный уровень логирования: %s", level)
	}
}
