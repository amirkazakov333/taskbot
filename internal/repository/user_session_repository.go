
package repository

import (
	"context"

	"telegram-bot/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserSessionRepository interface {
	GetByTelegramID(
		ctx context.Context,
		telegramID int64,
	) (*model.UserSession, error)

	Create(
		ctx context.Context,
		session model.UserSession,
	) error

	Update(
		ctx context.Context,
		session model.UserSession,
	) error

	Delete(
		ctx context.Context,
		telegramID int64,
	) error
}

type UserSessionPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserSessionPostgresRepository(
	pool *pgxpool.Pool,
) *UserSessionPostgresRepository {
	return &UserSessionPostgresRepository{
		pool: pool,
	}
}

func (r *UserSessionPostgresRepository) GetByTelegramID(
	ctx context.Context,
	telegramID int64,
) (*model.UserSession, error) {

	var session model.UserSession
	var sessionType string

	// Эти поля могут быть NULL в PostgreSQL.
	var username pgtype.Text
	var category pgtype.Text
	var title pgtype.Text
	var taskDate pgtype.Date
	var targetTaskID pgtype.Int8

	query := `
		SELECT
			telegram_id,
			session_type,
			username,
			first_name,
			last_name,
			preferred_name,
			category,
			title,
			task_date,
			target_task_id,
			current_step,
			created_at,
			updated_at
		FROM user_sessions
		WHERE telegram_id = $1
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		telegramID,
	).Scan(
		&session.TelegramID,
		&sessionType,
		&username,
		&session.FirstName,
		&session.LastName,
		&session.PreferredName,
		&category,
		&title,
		&taskDate,
		&targetTaskID,
		&session.CurrentStep,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	session.SessionType = model.UserSessionType(sessionType)

	if username.Valid {
		value := username.String
		session.Username = &value
	}

	if category.Valid {
		value := model.TaskCategory(category.String)
		session.Category = &value
	}

	if title.Valid {
		value := title.String
		session.Title = &value
	}

	if taskDate.Valid {
		value := taskDate.Time
		session.TaskDate = &value
	}

	if targetTaskID.Valid {
		value := targetTaskID.Int64
		session.TargetTaskID = &value
	}

	return &session, nil
}

func (r *UserSessionPostgresRepository) Create(
	ctx context.Context,
	session model.UserSession,
) error {

	query := `
		INSERT INTO user_sessions (
			telegram_id,
			session_type,
			username,
			first_name,
			last_name,
			preferred_name,
			category,
			title,
			task_date,
			target_task_id,
			current_step
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		session.TelegramID,
		session.SessionType,
		pointerValue(session.Username),
		session.FirstName,
		session.LastName,
		session.PreferredName,
		pointerValue(session.Category),
		pointerValue(session.Title),
		pointerValue(session.TaskDate),
		pointerValue(session.TargetTaskID),
		session.CurrentStep,
	)

	return err
}

func (r *UserSessionPostgresRepository) Update(
	ctx context.Context,
	session model.UserSession,
) error {

	query := `
		UPDATE user_sessions
		SET
			session_type = $2,
			username = $3,
			first_name = $4,
			last_name = $5,
			preferred_name = $6,
			category = $7,
			title = $8,
			task_date = $9,
			target_task_id = $10,
			current_step = $11,
			updated_at = CURRENT_TIMESTAMP
		WHERE telegram_id = $1
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		session.TelegramID,
		session.SessionType,
		pointerValue(session.Username),
		session.FirstName,
		session.LastName,
		session.PreferredName,
		pointerValue(session.Category),
		pointerValue(session.Title),
		pointerValue(session.TaskDate),
		pointerValue(session.TargetTaskID),
		session.CurrentStep,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *UserSessionPostgresRepository) Delete(
	ctx context.Context,
	telegramID int64,
) error {

	query := `
		DELETE FROM user_sessions
		WHERE telegram_id = $1
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		telegramID,
	)

	return err
}

// Преобразует указатель в значение.
// Если указатель nil, PostgreSQL получит NULL.
func pointerValue[T any](value *T) any {
	if value == nil {
		return nil
	}

	return *value
}
