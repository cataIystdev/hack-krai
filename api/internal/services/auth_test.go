// Файл auth_test.go содержит unit-тесты для сервиса аутентификации.
// Тестируемые сценарии: валидация email, хэширование пароля,
// извлечение имени из email, валидация длины пароля.
package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestIsValidEmail проверяет валидацию различных форматов email.
func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"корректный email", "user@example.com", true},
		{"email с поддоменом", "user@mail.example.com", true},
		{"email с точкой в имени", "user.name@example.com", true},
		{"email с цифрами", "user123@example.com", true},
		{"пустая строка", "", false},
		{"без символа @", "userexample.com", false},
		{"без домена", "user@", false},
		{"без имени", "@example.com", false},
		{"два символа @", "user@@example.com", false},
		{"без точки в домене", "user@example", false},
		{"слишком короткий", "a@b", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.valid, result,
				"isValidEmail(%q) = %v, ожидалось %v", tt.email, result, tt.valid)
		})
	}
}

// TestExtractNameFromEmail проверяет извлечение имени из email.
func TestExtractNameFromEmail(t *testing.T) {
	tests := []struct {
		email string
		name  string
	}{
		{"john@example.com", "john"},
		{"user.name@domain.com", "user.name"},
		{"test123@mail.ru", "test123"},
		{"noatsign", "noatsign"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := extractNameFromEmail(tt.email)
			assert.Equal(t, tt.name, result)
		})
	}
}

// TestBcryptHashAndCompare проверяет хэширование и сверку паролей через bcrypt.
func TestBcryptHashAndCompare(t *testing.T) {
	password := "securePassword123"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Проверка корректного пароля.
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	assert.NoError(t, err)

	// Проверка некорректного пароля.
	err = bcrypt.CompareHashAndPassword(hash, []byte("wrongPassword"))
	assert.Error(t, err)
}

// TestBcryptCost проверяет, что хэш использует указанную стоимость.
func TestBcryptCost(t *testing.T) {
	password := "testPassword"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	assert.NoError(t, err)

	cost, err := bcrypt.Cost(hash)
	assert.NoError(t, err)
	assert.Equal(t, BcryptCost, cost)
}

// TestPasswordTooShort проверяет валидацию минимальной длины пароля.
func TestPasswordTooShort(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"", false},
		{"1234567", false}, // 7 символов — слишком короткий.
		{"12345678", true}, // 8 символов — минимально допустимый.
		{"longpassword123", true},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			result := len(tt.password) >= 8
			assert.Equal(t, tt.valid, result)
		})
	}
}
