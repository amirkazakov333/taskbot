package repository

import (
	"context"

	"telegram-bot/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRegistrationRepository interface {
	GetByTelegramID(
		ctx context.Context,
		telegramID int64,
	) (*model.TaskCreationSession, error)

	RegisterCreate(
		ctx context.Context,
		session model.TaskCreationSession,
	) error

	Update(
		ctx context.Context,
		session model.TaskCreationSession,
	) error

	Delete(
		ctx context.Context,
		telegramID int64,
	) error
}

type TaskRegistrationPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRegistrationPostgresRepository(
	pool *pgxpool.Pool,
) *TaskRegistrationPostgresRepository {

	return &TaskRegistrationPostgresRepository{
		pool: pool,
	}
}

func (r *TaskRegistrationPostgresRepository) GetByTelegramID(
	ctx context.Context,
	telegramID int64,
) (*model.TaskCreationSession, error) {

	var session model.TaskCreationSession

	query := `
		SELECT
			telegram_id,
			category,
			title,
			task_date,
			current_step,
			created_at,
			updated_at
		FROM task_creation_sessions
		WHERE telegram_id = $1
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		telegramID,
	).Scan(
		&session.TelegramID,
		&session.Category,
		&session.Title,
		&session.TaskDate,
		&session.CurrentStep,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *TaskRegistrationPostgresRepository) RegisterCreate(
	ctx context.Context,
	session model.TaskCreationSession,
) error {

	query := `
		INSERT INTO task_creation_sessions (
			telegram_id,
			category,
			title,
			task_date,
			current_step
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		session.TelegramID,
		session.Category,
		session.Title,
		session.TaskDate,
		session.CurrentStep,
	)

	return err
}

func (r *TaskRegistrationPostgresRepository) Update(
	ctx context.Context,
	session model.TaskCreationSession,
) error {

	query := `
		UPDATE task_creation_sessions
		SET
			category = $2,
			title = $3,
			task_date = $4,
			current_step = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE telegram_id = $1
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		session.TelegramID,
		session.Category,
		session.Title,
		session.TaskDate,
		session.CurrentStep,
	)

	return err
}

func (r *TaskRegistrationPostgresRepository) Delete(
	ctx context.Context,
	telegramID int64,
) error {

	query := `
		DELETE FROM task_creation_sessions
		WHERE telegram_id = $1
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		telegramID,
	)

	return err
}
