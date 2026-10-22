package repository

import (
	"context"
	"time"

	"telegram-bot/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReminderRepository interface {
	GetPreReminderTasks(
		ctx context.Context,
		now time.Time,
		until time.Time,
	) ([]model.TaskReminder, error)

	GetDueReminderTasks(
		ctx context.Context,
		now time.Time,
	) ([]model.TaskReminder, error)

	GetOverdueReminderTasks(
		ctx context.Context,
		now time.Time,
	) ([]model.TaskReminder, error)

	MarkPreReminderSent(
		ctx context.Context,
		taskID int64,
	) error

	MarkDueReminderSent(
		ctx context.Context,
		taskID int64,
	) error

	UpdateLastOverdueReminder(
		ctx context.Context,
		taskID int64,
		sentAt time.Time,
	) error
}

type ReminderPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewReminderPostgresRepository(
	pool *pgxpool.Pool,
) *ReminderPostgresRepository {

	return &ReminderPostgresRepository{
		pool: pool,
	}
}

func (r *ReminderPostgresRepository) GetPreReminderTasks(
	ctx context.Context,
	now time.Time,
	until time.Time,
) ([]model.TaskReminder, error) {

	query := `
		SELECT
			t.id,
			u.telegram_id,
			t.title,
			t.due_at,
			t.pre_reminder_sent,
			t.due_reminder_sent,
			t.last_overdue_reminder_at
		FROM tasks t
		JOIN users u
			ON u.id = t.user_id
		WHERE t.status = 'todo'
		  AND t.pre_reminder_sent = FALSE
		  AND t.due_at > $1
		  AND t.due_at <= $2
		ORDER BY t.due_at
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		now,
		until,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []model.TaskReminder

	for rows.Next() {
		var reminder model.TaskReminder

		if err := rows.Scan(
			&reminder.TaskID,
			&reminder.TelegramID,
			&reminder.Title,
			&reminder.DueAt,
			&reminder.PreReminderSent,
			&reminder.DueReminderSent,
			&reminder.LastOverdueReminderAt,
		); err != nil {
			return nil, err
		}

		reminders = append(reminders, reminder)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reminders, nil
}

func (r *ReminderPostgresRepository) GetDueReminderTasks(
	ctx context.Context,
	now time.Time,
) ([]model.TaskReminder, error) {

	query := `
		SELECT
			t.id,
			u.telegram_id,
			t.title,
			t.due_at,
			t.pre_reminder_sent,
			t.due_reminder_sent,
			t.last_overdue_reminder_at
		FROM tasks t
		JOIN users u
			ON u.id = t.user_id
		WHERE t.status = 'todo'
		  AND t.due_reminder_sent = FALSE
		  AND t.due_at <= $1
		ORDER BY t.due_at
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		now,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []model.TaskReminder

	for rows.Next() {
		var reminder model.TaskReminder

		if err := rows.Scan(
			&reminder.TaskID,
			&reminder.TelegramID,
			&reminder.Title,
			&reminder.DueAt,
			&reminder.PreReminderSent,
			&reminder.DueReminderSent,
			&reminder.LastOverdueReminderAt,
		); err != nil {
			return nil, err
		}

		reminders = append(reminders, reminder)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reminders, nil
}

func (r *ReminderPostgresRepository) GetOverdueReminderTasks(
	ctx context.Context,
	now time.Time,
) ([]model.TaskReminder, error) {

	query := `
		SELECT
			t.id,
			u.telegram_id,
			t.title,
			t.due_at,
			t.pre_reminder_sent,
			t.due_reminder_sent,
			t.last_overdue_reminder_at
		FROM tasks t
		JOIN users u
			ON u.id = t.user_id
		WHERE t.status = 'todo'
		  AND t.due_at < $1
		  AND t.due_reminder_sent = TRUE
		  AND (
				t.last_overdue_reminder_at IS NULL
				OR t.last_overdue_reminder_at <= $1 - INTERVAL '1 hour'
		  )
		ORDER BY t.due_at
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		now,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []model.TaskReminder

	for rows.Next() {
		var reminder model.TaskReminder

		if err := rows.Scan(
			&reminder.TaskID,
			&reminder.TelegramID,
			&reminder.Title,
			&reminder.DueAt,
			&reminder.PreReminderSent,
			&reminder.DueReminderSent,
			&reminder.LastOverdueReminderAt,
		); err != nil {
			return nil, err
		}

		reminders = append(reminders, reminder)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reminders, nil
}

func (r *ReminderPostgresRepository) MarkPreReminderSent(
	ctx context.Context,
	taskID int64,
) error {

	query := `
		UPDATE tasks
		SET
			pre_reminder_sent = TRUE,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		taskID,
	)

	return err
}

func (r *ReminderPostgresRepository) MarkDueReminderSent(
	ctx context.Context,
	taskID int64,
) error {

	query := `
		UPDATE tasks
		SET
			due_reminder_sent = TRUE,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		taskID,
	)

	return err
}

func (r *ReminderPostgresRepository) UpdateLastOverdueReminder(
	ctx context.Context,
	taskID int64,
	sentAt time.Time,
) error {

	query := `
		UPDATE tasks
		SET
			last_overdue_reminder_at = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		taskID,
		sentAt,
	)

	return err
}
