package handler

import (
	"net/http"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DashboardHandler handles HTTP requests for dashboard data.
type DashboardHandler struct {
	usecase usecase.DashboardUseCase
	logger  *zap.Logger
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(uc usecase.DashboardUseCase, logger *zap.Logger) *DashboardHandler {
	return &DashboardHandler{usecase: uc, logger: logger}
}

// GetOverview returns the dashboard overview for a given month and year.
//
//	@Summary		Visão geral do mês
//	@Description	Retorna saldo líquido, receitas, despesas e top categorias
//	@Tags			Dashboard
//	@Produce		json
//	@Security		BearerAuth
//	@Param			month	query	int	false	"Mês (1-12, default: atual)"
//	@Param			year	query	int	false	"Ano (default: atual)"
//	@Success		200	{object}	DashboardOverviewResponse
//	@Router			/dashboard/overview [get]
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	month, year := parseMonthYear(c)

	overview, err := h.usecase.GetOverview(c.Request.Context(), userID, month, year)
	if err != nil {
		h.logger.Error("dashboard overview", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": overview,
		"meta": gin.H{"month": month, "year": year},
	})
}

// GetCashflow returns the last 12 months cashflow data.
//
//	@Summary		Cashflow dos últimos 12 meses
//	@Tags			Dashboard
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	CashflowResponse
//	@Router			/dashboard/cashflow [get]
func (h *DashboardHandler) GetCashflow(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	cashflow, err := h.usecase.GetCashflow(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("dashboard cashflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": cashflow,
		"meta": gin.H{"months": 12},
	})
}
