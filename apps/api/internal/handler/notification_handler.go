package handler

import (
	"net/http"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NotificationHandler handles HTTP requests for notification endpoints.
type NotificationHandler struct {
	usecase usecase.NotificationUseCase
	logger  *zap.Logger
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(uc usecase.NotificationUseCase, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{usecase: uc, logger: logger}
}

// List handles GET /api/v1/notifications
func (h *NotificationHandler) List(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	notifications, err := h.usecase.List(c.Request.Context(), userID, false)
	if err != nil {
		h.logger.Error("NotificationHandler.List", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	unread, _ := h.usecase.CountUnread(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{
		"data": notifications,
		"meta": gin.H{"unread_count": unread},
	})
}

// MarkAsRead handles PUT /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_ID", "message": "invalid notification ID"}})
		return
	}

	if err := h.usecase.MarkAsRead(c.Request.Context(), id, userID); err != nil {
		h.logger.Error("NotificationHandler.MarkAsRead", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"updated": true}})
}

// MarkAllAsRead handles PUT /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	if err := h.usecase.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		h.logger.Error("NotificationHandler.MarkAllAsRead", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"updated": true}})
}

// DeleteAll handles DELETE /api/v1/notifications
func (h *NotificationHandler) DeleteAll(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	if err := h.usecase.DeleteAll(c.Request.Context(), userID); err != nil {
		h.logger.Error("NotificationHandler.DeleteAll", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
