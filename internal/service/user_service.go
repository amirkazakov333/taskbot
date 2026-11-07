package service

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"

	"telegram-bot/internal/model"
	"telegram-bot/internal/repository"

	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5"
)

const (
	stepAwaitFirstName      = "await_first_name"
	stepAwaitLastName       = "await_last_name"
	stepAwaitPreferredName  = "await_preferred_name"
	settingsStepChooseField = "settings_choose_field"
)

type UserService struct {
	users    repository.UserRepository
	sessions repository.UserSessionRepository
}

func NewUserService(
	users repository.UserRepository,
	sessions repository.UserSessionRepository,
) *UserService {
	return &UserService{
		users:    users,
		sessions: sessions,
	}
}

func (s *UserService) HasActiveSettings(
	ctx context.Context,
	telegramID int64,
) (bool, error) {

	session, err := s.sessions.GetByTelegramID(
		ctx,
		telegramID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return session.SessionType == model.SessionTypeSettings, nil
}

func (s *UserService) Start(
	ctx context.Context,
	update *models.Update,
) (*model.BotResponse, error) {

	if update.Message == nil {
		return nil, errors.New("message is nil")
	}

	telegramID := update.Message.From.ID

	// 1. Проверяем, зарегистрирован ли пользователь.
	user, err := s.users.GetByTelegramID(ctx, telegramID)

	if err == nil {
		return &model.BotResponse{
			Text: fmt.Sprintf(
				"<b>👋 С возвращением, %s!</b>\n\n"+
					"📋 <b>TaskBot</b> — твой помощник по задачам.\n\n"+
					"Планируй дела, контролируй прогресс "+
					"и не забывай о важных задачах.\n\n"+
					"Выбери действие ниже 👇",
				html.EscapeString(user.PreferredName),
			),
			Menu: model.MenuMain,
		}, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// 2. Проверяем общую сессию пользователя.
	session, err := s.sessions.GetByTelegramID(
		ctx,
		telegramID,
	)

	if err == nil {
		switch session.SessionType {
		case model.SessionTypeRegistration:
			return &model.BotResponse{
				Text: registrationPrompt(session.CurrentStep),
				Menu: model.MenuNone,
			}, nil

		case model.SessionTypeTaskCreation:
			return &model.BotResponse{
				Text: "У тебя уже начато создание задачи. Продолжи этот процесс.",
				Menu: model.MenuMain,
			}, nil

		case model.SessionTypeTaskReschedule:
			return &model.BotResponse{
				Text: "У тебя уже начат перенос задачи. Продолжи этот процесс.",
				Menu: model.MenuMain,
			}, nil

		default:
			return nil, fmt.Errorf(
				"unknown session type: %q",
				session.SessionType,
			)
		}
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// 3. Создаём новую сессию регистрации.
	var username *string

	if name := update.Message.From.Username; name != "" {
		username = &name
	}

	newSession := model.UserSession{
		TelegramID:  telegramID,
		SessionType: model.SessionTypeRegistration,
		Username:    username,
		CurrentStep: stepAwaitFirstName,
	}

	if err := s.sessions.Create(ctx, newSession); err != nil {
		return nil, err
	}

	return &model.BotResponse{
		Text: "<b>👋 Привет!</b>\n\n" +
			"Добро пожаловать в <b>TaskBot</b> — твоего личного помощника по задачам! 📋\n\n" +
			"Я помогу тебе планировать дела, следить за их выполнением " +
			"и вовремя получать напоминания. ⏰\n\n" +
			"📝 <b>Давай познакомимся!</b>\n\n" +
			"Для начала регистрации напиши своё имя.",
		Menu: model.MenuNone,
	}, nil
}

func (s *UserService) StartSettings(
	ctx context.Context,
	telegramID int64,
) (*model.BotResponse, error) {

	user, err := s.users.GetByTelegramID(ctx, telegramID)

	if errors.Is(err, pgx.ErrNoRows) {
		return &model.BotResponse{
			Text: "Сначала зарегистрируйся через /start.",
			Menu: model.MenuMain,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	session, err := s.sessions.GetByTelegramID(
		ctx,
		telegramID,
	)

	if err == nil {
		// Если у пользователя уже есть другой процесс, не разрешаем одновременно открыть настройки.
		if session.SessionType != model.SessionTypeSettings {
			return &model.BotResponse{
				Text: "Сначала заверши текущий процесс.",
				Menu: model.MenuMain,
			}, nil
		}

		// Если настройки уже открыты и пользователь находится на этапе ввода, восстанавливаем его этап.
		if session.CurrentStep != settingsStepChooseField {
			return &model.BotResponse{
				Text: settingsInputPrompt(session.CurrentStep),
				Menu: model.MenuSettingsInput,
			}, nil
		}

		return settingsResponse(user, ""), nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	newSession := model.UserSession{
		TelegramID:  telegramID,
		SessionType: model.SessionTypeSettings,
		CurrentStep: settingsStepChooseField,
	}

	if err := s.sessions.Create(ctx, newSession); err != nil {
		return nil, err
	}

	return settingsResponse(user, ""), nil
}

func (s *UserService) HandleStartMessage(
	ctx context.Context,
	telegramID int64,
	text string,
) (*model.BotResponse, error) {

	// 1. Проверяем, зарегистрирован ли пользователь.
	_, err := s.users.GetByTelegramID(ctx, telegramID)

	if err == nil {
		return &model.BotResponse{
			Text: "Выбери нужное действие 👇",
			Menu: model.MenuMain,
		}, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// 2. Получаем общую сессию.
	session, err := s.sessions.GetByTelegramID(
		ctx,
		telegramID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return &model.BotResponse{
			Text: "Для начала работы напиши /start.",
			Menu: model.MenuNone,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	// Этот метод обрабатывает только регистрацию.
	if session.SessionType != model.SessionTypeRegistration {
		return &model.BotResponse{
			Text: "Сейчас у тебя активен другой процесс. Продолжи его, прежде чем начинать регистрацию.",
			Menu: model.MenuMain,
		}, nil
	}

	text = strings.TrimSpace(text)

	if text == "" || strings.HasPrefix(text, "/") {
		return &model.BotResponse{
			Text: registrationPrompt(session.CurrentStep),
			Menu: model.MenuNone,
		}, nil
	}

	// 3. Обрабатываем текущий этап регистрации.
	switch session.CurrentStep {

	case stepAwaitFirstName:
		session.FirstName = text
		session.CurrentStep = stepAwaitLastName

		if err := s.sessions.Update(ctx, *session); err != nil {
			return nil, err
		}

		return &model.BotResponse{
			Text: registrationPrompt(session.CurrentStep),
			Menu: model.MenuNone,
		}, nil

	case stepAwaitLastName:
		session.LastName = text
		session.CurrentStep = stepAwaitPreferredName

		if err := s.sessions.Update(ctx, *session); err != nil {
			return nil, err
		}

		return &model.BotResponse{
			Text: registrationPrompt(session.CurrentStep),
			Menu: model.MenuNone,
		}, nil

	case stepAwaitPreferredName:
		switch strings.ToLower(text) {
		case "по имени":
			session.PreferredName = session.FirstName

		case "по фамилии":
			session.PreferredName = session.LastName

		default:
			session.PreferredName = text
		}

		return s.completeRegistration(ctx, session)

	default:
		return nil, fmt.Errorf(
			"unknown registration step: %q",
			session.CurrentStep,
		)
	}
}

func (s *UserService) HandleSettingsMessage(
	ctx context.Context,
	telegramID int64,
	text string,
) (*model.BotResponse, error) {

	text = strings.TrimSpace(text)

	user, err := s.users.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	session, err := s.sessions.GetByTelegramID(
		ctx,
		telegramID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return &model.BotResponse{
			Text: "Открой настройки профиля заново.",
			Menu: model.MenuMain,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	if session.SessionType != model.SessionTypeSettings {
		return &model.BotResponse{
			Text: "Сейчас у тебя не активен процесс настройки профиля.",
			Menu: model.MenuMain,
		}, nil
	}

	// ВЫБОР ПОЛЯ

	if session.CurrentStep == settingsStepChooseField {

		switch text {

		case model.SettingsChangeFirstName:
			session.CurrentStep = stepAwaitFirstName

		case model.SettingsChangeLastName:
			session.CurrentStep = stepAwaitLastName

		case model.SettingsChangePreferredName:
			session.CurrentStep = stepAwaitPreferredName

		case model.SettingsBack:

			if err := s.sessions.Delete(ctx, telegramID); err != nil {
				return nil, err
			}

			return &model.BotResponse{
				Text: "Возвращаемся в главное меню 👋",
				Menu: model.MenuMain,
			}, nil

		default:
			return settingsResponse(
				user,
				"Выбери настройку с помощью кнопок 👇",
			), nil
		}

		if err := s.sessions.Update(ctx, *session); err != nil {
			return nil, err
		}

		return &model.BotResponse{
			Text: settingsInputPrompt(session.CurrentStep),
			Menu: model.MenuSettingsInput,
		}, nil
	}

	// ВОЗВРАТ ИЛИ ОТМЕНА

	if text == model.SettingsCancel {

		if err := s.sessions.Delete(ctx, telegramID); err != nil {
			return nil, err
		}

		return &model.BotResponse{
			Text: "Изменение настроек отменено.",
			Menu: model.MenuMain,
		}, nil
	}

	if text == model.SettingsBackToMenu {

		session.CurrentStep = settingsStepChooseField

		if err := s.sessions.Update(ctx, *session); err != nil {
			return nil, err
		}

		return settingsResponse(user, ""), nil
	}

	// ПРОВЕРКА ВВЕДЁННОГО ЗНАЧЕНИЯ

	if err := validateSettingsValue(text); err != nil {
		return &model.BotResponse{
			Text: err.Error(),
			Menu: model.MenuSettingsInput,
		}, nil
	}

	// ОПРЕДЕЛЯЕМ ИЗМЕНЯЕМОЕ ПОЛЕ

	var field model.UserData
	var successText string

	switch session.CurrentStep {

	case stepAwaitFirstName:
		field = model.UserFirstName
		successText = "✅ Имя успешно изменено."

	case stepAwaitLastName:
		field = model.UserLastName
		successText = "✅ Фамилия успешно изменена."

	case stepAwaitPreferredName:
		field = model.UserPreferredName
		successText = "✅ Имя для обращения успешно изменено."

	default:
		return nil, fmt.Errorf(
			"unknown settings step: %q",
			session.CurrentStep,
		)
	}

	// СОХРАНЯЕМ ИЗМЕНЕНИЯ

	if err := s.users.UpdateUserData(
		ctx,
		telegramID,
		field,
		text,
	); err != nil {
		return nil, err
	}

	// Возвращаемся к выбору настройки.
	session.CurrentStep = settingsStepChooseField

	if err := s.sessions.Update(ctx, *session); err != nil {
		return nil, err
	}

	// Обновляем локальные данные пользователя, чтобы показать актуальную информацию.
	switch field {

	case model.UserFirstName:
		user.FirstName = text

	case model.UserLastName:
		user.LastName = text

	case model.UserPreferredName:
		user.PreferredName = text
	}

	return settingsResponse(user, successText), nil
}

func settingsInputPrompt(step string) string {
	switch step {

	case stepAwaitFirstName:
		return "✏️ <b>Изменение имени</b>\n\nВведи новое имя."

	case stepAwaitLastName:
		return "✏️ <b>Изменение фамилии</b>\n\nВведи новую фамилию."

	case stepAwaitPreferredName:
		return "💬 <b>Имя для обращения</b>\n\nКак теперь к тебе обращаться?"

	default:
		return "⚠️ Не удалось определить настройку."
	}
}

func settingsResponse(
	user *model.User,
	notice string,
) *model.BotResponse {

	username := "Не указан"

	if user.Username != nil && *user.Username != "" {
		username = "@" + *user.Username
	}

	text := fmt.Sprintf(
		"⚙️ <b>Настройки профиля</b>\n\n"+
			"👤 Имя: %s\n"+
			"📋 Фамилия: %s\n"+
			"💬 Имя для обращения: %s\n"+
			"🔗 Telegram username: %s\n\n"+
			"Выбери, что хочешь изменить 👇",
		html.EscapeString(user.FirstName),
		html.EscapeString(user.LastName),
		html.EscapeString(user.PreferredName),
		html.EscapeString(username),
	)

	if notice != "" {
		text = html.EscapeString(notice) + "\n\n" + text
	}

	return &model.BotResponse{
		Text: text,
		Menu: model.MenuSettings,
	}
}

func validateSettingsValue(value string) error {
	value = strings.TrimSpace(value)

	if value == "" {
		return errors.New(
			"Поле не может быть пустым. Попробуй ещё раз.",
		)
	}

	if len([]rune(value)) > 50 {
		return errors.New(
			"Максимальная длина — 50 символов.",
		)
	}

	if strings.ContainsAny(value, "\r\n") {
		return errors.New(
			"Введи значение одной строкой.",
		)
	}

	if strings.ContainsRune(value, '\x00') {
		return errors.New(
			"Введённое значение содержит недопустимый символ.",
		)
	}

	return nil
}

func (s *UserService) completeRegistration(
	ctx context.Context,
	session *model.UserSession,
) (*model.BotResponse, error) {

	user := model.User{
		TelegramID:    session.TelegramID,
		Username:      session.Username,
		FirstName:     session.FirstName,
		LastName:      session.LastName,
		PreferredName: session.PreferredName,
	}

	if err := s.users.UserCreate(ctx, user); err != nil {
		return nil, err
	}

	if err := s.sessions.Delete(
		ctx,
		session.TelegramID,
	); err != nil {
		return nil, err
	}

	return &model.BotResponse{
		Text: fmt.Sprintf(
			"<b>🎉 Регистрация завершена!</b>\n\n"+
				"Приятно познакомиться, %s! 😊\n\n"+
				"<b>TaskBot</b> готов помогать тебе с задачами. 📋\n\n"+
				"Выбери действие ниже 👇",
			html.EscapeString(user.PreferredName),
		),
		Menu: model.MenuMain,
	}, nil
}

func registrationPrompt(step string) string {
	switch step {
	case stepAwaitFirstName:
		return "<b>📝 Шаг 1 из 3 · Имя</b>\n\n" +
			"Напиши своё имя. 👋"

	case stepAwaitLastName:
		return "<b>📋 Шаг 2 из 3 · Фамилия</b>\n\n" +
			"Напиши свою фамилию."

	case stepAwaitPreferredName:
		return "<b>💬 Шаг 3 из 3 · Как к тебе обращаться?</b>\n\n" +
			"Напиши:\n\n" +
			"👤 <b>По имени</b>\n" +
			"📋 <b>По фамилии</b>\n\n" +
			"Или просто введи своё обращение ✨"

	default:
		return "⚠️ Не удалось определить этап регистрации. Напиши /start."
	}
}
