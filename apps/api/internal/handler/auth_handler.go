package handler

import (
	"errors"
	"net/http"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	usecase usecase.AuthUseCase
	logger  *zap.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(uc usecase.AuthUseCase, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{usecase: uc, logger: logger}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req usecase.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	resp, err := h.usecase.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "EMAIL_EXISTS", "message": err.Error()}})
			return
		}
		h.logger.Error("register user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req usecase.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	resp, err := h.usecase.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "INVALID_CREDENTIALS", "message": err.Error()}})
			return
		}
		h.logger.Error("login user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// Refresh handles POST /auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	resp, err := h.usecase.RefreshToken(c.Request.Context(), body.RefreshToken)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidToken) || errors.Is(err, usecase.ErrTokenRevoked) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "INVALID_TOKEN", "message": err.Error()}})
			return
		}
		h.logger.Error("refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// Logout handles POST /auth/logout (protected)
func (h *AuthHandler) Logout(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	// Extract raw access token from Authorization header
	authHeader := c.GetHeader("Authorization")
	rawAccessToken := ""
	if len(authHeader) > 7 {
		rawAccessToken = authHeader[7:]
	}

	if err := h.usecase.Logout(c.Request.Context(), rawAccessToken, body.RefreshToken); err != nil {
		h.logger.Error("logout user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "logged out successfully"}})
}

// ForgotPassword handles POST /auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	if err := h.usecase.ForgotPassword(c.Request.Context(), body.Email); err != nil {
		h.logger.Error("forgot password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "if the email exists, a reset link has been sent"}})
}

// ResetPassword handles POST /auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var body struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	if err := h.usecase.ResetPassword(c.Request.Context(), body.Token, body.Password); err != nil {
		if errors.Is(err, usecase.ErrInvalidToken) {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_TOKEN", "message": err.Error()}})
			return
		}
		h.logger.Error("reset password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "password reset successfully"}})
}
