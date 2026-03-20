// Файл vibe.go реализует бизнес-логику мультимодального профилирования туриста.
// Содержит три основных метода:
//   - ProcessVoice: аудио -> Whisper STT -> LLM оси -> Embeddings -> Qdrant upsert
//   - Swipe: математический сдвиг вектора к/от сцены
//   - Finalize: Top-10 ближайших локаций из Qdrant
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/ai"
	"kudytudy-api/internal/database"
	"kudytudy-api/internal/models"
)

// swipeAlpha — коэффициент сохранения исходного вектора при свайпе.
// 0.85 означает 85% старого вектора + 15% вектора сцены (right) или
// 85% старого вектора - 15% вектора сцены (left).
const swipeAlpha = 0.85

// STTClient — интерфейс для сервиса распознавания речи.
// Реализуется VoskClient (локальный) и WhisperClient (API).
type STTClient interface {
	Transcribe(ctx context.Context, audioReader io.Reader, filename string, fileSize int64) (string, error)
}

// VibeService — сервис мультимодального vibe-профилирования.
type VibeService struct {
	stt           STTClient
	llm           *ai.LLMClient
	embeddings    *ai.EmbeddingsClient
	vibeRepo      *database.VibeRepository
	locationRepo  *database.LocationRepository
	storage       *StorageService
	demoLatencyMs int
	logger        *zap.Logger
}

// NewVibeService создаёт сервис vibe-профилирования.
// demoLatencyMs задаёт искусственную задержку в mock-режиме (по умолчанию 2000ms).
func NewVibeService(
	stt STTClient,
	llm *ai.LLMClient,
	embeddings *ai.EmbeddingsClient,
	vibeRepo *database.VibeRepository,
	locationRepo *database.LocationRepository,
	storage *StorageService,
	demoLatencyMs int,
	logger *zap.Logger,
) *VibeService {
	if demoLatencyMs <= 0 {
		demoLatencyMs = 2000
	}
	if demoLatencyMs < 800 {
		demoLatencyMs = 800
	}
	if demoLatencyMs > 5000 {
		demoLatencyMs = 5000
	}
	return &VibeService{
		stt:           stt,
		llm:           llm,
		embeddings:    embeddings,
		vibeRepo:      vibeRepo,
		locationRepo:  locationRepo,
		storage:       storage,
		demoLatencyMs: demoLatencyMs,
		logger:        logger.Named("vibe_service"),
	}
}

// ProcessVoice выполняет полный пайплайн голосового профилирования:
// 1. Сохранение аудио в MinIO
// 2. Распознавание речи через STT (Vosk/Whisper)
// 3. Извлечение осей vibe-профиля через LLM
// 4. Генерация эмбеддинга через Embeddings API
// 5. Upsert вектора в Qdrant (коллекция user_vibes)
// 6. Обновление vibe_vector_id в PostgreSQL
//
// В mock-режиме применяется controlled latency для естественной UX-анимации.
// Возвращает screenshot-ready VoiceProfileResponse с вложенными axes.
func (s *VibeService) ProcessVoice(ctx context.Context, userID uuid.UUID, audioReader io.Reader, filename string, fileSize int64) (*models.VoiceProfileResponse, error) {
	start := time.Now()

	s.logger.Info("начало голосового профилирования",
		zap.String("user_id", userID.String()),
		zap.String("filename", filename),
	)

	// Буферизация аудиоданных для переиспользования (MinIO + STT).
	audioData, err := io.ReadAll(audioReader)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения аудиоданных: %w", err)
	}
	actualSize := int64(len(audioData))

	// Шаг 1: Сохранение аудио в MinIO.
	objectName := GenerateObjectName(filename)
	if s.storage != nil {
		_, err := s.storage.Upload(ctx, objectName, bytes.NewReader(audioData), actualSize, "audio/mpeg")
		if err != nil {
			s.logger.Warn("не удалось загрузить аудио в MinIO", zap.Error(err))
		}
	}

	// Шаг 2: Распознавание речи через STT (Vosk или Whisper).
	transcription, err := s.stt.Transcribe(ctx, bytes.NewReader(audioData), filename, actualSize)
	if err != nil {
		return nil, fmt.Errorf("ошибка распознавания речи: %w", err)
	}

	s.logger.Info("транскрипция получена",
		zap.Int("text_length", len(transcription)),
	)

	if len(strings.TrimSpace(transcription)) == 0 {
		return nil, fmt.Errorf("ошибка: аудио не распознано или содержит тишину")
	}

	// Шаг 3: Извлечение осей vibe-профиля через LLM.
	axes, err := s.llm.ExtractVibeAxes(ctx, transcription)
	if err != nil {
		return nil, fmt.Errorf("ошибка извлечения осей: %w", err)
	}

	// Шаг 4: Генерация эмбеддинга.
	embeddingText := formatAxesForEmbedding(axes)
	vector, err := s.embeddings.Generate(ctx, embeddingText)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации эмбеддинга: %w", err)
	}

	// Шаг 5: Upsert вектора в Qdrant.
	payload := map[string]any{
		"stress_level":         axes.StressLevel,
		"solitude_vs_social":   axes.SolitudeVsSocial,
		"relax_vs_adrenaline":  axes.RelaxVsAdrenaline,
		"gastro_vs_nature":     axes.GastroVsNature,
		"culture_vs_adventure": axes.CultureVsAdventure,
		"vibe_summary":         axes.VibeSummary,
	}

	if err := s.vibeRepo.UpsertVibeVector(ctx, database.CollectionUserVibes, userID, vector, payload); err != nil {
		return nil, fmt.Errorf("ошибка upsert вектора: %w", err)
	}

	// Шаг 6: Обновление vibe_vector_id в PostgreSQL.
	if err := s.vibeRepo.UpdateVibeVectorID(ctx, userID, userID); err != nil {
		s.logger.Warn("не удалось обновить vibe_vector_id в PostgreSQL",
			zap.Error(err),
		)
	}

	// Шаг 7: Генерация заголовка vibe-паспорта.
	passportTitle := generatePassportTitle(axes)

	// Расчёт времени обработки.
	elapsed := time.Since(start)

	// Шаг 8: Controlled demo latency в mock-режиме.
	// Если реальная обработка заняла меньше целевого demo latency,
	// добавляем искусственную задержку для естественной анимации "ИИ думает".
	targetDuration := time.Duration(s.demoLatencyMs) * time.Millisecond
	if elapsed < targetDuration {
		remainder := targetDuration - elapsed
		s.logger.Debug("demo latency: добавлена задержка",
			zap.Duration("remainder", remainder),
		)
		time.Sleep(remainder)
		elapsed = time.Since(start)
	}

	s.logger.Info("голосовое профилирование завершено",
		zap.String("user_id", userID.String()),
		zap.Duration("processing_time", elapsed),
	)

	return &models.VoiceProfileResponse{
		Transcription: transcription,
		Axes: models.VibeAxesResponse{
			StressLevel:        axes.StressLevel,
			SolitudeVsSocial:   axes.SolitudeVsSocial,
			RelaxVsAdrenaline:  axes.RelaxVsAdrenaline,
			GastroVsNature:     axes.GastroVsNature,
			CultureVsAdventure: axes.CultureVsAdventure,
		},
		ExtractedTags:     axes.ExtractedTags,
		VibeSummary:       axes.VibeSummary,
		VibePassportTitle: passportTitle,
		VectorID:          userID.String(),
		ProcessingTimeMs:  elapsed.Milliseconds(),
	}, nil
}

// generatePassportTitle генерирует заголовок vibe-паспорта на основе осей профиля.
// Анализирует доминирующую ось и формирует описательный заголовок.
func generatePassportTitle(axes *ai.VibeAxes) string {
	// Определение доминирующей характеристики.
	if axes.GastroVsNature < -0.3 {
		return "Гастрономический путешественник"
	}
	if axes.CultureVsAdventure > 0.3 {
		return "Искатель приключений"
	}
	if axes.CultureVsAdventure < -0.3 {
		return "Культурный исследователь"
	}
	if axes.RelaxVsAdrenaline < -0.3 {
		return "Ценитель спокойствия"
	}
	if axes.RelaxVsAdrenaline > 0.3 {
		return "Адреналиновый турист"
	}
	if axes.SolitudeVsSocial < -0.3 {
		return "Созерцатель тишины"
	}
	if axes.SolitudeVsSocial > 0.3 {
		return "Душа компании"
	}
	return "Исследователь Кубани"
}

// Swipe выполняет математический сдвиг вектора пользователя на основе свайпа сцены.
// При свайпе вправо (right): вектор сдвигается к вектору сцены.
// При свайпе влево (left): вектор сдвигается от вектора сцены.
// Формула: new = normalize(alpha * user_vec +/- (1-alpha) * scene_vec)
func (s *VibeService) Swipe(ctx context.Context, userID uuid.UUID, sceneID string, direction models.SwipeDirection) error {
	s.logger.Info("обработка свайпа",
		zap.String("user_id", userID.String()),
		zap.String("scene_id", sceneID),
		zap.String("direction", string(direction)),
	)

	// Получение вектора пользователя из Qdrant.
	userVector, err := s.vibeRepo.GetVibeVector(ctx, database.CollectionUserVibes, userID)
	if err != nil {
		return fmt.Errorf("ошибка получения вектора пользователя: %w", err)
	}

	// Получение вектора сцены из Qdrant.
	sceneUUID, err := uuid.Parse(sceneID)
	if err != nil {
		return fmt.Errorf("некорректный UUID сцены: %w", err)
	}

	sceneVector, err := s.vibeRepo.GetVibeVector(ctx, database.CollectionUserVibes, sceneUUID)
	if err != nil {
		// Если вектора сцены нет в Qdrant — пропускаем свайп.
		s.logger.Warn("вектор сцены не найден в Qdrant, свайп пропущен",
			zap.String("scene_id", sceneID),
		)
		return nil
	}

	// Математический сдвиг вектора.
	newVector := shiftVector(userVector, sceneVector, direction)

	// Upsert обновлённого вектора.
	if err := s.vibeRepo.UpsertVibeVector(ctx, database.CollectionUserVibes, userID, newVector, nil); err != nil {
		return fmt.Errorf("ошибка обновления вектора после свайпа: %w", err)
	}

	s.logger.Info("свайп обработан",
		zap.String("user_id", userID.String()),
		zap.String("direction", string(direction)),
	)
	return nil
}

// Finalize выполняет финальный поиск Top-10 ближайших локаций на основе vibe-вектора.
// Берёт вектор пользователя из Qdrant (user_vibes) и ищет ближайшие в location_vibes.
// Обогащает результаты данными из PostgreSQL (preview image, tags, координаты).
func (s *VibeService) Finalize(ctx context.Context, userID uuid.UUID) (*models.FinalizeResponse, error) {
	s.logger.Info("финализация профиля",
		zap.String("user_id", userID.String()),
	)

	// Получение вектора пользователя.
	userVector, err := s.vibeRepo.GetVibeVector(ctx, database.CollectionUserVibes, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения вектора пользователя: %w", err)
	}

	// Поиск Top-10 ближайших локаций.
	recommendations, err := s.vibeRepo.SearchNearest(ctx, database.CollectionLocationVibes, userVector, 10)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска рекомендаций: %w", err)
	}

	// Обогащение рекомендаций данными из PostgreSQL.
	if len(recommendations) > 0 && s.locationRepo != nil {
		ids := make([]string, len(recommendations))
		for i, rec := range recommendations {
			ids[i] = rec.LocationID
		}

		locations, err := s.locationRepo.FindByIDs(ctx, ids)
		if err != nil {
			s.logger.Warn("ошибка обогащения рекомендаций, возвращаем базовые данные",
				zap.Error(err),
			)
		} else {
			// Построение карты для быстрого поиска.
			locMap := make(map[string]models.Location, len(locations))
			for _, loc := range locations {
				locMap[loc.ID.String()] = loc
			}

			for i, rec := range recommendations {
				if loc, ok := locMap[rec.LocationID]; ok {
					recommendations[i].Name = loc.Name
					recommendations[i].Category = loc.Category
					recommendations[i].DescriptionShort = loc.DescriptionShort
					recommendations[i].Tags = loc.Tags
					recommendations[i].PreviewImageURL = loc.PreviewImageURL
					recommendations[i].Latitude = loc.Latitude
					recommendations[i].Longitude = loc.Longitude
					recommendations[i].DensityLevel = string(loc.DensityLevel)
					if loc.SplatURL != nil {
						recommendations[i].SplatURL = *loc.SplatURL
					}
				}
			}
		}
	}

	return &models.FinalizeResponse{
		Recommendations: recommendations,
		TotalFound:      len(recommendations),
	}, nil
}

// GetScenes возвращает все сцены свайпа для анкеты.
func (s *VibeService) GetScenes(ctx context.Context) ([]models.SwipeScene, error) {
	return s.vibeRepo.GetAllScenes(ctx)
}

// shiftVector выполняет математический сдвиг вектора пользователя.
// Формула для right: new = normalize(alpha * user + (1-alpha) * scene)
// Формула для left:  new = normalize(alpha * user - (1-alpha) * scene)
func shiftVector(userVec, sceneVec []float32, direction models.SwipeDirection) []float32 {
	size := len(userVec)
	if len(sceneVec) < size {
		size = len(sceneVec)
	}

	result := make([]float32, size)
	beta := 1.0 - swipeAlpha

	for i := 0; i < size; i++ {
		userVal := float64(userVec[i]) * swipeAlpha
		sceneVal := float64(sceneVec[i]) * beta

		if direction == models.SwipeRight {
			result[i] = float32(userVal + sceneVal)
		} else {
			result[i] = float32(userVal - sceneVal)
		}
	}

	// Нормализация результата до единичной длины.
	return normalizeVector(result)
}

// normalizeVector нормализует вектор до единичной длины (L2 norm).
func normalizeVector(vec []float32) []float32 {
	var norm float64
	for _, v := range vec {
		norm += float64(v) * float64(v)
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		return vec
	}

	result := make([]float32, len(vec))
	for i, v := range vec {
		result[i] = float32(float64(v) / norm)
	}
	return result
}

// formatAxesForEmbedding форматирует оси vibe-профиля в текст для эмбеддинга.
// Текст содержит числовые значения осей и теги для генерации семантического вектора.
func formatAxesForEmbedding(axes *ai.VibeAxes) string {
	tagsJSON, _ := json.Marshal(axes.ExtractedTags)
	return fmt.Sprintf(
		"Турист: стресс=%.2f уединение=%.2f релакс=%.2f гастро=%.2f культура=%.2f теги=%s резюме: %s",
		axes.StressLevel,
		axes.SolitudeVsSocial,
		axes.RelaxVsAdrenaline,
		axes.GastroVsNature,
		axes.CultureVsAdventure,
		string(tagsJSON),
		axes.VibeSummary,
	)
}
