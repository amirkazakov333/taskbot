package handler

import (
	"context"
	"log"
	"strconv"
	"strings"

	"telegram-bot/internal/model"
	"telegram-bot/internal/reminder"
	"telegram-bot/internal/service"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	ButtonTaskDone      = "✅ Выполнено"
	ButtonTaskCancelled = "❌ Отменить задачу"
	ButtonTaskTransfer  = "🔄 Перенести задачу"
)

type ReminderHandler struct {
	reminderService *service.ReminderService
}

func NewReminderHandler(
	reminderService *service.ReminderService,
) *ReminderHandler {
	return &ReminderHandler{
		reminderService: reminderService,
	}
}

func (r ReminderHandler) ChangeTaskStatusToDone(ctx context.Context, b *tgBot.Bot, update *models.Update) {
	callback := update.CallbackQuery

	parts := strings.Split(callback.Data, ":")

	if len(parts) != 3 || parts[0] != "task" {
		return
	}

	action := parts[1]

	taskID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return
	}

	err = r.reminderService.ChangeTaskStatus(ctx, taskID, model.TaskStatus(action))

	if err != nil {
		send(
			ctx,
			b,
			update.Message.Chat.ID,
			&model.TaskReminderNotification{
				Text:   "⚠️ Не удалось изменить статус задачи.",
				Menu:   model.MenuTaskChangeStatus,
				TaskID: taskID,
			},
		)

		return
	}

	send(
		ctx,
		b,
		update.Message.Chat.ID,
		&model.TaskReminderNotification{
			Text:   "Отличная работа! Задача успешно отмечена как выполненная.",
			Menu:   model.MenuMain,
			TaskID: taskID,
		},
	)
}

func send(
	ctx context.Context,
	b *tgBot.Bot,
	chatID int64,
	response *model.TaskReminderNotification,
) {
	if response == nil || response.Text == "" {
		return
	}

	params := &tgBot.SendMessageParams{
		ChatID:    chatID,
		Text:      response.Text,
		ParseMode: models.ParseModeHTML,
	}

	switch response.Menu {

	case model.MenuMain:
		params.ReplyMarkup = mainMenu()

	case model.MenuTaskChangeStatus:
		params.ReplyMarkup = reminder.TaskReminderKeyboard(response.TaskID)

		_, err := b.SendMessage(ctx, params)
		if err != nil {
			log.Printf("send message error: %v", err)
		}
	}
}
