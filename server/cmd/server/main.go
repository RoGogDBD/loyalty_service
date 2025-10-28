package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RoGogDBD/loyalty_service/server/internal/app"
	"github.com/RoGogDBD/loyalty_service/server/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	cfg := config.NewConfig()
	cfg.ParseFlags()
	logger := config.NewLogger(cfg.Logger().Level, cfg.Env)

	if cfg.Database().Password == "" && cfg.Database().URL == "" {
		logger.Fatal("DB credentials not set: задайте DB_PASSWORD или DATABASE_URI")
	}

	config.LogConfig(cfg, logger)

	application := app.New(cfg, logger)

	if err := application.Initialize(); err != nil {
		logger.Fatalf("Failed to initialize application: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := application.Run(); err != nil {
			logger.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	logger.Info("Received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		logger.Errorf("Shutdown error: %v", err)
	}

	logger.Info("Server stopped gracefully")
}
