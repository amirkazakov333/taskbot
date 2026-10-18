package handler

import (
	"context"
	"telegram-bot/internal/service"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{
		userService: &service,
	}
}

func (h UserHandler) StartHandler(ctx context.Context, b *tgBot.Bot, update *models.Update) {
	message, err := h.userService.Start(ctx, update)
	if err != nil {

	}

	_, err = b.SendMessage(ctx, &tgBot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      message,
		ParseMode: models.ParseModeHTML,
	})
}
