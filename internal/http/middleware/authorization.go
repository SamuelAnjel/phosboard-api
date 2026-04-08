package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"phosboard/backend/internal/auth"
	"phosboard/backend/internal/repository"
)

var userRepo *repository.UserRepository

func SetUserRepository(repo *repository.UserRepository) {
	userRepo = repo
}

// AuthWithAuthorization combines authentication and authorization
func AuthWithAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Authentication
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		claims, err := auth.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Store claims in context
		c.Set("user", claims)

		// 2. Authorization (skip for super-admin with empty tenant_id)
		// Super-admin has global access, no need to check permissions
		if claims.TenantID == "" && claims.RoleID == "super-admin" {
			slog.Debug("super-admin access granted", "user_id", claims.UserID, "endpoint", c.Request.URL.Path)
			c.Next()
			return
		}

		// 3. Check permissions if userRepo is available
		if userRepo == nil {
			slog.Error("UserRepository not initialized for authorization")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "authorization system not ready"})
			return
		}

		// Get endpoint and method
		endpoint := c.Request.URL.Path
		method := c.Request.Method

		// For tenant-specific endpoints, we might need to validate tenant_id matches
		// For now, we just check if user has permission for this endpoint+method
		hasPermission, err := userRepo.HasPermission(c.Request.Context(), claims.UserID, endpoint, method)
		if err != nil {
			slog.Error("failed to check permissions", "user_id", claims.UserID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
			return
		}

		if !hasPermission {
			slog.Warn("access denied",
				"user_id", claims.UserID,
				"endpoint", endpoint,
				"method", method,
				"role", claims.RoleID,
				"tenant", claims.TenantID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		slog.Debug("access granted",
			"user_id", claims.UserID,
			"endpoint", endpoint,
			"method", method,
			"role", claims.RoleID)
		c.Next()
	}
}

// TenantContext middleware validates that user has access to the requested tenant
func TenantContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by Auth middleware)
		userValue, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		claims, ok := userValue.(*auth.UserClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid user claims"})
			return
		}

		// Super-admin can access any tenant
		if claims.TenantID == "" && claims.RoleID == "super-admin" {
			c.Next()
			return
		}

		// For tenant-specific users, check if they're accessing their own tenant
		// Extract tenant_id from URL path if present
		requestedTenantID := c.Param("tenant_id")
		if requestedTenantID == "" {
			// No tenant_id in path, use the one from token
			c.Next()
			return
		}

		// Check if user has access to this tenant
		if claims.TenantID != requestedTenantID {
			slog.Warn("tenant access denied",
				"user_id", claims.UserID,
				"user_tenant", claims.TenantID,
				"requested_tenant", requestedTenantID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access to this tenant is not allowed"})
			return
		}

		c.Next()
	}
}
