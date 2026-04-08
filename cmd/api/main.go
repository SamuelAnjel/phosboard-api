package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"phosboard/backend/internal/config"
	"phosboard/backend/internal/db"
	"phosboard/backend/internal/dispatcher"
	"phosboard/backend/internal/handler"
	"phosboard/backend/internal/http/middleware"
	"phosboard/backend/internal/publisher"
	"phosboard/backend/internal/repository"
)

func main() {
	_ = godotenv.Load(".env")

	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := config.Load(ctx); err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	cfg, err := config.LoadWithDefaults(ctx)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	dispatch, err := dispatcher.NewDispatcher(ctx, database.Pool(), dispatcher.Config{
		ProjectID:       cfg.ProjectID,
		PubSubEndpoint:  cfg.PubSubEmulatorHost,
		IntervalSeconds: cfg.DispatcherIntervalSeconds,
	})
	if err != nil {
		logger.Error("failed to create dispatcher", "error", err)
	} else {
		defer dispatch.Close()
		go func() {
			if err := dispatch.Start(ctx); err != nil {
				slog.Error("dispatcher error", "error", err)
			}
		}()
	}

	pub, err := publisher.NewPublisher(ctx, cfg.ProjectID, cfg.PubSubEmulatorHost)
	if err != nil {
		logger.Error("failed to create publisher", "error", err)
		os.Exit(1)
	}
	defer pub.Close()

	docHandler := handler.NewDocumentHandler(database.Pool(), pub)
	conceptRepo := repository.NewConceptRepository(database.Pool())
	conceptHandler := handler.NewConceptHandler(conceptRepo)
	sourceRepo := repository.NewSourceRepository(database.Pool())
	sourceHandler := handler.NewSourceHandler(sourceRepo)
	tenantRepo := repository.NewTenantRepository(database.Pool())
	tenantHandler := handler.NewTenantHandler(tenantRepo)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/api/auth/login", handler.Login)

	protected := r.Group("/api/v1")
	protected.Use(middleware.Auth())
	{
		protected.GET("/documents", docHandler.ListDocuments)

		protected.POST("/documents/track", docHandler.TrackGin)
		protected.GET("/tenants/:tenant_id/concepts", conceptHandler.HandleConceptsGin)
		protected.POST("/tenants/:tenant_id/concepts", conceptHandler.HandleConceptsGin)
		protected.DELETE("/tenants/:tenant_id/concepts/:concept_id", conceptHandler.HandleConceptGin)
		protected.GET("/tenants/:tenant_id/sources", sourceHandler.HandleSourcesGin)
		protected.POST("/tenants/:tenant_id/sources", sourceHandler.HandleSourcesGin)
		protected.GET("/tenants/:tenant_id/sources/:source_id", sourceHandler.HandleSourceGin)
		protected.PUT("/tenants/:tenant_id/sources/:source_id", sourceHandler.HandleSourceGin)
		protected.DELETE("/tenants/:tenant_id/sources/:source_id", sourceHandler.HandleSourceGin)

		// Tenant management (authenticated)
		protected.POST("/tenants", tenantHandler.CreateTenantGin)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		logger.Info("server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}
	logger.Info("server exited")
}
