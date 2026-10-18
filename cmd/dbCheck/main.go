package main

import (
	"context"
	"log"
	"time"

	"telegram-bot/internal/config"
	"telegram-bot/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	pool, err := repository.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("PostgreSQL connection failed: %v", err)
	}
	defer pool.Close()

	log.Println("PostgreSQL connection successful")
}
