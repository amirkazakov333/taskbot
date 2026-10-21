package main

import (
	"context"
	"log"
	"time"

	"telegram-bot/internal/config"
	"telegram-bot/internal/handler"
	"telegram-bot/internal/repository"
	"telegram-bot/internal/service"

	tgbot "github.com/go-telegram/bot"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// Контекст только для первоначального подключения к PostgreSQL
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

	// =========================
	// Repositories
	// =========================

	usersRepository :=
		repository.NewUserPostgresRepository(pool)

	registrationSessionRepository :=
		repository.NewRegistrationSessionPostgresRepository(pool)

	taskRegistrationRepository :=
		repository.NewTaskRegistrationPostgresRepository(pool)

	taskRepository :=
		repository.NewTaskPostgresRepository(pool)

	// =========================
	// Services
	// =========================

	userService := service.NewUserService(
		usersRepository,
		registrationSessionRepository,
	)

	taskService := service.NewTaskService(
		usersRepository,
		taskRepository,
		taskRegistrationRepository,
	)

	// =========================
	// Handlers
	// =========================

	userHandler := handler.NewUserHandler(
		*userService,
	)

	taskHandler := handler.NewTaskHandler(
		taskService,
	)

	messageRouter := handler.NewMessageRouter(
		userHandler,
		taskHandler,
		taskService,
	)

	// =========================
	// Telegram Bot
	// =========================

	tgBot, err := tgbot.New(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf(
			"не удалось создать Telegram-бота: %v",
			err,
		)
	}

	// /start обрабатывается UserHandler
	tgBot.RegisterHandler(
		tgbot.HandlerTypeMessageText,
		"/start",
		tgbot.MatchTypeExact,
		userHandler.StartHandler,
	)

	// Все остальные сообщения идут через Router
	tgBot.RegisterHandler(
		tgbot.HandlerTypeMessageText,
		"",
		tgbot.MatchTypePrefix,
		messageRouter.Message,
	)

	log.Println("Telegram-бот запущен")

	tgBot.Start(context.Background())
}
