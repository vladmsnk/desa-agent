package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"desa-agent/internal/app"
	"desa-agent/internal/config"
)

func main() {
	_ = godotenv.Load() // Load .env file if exists

	if err := run(); err != nil {
		log.Printf("application error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return err
	}

	application, err := app.New(cfg)
	if err != nil {
		return err
	}

	return application.Run(context.Background())
}
