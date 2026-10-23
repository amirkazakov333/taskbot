package repository

import (
	"context"
	"log"
	"time"

	"telegram-bot/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository interface {
	Create(
		ctx context.Context,
		task *model.Task,
	) error

	GetByUserID(
		ctx context.Context,
		userID int64,
	) ([]model.Task, error)

	GetByUserIDAndCategory(
		ctx context.Context,
		userID int64,
		category model.TaskCategory,
	) ([]model.Task, error)

	GetByIDAndUserID(
		ctx context.Context,
		taskID int64,
		userID int64,
	) (*model.Task, error)

	UpdateStatusByUserID(
		ctx context.Context,
		taskID int64,
		userID int64,
		status model.TaskStatus,
	) error

	RescheduleByUserID(
		ctx context.Context,
		taskID int64,
		userID int64,
		dueAt time.Time,
	) error
}

type TaskPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewTaskPostgresRepository(
	pool *pgxpool.Pool,
) *TaskPostgresRepository {
	return &TaskPostgresRepository{
		pool: pool,
	}
}

func (r *TaskPostgresRepository) Create(
	ctx context.Context,
	task *model.Task,
) error {

	query := `
		INSERT INTO tasks (
			user_id,
			category,
			title,
			status,
			due_at
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			created_at,
			updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		task.UserID,
		task.Category,
		task.Title,
		task.Status,
		task.DueAt,
	).Scan(
		&task.ID,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	return err
}

func (r *TaskPostgresRepository) GetByUserID(
	ctx context.Context,
	userID int64,
) ([]model.Task, error) {

	todoQuery := `
		SELECT
			id,
			user_id,
			category,
			title,
			status,
			due_at,
			pre_reminder_sent,
			due_reminder_sent,
			last_overdue_reminder_at,
			completed_at,
			created_at,
			updated_at
		FROM tasks
		WHERE user_id = $1
		   AND status = 'todo'
		ORDER BY due_at
	`

	rows, err := r.pool.Query(
		ctx,
		todoQuery,
		userID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task

	for rows.Next() {
		var task model.Task

		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Category,
			&task.Title,
			&task.Status,
			&task.DueAt,
			&task.PreReminderSent,
			&task.DueReminderSent,
			&task.LastOverdueReminderAt,
			&task.CompletedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskPostgresRepository) GetByUserIDAndCategory(
	ctx context.Context,
	userID int64,
	category model.TaskCategory,
) ([]model.Task, error) {

	query := `
		SELECT
			id,
			user_id,
			category,
			title,
			status,
			due_at,
			pre_reminder_sent,
			due_reminder_sent,
			last_overdue_reminder_at,
			completed_at,
			created_at,
			updated_at
		FROM tasks
		WHERE user_id = $1
		  AND category = $2
		  AND status = 'todo'
		ORDER BY due_at
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		userID,
		category,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task

	for rows.Next() {
		var task model.Task

		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Category,
			&task.Title,
			&task.Status,
			&task.DueAt,
			&task.PreReminderSent,
			&task.DueReminderSent,
			&task.LastOverdueReminderAt,
			&task.CompletedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskPostgresRepository) GetByIDAndUserID(
	ctx context.Context,
	taskID int64,
	userID int64,
) (*model.Task, error) {
	var task model.Task

	query := `
		SELECT
			id,
			user_id,
			category,
			title,
			status,
			due_at,
			pre_reminder_sent,
			due_reminder_sent,
			last_overdue_reminder_at,
			completed_at,
			created_at,
			updated_at
		FROM tasks
		WHERE id = $1
		  AND user_id = $2
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		taskID,
		userID,
	).Scan(
		&task.ID,
		&task.UserID,
		&task.Category,
		&task.Title,
		&task.Status,
		&task.DueAt,
		&task.PreReminderSent,
		&task.DueReminderSent,
		&task.LastOverdueReminderAt,
		&task.CompletedAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *TaskPostgresRepository) UpdateStatusByUserID(
	ctx context.Context,
	taskID int64,
	userID int64,
	status model.TaskStatus,
) error {
	query := `
		UPDATE tasks
		SET
			status = $3,
			completed_at = CASE
				WHEN $3 = 'done' THEN CURRENT_TIMESTAMP
				ELSE NULL
			END,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND user_id = $2
		  AND status = 'todo'
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		taskID,
		userID,
		status,
	)
	if err != nil {
		log.Printf(
			"UPDATE TASK ERROR: task_id=%d user_id=%d status=%s err=%v",
			taskID,
			userID,
			status,
			err,
		)
		return err
	}

	log.Printf(
		"UPDATE TASK RESULT: task_id=%d user_id=%d status=%s affected=%d",
		taskID,
		userID,
		status,
		tag.RowsAffected(),
	)

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *TaskPostgresRepository) RescheduleByUserID(
	ctx context.Context,
	taskID int64,
	userID int64,
	dueAt time.Time,
) error {
	query := `
		UPDATE tasks
		SET
			due_at = $3,
			pre_reminder_sent = FALSE,
			due_reminder_sent = FALSE,
			last_overdue_reminder_at = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND user_id = $2
		  AND status = 'todo'
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		taskID,
		userID,
		dueAt,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
