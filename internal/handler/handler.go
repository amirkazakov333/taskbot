package handler

import (
	"context"
	"log"
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
		log.Printf("start handler error: %v", err)
		h.send(ctx, b, update.Message.Chat.ID, "⚠️ Не удалось обработать запрос. Попробуй ещё раз.")
		return
	}

	_, err = b.SendMessage(ctx, &tgBot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      message,
		ParseMode: models.ParseModeHTML,
	})
}

func (h *UserHandler) Message(ctx context.Context, b *tgBot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	text, err := h.userService.HandleMessage(
		ctx,
		update.Message.From.ID,
		update.Message.Text,
	)

	if err != nil {
		log.Printf("message handler error: %v", err)
		h.send(ctx, b, update.Message.Chat.ID,
			"⚠️ Произошла ошибка при обработке сообщения. Попробуй ещё раз.")
		return
	}

	h.send(ctx, b, update.Message.Chat.ID, text)
}

func (h *UserHandler) send(ctx context.Context, b *tgBot.Bot, chatID int64, text string) {
	if text == "" {
		return
	}

	_, err := b.SendMessage(ctx, &tgBot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})

	if err != nil {
		log.Printf("send message error: %v", err)
	}
}
