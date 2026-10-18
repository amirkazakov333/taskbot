package model

type User struct {
	ID            int64
	TelegramID    int64
	Username      *string
	FirstName     string
	LastName      string
	PreferredName string
}
