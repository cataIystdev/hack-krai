// Package ai предоставляет клиенты для внешних AI-сервисов.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	elevenLabsBaseURL  = "https://api.elevenlabs.io/v1"
	elevenLabsTimeout  = 30 * time.Second
	defaultModelID     = "eleven_multilingual_v2"
	defaultOutputFormat = "mp3_44100_128"
)

// ElevenLabsClient — HTTP-клиент для ElevenLabs Text-to-Speech API.
type ElevenLabsClient struct {
	apiKey  string
	voiceID string
	client  *http.Client
}

// NewElevenLabsClient создаёт нового клиента ElevenLabs TTS.
// voiceID — идентификатор голоса из библиотеки ElevenLabs.
func NewElevenLabsClient(apiKey, voiceID string) *ElevenLabsClient {
	return &ElevenLabsClient{
		apiKey:  apiKey,
		voiceID: voiceID,
		client:  &http.Client{Timeout: elevenLabsTimeout},
	}
}

// ttsRequest — тело запроса к ElevenLabs TTS endpoint.
type ttsRequest struct {
	Text          string      `json:"text"`
	ModelID       string      `json:"model_id"`
	VoiceSettings voiceConfig `json:"voice_settings"`
}

// voiceConfig — настройки голоса: стабильность и схожесть с оригинальным голосом.
type voiceConfig struct {
	Stability       float64 `json:"stability"`
	SimilarityBoost float64 `json:"similarity_boost"`
	Style           float64 `json:"style"`
	UseSpeakerBoost bool    `json:"use_speaker_boost"`
}

// TextToSpeech конвертирует текст в mp3-аудио через ElevenLabs API.
// Возвращает raw bytes mp3-файла.
func (c *ElevenLabsClient) TextToSpeech(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("text must not be empty")
	}

	body := ttsRequest{
		Text:    text,
		ModelID: defaultModelID,
		VoiceSettings: voiceConfig{
			Stability:       0.5,
			SimilarityBoost: 0.75,
			Style:           0.3,
			UseSpeakerBoost: true,
		},
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/text-to-speech/%s?output_format=%s", elevenLabsBaseURL, c.voiceID, defaultOutputFormat)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: create request: %w", err)
	}

	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("elevenlabs: API error %d: %s", resp.StatusCode, string(errBody))
	}

	mp3Bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: read response: %w", err)
	}

	if len(mp3Bytes) == 0 {
		return nil, fmt.Errorf("elevenlabs: empty audio response")
	}

	return mp3Bytes, nil
}

// EstimateDurationSec оценивает длительность аудио по количеству символов.
// Средняя скорость речи ~150 слов/мин, ~5 символов на слово.
func EstimateDurationSec(text string) int {
	wordsEstimate := float64(len(text)) / 5.0
	seconds := int(wordsEstimate / 150.0 * 60.0)
	if seconds < 5 {
		return 5
	}
	return seconds
}
