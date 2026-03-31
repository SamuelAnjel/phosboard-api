package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"phosboard/backend/internal/models"
	"phosboard/backend/internal/repository"
)

type ConceptHandler struct {
	repo repository.ConceptRepository
}

func NewConceptHandler(repo repository.ConceptRepository) *ConceptHandler {
	return &ConceptHandler{repo: repo}
}

func (h *ConceptHandler) HandleConcepts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetConcepts(w, r)
	case http.MethodPost:
		h.CreateConcept(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Error: "method not allowed"})
	}
}

func (h *ConceptHandler) HandleConcept(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		h.DeleteConcept(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Error: "method not allowed"})
	}
}

func (h *ConceptHandler) HandleConceptsGin(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	switch c.Request.Method {
	case http.MethodGet:
		concepts, err := h.repo.GetActiveConceptsByTenant(c.Request.Context(), tenantID)
		if err != nil {
			slog.Error("failed to get concepts", "error", err, "tenant_id", tenantID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": concepts})

	case http.MethodPost:
		var req models.CreateConceptRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if req.ConceptTerm == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "concept_term is required"})
			return
		}

		concept, err := h.repo.CreateConcept(c.Request.Context(), tenantID, req.ConceptTerm)
		if err != nil {
			slog.Error("failed to create concept", "error", err, "tenant_id", tenantID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": concept})

	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

func (h *ConceptHandler) HandleConceptGin(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	conceptID := c.Param("concept_id")

	if tenantID == "" || conceptID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id and concept_id are required"})
		return
	}

	if err := h.repo.DeleteConcept(c.Request.Context(), conceptID); err != nil {
		slog.Error("failed to delete concept", "error", err, "concept_id", conceptID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ConceptHandler) GetConcepts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Error: "method not allowed"})
		return
	}

	tenantID := r.PathValue("tenant_id")
	if tenantID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Error: "tenant_id is required"})
		return
	}

	concepts, err := h.repo.GetActiveConceptsByTenant(r.Context(), tenantID)
	if err != nil {
		slog.Error("failed to get concepts", "error", err, "tenant_id", tenantID)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Error: "internal server error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Data: concepts})
}

func (h *ConceptHandler) CreateConcept(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Error: "method not allowed"})
		return
	}

	tenantID := r.PathValue("tenant_id")
	if tenantID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Error: "tenant_id is required"})
		return
	}

	var req models.CreateConceptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Error: "invalid request body"})
		return
	}

	if req.ConceptTerm == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Error: "concept_term is required"})
		return
	}

	concept, err := h.repo.CreateConcept(r.Context(), tenantID, req.ConceptTerm)
	if err != nil {
		slog.Error("failed to create concept", "error", err, "tenant_id", tenantID)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Error: "internal server error"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Data: concept})
}

func (h *ConceptHandler) DeleteConcept(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Error: "method not allowed"})
		return
	}

	tenantID := r.PathValue("tenant_id")
	conceptID := r.PathValue("concept_id")

	if tenantID == "" || conceptID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Error: "tenant_id and concept_id are required"})
		return
	}

	if err := h.repo.DeleteConcept(r.Context(), conceptID); err != nil {
		slog.Error("failed to delete concept", "error", err, "concept_id", conceptID)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Error: "internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
