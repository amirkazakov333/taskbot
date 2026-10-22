package handler

import (
	"context"
	"log"
	"strings"

	"telegram-bot/internal/model"
	"telegram-bot/internal/service"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type MessageRouter struct {
	userHandler *UserHandler
	taskHandler *TaskHandler
	taskService *service.TaskService
}

func NewMessageRouter(
	userHandler *UserHandler,
	taskHandler *TaskHandler,
	taskService *service.TaskService,
) *MessageRouter {

	return &MessageRouter{
		userHandler: userHandler,
		taskHandler: taskHandler,
		taskService: taskService,
	}
}

func (r *MessageRouter) Message(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	text := strings.TrimSpace(update.Message.Text)

	switch text {

	// =========================
	// Главное меню
	// =========================

	case ButtonAddTask:
		r.taskHandler.StartCreateTask(
			ctx,
			b,
			update,
		)
		return

	case ButtonMyTasks:
		r.taskHandler.ShowTasksMenu(
			ctx,
			b,
			update,
		)
		return

	case ButtonSettings:
		sendResponse(
			ctx,
			b,
			update.Message.Chat.ID,
			&model.BotResponse{
				Text: "⚙️ Настройки пока в разработке.",
				Menu: model.MenuMain,
			},
		)
		return

	// =========================
	// Фильтр "Мои задачи"
	// =========================

	case ButtonWorkTasks:
		r.taskHandler.GetWorkTasks(
			ctx,
			b,
			update,
		)
		return

	case ButtonPersonalTasks:
		r.taskHandler.GetPersonalTasks(
			ctx,
			b,
			update,
		)
		return

	case ButtonAllTasks:
		r.taskHandler.GetAllTasks(
			ctx,
			b,
			update,
		)
		return
	}

	// =========================
	// Создание задачи
	// =========================

	activeTaskCreation, err := r.taskService.HasActiveCreation(
		ctx,
		update.Message.From.ID,
	)

	if err != nil {
		log.Printf(
			"check task creation session error: %v",
			err,
		)

		sendResponse(
			ctx,
			b,
			update.Message.Chat.ID,
			&model.BotResponse{
				Text: "⚠️ Не удалось обработать запрос.",
				Menu: model.MenuMain,
			},
		)

		return
	}

	if activeTaskCreation {
		r.taskHandler.Message(
			ctx,
			b,
			update,
		)
		return
	}

	// =========================
	// Регистрация пользователя
	// =========================

	r.userHandler.Message(
		ctx,
		b,
		update,
	)
}
