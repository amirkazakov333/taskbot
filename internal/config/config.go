package config

import (
	"fmt"
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
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
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
