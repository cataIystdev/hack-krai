// Файл onboarding.go реализует бизнес-логику Zero-UI онбординга хостов.
// Содержит pipeline обработки голосового описания и медиафайлов
// для автоматического создания черновика локации (GDD Feature 5).
// Pipeline: аудио -> STT -> LLM -> embedding -> location draft -> PostgreSQL.
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kudytudy-api/internal/ai"
	"kudytudy-api/internal/database"
	"kudytudy-api/internal/models"
)

// OnboardingService — сервис Zero-UI онбординга хостов.
// Управляет pipeline создания локации из голосового описания и медиа.
type OnboardingService struct {
	stt          STTClient
	llm          *ai.LLMClient
	embeddings   *ai.EmbeddingsClient
	taskRepo     *database.AITaskRepository
	locationRepo *database.LocationRepository
	vibeRepo     *database.VibeRepository
	storage      *StorageService
	demoLatency  int
	logger       *zap.Logger
}

// NewOnboardingService создаёт экземпляр сервиса онбординга.
func NewOnboardingService(
	stt STTClient,
	llm *ai.LLMClient,
	embeddings *ai.EmbeddingsClient,
	taskRepo *database.AITaskRepository,
	locationRepo *database.LocationRepository,
	vibeRepo *database.VibeRepository,
	storage *StorageService,
	demoLatencyMs int,
	logger *zap.Logger,
) *OnboardingService {
	if demoLatencyMs <= 0 {
		demoLatencyMs = 2000
	}
	return &OnboardingService{
		stt:          stt,
		llm:          llm,
		embeddings:   embeddings,
		taskRepo:     taskRepo,
		locationRepo: locationRepo,
		vibeRepo:     vibeRepo,
		storage:      storage,
		demoLatency:  demoLatencyMs,
		logger:       logger.Named("onboarding_service"),
	}
}

// StartOnboarding запускает pipeline онбординга хоста.
// Принимает аудиоданные и опциональные URL медиафайлов.
// Выполняет pipeline синхронно (MVP):
//  1. Сохранение аудио в MinIO
//  2. STT: распознавание речи через Vosk
//  3. LLM: извлечение структурированных данных локации
//  4. Embedding: генерация vibe-вектора локации -> Qdrant
//  5. Location draft: создание черновика в PostgreSQL
//
// Возвращает OnboardResponse с task_id и результатом.
func (s *OnboardingService) StartOnboarding(
	ctx context.Context,
	userID uuid.UUID,
	audioData []byte,
	filename string,
	mediaURLs []string,
	latitude *float64,
	longitude *float64,
	address string,
) (*models.OnboardResponse, error) {
	start := time.Now()

	s.logger.Info("начало онбординга хоста",
		zap.String("user_id", userID.String()),
		zap.String("filename", filename),
		zap.Int("audio_bytes", len(audioData)),
		zap.Int("media_count", len(mediaURLs)),
	)

	// Подготовка входных данных для сохранения в ai_tasks.input_data.
	inputData := models.OnboardingInputData{
		MediaURLs: mediaURLs,
		Latitude:  latitude,
		Longitude: longitude,
		Address:   address,
	}

	// Сохранение аудио в MinIO (если сервис доступен).
	if s.storage != nil && len(audioData) > 0 {
		objectName := GenerateObjectName(filename)
		result, err := s.storage.Upload(ctx, objectName, bytes.NewReader(audioData), int64(len(audioData)), "audio/webm")
		if err != nil {
			s.logger.Warn("ошибка загрузки аудио в MinIO", zap.Error(err))
		} else if result != nil {
			inputData.AudioURL = result.URL
		}
	}

	// Создание задачи AI-обработки в БД.
	inputJSON, _ := json.Marshal(inputData)
	task, err := s.taskRepo.Create(ctx, userID, models.TaskTypeOnboarding, inputJSON)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания задачи онбординга: %w", err)
	}

	s.logger.Info("задача онбординга создана",
		zap.String("task_id", task.ID.String()),
	)

	// Pipeline выполняется синхронно (MVP).
	// При ошибке на любом шаге задача помечается как failed.
	result, err := s.executePipeline(ctx, task.ID, userID, audioData, filename, inputData)
	if err != nil {
		errMsg := err.Error()
		if failErr := s.taskRepo.Fail(ctx, task.ID, errMsg); failErr != nil {
			s.logger.Error("ошибка маркировки задачи как failed", zap.Error(failErr))
		}
		return &models.OnboardResponse{
			TaskID:   task.ID,
			Status:   models.TaskStatusFailed,
			Progress: task.Progress,
			Message:  "ошибка обработки: " + errMsg,
		}, nil
	}

	s.logger.Info("онбординг завершён",
		zap.String("task_id", task.ID.String()),
		zap.Duration("total_duration", time.Since(start)),
	)

	return result, nil
}

// executePipeline выполняет полный pipeline онбординга.
func (s *OnboardingService) executePipeline(
	ctx context.Context,
	taskID uuid.UUID,
	userID uuid.UUID,
	audioData []byte,
	filename string,
	inputData models.OnboardingInputData,
) (*models.OnboardResponse, error) {

	// Шаг 1 (progress=20): STT — распознавание речи.
	if err := s.taskRepo.UpdateProgress(ctx, taskID, models.TaskStatusProcessing, 10); err != nil {
		s.logger.Warn("ошибка обновления прогресса", zap.Error(err))
	}

	s.logger.Info("STT: начало распознавания речи")
	transcription, err := s.stt.Transcribe(ctx, bytes.NewReader(audioData), filename, int64(len(audioData)))
	if err != nil {
		return nil, fmt.Errorf("ошибка распознавания речи: %w", err)
	}

	if err := s.taskRepo.UpdateProgress(ctx, taskID, models.TaskStatusProcessing, 20); err != nil {
		s.logger.Warn("ошибка обновления прогресса", zap.Error(err))
	}

	s.logger.Info("STT: речь распознана",
		zap.Int("text_length", len(transcription)),
	)

	// Шаг 2 (progress=50): LLM — извлечение структурированных данных локации.
	s.logger.Info("LLM: извлечение данных локации")
	locationData, err := s.llm.ExtractLocationData(ctx, transcription)
	if err != nil {
		return nil, fmt.Errorf("ошибка извлечения данных локации: %w", err)
	}

	if err := s.taskRepo.UpdateProgress(ctx, taskID, models.TaskStatusProcessing, 50); err != nil {
		s.logger.Warn("ошибка обновления прогресса", zap.Error(err))
	}

	s.logger.Info("LLM: данные локации извлечены",
		zap.String("name", locationData.NameSuggestion),
	)

	// Шаг 3 (progress=70): Embedding — генерация vibe-вектора.
	var vibeVectorID *uuid.UUID
	if s.embeddings != nil && s.vibeRepo != nil {
		embeddingText := buildLocationEmbeddingText(locationData)
		vector, err := s.embeddings.Generate(ctx, embeddingText)
		if err != nil {
			s.logger.Warn("ошибка генерации эмбеддинга локации", zap.Error(err))
		} else {
			vecID := uuid.New()
			payload := map[string]any{
				"type":     "location",
				"name":     locationData.NameSuggestion,
				"category": locationData.Category,
				"tags":     locationData.Tags,
			}
			if err := s.vibeRepo.UpsertVibeVector(ctx, "location_vibes", vecID, vector, payload); err != nil {
				s.logger.Warn("ошибка сохранения вектора в Qdrant", zap.Error(err))
			} else {
				vibeVectorID = &vecID
			}
		}
	}

	if err := s.taskRepo.UpdateProgress(ctx, taskID, models.TaskStatusProcessing, 70); err != nil {
		s.logger.Warn("ошибка обновления прогресса", zap.Error(err))
	}

	// Шаг 4 (progress=90): Создание черновика локации в PostgreSQL.
	slug := models.GenerateSlug(locationData.NameSuggestion)
	if slug == "" {
		slug = "location"
	}

	// Проверка уникальности slug.
	baseSlug := slug
	suffix := 1
	for {
		_, err := s.locationRepo.FindBySlug(ctx, slug)
		if err != nil {
			break
		}
		suffix++
		slug = fmt.Sprintf("%s-%d", baseSlug, suffix)
	}

	// Координаты по умолчанию — центр Краснодарского края.
	lat := 45.035
	lon := 38.975
	if inputData.Latitude != nil {
		lat = *inputData.Latitude
	}
	if inputData.Longitude != nil {
		lon = *inputData.Longitude
	}

	// Формирование объекта локации.
	loc := &models.Location{
		OwnerID:          userID,
		Slug:             slug,
		Name:             locationData.NameSuggestion,
		DescriptionShort: locationData.DescriptionShort,
		DescriptionFull:  locationData.DescriptionLiterary,
		Category:         locationData.Category,
		Tags:             locationData.Tags,
		PricePerNight:    locationData.PricePerNight,
		Capacity:         locationData.Capacity,
		AccessLevel:      models.AccessLevelOpen,
		DensityLevel:     models.DensityGreen,
		ChildFriendly:    false,
		Address:          inputData.Address,
		IsPublished:      false,
		Latitude:         lat,
		Longitude:        lon,
		VibeVectorID:     vibeVectorID,
	}

	// Установка hero image и галереи.
	if len(inputData.MediaURLs) > 0 {
		loc.PreviewImageURL = inputData.MediaURLs[0]
		if len(inputData.MediaURLs) > 1 {
			loc.GalleryURLs = inputData.MediaURLs[1:]
		}
	}

	if loc.Tags == nil {
		loc.Tags = []string{}
	}
	if loc.GalleryURLs == nil {
		loc.GalleryURLs = []string{}
	}

	location, err := s.locationRepo.Create(ctx, loc)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания черновика локации: %w", err)
	}

	if err := s.taskRepo.UpdateProgress(ctx, taskID, models.TaskStatusProcessing, 90); err != nil {
		s.logger.Warn("ошибка обновления прогресса", zap.Error(err))
	}

	s.logger.Info("черновик локации создан",
		zap.String("location_id", location.ID.String()),
		zap.String("name", location.Name),
	)

	// Шаг 5 (progress=100): Завершение задачи.
	extractedData := &models.OnboardingResult{
		NameSuggestion:      locationData.NameSuggestion,
		DescriptionShort:    locationData.DescriptionShort,
		DescriptionLiterary: locationData.DescriptionLiterary,
		Tags:                locationData.Tags,
		Category:            locationData.Category,
		PricePerNight:       locationData.PricePerNight,
		Amenities:           locationData.Amenities,
		Capacity:            locationData.Capacity,
	}

	outputData := models.OnboardingOutputData{
		LocationID:    location.ID,
		Transcription: transcription,
		ExtractedData: extractedData,
	}

	outputJSON, _ := json.Marshal(outputData)
	if err := s.taskRepo.Complete(ctx, taskID, outputJSON); err != nil {
		s.logger.Warn("ошибка завершения задачи", zap.Error(err))
	}

	return &models.OnboardResponse{
		TaskID:        taskID,
		Status:        models.TaskStatusCompleted,
		Progress:      100,
		Message:       models.ProgressMessage(100),
		LocationID:    &location.ID,
		ExtractedData: extractedData,
	}, nil
}

// GetTaskStatus возвращает текущий статус задачи онбординга.
// Проверяет принадлежность задачи пользователю.
func (s *OnboardingService) GetTaskStatus(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*models.OnboardResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("задача не найдена: %w", err)
	}

	// Проверка принадлежности задачи пользователю.
	if task.UserID != userID {
		return nil, fmt.Errorf("задача не принадлежит пользователю")
	}

	resp := &models.OnboardResponse{
		TaskID:   task.ID,
		Status:   task.Status,
		Progress: task.Progress,
		Message:  models.ProgressMessage(task.Progress),
	}

	// Извлечение output_data если задача завершена.
	if task.Status == models.TaskStatusCompleted && len(task.OutputData) > 0 {
		var output models.OnboardingOutputData
		if err := json.Unmarshal(task.OutputData, &output); err == nil {
			resp.LocationID = &output.LocationID
			resp.ExtractedData = output.ExtractedData
		}
	}

	// Если задача failed — добавляем ошибку в message.
	if task.Status == models.TaskStatusFailed && task.Error != nil {
		resp.Message = "ошибка: " + *task.Error
	}

	return resp, nil
}

// GetHostLocations возвращает локации текущего хоста.
func (s *OnboardingService) GetHostLocations(ctx context.Context, userID uuid.UUID) ([]models.Location, error) {
	locations, err := s.locationRepo.FindByOwnerID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения локаций хоста: %w", err)
	}
	if locations == nil {
		locations = []models.Location{}
	}
	return locations, nil
}

// buildLocationEmbeddingText формирует текст для генерации эмбеддинга локации.
// Объединяет название, описание, теги и категорию.
func buildLocationEmbeddingText(data *ai.LocationData) string {
	parts := []string{
		data.NameSuggestion,
		data.DescriptionLiterary,
	}
	if len(data.Tags) > 0 {
		parts = append(parts, strings.Join(data.Tags, ", "))
	}
	if data.Category != "" {
		parts = append(parts, data.Category)
	}
	return strings.Join(parts, ". ")
}

// --- 3D Gaussian Splatting (Mock Pipeline) ---

// StartSplatting запускает mock pipeline генерации 3D-сцены из видео.
// В MVP — выбирает релевантный существующий .splat файл по категории локации.
// Pipeline: загрузка видео → извлечение кадров → облако точек → обучение GS → экспорт → оптимизация → привязка.
func (s *OnboardingService) StartSplatting(
	ctx context.Context,
	userID uuid.UUID,
	locationID uuid.UUID,
	videoURL string,
) (*models.SplattingResponse, error) {
	start := time.Now()

	s.logger.Info("начало 3D splatting pipeline",
		zap.String("user_id", userID.String()),
		zap.String("location_id", locationID.String()),
		zap.String("video_url", videoURL),
	)

	// Проверка существования локации и прав владельца.
	location, err := s.locationRepo.FindByID(ctx, locationID.String())
	if err != nil {
		return nil, fmt.Errorf("локация не найдена: %w", err)
	}
	if location.OwnerID != userID {
		return nil, fmt.Errorf("нет прав на эту локацию")
	}

	// Создание задачи.
	inputData := models.SplattingInputData{
		VideoURL:   videoURL,
		LocationID: locationID,
	}
	inputJSON, _ := json.Marshal(inputData)
	task, err := s.taskRepo.Create(ctx, userID, models.TaskTypeSplatting, inputJSON)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания задачи splatting: %w", err)
	}

	// Mock pipeline: симуляция шагов с прогрессом.
	steps := []struct {
		progress int
		delay    time.Duration
	}{
		{15, time.Duration(s.demoLatency/3) * time.Millisecond},
		{30, time.Duration(s.demoLatency/3) * time.Millisecond},
		{50, time.Duration(s.demoLatency/2) * time.Millisecond},
		{70, time.Duration(s.demoLatency) * time.Millisecond},
		{85, time.Duration(s.demoLatency/2) * time.Millisecond},
		{95, time.Duration(s.demoLatency/4) * time.Millisecond},
	}

	for _, step := range steps {
		time.Sleep(step.delay)
		if err := s.taskRepo.UpdateProgress(ctx, task.ID, models.TaskStatusProcessing, step.progress); err != nil {
			s.logger.Warn("ошибка обновления прогресса splatting", zap.Error(err))
		}
	}

	// Выбор релевантного mock .splat файла по категории.
	splatURL := selectMockSplatURL(location.Category)

	// Обновление splat_url локации.
	if err := s.locationRepo.UpdateSplatURL(ctx, locationID, splatURL); err != nil {
		errMsg := err.Error()
		_ = s.taskRepo.Fail(ctx, task.ID, errMsg)
		return &models.SplattingResponse{
			TaskID:   task.ID,
			Status:   models.TaskStatusFailed,
			Progress: 95,
			Message:  "ошибка привязки 3D-сцены: " + errMsg,
		}, nil
	}

	// Завершение задачи.
	outputData := models.SplattingOutputData{
		SplatURL:   splatURL,
		LocationID: locationID,
	}
	outputJSON, _ := json.Marshal(outputData)
	if err := s.taskRepo.Complete(ctx, task.ID, outputJSON); err != nil {
		s.logger.Warn("ошибка завершения задачи splatting", zap.Error(err))
	}

	s.logger.Info("3D splatting pipeline завершён",
		zap.String("task_id", task.ID.String()),
		zap.String("splat_url", splatURL),
		zap.Duration("total_duration", time.Since(start)),
	)

	return &models.SplattingResponse{
		TaskID:   task.ID,
		Status:   models.TaskStatusCompleted,
		Progress: 100,
		Message:  models.SplattingProgressMessage(100),
		SplatURL: splatURL,
	}, nil
}

// selectMockSplatURL выбирает релевантный mock .splat файл на основе категории локации.
// Маппинг основан на реальных данных сервера:
//   - winery → garden.splat (виноградники, открытые пространства)
//   - farm → kitchen.splat (фермерские постройки, внутренние помещения)
//   - nature → stump.splat (природные объекты, деревья)
//   - gastro → counter.splat (прилавки, интерьеры)
//   - resort/guesthouse → room.splat (комнаты, помещения)
//   - extreme → bicycle.splat (активный отдых)
//   - cultural → bonsai.splat (культурные объекты)
//   - trail → stump.splat (тропы, природа)
func selectMockSplatURL(category string) string {
	const base = "http://141.98.7.225:9102/deepkrai-media/locations/splat/"

	categoryMap := map[string]string{
		"winery":     "garden.splat",
		"farm":       "kitchen.splat",
		"nature":     "stump.splat",
		"gastro":     "counter.splat",
		"resort":     "room.splat",
		"guesthouse": "room.splat",
		"extreme":    "bicycle.splat",
		"cultural":   "bonsai.splat",
		"trail":      "stump.splat",
	}

	if splat, ok := categoryMap[strings.ToLower(category)]; ok {
		return base + splat
	}

	// Fallback: garden.splat — наиболее универсальный.
	return base + "garden.splat"
}

