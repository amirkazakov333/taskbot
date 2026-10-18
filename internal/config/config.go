package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	DBHost           string
	DBPort           int
	DBName           string
	DBUser           string
	DBPassword       string
}

func Load() (*Config, error) {
	port, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	return &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		DBHost:           os.Getenv("DB_HOST"),
		DBPort:           port,
		DBName:           os.Getenv("POSTGRES_DB"),
		DBUser:           os.Getenv("POSTGRES_USER"),
		DBPassword:       os.Getenv("POSTGRES_PASSWORD"),
	}, nil
}
