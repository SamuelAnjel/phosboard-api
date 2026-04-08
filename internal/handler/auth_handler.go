package handler

import (
	"fmt"
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
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	slog.Info("Login attempt", "email", req.Email, "db_auth", "enabled", "debug", "rbac_v1.1.5", "password_length", len(req.Password))

	if userRepo == nil {
		slog.Error("userRepo is nil - SetUserRepository not called")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication system not initialized"})
		return
	}

	// Validate credentials against database
	slog.Debug("Calling ValidateCredentials", "email", req.Email)
	user, err := userRepo.ValidateCredentials(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		slog.Warn("Failed login attempt - DB validation failed",
			"email", req.Email,
			"error", err.Error(),
			"error_type", fmt.Sprintf("%T", err),
			"debug", "checking hash mismatch",
			"stack", "ValidateCredentials returned error")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	slog.Debug("ValidateCredentials succeeded", "user_id", user.ID, "roles_count", len(user.Roles))

	// Determine primary role and tenant for token
	// For super-admin (global role), tenantID is empty
	// For tenant roles, use the first tenant-specific role
	var tenantID, roleID string

	slog.Debug("Processing user roles", "total_roles", len(user.Roles))
	for i, role := range user.Roles {
		slog.Debug("Role info",
			"index", i,
			"role_name", role.RoleName,
			"tenant_id", role.TenantID,
			"is_global", role.TenantID == nil)
		if role.TenantID == nil {
			// Global role (super-admin)
			tenantID = ""
			roleID = role.RoleName
			slog.Debug("Found global role", "role_name", roleID)
			break
		} else if tenantID == "" {
			// First tenant-specific role
			tenantID = *role.TenantID
			roleID = role.RoleName
			slog.Debug("Found tenant role", "role_name", roleID, "tenant_id", tenantID)
		}
	}

	if roleID == "" {
		slog.Error("user has no roles", "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user has no assigned roles"})
		return
	}

	slog.Debug("Selected role for token", "role_id", roleID, "tenant_id", tenantID)

	// Generate JWT token
	slog.Debug("Generating JWT token")
	token, err := auth.GenerateToken(c.Request.Context(), user.ID, tenantID, roleID)
	if err != nil {
		slog.Error("failed to generate token", "user_id", user.ID, "error", err, "error_details", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	slog.Debug("Token generated successfully", "token_length", len(token))

	slog.Info("Successful login - RBAC system",
		"user_id", user.ID,
		"email", user.Email,
		"role", roleID,
		"tenant", tenantID,
		"rbac", "enabled",
		"user_roles_count", len(user.Roles))

	// Debug: log all roles
	for i, role := range user.Roles {
		slog.Debug("User role",
			"index", i,
			"role_name", role.RoleName,
			"role_id", role.RoleID,
			"tenant_id", role.TenantID)
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}
