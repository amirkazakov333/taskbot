package service

import (
	"context"
	"fmt"
	"time"

	"telegram-bot/internal/model"
	"telegram-bot/internal/repository"
)

type ReminderService struct {
	reminders repository.ReminderRepository
}

func NewReminderService(
	reminders repository.ReminderRepository,
) *ReminderService {

	return &ReminderService{
		reminders: reminders,
	}
}

func (s *ReminderService) GetReminders(
	ctx context.Context,
	now time.Time,
	includePre bool,
	includeOverdue bool,
) ([]model.ReminderMessage, error) {

	var messages []model.ReminderMessage

	// =========================
	// За 30 минут
	// =========================

	until := now.Add(30 * time.Minute)

	preTasks, err := s.reminders.GetPreReminderTasks(
		ctx,
		now,
		until,
	)
	if err != nil {
		return nil, err
	}

	for _, task := range preTasks {
		dueAt := task.DueAt.In(appLocation())
		category := reminderCategory(task.Category)

		messages = append(
			messages,
			model.ReminderMessage{
				TaskID:     task.TaskID,
				TelegramID: task.TelegramID,
				Type:       model.ReminderTypePre,
				Text: fmt.Sprintf(
					"🔔 <b>Напоминание</b>\n\n"+
						"%s\n"+
						"📝 %s\n"+
						"🕐 Через 30 минут — %s",
					category,
					task.Title,
					dueAt.Format("15:04"),
				),
			},
		)
	}
	// =========================
	// Просроченные
	// =========================

	if includeOverdue {

		overdueTasks, err :=
			s.reminders.GetOverdueReminderTasks(
				ctx,
				now,
			)

		if err != nil {
			return nil, err
		}

		for _, task := range overdueTasks {
			dueAt := task.DueAt.In(appLocation())
			category := reminderCategory(task.Category)

			messages = append(
				messages,
				model.ReminderMessage{
					TaskID:     task.TaskID,
					TelegramID: task.TelegramID,
					Type:       model.ReminderTypeOverdue,
					Text: fmt.Sprintf(
						"⚠️ <b>Задача просрочена</b>\n\n"+
							"%s\n"+
							"📝 %s\n"+
							"🕐 Срок был: %s",
						category,
						task.Title,
						dueAt.Format("15:04"),
					),
				},
			)
		}
	}

	// =========================
	// Наступивший срок
	// =========================

	dueTasks, err :=
		s.reminders.GetDueReminderTasks(
			ctx,
			now,
		)

	if err != nil {
		return nil, err
	}

	for _, task := range dueTasks {
		dueAt := task.DueAt.In(appLocation())
		category := reminderCategory(task.Category)

		messages = append(
			messages,
			model.ReminderMessage{
				TaskID:     task.TaskID,
				TelegramID: task.TelegramID,
				Type:       model.ReminderTypeDue,
				Text: fmt.Sprintf(
					"⏰ <b>Время выполнить задачу</b>\n\n"+
						"%s\n"+
						"📝 %s\n"+
						"🕐 %s",
					category,
					task.Title,
					dueAt.Format("15:04"),
				),
			},
		)
	}

	return messages, nil
}

func (s *ReminderService) MarkSent(
	ctx context.Context,
	message model.ReminderMessage,
	sentAt time.Time,
) error {

	switch message.Type {

	case model.ReminderTypePre:

		return s.reminders.MarkPreReminderSent(
			ctx,
			message.TaskID,
		)

	case model.ReminderTypeDue:

		return s.reminders.MarkDueReminderSent(
			ctx,
			message.TaskID,
		)

	case model.ReminderTypeOverdue:

		return s.reminders.UpdateLastOverdueReminder(
			ctx,
			message.TaskID,
			sentAt,
		)

	default:
		return fmt.Errorf(
			"unknown reminder type: %s",
			message.Type,
		)
	}
}

func reminderCategory(
	category model.TaskCategory,
) string {

	switch category {

	case model.TaskCategoryWork:
		return "💼 Работа"

	case model.TaskCategoryPersonal:
		return "🏠 Личное"

	default:
		return ""
	}
}
