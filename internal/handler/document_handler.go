package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"phosboard/backend/internal/auth"
	"phosboard/backend/internal/publisher"
	"phosboard/backend/internal/repository"
)

type DocumentHandler struct {
	pool      *pgxpool.Pool
	publisher *publisher.Publisher
	repo      *repository.PostgresDocumentRepository
}

type TrackRequest struct {
	URL      string `json:"url"`
	SourceID string `json:"source_id,omitempty"`
	Priority int    `json:"priority,omitempty"`
	TenantID string `json:"tenant_id,omitempty"` // Required for super-admin, optional for tenant users
}

type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func NewDocumentHandler(pool *pgxpool.Pool, pub *publisher.Publisher) *DocumentHandler {
	return &DocumentHandler{
		pool:      pool,
		publisher: pub,
		repo:      repository.NewDocumentRepository(pool),
	}
}

func (h *DocumentHandler) Track(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(Response{Error: "method not allowed"})
		return
	}

	var req TrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(Response{Error: "invalid request body"})
		return
	}

	h.trackHandler(w, r, req)
}

func (h *DocumentHandler) TrackGin(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	claims, ok := user.(*auth.UserClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	var req TrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	// Determine tenantID:
	// 1. For super-admin (tenantID empty in token), use tenant_id from request
	// 2. For tenant users, use tenant_id from token (ignore request tenant_id for security)
	tenantID := claims.TenantID
	if tenantID == "" {
		// Super-admin: require tenant_id in request
		if req.TenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required for super-admin"})
			return
		}
		tenantID = req.TenantID
	} else {
		// Tenant user: ensure they're not trying to access another tenant
		if req.TenantID != "" && req.TenantID != tenantID {
			c.JSON(http.StatusForbidden, gin.H{"error": "cannot track documents for other tenants"})
			return
		}
	}

	parsedURL, err := url.Parse(req.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid URL format"})
		return
	}

	sourceType := req.SourceID
	if sourceType == "" {
		sourceType = "manual"
	}

	docID, taskID, err := h.repo.TrackDocument(c.Request.Context(), tenantID, req.URL, sourceType, req.Priority)
	if err != nil {
		slog.Error("failed to track document", "error", err, "url", req.URL, "tenant_id", tenantID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to track document"})
		return
	}

	if h.publisher != nil {
		if err := h.publisher.PublishURLScrape(c.Request.Context(), docID, req.URL); err != nil {
			slog.Error("failed to publish url scrape", "error", err)
			c.JSON(http.StatusInternalServerError, Response{Error: "failed to publish task"})
			return
		}
	} else {
		slog.Warn("publisher not available, skipping Pub/Sub publish")
	}

	c.JSON(http.StatusAccepted, gin.H{
		"data": map[string]string{
			"task_id": taskID,
			"url":     req.URL,
			"status":  "queued",
		},
	})
}

func (h *DocumentHandler) trackHandler(w http.ResponseWriter, r *http.Request, req TrackRequest) {
	if req.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(Response{Error: "url is required"})
		return
	}

	parsedURL, err := url.Parse(req.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(Response{Error: "invalid URL format"})
		return
	}

	sourceID := req.SourceID
	if sourceID == "" {
		sourceID = "manual"
	}

	// For this legacy endpoint, we need to create a document and get its ID
	// Since we don't have tenant context here, we'll use a default tenant
	defaultTenantID := "85c5f582-86b1-4217-bd4a-e1b1d0aac195" // Default tenant

	docID, taskID, err := h.repo.TrackDocument(r.Context(), defaultTenantID, req.URL, sourceID, req.Priority)
	if err != nil {
		slog.Error("failed to track document", "error", err, "url", req.URL, "tenant_id", defaultTenantID)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(Response{Error: "failed to track document"})
		return
	}

	if h.publisher != nil {
		if err := h.publisher.PublishURLScrape(r.Context(), docID, req.URL); err != nil {
			slog.Error("failed to publish url scrape", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(Response{Error: "failed to publish task"})
			return
		}
	} else {
		slog.Warn("publisher not available, skipping Pub/Sub publish")
	}

	_ = taskID          // taskID is not used in response but we need to handle it
	documentID := docID // Update the variable for response

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(Response{
		Data: map[string]string{
			"document_id": documentID,
			"url":         req.URL,
			"status":      "queued",
		},
	})
}

func (h *DocumentHandler) insertDiscoveryTask(ctx context.Context, url, sourceType string, priority int) (string, error) {
	var id string
	err := h.pool.QueryRow(ctx, `
		INSERT INTO discovery_tasks (url, source_type, status, priority)
		VALUES ($1, $2, 'pending', $3)
		ON CONFLICT (url) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, url, sourceType, priority).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("insert discovery task: %w", err)
	}

	return id, nil
}

func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	claims, ok := user.(*auth.UserClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	tenantID := claims.TenantID
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id not found in token"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	docs, total, err := h.repo.GetDocumentsByTenant(c.Request.Context(), tenantID, limit, offset)
	if err != nil {
		slog.Error("failed to get documents", "error", err, "tenant_id", tenantID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": docs,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}
