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
//
//	@Summary		Listar recorrências
//	@Tags			Recurrences
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	RecurrenceListResponse
//	@Router			/recurrences [get]
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
//
//	@Summary		Criar recorrência
//	@Tags			Recurrences
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body	usecase.CreateRecurrenceRequest	true	"Dados da recorrência"
//	@Success		201	{object}	RecurrenceResponse
//	@Failure		400	{object}	ErrorResponse
//	@Router			/recurrences [post]
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
//
//	@Summary		Atualizar recorrência
//	@Tags			Recurrences
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string							true	"UUID da recorrência"
//	@Param			body	body	usecase.UpdateRecurrenceRequest	true	"Campos a atualizar"
//	@Success		200	{object}	RecurrenceResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/recurrences/{id} [put]
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
//
//	@Summary		Remover recorrência
//	@Tags			Recurrences
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID da recorrência"
//	@Success		204	"No Content"
//	@Failure		404	{object}	ErrorResponse
//	@Router			/recurrences/{id} [delete]
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
