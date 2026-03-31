package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"phosboard/backend/internal/auth"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, tenantID, roleID, err := validateCredentials(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(c.Request.Context(), userID, tenantID, roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

func validateCredentials(email, password string) (userID, tenantID, roleID string, err error) {
	if email == "admin@phosboard.cl" && password == "password123" {
		return "550e8400-e29b-41d4-a716-446655440000", "85c5f582-86b1-4217-bd4a-e1b1d0aac195", "admin", nil
	}

	if email == "user@phosboard.cl" && password == "password123" {
		return "550e8400-e29b-41d4-a716-446655440001", "85c5f582-86b1-4217-bd4a-e1b1d0aac195", "user", nil
	}

	return "", "", "", auth.ErrInvalidCredentials
}
