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
	stepAwaitFirstName     = "await_first_name"
	stepAwaitLastName      = "await_last_name"
	stepAwaitPreferredName = "await_preferred_name"
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
