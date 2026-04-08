package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"phosboard/backend/internal/repository"
)

type TenantHandler struct {
	repo *repository.TenantRepository
}

func NewTenantHandler(repo *repository.TenantRepository) *TenantHandler {
	return &TenantHandler{repo: repo}
}

type CreateTenantRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateTenantResponse struct {
	Data  *repository.Tenant `json:"data,omitempty"`
	Error string             `json:"error,omitempty"`
}

func (h *TenantHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(CreateTenantResponse{Error: "method not allowed"})
		return
	}

	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(CreateTenantResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(CreateTenantResponse{Error: "name is required"})
		return
	}

	tenant, err := h.repo.CreateTenant(r.Context(), req.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(CreateTenantResponse{Error: "failed to create tenant"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(CreateTenantResponse{Data: tenant})
}

func (h *TenantHandler) CreateTenantGin(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	tenant, err := h.repo.CreateTenant(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tenant"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": tenant})
}
