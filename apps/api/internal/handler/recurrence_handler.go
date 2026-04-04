package handler

import (
	"net/http"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RecurrenceHandler handles HTTP requests for recurrences.
type RecurrenceHandler struct {
	usecase usecase.RecurrenceUseCase
	logger  *zap.Logger
}

// NewRecurrenceHandler creates a new RecurrenceHandler.
func NewRecurrenceHandler(uc usecase.RecurrenceUseCase, logger *zap.Logger) *RecurrenceHandler {
	return &RecurrenceHandler{usecase: uc, logger: logger}
}

// List returns all recurrences for the authenticated user.
func (h *RecurrenceHandler) List(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	recs, err := h.usecase.List(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("recurrence list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": recs, "meta": gin.H{"total": len(recs)}})
}

// Create creates a new recurrence for the authenticated user.
func (h *RecurrenceHandler) Create(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	var req usecase.CreateRecurrenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	rec, err := h.usecase.Create(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("recurrence create", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": rec})
}

// Update updates an existing recurrence.
func (h *RecurrenceHandler) Update(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid id"}})
		return
	}

	var req usecase.UpdateRecurrenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	rec, err := h.usecase.Update(c.Request.Context(), id, userID, req)
	if err != nil {
		if err == usecase.ErrRecurrenceNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "recurrence not found"}})
			return
		}
		h.logger.Error("recurrence update", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rec})
}

// Delete removes a recurrence.
func (h *RecurrenceHandler) Delete(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid id"}})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id, userID); err != nil {
		if err == usecase.ErrRecurrenceNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "recurrence not found"}})
			return
		}
		h.logger.Error("recurrence delete", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
