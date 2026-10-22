package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"telegram-bot/internal/model"
	"telegram-bot/internal/repository"

	"github.com/jackc/pgx/v5"
)

const (
	TaskStepAwaitCategory   = "await_task_category"
	TaskStepAwaitTitle      = "await_task_title"
	TaskStepAwaitDate       = "await_task_date"
	TaskStepAwaitCustomDate = "await_task_custom_date"
	TaskStepAwaitTime       = "await_task_time"
)

type TaskService struct {
	users             repository.UserRepository
	tasks             repository.TaskRepository
	registrationTasks repository.TaskRegistrationRepository
}

func NewTaskService(users repository.UserRepository, tasks repository.TaskRepository, registrationTasks repository.TaskRegistrationRepository) *TaskService {
	return &TaskService{
		users:             users,
		tasks:             tasks,
		registrationTasks: registrationTasks,
	}
}

func (s *TaskService) StartCreateTask(ctx context.Context, telegramID int64) (*model.BotResponse, error) {
	session, err := s.registrationTasks.GetByTelegramID(
		ctx,
		telegramID,
	)

	// Если создание задачи уже было начато
	if err == nil {
		return taskCreationPrompt(session.CurrentStep), nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	newSession := model.TaskCreationSession{
		TelegramID:  telegramID,
		CurrentStep: TaskStepAwaitCategory,
	}

	if err := s.registrationTasks.RegisterCreate(
		ctx,
		newSession,
	); err != nil {
		return nil, err
	}

	return &model.BotResponse{
		Text: "📂 <b>Выбери категорию задачи</b>",
		Menu: model.MenuTaskCategory,
	}, nil
}

func (s *TaskService) GetTasks(
	ctx context.Context,
	telegramID int64,
) (*model.BotResponse, error) {

	user, err := s.users.GetByTelegramID(
		ctx,
		telegramID,
	)
	if err != nil {
		return nil, err
	}

	tasks, err := s.tasks.GetByUserID(
		ctx,
		user.ID,
	)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return &model.BotResponse{
			Text: "📋 У вас пока нет активных задач.",
			Menu: model.MenuMain,
		}, nil
	}

	return &model.BotResponse{
		Text: formatTasks(tasks),
		Menu: model.MenuMain,
	}, nil
}

func (s *TaskService) GetTasksByCategory(
	ctx context.Context,
	telegramID int64,
	category model.TaskCategory,
) (*model.BotResponse, error) {

	user, err := s.users.GetByTelegramID(
		ctx,
		telegramID,
	)
	if err != nil {
		return nil, err
	}

	tasks, err := s.tasks.GetByUserIDAndCategory(
		ctx,
		user.ID,
		category,
	)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {

		var categoryName string

		switch category {
		case model.TaskCategoryWork:
			categoryName = "рабочих"
		case model.TaskCategoryPersonal:
			categoryName = "личных"
		}

		return &model.BotResponse{
			Text: fmt.Sprintf(
				"📋 У вас пока нет %s задач.",
				categoryName,
			),
			Menu: model.MenuMain,
		}, nil
	}

	return &model.BotResponse{
		Text: formatTasks(tasks),
		Menu: model.MenuMain,
	}, nil
}

func (s *TaskService) HandleTaskMessage(ctx context.Context, telegramID int64, text string) (*model.BotResponse, error) {
	session, err := s.registrationTasks.GetByTelegramID(
		ctx,
		telegramID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return &model.BotResponse{
			Text: "Сначала нажми «➕ Добавить задачу».",
			Menu: model.MenuMain,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	text = strings.TrimSpace(text)

	if text == "" {
		return taskCreationPrompt(session.CurrentStep), nil
	}

	switch session.CurrentStep {

	case TaskStepAwaitCategory:
		return s.handleCategory(ctx, session, text)

	case TaskStepAwaitTitle:
		return s.handleTitle(ctx, session, text)

	case TaskStepAwaitDate:
		return s.handleDate(ctx, session, text)

	case TaskStepAwaitCustomDate:
		return s.handleCustomDate(ctx, session, text)

	case TaskStepAwaitTime:
		return s.handleTime(ctx, session, text)

	default:
		return nil, fmt.Errorf(
			"unknown task creation step: %q",
			session.CurrentStep,
		)
	}
}

func (s *TaskService) HasActiveCreation(
	ctx context.Context,
	telegramID int64,
) (bool, error) {

	_, err := s.registrationTasks.GetByTelegramID(
		ctx,
		telegramID,
	)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	return false, err
}

func taskCreationPrompt(step string) *model.BotResponse {
	switch step {

	case TaskStepAwaitCategory:
		return &model.BotResponse{
			Text: "📂 <b>Выбери категорию задачи</b>",
			Menu: model.MenuTaskCategory,
		}

	case TaskStepAwaitTitle:
		return &model.BotResponse{
			Text: "📝 <b>Что нужно сделать?</b>\n\nНапиши название задачи.",
			Menu: model.MenuNone,
		}

	case TaskStepAwaitDate:
		return &model.BotResponse{
			Text: "📅 <b>Когда выполнить задачу?</b>",
			Menu: model.MenuTaskDate,
		}

	case TaskStepAwaitCustomDate:
		return &model.BotResponse{
			Text: "📅 Введи дату в формате <b>ДД.ММ.ГГГГ</b>.",
			Menu: model.MenuNone,
		}

	case TaskStepAwaitTime:
		return &model.BotResponse{
			Text: "🕐 <b>Выбери время задачи</b>",
			Menu: model.MenuTaskTime,
		}

	default:
		return &model.BotResponse{
			Text: "⚠️ Не удалось определить этап создания задачи.",
			Menu: model.MenuMain,
		}
	}
}

func formatTasks(tasks []model.Task) string {
	var builder strings.Builder

	builder.WriteString(
		"📋 <b>Ваши задачи:</b>\n\n",
	)

	for i, task := range tasks {

		var category string

		switch task.Category {

		case model.TaskCategoryWork:
			category = "💼 Работа"

		case model.TaskCategoryPersonal:
			category = "🏠 Личное"
		}

		builder.WriteString(
			fmt.Sprintf(
				"%d. <b>%s</b>\n"+
					"%s\n"+
					"📅 %s\n\n",
				i+1,
				task.Title,
				category,
				task.DueAt.Format("02.01.2006 15:04"),
			),
		)
	}

	return builder.String()
}

func (s *TaskService) handleCategory(ctx context.Context, session *model.TaskCreationSession, text string) (*model.BotResponse, error) {
	var category model.TaskCategory

	switch text {

	case "💼 Работа":
		category = model.TaskCategoryWork

	case "🏠 Личное":
		category = model.TaskCategoryPersonal

	default:
		return &model.BotResponse{
			Text: "Выбери категорию с помощью кнопок 👇",
			Menu: model.MenuTaskCategory,
		}, nil
	}

	session.Category = &category
	session.CurrentStep = TaskStepAwaitTitle

	if err := s.registrationTasks.Update(
		ctx,
		*session,
	); err != nil {
		return nil, err
	}

	return &model.BotResponse{
		Text: "📝 <b>Что нужно сделать?</b>\n\nНапиши название задачи.",
		Menu: model.MenuNone,
	}, nil
}

func (s *TaskService) handleTitle(ctx context.Context, session *model.TaskCreationSession, text string) (*model.BotResponse, error) {
	if len([]rune(text)) < 2 {
		return &model.BotResponse{
			Text: "Название слишком короткое. Напиши задачу подробнее.",
			Menu: model.MenuNone,
		}, nil
	}

	title := text

	session.Title = &title
	session.CurrentStep = TaskStepAwaitDate

	if err := s.registrationTasks.Update(
		ctx,
		*session,
	); err != nil {
		return nil, err
	}

	return &model.BotResponse{
		Text: "📅 <b>Когда выполнить задачу?</b>",
		Menu: model.MenuTaskDate,
	}, nil
}

func (s *TaskService) handleDate(
	ctx context.Context,
	session *model.TaskCreationSession,
	text string,
) (*model.BotResponse, error) {
	now := time.Now()

	var taskDate time.Time

	switch text {

	case "Сегодня":
		taskDate = now

	case "Завтра":
		taskDate = now.AddDate(0, 0, 1)

	case "Другая дата":
		session.CurrentStep = TaskStepAwaitCustomDate

		if err := s.registrationTasks.Update(
			ctx,
			*session,
		); err != nil {
			return nil, err
		}

		return &model.BotResponse{
			Text: "📅 Введи дату в формате <b>ДД.ММ.ГГГГ</b>\n\nНапример: <code>25.10.2026</code>",
			Menu: model.MenuNone,
		}, nil

	default:
		return &model.BotResponse{
			Text: "Выбери дату с помощью кнопок 👇",
			Menu: model.MenuTaskDate,
		}, nil
	}

	taskDate = dateOnly(taskDate)

	session.TaskDate = &taskDate
	session.CurrentStep = TaskStepAwaitTime

	if err := s.registrationTasks.Update(
		ctx,
		*session,
	); err != nil {
		return nil, err
	}

	return &model.BotResponse{
		Text: "🕐 <b>Выбери время задачи</b>\n\nДоступное время: с 07:00 до 24:00.",
		Menu: model.MenuTaskTime,
	}, nil
}

func (s *TaskService) handleCustomDate(
	ctx context.Context,
	session *model.TaskCreationSession,
	text string,
) (*model.BotResponse, error) {

	location := appLocation()

	taskDate, err := time.ParseInLocation(
		"02.01.2006",
		text,
		location,
	)
	if err != nil {
		return &model.BotResponse{
			Text: "Некорректная дата.\n\nВведи её в формате <b>ДД.ММ.ГГГГ</b>.\nНапример: <code>25.10.2026</code>",
			Menu: model.MenuNone,
		}, nil
	}

	now := time.Now().In(location)

	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		location,
	)

	if taskDate.Before(today) {
		return &model.BotResponse{
			Text: "Нельзя создать задачу на прошедшую дату.",
			Menu: model.MenuNone,
		}, nil
	}

	session.TaskDate = &taskDate
	session.CurrentStep = TaskStepAwaitTime

	if err := s.registrationTasks.Update(
		ctx,
		*session,
	); err != nil {
		return nil, err
	}

	menu := model.MenuTaskTime

	if taskDate.Equal(today) {
		menu = model.MenuTaskTimeToday
	}

	return &model.BotResponse{
		Text: "🕐 <b>Выбери время задачи</b>",
		Menu: menu,
	}, nil
}

func (s *TaskService) handleTime(
	ctx context.Context,
	session *model.TaskCreationSession,
	text string,
) (*model.BotResponse, error) {

	if session.Category == nil ||
		session.Title == nil ||
		session.TaskDate == nil {

		return nil, errors.New("incomplete task creation session")
	}

	hour, minute, err := parseTaskTime(text)
	if err != nil {
		return &model.BotResponse{
			Text: "Некорректное время.\n\nВыбери время с 07:00 до 24:00.",
			Menu: model.MenuTaskTime,
		}, nil
	}

	dueAt := time.Date(
		session.TaskDate.Year(),
		session.TaskDate.Month(),
		session.TaskDate.Day(),
		hour,
		minute,
		0,
		0,
		time.Local,
	)

	// 24:00 фактически является 00:00 следующего дня
	if hour == 24 {
		dueAt = time.Date(
			session.TaskDate.Year(),
			session.TaskDate.Month(),
			session.TaskDate.Day(),
			0,
			0,
			0,
			0,
			time.Local,
		).AddDate(0, 0, 1)
	}

	if dueAt.Before(time.Now()) {
		return &model.BotResponse{
			Text: "Это время уже прошло. Выбери другое.",
			Menu: model.MenuTaskTime,
		}, nil
	}

	user, err := s.users.GetByTelegramID(
		ctx,
		session.TelegramID,
	)
	if err != nil {
		return nil, err
	}

	task := model.Task{
		UserID:   user.ID,
		Category: *session.Category,
		Title:    *session.Title,
		Status:   model.TaskStatusTodo,
		DueAt:    dueAt,
	}

	if err := s.tasks.Create(ctx, &task); err != nil {
		return nil, err
	}

	if err := s.registrationTasks.Delete(
		ctx,
		session.TelegramID,
	); err != nil {
		return nil, err
	}

	return &model.BotResponse{
		Text: fmt.Sprintf(
			"✅ <b>Задача добавлена!</b>\n\n"+
				"📝 %s\n"+
				"📅 %s\n"+
				"🕐 %s",
			*session.Title,
			dueAt.Format("02.01.2006"),
			dueAt.Format("15:04"),
		),
		Menu: model.MenuMain,
	}, nil
}

func dateOnly(t time.Time) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		0,
		0,
		0,
		0,
		t.Location(),
	)
}

func parseTaskTime(value string) (int, int, error) {

	value = strings.TrimSpace(value)

	if value == "24:00" {
		return 24, 0, nil
	}

	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return 0, 0, err
	}

	hour := parsed.Hour()
	minute := parsed.Minute()

	if hour < 7 {
		return 0, 0, errors.New("time must be after 07:00")
	}

	return hour, minute, nil
}

func appLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Dushanbe")
	if err != nil {
		return time.FixedZone("UTC+5", 5*60*60)
	}

	return location
}
