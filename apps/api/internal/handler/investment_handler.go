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
//
//	@Summary		Listar portfólios
//	@Tags			Investments
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	PortfolioListResponse
//	@Router			/portfolios [get]
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
//
//	@Summary		Criar portfólio
//	@Tags			Investments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body	usecase.CreatePortfolioRequest	true	"Dados do portfólio"
//	@Success		201	{object}	PortfolioResponse
//	@Failure		400	{object}	ErrorResponse
//	@Router			/portfolios [post]
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
//
//	@Summary		Atualizar portfólio
//	@Tags			Investments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string							true	"UUID do portfólio"
//	@Param			body	body	usecase.UpdatePortfolioRequest	true	"Campos a atualizar"
//	@Success		200	{object}	PortfolioResponse
//	@Router			/portfolios/{id} [put]
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
//
//	@Summary		Remover portfólio
//	@Tags			Investments
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID do portfólio"
//	@Success		200	{object}	MessageResponse
//	@Router			/portfolios/{id} [delete]
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
//
//	@Summary		Listar holdings do portfólio
//	@Description	Retorna holdings com P&L calculado
//	@Tags			Investments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID do portfólio"
//	@Success		200	{object}	HoldingListResponse
//	@Router			/portfolios/{id}/holdings [get]
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
//
//	@Summary		Adicionar holding ao portfólio
//	@Tags			Investments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string						true	"UUID do portfólio"
//	@Param			body	body	usecase.CreateHoldingRequest	true	"Dados do ativo"
//	@Success		201	{object}	HoldingResponse
//	@Router			/portfolios/{id}/holdings [post]
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
//
//	@Summary		Atualizar holding
//	@Tags			Investments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string						true	"UUID do holding"
//	@Param			body	body	usecase.UpdateHoldingRequest	true	"Campos a atualizar"
//	@Success		200	{object}	HoldingResponse
//	@Router			/holdings/{id} [put]
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
//
//	@Summary		Remover holding
//	@Tags			Investments
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID do holding"
//	@Success		200	{object}	MessageResponse
//	@Router			/holdings/{id} [delete]
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
//
//	@Summary		Listar movimentações do holding
//	@Tags			Investments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID do holding"
//	@Success		200	{object}	InvestmentTransactionListResponse
//	@Router			/holdings/{id}/transactions [get]
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
//
//	@Summary		Registrar movimentação (compra/venda/dividendo)
//	@Tags			Investments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string										true	"UUID do holding"
//	@Param			body	body	usecase.CreateInvestmentTransactionRequest	true	"Dados da movimentação"
//	@Success		201	{object}	InvestmentTransactionResponse
//	@Router			/holdings/{id}/transactions [post]
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
//
//	@Summary		Remover movimentação de investimento
//	@Tags			Investments
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID da movimentação"
//	@Success		200	{object}	MessageResponse
//	@Router			/investment-transactions/{id} [delete]
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
//
//	@Summary		Buscar ativos por ticker ou nome
//	@Tags			Investments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q	query	string	true	"Ticker ou nome (ex: PETR4)"
//	@Success		200	{object}	AssetSearchResponse
//	@Failure		400	{object}	ErrorResponse
//	@Router			/assets/search [get]
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
//
//	@Summary		Listar ativos personalizados
//	@Description	Retorna imóveis, veículos e outros ativos cadastrados manualmente
//	@Tags			Investments
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	CustomAssetListResponse
//	@Router			/custom-assets [get]
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
//
//	@Summary		Criar ativo personalizado
//	@Tags			Investments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body	usecase.CreateCustomAssetRequest	true	"Dados do ativo"
//	@Success		201	{object}	CustomAssetResponse
//	@Router			/custom-assets [post]
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
//
//	@Summary		Atualizar ativo personalizado
//	@Tags			Investments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string							true	"UUID do ativo"
//	@Param			body	body	usecase.UpdateCustomAssetRequest	true	"Campos a atualizar"
//	@Success		200	{object}	CustomAssetResponse
//	@Router			/custom-assets/{id} [put]
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
//
//	@Summary		Remover ativo personalizado
//	@Tags			Investments
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID do ativo"
//	@Success		200	{object}	MessageResponse
//	@Router			/custom-assets/{id} [delete]
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
