package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"phosboard/backend/internal/auth"
	"phosboard/backend/internal/repository"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

var userRepo *repository.UserRepository

func SetUserRepository(repo *repository.UserRepository) {
	userRepo = repo
}

func Login(c *gin.Context) {
	if userRepo == nil {
		slog.Error("UserRepository not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication system not ready"})
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate credentials against database
	user, err := userRepo.ValidateCredentials(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		slog.Warn("failed login attempt", "email", req.Email, "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Determine primary role and tenant for token
	// For super-admin (global role), tenantID is empty
	// For tenant roles, use the first tenant-specific role
	var tenantID, roleID string

	for _, role := range user.Roles {
		if role.TenantID == nil {
			// Global role (super-admin)
			tenantID = ""
			roleID = role.RoleName
			break
		} else if tenantID == "" {
			// First tenant-specific role
			tenantID = *role.TenantID
			roleID = role.RoleName
		}
	}

	if roleID == "" {
		slog.Error("user has no roles", "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user has no assigned roles"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(c.Request.Context(), user.ID, tenantID, roleID)
	if err != nil {
		slog.Error("failed to generate token", "user_id", user.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	slog.Info("successful login", "user_id", user.ID, "email", user.Email, "role", roleID)
	c.JSON(http.StatusOK, LoginResponse{Token: token})
}
