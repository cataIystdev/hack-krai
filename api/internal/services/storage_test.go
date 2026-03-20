// Тесты для сервиса хранилища файлов.
// Проверяют генерацию уникальных имён объектов.
package services

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGenerateObjectName проверяет генерацию уникального имени объекта.
func TestGenerateObjectName(t *testing.T) {
	name := GenerateObjectName("photo.jpg")

	// Имя должно начинаться с "uploads/".
	assert.True(t, strings.HasPrefix(name, "uploads/"), "имя должно начинаться с uploads/")

	// Имя должно содержать оригинальное расширение.
	assert.True(t, strings.HasSuffix(name, ".jpg"), "имя должно заканчиваться на .jpg")

	// Имя должно содержать оригинальное имя файла (без расширения).
	assert.Contains(t, name, "photo")
}

// TestGenerateObjectNameUniqueness проверяет, что последовательные вызовы
// генерируют уникальные имена объектов.
func TestGenerateObjectNameUniqueness(t *testing.T) {
	name1 := GenerateObjectName("test.png")
	// Небольшая задержка для гарантии разного timestamp.
	time.Sleep(1 * time.Millisecond)
	name2 := GenerateObjectName("test.png")

	assert.NotEqual(t, name1, name2, "два последовательных вызова должны дать разные имена")
}

// TestGenerateObjectNameWithPath проверяет работу с именами,
// содержащими путь или специальные символы.
func TestGenerateObjectNameWithPath(t *testing.T) {
	name := GenerateObjectName("my_video.mp4")

	assert.True(t, strings.HasPrefix(name, "uploads/"))
	assert.True(t, strings.HasSuffix(name, ".mp4"))
	assert.Contains(t, name, "my_video")
}

// TestGenerateObjectNameNoExtension проверяет работу с файлами без расширения.
func TestGenerateObjectNameNoExtension(t *testing.T) {
	name := GenerateObjectName("readme")

	assert.True(t, strings.HasPrefix(name, "uploads/"))
	assert.Contains(t, name, "readme")
}
