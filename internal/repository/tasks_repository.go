package repository

import (
	"context"

	"telegram-bot/internal/model"

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
