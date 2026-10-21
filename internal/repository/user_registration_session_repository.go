package repository

import (
	"context"

	"telegram-bot/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RegistrationSessionRepository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (*model.RegistrationSession, error)
	RegisterCreate(ctx context.Context, session model.RegistrationSession) error
	Update(ctx context.Context, session model.RegistrationSession) error
	Delete(ctx context.Context, telegramID int64) error
}

type RegistrationSessionPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewRegistrationSessionPostgresRepository(pool *pgxpool.Pool) *RegistrationSessionPostgresRepository {
	return &RegistrationSessionPostgresRepository{
		pool: pool,
	}
}

func (r *RegistrationSessionPostgresRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*model.RegistrationSession, error) {
	var session model.RegistrationSession

	query := `
		SELECT
			telegram_id,
			username,
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			COALESCE(preferred_name, ''),
			current_step,
			created_at,
			updated_at
		FROM registration_sessions
		WHERE telegram_id = $1
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		telegramID,
	).Scan(
		&session.TelegramID,
		&session.Username,
		&session.FirstName,
		&session.LastName,
		&session.PreferredName,
		&session.CurrentStep,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *RegistrationSessionPostgresRepository) RegisterCreate(ctx context.Context, session model.RegistrationSession) error {
	query := `
		INSERT INTO registration_sessions (
			telegram_id,
			username,
			first_name,
			last_name,
			preferred_name,
			current_step,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		session.TelegramID,
		session.Username,
		session.FirstName,
		session.LastName,
		session.PreferredName,
		session.CurrentStep,
	)

	return err
}

func (r *RegistrationSessionPostgresRepository) Update(ctx context.Context, session model.RegistrationSession) error {
	query := `
		UPDATE registration_sessions
		SET
			username = $2,
			first_name = $3,
			last_name = $4,
			preferred_name = $5,
			current_step = $6,
			updated_at = CURRENT_TIMESTAMP
		WHERE telegram_id = $1
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		session.TelegramID,
		session.Username,
		session.FirstName,
		session.LastName,
		session.PreferredName,
		session.CurrentStep,
	)

	return err
}

func (r *RegistrationSessionPostgresRepository) Delete(ctx context.Context, telegramID int64) error {
	query := `
		DELETE FROM registration_sessions
		WHERE telegram_id = $1
	`

	_, err := r.pool.Exec(ctx, query, telegramID)
	return err
}
