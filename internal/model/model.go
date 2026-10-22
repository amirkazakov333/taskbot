package model

import "time"

type MenuType string

type TaskCategory string

type TaskStatus string

const (
	MenuNone          MenuType = "none"
	MenuMain          MenuType = "main"
	MenuTaskCategory  MenuType = "task_category"
	MenuTaskDate      MenuType = "task_date"
	MenuTaskTime      MenuType = "task_time"
	MenuTaskFilter    MenuType = "task_filter"
	MenuTaskTimeToday MenuType = "task_time_today"
)

const (
	TaskCategoryWork     TaskCategory = "work"
	TaskCategoryPersonal TaskCategory = "personal"
)

const (
	TaskStatusTodo TaskStatus = "todo"
	TaskStatusDone TaskStatus = "done"
)

type BotResponse struct {
	Text string
	Menu MenuType
}

type User struct {
	ID            int64
	TelegramID    int64
	Username      *string
	FirstName     string
	LastName      string
	PreferredName string
}

type RegistrationSession struct {
	TelegramID    int64
	Username      *string
	FirstName     string
	LastName      string
	PreferredName string
	CurrentStep   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Task struct {
	ID                    int64
	UserID                int64
	Category              TaskCategory
	Title                 string
	Status                TaskStatus
	DueAt                 time.Time
	PreReminderSent       bool
	DueReminderSent       bool
	LastOverdueReminderAt *time.Time
	CompletedAt           *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type TaskCreationSession struct {
	TelegramID  int64
	Category    *TaskCategory
	Title       *string
	TaskDate    *time.Time
	CurrentStep string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
