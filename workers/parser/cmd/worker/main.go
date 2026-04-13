package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"phosboard/workers/parser/internal/processor"
	"phosboard/workers/parser/internal/repository"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	gcsBucket := getEnv("GCS_BUCKET", "phosboard-documents")
	databaseURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/phosboard")
	tenantID := getEnv("TENANT_ID", "")

	if tenantID == "" {
		logger.Error("TENANT_ID environment variable is required")
		os.Exit(1)
	}

	logger.Info("starting parser worker", "bucket", gcsBucket, "tenant_id", tenantID)

	repo, err := repository.NewDocumentRepository(ctx, databaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	proc, err := processor.NewDocumentProcessor(
		ctx,
		processor.Config{
			GCSBucket: gcsBucket,
			TenantID:  tenantID,
		},
		repo,
	)
	if err != nil {
		logger.Error("failed to create processor", "error", err)
		os.Exit(1)
	}

	objectName := "phos-raw-data/test_1234567890.parquet"
	count, err := proc.ProcessFile(ctx, objectName)
	if err != nil {
		logger.Error("failed to process file", "error", err)
		os.Exit(1)
	}

	logger.Info("parser worker completed", "processed", count)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
