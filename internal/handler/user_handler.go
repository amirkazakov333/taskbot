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

func NewUserHandler(
	userService *service.UserService,
) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) HasActiveSettings(
	ctx context.Context,
	telegramID int64,
) (bool, error) {
	return h.userService.HasActiveSettings(ctx, telegramID)
}

func (h *UserHandler) StartHandler(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
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

	h.send(ctx, b, update.Message.Chat.ID, response)
}

func (h *UserHandler) Message(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	response, err := h.userService.HandleStartMessage(
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

	h.send(ctx, b, update.Message.Chat.ID, response)
}

// Открытие настроек профиля.
func (h *UserHandler) SettingsHandler(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	response, err := h.userService.StartSettings(
		ctx,
		update.Message.From.ID,
	)

	if err != nil {
		log.Printf("settings handler error: %v", err)

		response = &model.BotResponse{
			Text: "⚠️ Не удалось открыть настройки. Попробуй ещё раз.",
			Menu: model.MenuMain,
		}
	}

	h.send(ctx, b, update.Message.Chat.ID, response)
}

// Обработка ввода новых данных пользователя.
func (h *UserHandler) SettingsMessage(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	response, err := h.userService.HandleSettingsMessage(
		ctx,
		update.Message.From.ID,
		update.Message.Text,
	)

	if err != nil {
		log.Printf("settings message error: %v", err)

		response = &model.BotResponse{
			Text: "⚠️ Не удалось сохранить изменения. Попробуй ещё раз.",
			Menu: model.MenuSettingsInput,
		}
	}

	h.send(ctx, b, update.Message.Chat.ID, response)
}

// Универсальная отправка сообщений с выбором клавиатуры.
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

	case model.MenuSettings:
		params.ReplyMarkup = settingsMenu()

	case model.MenuSettingsInput:
		params.ReplyMarkup = settingsInputMenu()

	case model.MenuNone:
		// Клавиатуру не добавляем.
	}

	_, err := b.SendMessage(ctx, params)
	if err != nil {
		log.Printf("send message error: %v", err)
	}
}

// Главное меню.
func mainMenu() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: ButtonAddTask},
			},
			{
				{Text: ButtonMyTasks},
				{Text: ButtonSettings},
			},
		},
		ResizeKeyboard:        true,
		IsPersistent:          true,
		InputFieldPlaceholder: "Выбери действие",
	}
}

// Меню настроек профиля.
func settingsMenu() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: model.SettingsChangeFirstName},
			},
			{
				{Text: model.SettingsChangeLastName},
			},
			{
				{Text: model.SettingsChangePreferredName},
			},
			{
				{Text: model.SettingsBack},
			},
		},
		ResizeKeyboard:        true,
		IsPersistent:          true,
		InputFieldPlaceholder: "Выбери настройку",
	}
}

// Клавиатура при вводе нового значения.
func settingsInputMenu() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: model.SettingsBackToMenu},
			},
			{
				{Text: model.SettingsCancel},
			},
		},
		ResizeKeyboard:        true,
		IsPersistent:          true,
		InputFieldPlaceholder: "Введи новое значение",
	}
}
