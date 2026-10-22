package main

import (
	"context"
	"log"
	"time"

	"telegram-bot/internal/config"
	"telegram-bot/internal/handler"
	"telegram-bot/internal/reminder"
	"telegram-bot/internal/repository"
	"telegram-bot/internal/service"

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

	// =========================
	// Repositories
	// =========================

	usersRepository :=
		repository.NewUserPostgresRepository(pool)

	userSessionRepository :=
		repository.NewUserSessionPostgresRepository(pool)

	taskRepository :=
		repository.NewTaskPostgresRepository(pool)

	reminderRepository :=
		repository.NewReminderPostgresRepository(pool)

	// =========================
	// Services
	// =========================

	userService := service.NewUserService(
		usersRepository,
		userSessionRepository,
	)

	taskService := service.NewTaskService(
		usersRepository,
		taskRepository,
		userSessionRepository,
	)

	reminderService := service.NewReminderService(
		reminderRepository,
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

	reminderWorker := reminder.NewWorker(
		reminderService,
		tgBot,
	)

	// /start обрабатывается UserHandler.
	tgBot.RegisterHandler(
		tgbot.HandlerTypeMessageText,
		"/start",
		tgbot.MatchTypeExact,
		userHandler.StartHandler,
	)

	// Все остальные текстовые сообщения идут через Router.
	tgBot.RegisterHandler(
		tgbot.HandlerTypeMessageText,
		"",
		tgbot.MatchTypePrefix,
		messageRouter.Message,
	)

	appCtx, appCancel := context.WithCancel(
		context.Background(),
	)
	defer appCancel()

	go reminderWorker.Start(appCtx)

	log.Println("Telegram-бот запущен")

	tgBot.Start(appCtx)
}
