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
	ButtonAddTask  = "➕ Добавить задачу"
	ButtonMyTasks  = "📋 Мои задачи"
	ButtonSettings = "⚙️ Настройки"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{
		userService: &service,
	}
}

func (h *UserHandler) StartHandler(ctx context.Context, b *tgBot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	response, err := h.userService.Start(ctx, update)
	if err != nil {
		log.Printf("start handler error: %v", err)

		h.send(
			ctx,
			b,
			update.Message.Chat.ID,
			&model.BotResponse{
				Text: "⚠️ Не удалось обработать запрос. Попробуй ещё раз.",
				Menu: model.MenuNone,
			},
		)

		return
	}

	h.send(
		ctx,
		b,
		update.Message.Chat.ID,
		response,
	)
}

func (h *UserHandler) Message(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	response, err := h.userService.HandleMessage(
		ctx,
		update.Message.From.ID,
		update.Message.Text,
	)

	if err != nil {
		log.Printf("message handler error: %v", err)

		h.send(
			ctx,
			b,
			update.Message.Chat.ID,
			&model.BotResponse{
				Text: "⚠️ Произошла ошибка при обработке сообщения. Попробуй ещё раз.",
				Menu: model.MenuNone,
			},
		)

		return
	}

	h.send(
		ctx,
		b,
		update.Message.Chat.ID,
		response,
	)
}

func (h *UserHandler) send(
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

	case model.MenuNone:
		// Клавиатуру не добавляем
	}

	_, err := b.SendMessage(ctx, params)
	if err != nil {
		log.Printf("send message error: %v", err)
	}
}

func mainMenu() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{
					Text: ButtonAddTask,
				},
			},
			{
				{
					Text: ButtonMyTasks,
				},
				{
					Text: ButtonSettings,
				},
			},
		},

		ResizeKeyboard:        true,
		IsPersistent:          true,
		InputFieldPlaceholder: "Выбери действие",
	}
}
