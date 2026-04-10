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

	var dispatch *dispatcher.Dispatcher
	dispatch, err = dispatcher.NewDispatcher(ctx, database.Pool(), dispatcher.Config{
		ProjectID:       cfg.ProjectID,
		PubSubEndpoint:  cfg.PubSubEmulatorHost,
		IntervalSeconds: cfg.DispatcherIntervalSeconds,
	})
	if err != nil {
		logger.Warn("failed to create dispatcher, continuing without dispatcher", "error", err)
		dispatch = nil
	} else {
		defer dispatch.Close()
		go func() {
			if err := dispatch.Start(ctx); err != nil {
				slog.Error("dispatcher error", "error", err)
			}
		}()
	}

	var pub *publisher.Publisher
	pub, err = publisher.NewPublisher(ctx, cfg.ProjectID, cfg.PubSubEmulatorHost)
	if err != nil {
		logger.Warn("failed to create publisher, continuing without Pub/Sub", "error", err)
		pub = nil
	} else {
		defer pub.Close()
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(database.Pool())
	docRepo := repository.NewDocumentRepository(database.Pool())
	docHandler := handler.NewDocumentHandler(database.Pool(), pub)
	conceptRepo := repository.NewConceptRepository(database.Pool())
	conceptHandler := handler.NewConceptHandler(conceptRepo)
	sourceRepo := repository.NewSourceRepository(database.Pool())
	sourceHandler := handler.NewSourceHandler(sourceRepo)
	tenantRepo := repository.NewTenantRepository(database.Pool())
	tenantHandler := handler.NewTenantHandler(tenantRepo)

	// Setup authentication and authorization
	handler.SetUserRepository(userRepo)
	middleware.SetUserRepository(userRepo)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "version": "v1.1.9", "rbac": "enabled"})
	})

	r.GET("/debug/version", func(c *gin.Context) {
		// Test simple response
		c.JSON(http.StatusOK, gin.H{
			"version":   "v1.2.1",
			"rbac":      "enabled",
			"timestamp": time.Now().Format(time.RFC3339),
			"debug":     "auth_with_debug_logs",
		})
	})

	r.GET("/debug/simple", func(c *gin.Context) {
		// Even simpler response
		c.JSON(http.StatusOK, map[string]string{
			"test":   "working",
			"status": "ok",
		})
	})

	r.GET("/debug/plain", func(c *gin.Context) {
		// Plain text response
		c.String(http.StatusOK, "Plain text response - v1.2.1")
	})

	// Test endpoint without authentication
	r.POST("/test/track", func(c *gin.Context) {
		var req struct {
			URL string `json:"url"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if req.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
			return
		}

		// Use a test tenant ID
		tenantID := "test-tenant"
		sourceType := "manual"
		priority := 0

		docID, taskID, err := docRepo.TrackDocument(c.Request.Context(), tenantID, req.URL, sourceType, priority)
		if err != nil {
			slog.Error("failed to track document", "error", err, "url", req.URL)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to track document"})
			return
		}

		if pub != nil {
			if err := pub.PublishURLScrape(c.Request.Context(), docID, req.URL); err != nil {
				slog.Error("failed to publish url scrape", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish task"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"document_id": docID,
			"task_id":     taskID,
			"url":         req.URL,
			"message":     "tracking started",
		})
	})

	r.POST("/api/auth/login", handler.Login)

	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthWithAuthorization())
	{
		protected.GET("/documents", docHandler.ListDocuments)

		protected.POST("/documents/track", docHandler.TrackGin)

		// Tenant-specific routes with tenant context validation
		tenantRoutes := protected.Group("/tenants/:tenant_id")
		tenantRoutes.Use(middleware.TenantContext())
		{
			tenantRoutes.GET("/concepts", conceptHandler.HandleConceptsGin)
			tenantRoutes.POST("/concepts", conceptHandler.HandleConceptsGin)
			tenantRoutes.DELETE("/concepts/:concept_id", conceptHandler.HandleConceptGin)
			tenantRoutes.GET("/sources", sourceHandler.HandleSourcesGin)
			tenantRoutes.POST("/sources", sourceHandler.HandleSourcesGin)
			tenantRoutes.GET("/sources/:source_id", sourceHandler.HandleSourceGin)
			tenantRoutes.PUT("/sources/:source_id", sourceHandler.HandleSourceGin)
			tenantRoutes.DELETE("/sources/:source_id", sourceHandler.HandleSourceGin)
		}

		// Tenant management (authenticated, for super-admin)
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
