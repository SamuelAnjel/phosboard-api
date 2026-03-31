package config

import (
	"context"
	"os"
	"testing"
)

func TestLoadWithDefaults_Success(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("GOOGLE_PROJECT_ID", "test-project")
	t.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")

	cfg, err := LoadWithDefaults(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.DatabaseURL != "postgres://user:pass@localhost:5432/db" {
		t.Errorf("expected database URL, got %s", cfg.DatabaseURL)
	}
	if cfg.ProjectID != "test-project" {
		t.Errorf("expected project ID, got %s", cfg.ProjectID)
	}
}

func TestLoadWithDefaults_MissingEnvVar(t *testing.T) {
	_ = os.Unsetenv("DATABASE_URL")
	_ = os.Unsetenv("GOOGLE_PROJECT_ID")
	_ = os.Unsetenv("PUBSUB_EMULATOR_HOST")

	_, err := LoadWithDefaults(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
