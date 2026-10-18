package repository

import (
	"context"
	"telegram-bot/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (*model.User, error)
}

type UserPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserPostgresRepository(pool *pgxpool.Pool) *UserPostgresRepository {
	return &UserPostgresRepository{
		pool: pool,
	}
}

func (r UserPostgresRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*model.User, error) {
	var user model.User

	query := `
		SELECT id, telegram_id, username, first_name, last_name, preferred_name
		FROM users
		WHERE telegram_id = $1`

	err := r.pool.QueryRow(
		ctx,
		query,
		telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.PreferredName,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
