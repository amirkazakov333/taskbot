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

func (s UserService) Start(ctx context.Context, update *models.Update) (string, error) {
	telegramID := update.Message.From.ID

	user, err := s.users.GetByTelegramID(ctx, telegramID)

	if err == nil {
		return fmt.Sprintf(
			"<b>👋 С возвращением, %s!</b>\n\n"+
				"📋 <b>TaskBot</b> — твой помощник по задачам.\n\n"+
				"Планируй дела, контролируй прогресс "+
				"и не забывай о важных задачах.\n\n"+
				"Выбери действие ниже 👇",
			html.EscapeString(user.PreferredName),
		), nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	session, err := s.registrationSessions.GetByTelegramID(
		ctx,
		telegramID,
	)

	if err == nil {
		switch session.CurrentStep {
		case "await_first_name":
			return "<b>📝 Давай познакомимся!</b>\n\n" +
				"Привет! Чтобы начать, напиши своё <b>имя</b>. 👋\n\n" +
				"<code>Шаг 1 из 3</code>", nil
		case "await_last_name":
			return "<b>✅ Имя сохранено!</b>\n\n" +
				"<b>📋 Шаг 2 из 3 · Фамилия</b>\n\n" +
				"Теперь напиши свою <b>фамилию</b>.\n\n", nil
		case "await_preferred_name":
			return "<b>✅ Основные данные почти готовы!</b>\n\n" +
				"<b>💬 Шаг 3 из 3 · Как к тебе обращаться?</b>\n\n" +
				"Выбери, как мне называть тебя. 😊\n\n" +
				"👤 По имени\n" +
				"📋 По фамилии\n" +
				"✨ Другое имя", nil
		case "await_custom_preferred_name":
			return "<b>✨ Придумай своё обращение</b>\n\n" +
				"Напиши, как тебе будет приятнее, чтобы я к тебе обращался.\n\n", nil
		default:
			return "", fmt.Errorf("unknown registration step: %q", session.CurrentStep)
		}
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	var username *string

	if name := update.Message.From.Username; name != "" {
		username = &name
	}

	newSession := model.RegistrationSession{
		TelegramID:  telegramID,
		Username:    username,
		CurrentStep: "await_first_name",
	}

	if err := s.registrationSessions.RegisterCreate(ctx, newSession); err != nil {
		return "", err
	}

	return "<b>👋 Привет!</b>\n\n" +
		"Добро пожаловать в <b>TaskBot</b> — твоего личного помощника по задачам! 📋\n\n" +
		"Я помогу тебе планировать дела, следить за их выполнением " +
		"и вовремя получать напоминания. ⏰\n\n" +
		"📝 <b>Давай познакомимся!</b>\n\n" +
		"Для начала регистрации напиши своё имя.", nil
}

func (s *UserService) HandleMessage(ctx context.Context, telegramID int64, text string) (string, error) {
	_, err := s.users.GetByTelegramID(ctx, telegramID)

	if err == nil {
		return "", nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	session, err := s.registrationSessions.GetByTelegramID(ctx, telegramID)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	text = strings.TrimSpace(text)

	if text == "" || strings.HasPrefix(text, "/") {
		return registrationPrompt(session.CurrentStep), nil
	}

	switch session.CurrentStep {
	case "await_first_name":
		session.FirstName = text
		session.CurrentStep = "await_last_name"

		if err := s.registrationSessions.Update(ctx, *session); err != nil {
			return "", err
		}

		return "<b>✅ Имя сохранено!</b>\n\n" +
			"<b>📋 Шаг 2 из 3 · Фамилия</b>\n\n" +
			"Теперь напиши свою <b>фамилию</b>.\n\n", nil

	case "await_last_name":
		session.LastName = text
		session.CurrentStep = "await_preferred_name"

		if err := s.registrationSessions.Update(ctx, *session); err != nil {
			return "", err
		}

		return registrationPrompt(session.CurrentStep), nil

	case "await_preferred_name":
		switch strings.ToLower(text) {
		case "по имени":
			session.PreferredName = session.FirstName
			return s.completeRegistration(ctx, session)
		case "по фамилии":
			session.PreferredName = session.LastName
			return s.completeRegistration(ctx, session)
		case "другое имя":
			session.CurrentStep = "await_custom_preferred_name"

			if err := s.registrationSessions.Update(ctx, *session); err != nil {
				return "", err
			}

			return registrationPrompt(session.CurrentStep), nil

		default:
			return "Пожалуйста, выбери один из вариантов:\n\n" +
				"• По имени\n" +
				"• По фамилии\n" +
				"• Другое имя", nil
		}

	case "await_custom_preferred_name":
		session.PreferredName = text
		return s.completeRegistration(ctx, session)

	default:
		return "", fmt.Errorf(
			"unknown registration step: %q",
			session.CurrentStep,
		)
	}
}

func (s *UserService) completeRegistration(ctx context.Context, session *model.RegistrationSession) (string, error) {
	user := model.User{
		TelegramID:    session.TelegramID,
		Username:      session.Username,
		FirstName:     session.FirstName,
		LastName:      session.LastName,
		PreferredName: session.PreferredName,
	}

	if err := s.users.UserCreate(ctx, user); err != nil {
		return "", err
	}

	if err := s.registrationSessions.Delete(ctx, session.TelegramID); err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"<b>🎉 Регистрация завершена!</b>\n\n"+
			"Приятно познакомиться, %s! 😊\n\n"+
			"<b>TaskBot</b> готов помогать тебе с задачами. 📋",
		html.EscapeString(user.PreferredName),
	), nil
}

func registrationPrompt(step string) string {
	switch step {
	case "await_first_name":
		return "<b>📝 Шаг 1 из 3 · Имя</b>\n\n" +
			"Напиши своё имя. 👋\n\n" +
			"<i>Например: Амир</i>"

	case "await_last_name":
		return "<b>📋 Шаг 2 из 3 · Фамилия</b>\n\n" +
			"Напиши свою фамилию.\n\n" +
			"<i>Например: Каримов</i>"

	case "await_preferred_name":
		return "<b>💬 Шаг 3 из 3 · Как к тебе обращаться?</b>\n\n" +
			"Выбери один вариант и отправь его сообщением:\n\n" +
			"👤 По имени\n" +
			"📋 По фамилии\n" +
			"✨ Другое имя"

	case "await_custom_preferred_name":
		return "<b>✨ Твоё имя для TaskBot</b>\n\n" +
			"Напиши, как тебе будет приятнее, чтобы я к тебе обращался.\n\n" +
			"<i>Например: Амирчик</i>"

	default:
		return "⚠️ Не удалось определить этап регистрации. Напиши /start."
	}
}
