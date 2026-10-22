package model

import "time"

const (
	SessionTypeRegistration   UserSessionType = "registration"
	SessionTypeTaskCreation   UserSessionType = "task_creation"
	SessionTypeTaskReschedule UserSessionType = "task_reschedule"
)

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
	TaskStatusTodo      TaskStatus = "todo"
	TaskStatusDone      TaskStatus = "done"
	TaskStatusCancelled TaskStatus = "cancelled"
)

const (
	ReminderTypePre     ReminderType = "pre"
	ReminderTypeDue     ReminderType = "due"
	ReminderTypeOverdue ReminderType = "overdue"
)

type MenuType string

type TaskCategory string

type TaskStatus string

type UserSessionType string

type UserSession struct {
	TelegramID  int64
	SessionType UserSessionType

	Username      *string
	FirstName     string
	LastName      string
	PreferredName string

	Category *TaskCategory
	Title    *string
	TaskDate *time.Time

	TargetTaskID *int64

	CurrentStep string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

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

type TaskReminder struct {
	TaskID     int64
	TelegramID int64

	Category TaskCategory
	Title    string
	DueAt    time.Time

	PreReminderSent bool
	DueReminderSent bool

	LastOverdueReminderAt *time.Time
}
type ReminderType string

type ReminderMessage struct {
	TaskID     int64
	TelegramID int64
	Type       ReminderType
	Text       string
}
