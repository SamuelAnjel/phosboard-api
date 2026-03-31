package config

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL               string
	ProjectID                 string
	PubSubEmulatorHost        string
	DispatcherIntervalSeconds int
}

func Load(ctx context.Context) error {
	paths := []string{".env", "../.env", "../../.env"}

	for _, path := range paths {
		if err := godotenv.Load(path); err == nil {
			return nil
		}
	}

	if os.Getenv("DATABASE_URL") == "" {
		return fmt.Errorf("DATABASE_URL environment variable not set and no .env file found: %w", context.DeadlineExceeded)
	}

	return nil
}

func LoadWithDefaults(ctx context.Context) (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set: %w", context.DeadlineExceeded)
	}

	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	pubSubHost := os.Getenv("PUBSUB_EMULATOR_HOST")

	intervalStr := os.Getenv("DISPATCHER_INTERVAL_SECONDS")
	interval := 900 // default 15 minutes
	if intervalStr != "" {
		if parsed, err := strconv.Atoi(intervalStr); err == nil {
			interval = parsed
		}
	}

	return &Config{
		DatabaseURL:               dbURL,
		ProjectID:                 projectID,
		PubSubEmulatorHost:        pubSubHost,
		DispatcherIntervalSeconds: interval,
	}, nil
}
