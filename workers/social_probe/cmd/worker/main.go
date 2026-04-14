package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"phosboard/workers/social_probe/internal/handler"
	"phosboard/workers/social_probe/internal/publisher"
	"phosboard/workers/social_probe/internal/scraper"
	"phosboard/workers/social_probe/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	godotenv.Load(".env")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bucketName := getEnvOrDefault("GCS_BUCKET_NAME", "phosboard-documents")
	logger.Info("initializing GCS storage", "bucket", bucketName)

	gcsStorage, err := storage.NewGCSStorage(ctx, bucketName)
	if err != nil {
		logger.Error("failed to create GCS storage", "error", err)
		os.Exit(1)
	}

	// Use Apify scraper if configured, otherwise fall back to mock
	var scraperClient scraper.SocialScraper
	apifyScraper, err := scraper.NewApifyScraper()
	if err != nil {
		logger.Warn("Apify scraper not configured, using mock scraper", "error", err)
		scraperClient = scraper.NewMockScraper()
	} else {
		logger.Info("Using Apify scraper for social media monitoring")
		scraperClient = apifyScraper
	}

	projectID := getEnvOrDefault("GOOGLE_PROJECT_ID", "phosboard")
	pubsubEndpoint := os.Getenv("PUBSUB_EMULATOR_HOST")

	logger.Info("initializing publisher", "project", projectID, "pubsub_endpoint", pubsubEndpoint)

	pub, err := publisher.NewPublisher(ctx, projectID, pubsubEndpoint)
	if err != nil {
		logger.Error("failed to create publisher", "error", err)
		os.Exit(1)
	}
	defer pub.Close()

	logger.Info("publisher initialized")

	h := handler.NewHandler(scraperClient, gcsStorage, pub)

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.HandlePubSubPush)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			logger.Error("failed to write health response", "error", err)
		}
	})

	port := getEnvOrDefault("PORT", "8080")
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("starting HTTP server", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	<-sigChan
	logger.Info("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	logger.Info("social-probe worker stopped")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func maskDBURL(dbURL string) string {
	if dbURL == "" {
		return ""
	}

	parsed, err := url.Parse(dbURL)
	if err != nil {
		return "[invalid URL]"
	}

	if parsed.User != nil {
		parsed.User = url.UserPassword("***", "***")
	}

	return parsed.String()
}
