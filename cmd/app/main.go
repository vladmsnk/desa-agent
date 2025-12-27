package main

import (
	"context"
	"log"
	"os"

	"desa-agent/internal/app"
	"desa-agent/internal/config"
)

func main() {
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
