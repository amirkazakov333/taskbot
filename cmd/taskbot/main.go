package main

import (
	"context"
	"log"
	"telegram-bot/internal/config"
	"telegram-bot/internal/handler"
	"telegram-bot/internal/repository"
	"time"

	tgbot "github.com/go-telegram/bot"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	pool, err := repository.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("PostgreSQL connection failed: %v", err)
	}
	defer pool.Close()

	log.Println("PostgreSQL connection successful")

	tgBot, err := tgbot.New(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("не удалось создать Telegram-бота: %v", err)
	}

	tgBot.RegisterHandler(
		tgbot.HandlerTypeMessageText,
		"/start",
		tgbot.MatchTypeExact,
		handler.StartHandler,
	)

	log.Println("Telegram-бот запущен")
	tgBot.Start(context.Background())
}
