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
	users                repository.UserRepository
	registrationSessions repository.RegistrationSessionRepository
}

func NewUserService(users repository.UserRepository, registrationSessions repository.RegistrationSessionRepository) *UserService {
	return &UserService{
		users:                users,
		registrationSessions: registrationSessions,
	}
}

func (s *UserService) Start(ctx context.Context, update *models.Update) (*model.BotResponse, error) {
	telegramID := update.Message.From.ID

	// 1. Проверяем, зарегистрирован ли пользователь
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

	// 2. Проверяем, есть ли незавершённая регистрация
	session, err := s.registrationSessions.GetByTelegramID(
		ctx,
		telegramID,
	)

	if err == nil {
		switch session.CurrentStep {

		case stepAwaitFirstName,
			stepAwaitLastName,
			stepAwaitPreferredName:

			return &model.BotResponse{
				Text: registrationPrompt(session.CurrentStep),
				Menu: model.MenuNone,
			}, nil

		default:
			return nil, fmt.Errorf(
				"unknown registration step: %q",
				session.CurrentStep,
			)
		}
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// 3. Новый пользователь — создаём registration session
	var username *string

	if name := update.Message.From.Username; name != "" {
		username = &name
	}

	newSession := model.RegistrationSession{
		TelegramID:  telegramID,
		Username:    username,
		CurrentStep: stepAwaitFirstName,
	}

	if err := s.registrationSessions.RegisterCreate(
		ctx,
		newSession,
	); err != nil {
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

func (s *UserService) HandleMessage(ctx context.Context, telegramID int64, text string) (*model.BotResponse, error) {
	// 1. Проверяем, зарегистрирован ли пользователь
	_, err := s.users.GetByTelegramID(ctx, telegramID)

	if err == nil {
		// Пока обработку кнопок главного меню ещё не сделали.
		return &model.BotResponse{
			Text: "Выбери нужное действие 👇",
			Menu: model.MenuMain,
		}, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// 2. Получаем текущую registration session
	session, err := s.registrationSessions.GetByTelegramID(
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

	text = strings.TrimSpace(text)

	// Пустое сообщение или команда во время регистрации
	if text == "" || strings.HasPrefix(text, "/") {
		return &model.BotResponse{
			Text: registrationPrompt(session.CurrentStep),
			Menu: model.MenuNone,
		}, nil
	}

	// 3. Обрабатываем текущий этап регистрации
	switch session.CurrentStep {

	case stepAwaitFirstName:

		session.FirstName = text
		session.CurrentStep = stepAwaitLastName

		if err := s.registrationSessions.Update(
			ctx,
			*session,
		); err != nil {
			return nil, err
		}

		return &model.BotResponse{
			Text: registrationPrompt(session.CurrentStep),
			Menu: model.MenuNone,
		}, nil

	case stepAwaitLastName:

		session.LastName = text
		session.CurrentStep = stepAwaitPreferredName

		if err := s.registrationSessions.Update(
			ctx,
			*session,
		); err != nil {
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

func (s *UserService) completeRegistration(ctx context.Context, session *model.RegistrationSession) (*model.BotResponse, error) {
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

	if err := s.registrationSessions.Delete(
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
