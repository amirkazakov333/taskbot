package model

import "time"

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
