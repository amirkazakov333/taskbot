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

	log.Printf("ROUTER: received %q", text)

	if text == ButtonAddTask {
		r.taskHandler.StartCreateTask(
			ctx,
			b,
			update,
		)
		return
	}
	// Проверяем, находится ли пользователь
	// сейчас внутри создания задачи
	activeTaskCreation, err := r.taskService.HasActiveCreation(
		ctx,
		update.Message.From.ID,
	)

	if err != nil {
		log.Printf("check task creation session error: %v", err)

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

	// Пока эти разделы ещё не реализованы
	switch text {

	case ButtonMyTasks:

		sendResponse(
			ctx,
			b,
			update.Message.Chat.ID,
			&model.BotResponse{
				Text: "📋 Раздел «Мои задачи» скоро добавим.",
				Menu: model.MenuMain,
			},
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
	}

	// Если task creation нет,
	// сообщение отправляется в регистрацию
	r.userHandler.Message(
		ctx,
		b,
		update,
	)
}
