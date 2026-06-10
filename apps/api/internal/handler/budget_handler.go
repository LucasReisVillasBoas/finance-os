package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BudgetHandler handles HTTP requests for budgets.
type BudgetHandler struct {
	usecase usecase.BudgetUseCase
	logger  *zap.Logger
}

// NewBudgetHandler creates a new BudgetHandler.
func NewBudgetHandler(uc usecase.BudgetUseCase, logger *zap.Logger) *BudgetHandler {
	return &BudgetHandler{usecase: uc, logger: logger}
}

// parseMonthYear extracts month and year query params, defaulting to current month/year.
func parseMonthYear(c *gin.Context) (int, int) {
	now := time.Now()
	month := now.Month()
	year := now.Year()

	if m := c.Query("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v >= 1 && v <= 12 {
			month = time.Month(v)
		}
	}
	if y := c.Query("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil && v > 0 {
			year = v
		}
	}
	return int(month), year
}

// List returns budgets for the given month/year.
//
//	@Summary		Listar orçamentos
//	@Tags			Budgets
//	@Produce		json
//	@Security		BearerAuth
//	@Param			month	query	int	false	"Mês (1-12, default: atual)"
//	@Param			year	query	int	false	"Ano (default: atual)"
//	@Success		200	{object}	BudgetListResponse
//	@Router			/budgets [get]
func (h *BudgetHandler) List(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	month, year := parseMonthYear(c)

	budgets, err := h.usecase.List(c.Request.Context(), userID, month, year)
	if err != nil {
		h.logger.Error("budget list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": budgets, "meta": gin.H{"total": len(budgets), "month": month, "year": year}})
}

// Create creates a new budget.
//
//	@Summary		Criar orçamento
//	@Tags			Budgets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body	usecase.CreateBudgetRequest	true	"Dados do orçamento"
//	@Success		201	{object}	BudgetResponse
//	@Failure		400	{object}	ErrorResponse
//	@Router			/budgets [post]
func (h *BudgetHandler) Create(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	var req usecase.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	budget, err := h.usecase.Create(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("budget create", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": budget})
}

// GetProgress returns budget progress for the given month/year.
//
//	@Summary		Progresso dos orçamentos
//	@Description	Retorna planejado vs real por categoria
//	@Tags			Budgets
//	@Produce		json
//	@Security		BearerAuth
//	@Param			month	query	int	false	"Mês (1-12, default: atual)"
//	@Param			year	query	int	false	"Ano (default: atual)"
//	@Success		200	{object}	BudgetProgressResponse
//	@Router			/budgets/progress [get]
func (h *BudgetHandler) GetProgress(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	month, year := parseMonthYear(c)

	progress, err := h.usecase.GetProgress(c.Request.Context(), userID, month, year)
	if err != nil {
		h.logger.Error("budget progress", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": progress, "meta": gin.H{"month": month, "year": year}})
}

// Update updates an existing budget.
//
//	@Summary		Atualizar orçamento
//	@Tags			Budgets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string						true	"UUID do orçamento"
//	@Param			body	body	usecase.UpdateBudgetRequest	true	"Campos a atualizar"
//	@Success		200	{object}	BudgetResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/budgets/{id} [put]
func (h *BudgetHandler) Update(c *gin.Context) {
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

	var req usecase.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	budget, err := h.usecase.Update(c.Request.Context(), id, userID, req)
	if err != nil {
		if err == usecase.ErrBudgetNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "budget not found"}})
			return
		}
		h.logger.Error("budget update", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": budget})
}

// Delete removes a budget.
//
//	@Summary		Remover orçamento
//	@Tags			Budgets
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID do orçamento"
//	@Success		204	"No Content"
//	@Failure		404	{object}	ErrorResponse
//	@Router			/budgets/{id} [delete]
func (h *BudgetHandler) Delete(c *gin.Context) {
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
		if err == usecase.ErrBudgetNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "budget not found"}})
			return
		}
		h.logger.Error("budget delete", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
