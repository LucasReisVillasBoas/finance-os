package handler

import (
	"net/http"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GoalHandler handles HTTP requests for goals.
type GoalHandler struct {
	usecase usecase.GoalUseCase
	logger  *zap.Logger
}

// NewGoalHandler creates a new GoalHandler.
func NewGoalHandler(uc usecase.GoalUseCase, logger *zap.Logger) *GoalHandler {
	return &GoalHandler{usecase: uc, logger: logger}
}

// List returns all goals for the authenticated user.
func (h *GoalHandler) List(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	goals, err := h.usecase.List(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("goal list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": goals, "meta": gin.H{"total": len(goals)}})
}

// Create creates a new goal.
func (h *GoalHandler) Create(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	var req usecase.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	goal, err := h.usecase.Create(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("goal create", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": goal})
}

// Update updates an existing goal.
func (h *GoalHandler) Update(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid id"}})
		return
	}

	var req usecase.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	goal, err := h.usecase.Update(c.Request.Context(), id, userID, req)
	if err != nil {
		if err == usecase.ErrGoalNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "goal not found"}})
			return
		}
		h.logger.Error("goal update", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": goal})
}

// Delete removes a goal.
func (h *GoalHandler) Delete(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid id"}})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id, userID); err != nil {
		if err == usecase.ErrGoalNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "goal not found"}})
			return
		}
		h.logger.Error("goal delete", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Contribute adds a contribution to a goal.
func (h *GoalHandler) Contribute(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid goal id"}})
		return
	}

	var req usecase.ContributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	contribution, err := h.usecase.Contribute(c.Request.Context(), goalID, userID, req)
	if err != nil {
		if err == usecase.ErrGoalNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "goal not found"}})
			return
		}
		h.logger.Error("goal contribute", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": contribution})
}

// GetProjections returns projections for all goals.
func (h *GoalHandler) GetProjections(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	projections, err := h.usecase.GetProjections(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("goal projections", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": projections, "meta": gin.H{"total": len(projections)}})
}
