package handler

import (
	"net/http"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// FamilyHandler handles HTTP requests for family group endpoints.
type FamilyHandler struct {
	usecase usecase.FamilyUseCase
	logger  *zap.Logger
}

// NewFamilyHandler creates a new FamilyHandler.
func NewFamilyHandler(uc usecase.FamilyUseCase, logger *zap.Logger) *FamilyHandler {
	return &FamilyHandler{usecase: uc, logger: logger}
}

// Create handles POST /api/v1/family
func (h *FamilyHandler) Create(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required,min=1,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	group, err := h.usecase.CreateGroup(c.Request.Context(), userID, req.Name)
	if err != nil {
		h.logger.Error("FamilyHandler.Create", zap.Error(err))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "UNPROCESSABLE", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": group})
}

// Get handles GET /api/v1/family
func (h *FamilyHandler) Get(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	group, err := h.usecase.GetGroup(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("FamilyHandler.Get", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}
	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "no family group found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": group})
}

// GetInvite handles POST /api/v1/family/invite
func (h *FamilyHandler) GetInvite(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	code, err := h.usecase.GenerateInvite(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("FamilyHandler.GetInvite", zap.Error(err))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "UNPROCESSABLE", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"invite_code": code}})
}

// Join handles POST /api/v1/family/join
func (h *FamilyHandler) Join(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	var req struct {
		InviteCode string `json:"invite_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	group, err := h.usecase.JoinGroup(c.Request.Context(), userID, req.InviteCode)
	if err != nil {
		h.logger.Error("FamilyHandler.Join", zap.Error(err))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "UNPROCESSABLE", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": group})
}

// RemoveMember handles DELETE /api/v1/family/members/:id
func (h *FamilyHandler) RemoveMember(c *gin.Context) {
	ownerID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	memberID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_ID", "message": "invalid member ID"}})
		return
	}

	if err := h.usecase.RemoveMember(c.Request.Context(), memberID, ownerID); err != nil {
		h.logger.Error("FamilyHandler.RemoveMember", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// GetDashboard handles GET /api/v1/family/dashboard
func (h *FamilyHandler) GetDashboard(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	dashboard, err := h.usecase.GetDashboard(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("FamilyHandler.GetDashboard", zap.Error(err))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "UNPROCESSABLE", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dashboard})
}
