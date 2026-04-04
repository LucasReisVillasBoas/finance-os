package handler

import (
	"errors"
	"net/http"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// InvestmentHandler handles HTTP requests for investment endpoints.
type InvestmentHandler struct {
	usecase usecase.InvestmentUseCase
	logger  *zap.Logger
}

// NewInvestmentHandler creates a new InvestmentHandler.
func NewInvestmentHandler(uc usecase.InvestmentUseCase, logger *zap.Logger) *InvestmentHandler {
	return &InvestmentHandler{usecase: uc, logger: logger}
}

// ListPortfolios handles GET /api/v1/portfolios
func (h *InvestmentHandler) ListPortfolios(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	portfolios, err := h.usecase.GetPortfolios(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("list portfolios", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": portfolios})
}

// CreatePortfolio handles POST /api/v1/portfolios
func (h *InvestmentHandler) CreatePortfolio(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	var req usecase.CreatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	p, err := h.usecase.CreatePortfolio(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("create portfolio", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": p})
}

// UpdatePortfolio handles PUT /api/v1/portfolios/:id
func (h *InvestmentHandler) UpdatePortfolio(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	portfolioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid portfolio id"}})
		return
	}

	var req usecase.UpdatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	p, err := h.usecase.UpdatePortfolio(c.Request.Context(), portfolioID, userID, req)
	if err != nil {
		if errors.Is(err, usecase.ErrPortfolioNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("update portfolio", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": p})
}

// DeletePortfolio handles DELETE /api/v1/portfolios/:id
func (h *InvestmentHandler) DeletePortfolio(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	portfolioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid portfolio id"}})
		return
	}

	if err := h.usecase.DeletePortfolio(c.Request.Context(), portfolioID, userID); err != nil {
		if errors.Is(err, usecase.ErrPortfolioNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("delete portfolio", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "portfolio deleted successfully"}})
}

// ListHoldings handles GET /api/v1/portfolios/:id/holdings
func (h *InvestmentHandler) ListHoldings(c *gin.Context) {
	portfolioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid portfolio id"}})
		return
	}

	holdings, err := h.usecase.GetHoldings(c.Request.Context(), portfolioID)
	if err != nil {
		h.logger.Error("list holdings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": holdings})
}

// CreateHolding handles POST /api/v1/portfolios/:id/holdings
func (h *InvestmentHandler) CreateHolding(c *gin.Context) {
	portfolioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid portfolio id"}})
		return
	}

	var req usecase.CreateHoldingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	holding, err := h.usecase.CreateHolding(c.Request.Context(), portfolioID, req)
	if err != nil {
		h.logger.Error("create holding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": holding})
}

// UpdateHolding handles PUT /api/v1/holdings/:id
func (h *InvestmentHandler) UpdateHolding(c *gin.Context) {
	holdingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid holding id"}})
		return
	}

	var req usecase.UpdateHoldingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	holding, err := h.usecase.UpdateHolding(c.Request.Context(), holdingID, req)
	if err != nil {
		if errors.Is(err, usecase.ErrHoldingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("update holding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": holding})
}

// DeleteHolding handles DELETE /api/v1/holdings/:id
func (h *InvestmentHandler) DeleteHolding(c *gin.Context) {
	holdingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid holding id"}})
		return
	}

	if err := h.usecase.DeleteHolding(c.Request.Context(), holdingID); err != nil {
		if errors.Is(err, usecase.ErrHoldingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("delete holding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "holding deleted successfully"}})
}

// ListInvestmentTransactions handles GET /api/v1/holdings/:id/transactions
func (h *InvestmentHandler) ListInvestmentTransactions(c *gin.Context) {
	holdingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid holding id"}})
		return
	}

	txs, err := h.usecase.GetInvestmentTransactions(c.Request.Context(), holdingID)
	if err != nil {
		h.logger.Error("list investment transactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": txs})
}

// CreateInvestmentTransaction handles POST /api/v1/holdings/:id/transactions
func (h *InvestmentHandler) CreateInvestmentTransaction(c *gin.Context) {
	holdingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid holding id"}})
		return
	}

	var req usecase.CreateInvestmentTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	tx, err := h.usecase.CreateInvestmentTransaction(c.Request.Context(), holdingID, req)
	if err != nil {
		if errors.Is(err, usecase.ErrHoldingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("create investment transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": tx})
}

// DeleteInvestmentTransaction handles DELETE /api/v1/investment-transactions/:id
func (h *InvestmentHandler) DeleteInvestmentTransaction(c *gin.Context) {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid transaction id"}})
		return
	}

	if err := h.usecase.DeleteInvestmentTransaction(c.Request.Context(), txID); err != nil {
		h.logger.Error("delete investment transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "transaction deleted successfully"}})
}

// SearchAssets handles GET /api/v1/assets/search?q=PETR4
func (h *InvestmentHandler) SearchAssets(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "query parameter 'q' is required"}})
		return
	}

	assets, err := h.usecase.SearchAssets(c.Request.Context(), q)
	if err != nil {
		h.logger.Error("search assets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": assets})
}

// ListCustomAssets handles GET /api/v1/custom-assets
func (h *InvestmentHandler) ListCustomAssets(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	assets, err := h.usecase.GetCustomAssets(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("list custom assets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": assets})
}

// CreateCustomAsset handles POST /api/v1/custom-assets
func (h *InvestmentHandler) CreateCustomAsset(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	var req usecase.CreateCustomAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	asset, err := h.usecase.CreateCustomAsset(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("create custom asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": asset})
}

// UpdateCustomAsset handles PUT /api/v1/custom-assets/:id
func (h *InvestmentHandler) UpdateCustomAsset(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	assetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid asset id"}})
		return
	}

	var req usecase.UpdateCustomAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	asset, err := h.usecase.UpdateCustomAsset(c.Request.Context(), assetID, userID, req)
	if err != nil {
		if errors.Is(err, usecase.ErrCustomAssetNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("update custom asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": asset})
}

// DeleteCustomAsset handles DELETE /api/v1/custom-assets/:id
func (h *InvestmentHandler) DeleteCustomAsset(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	assetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid asset id"}})
		return
	}

	if err := h.usecase.DeleteCustomAsset(c.Request.Context(), assetID, userID); err != nil {
		if errors.Is(err, usecase.ErrCustomAssetNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("delete custom asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "custom asset deleted successfully"}})
}
