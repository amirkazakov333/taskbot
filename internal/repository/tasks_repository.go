package repository

import (
	"context"

	"telegram-bot/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
}

type TaskPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewTaskPostgresRepository(pool *pgxpool.Pool) *TaskPostgresRepository {
	return &TaskPostgresRepository{
		pool: pool,
	}
}

func (r *TaskPostgresRepository) Create(ctx context.Context, task *model.Task) error {
	query := `
		INSERT INTO tasks (
			user_id,
			category,
			title,
			status,
			due_at
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
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
