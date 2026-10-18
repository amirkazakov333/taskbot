package handler

import (
	"context"
	"log"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func StartHandler(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Привет!",
	})

	if err != nil {
		log.Printf("ошибка при отправке сообщения: %s", err.Error())
	}
}
