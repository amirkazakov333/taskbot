package handler

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"telegram-bot/internal/model"
	"telegram-bot/internal/service"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	ButtonWorkTasks     = "💼 Рабочие"
	ButtonPersonalTasks = "🏠 Личные"
	ButtonAllTasks      = "📋 Все задачи"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(
	taskService *service.TaskService,
) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func (h *TaskHandler) StartCreateTask(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	log.Println("TASK HANDLER: StartCreateTask called")
	if update.Message == nil {
		return
	}

	response, err := h.taskService.StartCreateTask(
		ctx,
		update.Message.From.ID,
	)

	if err != nil {
		log.Printf("start create task error: %v", err)

		sendResponse(
			ctx,
			b,
			update.Message.Chat.ID,
			&model.BotResponse{
				Text: "⚠️ Не удалось начать создание задачи.",
				Menu: model.MenuMain,
			},
		)

		return
	}

	sendResponse(
		ctx,
		b,
		update.Message.Chat.ID,
		response,
	)
}

func (h *TaskHandler) ShowTasksMenu(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	sendResponse(
		ctx,
		b,
		update.Message.Chat.ID,
		&model.BotResponse{
			Text: "📋 <b>Какие задачи показать?</b>",
			Menu: model.MenuTaskFilter,
		},
	)
}

func (h *TaskHandler) GetAllTasks(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	response, err := h.taskService.GetTasks(
		ctx,
		update.Message.From.ID,
	)

	if err != nil {
		log.Printf("get all tasks error: %v", err)

		sendResponse(
			ctx,
			b,
			update.Message.Chat.ID,
			&model.BotResponse{
				Text: "⚠️ Не удалось получить задачи.",
				Menu: model.MenuMain,
			},
		)

		return
	}

	sendResponse(
		ctx,
		b,
		update.Message.Chat.ID,
		response,
	)
}

func (h *TaskHandler) GetWorkTasks(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	response, err := h.taskService.GetTasksByCategory(
		ctx,
		update.Message.From.ID,
		model.TaskCategoryWork,
	)

	if err != nil {
		log.Printf("get work tasks error: %v", err)

		sendResponse(
			ctx,
			b,
			update.Message.Chat.ID,
			&model.BotResponse{
				Text: "⚠️ Не удалось получить рабочие задачи.",
				Menu: model.MenuMain,
			},
		)

		return
	}

	sendResponse(
		ctx,
		b,
		update.Message.Chat.ID,
		response,
	)
}

func (h *TaskHandler) GetPersonalTasks(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	response, err := h.taskService.GetTasksByCategory(
		ctx,
		update.Message.From.ID,
		model.TaskCategoryPersonal,
	)

	if err != nil {
		log.Printf("get personal tasks error: %v", err)

		sendResponse(
			ctx,
			b,
			update.Message.Chat.ID,
			&model.BotResponse{
				Text: "⚠️ Не удалось получить личные задачи.",
				Menu: model.MenuMain,
			},
		)

		return
	}

	sendResponse(
		ctx,
		b,
		update.Message.Chat.ID,
		response,
	)
}

func (h *TaskHandler) Message(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	response, err := h.taskService.HandleTaskMessage(
		ctx,
		update.Message.From.ID,
		update.Message.Text,
	)

	if err != nil {
		log.Printf("task message error: %v", err)

		sendResponse(
			ctx,
			b,
			update.Message.Chat.ID,
			&model.BotResponse{
				Text: "⚠️ Не удалось обработать задачу. Попробуй ещё раз.",
				Menu: model.MenuMain,
			},
		)

		return
	}

	sendResponse(
		ctx,
		b,
		update.Message.Chat.ID,
		response,
	)
}

func (h *TaskHandler) HandleTaskCallback(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	if update.CallbackQuery == nil {
		return
	}

	callback := update.CallbackQuery

	// Сразу сообщаем Telegram, что получили нажатие.
	_, err := b.AnswerCallbackQuery(
		ctx,
		&tgBot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
		},
	)
	if err != nil {
		log.Printf("answer callback query error: %v", err)
	}

	// Наши кнопки прикреплены к обычным сообщениям бота.
	if callback.Message.Message == nil {
		log.Println("callback has no accessible message")
		return
	}

	chatID := callback.Message.Message.Chat.ID

	// Формат: task:done:42
	parts := strings.Split(callback.Data, ":")

	if len(parts) != 3 || parts[0] != "task" {
		sendResponse(
			ctx,
			b,
			chatID,
			&model.BotResponse{
				Text: "⚠️ Не удалось определить действие.",
				Menu: model.MenuMain,
			},
		)
		return
	}

	taskID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || taskID <= 0 {
		sendResponse(
			ctx,
			b,
			chatID,
			&model.BotResponse{
				Text: "⚠️ Некорректный идентификатор задачи.",
				Menu: model.MenuMain,
			},
		)
		return
	}

	telegramID := callback.From.ID

	var response *model.BotResponse

	switch parts[1] {
	case "done":
		response, err = h.taskService.CompleteTask(
			ctx,
			telegramID,
			taskID,
		)

	case "cancel":
		response, err = h.taskService.CancelTask(
			ctx,
			telegramID,
			taskID,
		)

	case "reschedule":
		response, err = h.taskService.StartRescheduleTask(
			ctx,
			telegramID,
			taskID,
		)

	default:
		response = &model.BotResponse{
			Text: "⚠️ Неизвестное действие с задачей.",
			Menu: model.MenuMain,
		}
	}

	if err != nil {
		log.Printf(
			"task callback error: action=%s task_id=%d: %v",
			parts[1],
			taskID,
			err,
		)

		response = &model.BotResponse{
			Text: "⚠️ Не удалось обработать действие. Попробуй ещё раз.",
			Menu: model.MenuMain,
		}
	}

	sendResponse(ctx, b, chatID, response)
}

func sendResponse(
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

	case model.MenuTaskCategory:
		params.ReplyMarkup = taskCategoryMenu()

	case model.MenuTaskDate:
		params.ReplyMarkup = taskDateMenu()

	case model.MenuTaskTime:
		params.ReplyMarkup = taskTimeMenu(false)

	case model.MenuTaskTimeToday:
		params.ReplyMarkup = taskTimeMenu(true)

	case model.MenuTaskFilter:
		params.ReplyMarkup = taskFilterMenu()

	case model.MenuNone:
		params.ReplyMarkup = &models.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		}
	}

	_, err := b.SendMessage(ctx, params)
	if err != nil {
		log.Printf("send message error: %v", err)
	}
}

func taskCategoryMenu() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: "💼 Работа"},
				{Text: "🏠 Личное"},
			},
		},
		ResizeKeyboard: true,
		IsPersistent:   true,
	}
}

func taskDateMenu() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: "Сегодня"},
				{Text: "Завтра"},
			},
			{
				{Text: "Другая дата"},
			},
		},
		ResizeKeyboard: true,
		IsPersistent:   true,
	}
}

func taskTimeMenu(
	fromCurrentTime bool,
) *models.ReplyKeyboardMarkup {

	var keyboard [][]models.KeyboardButton
	var row []models.KeyboardButton

	startMinutes := 7 * 60

	if fromCurrentTime {

		location, err := time.LoadLocation("Asia/Dushanbe")
		if err != nil {
			location = time.FixedZone(
				"UTC+5",
				5*60*60,
			)
		}

		now := time.Now().In(location)

		currentMinutes :=
			now.Hour()*60 + now.Minute()

		// Берём следующий 30-минутный интервал.
		// 12:00 -> 12:30
		// 12:10 -> 12:30
		// 12:30 -> 13:00
		startMinutes =
			(currentMinutes/30 + 1) * 30

		// Раньше 07:00 задачи создавать нельзя
		if startMinutes < 7*60 {
			startMinutes = 7 * 60
		}
	}

	for totalMinutes := startMinutes; totalMinutes <= 24*60; totalMinutes += 30 {

		hour := totalMinutes / 60
		minute := totalMinutes % 60

		text := fmt.Sprintf(
			"%02d:%02d",
			hour,
			minute,
		)

		row = append(
			row,
			models.KeyboardButton{
				Text: text,
			},
		)

		if len(row) == 4 {
			keyboard = append(
				keyboard,
				row,
			)

			row = nil
		}
	}

	if len(row) > 0 {
		keyboard = append(
			keyboard,
			row,
		)
	}

	return &models.ReplyKeyboardMarkup{
		Keyboard:       keyboard,
		ResizeKeyboard: true,
		IsPersistent:   true,
	}
}

func taskFilterMenu() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: ButtonWorkTasks},
				{Text: ButtonPersonalTasks},
			},
			{
				{Text: ButtonAllTasks},
			},
		},
		ResizeKeyboard: true,
		IsPersistent:   true,
	}
}
