package service

import (
	"context"
	"errors"
	"fmt"
	"html"
	"telegram-bot/internal/repository"

	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5"
)

type UserService struct {
	repository repository.UserRepository
}

func NewUserService(repository repository.UserRepository) *UserService {
	return &UserService{
		repository: repository,
	}
}

func (s UserService) Start(ctx context.Context, update *models.Update) (string, error) {
	user, err := s.repository.GetByTelegramID(ctx, update.Message.From.ID)

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

	if errors.Is(err, pgx.ErrNoRows) {
		return "<b>👋 Привет!</b>\n\n" +
			"Добро пожаловать в <b>TaskBot</b>!\n\n" +
			"Я помогу тебе планировать дела, " +
			"следить за задачами и вовремя получать напоминания.\n\n" +
			"Давай начнём с регистрации!", nil
	}

	return "", err
}
