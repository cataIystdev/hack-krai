package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// EstimateDurationSec — чистая функция, тестируем напрямую
// ---------------------------------------------------------------------------

func TestEstimateDurationSec_Empty(t *testing.T) {
	// Минимум 5 секунд даже для пустой строки
	dur := EstimateDurationSec("")
	assert.Equal(t, 5, dur)
}

func TestEstimateDurationSec_ShortText(t *testing.T) {
	// Очень короткий текст → минимум 5 секунд
	dur := EstimateDurationSec("Привет!")
	assert.Equal(t, 5, dur)
}

func TestEstimateDurationSec_TypicalStory(t *testing.T) {
	// Типичная история ~200 символов → ~3-4 секунды (будет min=5)
	text := "Абрау-Дюрсо — одна из точек вашего маршрута по Краснодарскому краю. " +
		"Знаменитая винодельня на берегу озера. Это место категории winery, " +
		"поэтому здесь особенно хорошо задержаться на 90 минут."
	dur := EstimateDurationSec(text)
	// ~200 символов / 5 символов/слово = 40 слов; 40/150*60 = ~16 сек
	assert.GreaterOrEqual(t, dur, 5)
	assert.LessOrEqual(t, dur, 60)
}

func TestEstimateDurationSec_LongText(t *testing.T) {
	// Длинный текст → заметно больше 5 секунд
	long := make([]byte, 3000)
	for i := range long {
		long[i] = 'а'
	}
	dur := EstimateDurationSec(string(long))
	// 3000 / 5 = 600 слов; 600/150*60 = 240 сек
	assert.Greater(t, dur, 60)
}

func TestEstimateDurationSec_Monotonic(t *testing.T) {
	// Чем длиннее текст — тем больше длительность
	short := "Короткий текст."
	medium := "Это средний текст, который немного длиннее чем предыдущий вариант для данного теста."
	longText := medium + medium + medium

	durShort := EstimateDurationSec(short)
	durMedium := EstimateDurationSec(medium)
	durLong := EstimateDurationSec(longText)

	assert.LessOrEqual(t, durShort, durMedium)
	assert.Less(t, durMedium, durLong)
}

// ---------------------------------------------------------------------------
// NewElevenLabsClient — конструктор
// ---------------------------------------------------------------------------

func TestNewElevenLabsClient_NotNil(t *testing.T) {
	client := NewElevenLabsClient("test-api-key", "test-voice-id")
	assert.NotNil(t, client)
}

func TestNewElevenLabsClient_StoresValues(t *testing.T) {
	client := NewElevenLabsClient("sk_mykey123", "voice_abc")
	assert.Equal(t, "sk_mykey123", client.apiKey)
	assert.Equal(t, "voice_abc", client.voiceID)
	assert.NotNil(t, client.client)
}

func TestNewElevenLabsClient_DefaultTimeout(t *testing.T) {
	client := NewElevenLabsClient("key", "voice")
	assert.Equal(t, elevenLabsTimeout, client.client.Timeout)
}
