package handler

import (
	"context"
	"log"

	"telegram-bot/internal/model"
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
	userService     *service.UserService
}

func NewReminderHandler(
	reminderService *service.ReminderService,
) *ReminderHandler {
	return &ReminderHandler{
		reminderService: reminderService,
	}
}

func (r ReminderHandler) ChangeTaskStatusToDone(ctx context.Context, b *tgBot.Bot, update *models.Update) {
	// response, err :=
}

func send(
	ctx context.Context,
	b *tgBot.Bot,
	chatID int64,
	response *model.BotResponse,
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

		_, err := b.SendMessage(ctx, params)
		if err != nil {
			log.Printf("send message error: %v", err)
		}
	}
}
