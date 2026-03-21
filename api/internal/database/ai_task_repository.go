// Файл ai_task_repository.go реализует data access слой для таблицы ai_tasks.
// Предоставляет CRUD-операции для задач AI-обработки, используемых
// в pipeline онбординга хостов (GDD Feature 5).
package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
)

// AITaskRepository — репозиторий для работы с таблицей ai_tasks.
type AITaskRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewAITaskRepository создаёт экземпляр репозитория задач AI-обработки.
func NewAITaskRepository(pool *pgxpool.Pool, logger *zap.Logger) *AITaskRepository {
	return &AITaskRepository{
		pool:   pool,
		logger: logger.Named("ai_task_repo"),
	}
}

// Create создаёт новую задачу AI-обработки в БД.
// Возвращает созданную задачу с заполненным ID и временными метками.
func (r *AITaskRepository) Create(ctx context.Context, userID uuid.UUID, taskType models.TaskType, inputData json.RawMessage) (*models.AITask, error) {
	task := &models.AITask{}

	err := r.pool.QueryRow(ctx,
		`INSERT INTO ai_tasks (user_id, type, status, progress, input_data)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, type, status, progress, input_data, output_data,
		           error, created_at, updated_at, completed_at`,
		userID, string(taskType), string(models.TaskStatusProcessing), 0, inputData,
	).Scan(
		&task.ID, &task.UserID, &task.Type, &task.Status, &task.Progress,
		&task.InputData, &task.OutputData, &task.Error,
		&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания задачи AI: %w", err)
	}

	r.logger.Debug("задача AI создана",
		zap.String("task_id", task.ID.String()),
		zap.String("type", string(taskType)),
	)

	return task, nil
}

// GetByID возвращает задачу по ID.
// Возвращает ErrTaskNotFound, если задача не найдена.
func (r *AITaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AITask, error) {
	task := &models.AITask{}

	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, type, status, progress, input_data, output_data,
		        error, created_at, updated_at, completed_at
		 FROM ai_tasks WHERE id = $1`,
		id,
	).Scan(
		&task.ID, &task.UserID, &task.Type, &task.Status, &task.Progress,
		&task.InputData, &task.OutputData, &task.Error,
		&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("задача AI не найдена (id=%s): %w", id.String(), err)
	}

	return task, nil
}

// UpdateProgress обновляет прогресс и статус задачи.
// Используется для инкрементального обновления прогресс-бара на фронтенде.
func (r *AITaskRepository) UpdateProgress(ctx context.Context, id uuid.UUID, status models.TaskStatus, progress int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE ai_tasks SET status = $1, progress = $2, updated_at = NOW()
		 WHERE id = $3`,
		string(status), progress, id,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления прогресса задачи: %w", err)
	}
	return nil
}

// Complete завершает задачу с результатом.
// Устанавливает status=completed, progress=100, output_data и completed_at.
func (r *AITaskRepository) Complete(ctx context.Context, id uuid.UUID, outputData json.RawMessage) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE ai_tasks
		 SET status = $1, progress = 100, output_data = $2,
		     updated_at = $3, completed_at = $3
		 WHERE id = $4`,
		string(models.TaskStatusCompleted), outputData, now, id,
	)
	if err != nil {
		return fmt.Errorf("ошибка завершения задачи: %w", err)
	}
	return nil
}

// Fail помечает задачу как провалившуюся.
// Устанавливает status=failed, error и completed_at.
func (r *AITaskRepository) Fail(ctx context.Context, id uuid.UUID, errMsg string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE ai_tasks
		 SET status = $1, error = $2, updated_at = $3, completed_at = $3
		 WHERE id = $4`,
		string(models.TaskStatusFailed), errMsg, now, id,
	)
	if err != nil {
		return fmt.Errorf("ошибка маркировки задачи как failed: %w", err)
	}
	return nil
}

// FindByUserID возвращает задачи пользователя определённого типа.
// Результаты отсортированы по дате создания (новые первые).
func (r *AITaskRepository) FindByUserID(ctx context.Context, userID uuid.UUID, taskType models.TaskType) ([]models.AITask, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, type, status, progress, input_data, output_data,
		        error, created_at, updated_at, completed_at
		 FROM ai_tasks
		 WHERE user_id = $1 AND type = $2
		 ORDER BY created_at DESC
		 LIMIT 50`,
		userID, string(taskType),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска задач AI: %w", err)
	}
	defer rows.Close()

	var tasks []models.AITask
	for rows.Next() {
		var task models.AITask
		if err := rows.Scan(
			&task.ID, &task.UserID, &task.Type, &task.Status, &task.Progress,
			&task.InputData, &task.OutputData, &task.Error,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования задачи AI: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
