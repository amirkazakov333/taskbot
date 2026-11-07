package repository

import (
	"context"
	"fmt"

	"telegram-bot/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetByTelegramID(
		ctx context.Context,
		telegramID int64,
	) (*model.User, error)

	UserCreate(
		ctx context.Context,
		user model.User,
	) error

	UpdateUserData(
		ctx context.Context,
		telegramID int64,
		field model.UserData,
		value string,
	) error
}
type UserPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserPostgresRepository(pool *pgxpool.Pool) *UserPostgresRepository {
	return &UserPostgresRepository{
		pool: pool,
	}
}

func (r *UserPostgresRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*model.User, error) {
	var user model.User

	query := `
		SELECT id, telegram_id, username, first_name, last_name, preferred_name
		FROM users
		WHERE telegram_id = $1
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		telegramID,
	).Scan(
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

func (r *UserPostgresRepository) UserCreate(ctx context.Context, user model.User) error {
	query := `
		INSERT INTO users (
			telegram_id,
			username,
			first_name,
			last_name,
			preferred_name
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		user.TelegramID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.PreferredName,
	)

	return err
}

func (r *UserPostgresRepository) UpdateUserData(
	ctx context.Context,
	telegramID int64,
	field model.UserData,
	value string,
) error {

	var query string

	switch field {
	case model.UserFirstName:
		query = `
            UPDATE users
            SET first_name = $2
            WHERE telegram_id = $1
        `

	case model.UserLastName:
		query = `
            UPDATE users
            SET last_name = $2
            WHERE telegram_id = $1
        `

	case model.UserPreferredName:
		query = `
            UPDATE users
            SET preferred_name = $2
            WHERE telegram_id = $1
        `

	default:
		return fmt.Errorf("unsupported user data field: %q", field)
	}

	tag, err := r.pool.Exec(
		ctx,
		query,
		telegramID,
		value,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
