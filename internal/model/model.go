package model

import "time"

type MenuType string

const (
	MenuNone MenuType = "none"
	MenuMain MenuType = "main"
	MenuTask MenuType = "task"
)

type BotResponse struct {
	Text string
	Menu MenuType
}

type User struct {
	ID            int64
	TelegramID    int64
	Username      *string
	FirstName     string
	LastName      string
	PreferredName string
}

type RegistrationSession struct {
	TelegramID    int64
	Username      *string
	FirstName     string
	LastName      string
	PreferredName string
	CurrentStep   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
