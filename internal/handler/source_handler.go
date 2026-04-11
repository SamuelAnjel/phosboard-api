package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"phosboard/backend/internal/models"
	"phosboard/backend/internal/repository"
)

type SourceHandler struct {
	repo repository.SourceRepository
}

func NewSourceHandler(repo repository.SourceRepository) *SourceHandler {
	return &SourceHandler{repo: repo}
}

func (h *SourceHandler) HandleSourcesGin(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	switch c.Request.Method {
	case http.MethodGet:
		sources, err := h.repo.GetSources(c.Request.Context())
		if err != nil {
			slog.Error("failed to get sources", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": sources})

	case http.MethodPost:
		var req models.CreateSourceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}

		if req.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
			return
		}

		sourceType := req.Type
		if sourceType == "" {
			sourceType = "rss"
		}

		// Validar tipo de source
		if sourceType != "rss" && sourceType != "web-crawl" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type must be 'rss' or 'web-crawl'"})
			return
		}

		// Validar configuración de crawling si es tipo web-crawl
		if sourceType == "web-crawl" && req.Crawl == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "crawl configuration is required for web-crawl type"})
			return
		}

		source, err := h.repo.CreateSource(c.Request.Context(), req.Name, sourceType, req.URL)
		if err != nil {
			slog.Error("failed to create source", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		// Construir configuración completa
		config := make(map[string]interface{})

		// Siempre incluir el tipo en la configuración para que el discovery worker lo encuentre
		config["type"] = sourceType

		if req.MaxLinks > 0 {
			config["max_links"] = req.MaxLinks
		}

		// Agregar configuración de crawling si es tipo web-crawl
		if sourceType == "web-crawl" && req.Crawl != nil {
			config["crawl"] = map[string]interface{}{
				"max_depth":      req.Crawl.MaxDepth,
				"max_pages":      req.Crawl.MaxPages,
				"same_domain":    req.Crawl.SameDomain,
				"include_paths":  req.Crawl.IncludePaths,
				"exclude_paths":  req.Crawl.ExcludePaths,
				"respect_robots": req.Crawl.RespectRobots,
				"crawl_delay_ms": req.Crawl.CrawlDelayMS,
			}
		}

		// Solo actualizar si hay configuración
		if len(config) > 0 {
			source, err = h.repo.UpdateSourceConfig(c.Request.Context(), source.ID, config)
			if err != nil {
				slog.Error("failed to update source config", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}

		c.JSON(http.StatusCreated, gin.H{"data": source})

	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

func (h *SourceHandler) HandleSourceGin(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	sourceID := c.Param("source_id")

	if tenantID == "" || sourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id and source_id are required"})
		return
	}

	switch c.Request.Method {
	case http.MethodGet:
		source, err := h.repo.GetSourceByID(c.Request.Context(), sourceID)
		if err != nil {
			slog.Error("failed to get source", "error", err, "source_id", sourceID)
			c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": source})

	case http.MethodPut:
		var req models.UpdateSourceConfigRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if req.Config == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "config is required"})
			return
		}

		source, err := h.repo.UpdateSourceConfig(c.Request.Context(), sourceID, req.Config)
		if err != nil {
			slog.Error("failed to update source config", "error", err, "source_id", sourceID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": source})

	case http.MethodDelete:
		if err := h.repo.DeleteSource(c.Request.Context(), sourceID); err != nil {
			slog.Error("failed to delete source", "error", err, "source_id", sourceID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.Status(http.StatusNoContent)

	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}
