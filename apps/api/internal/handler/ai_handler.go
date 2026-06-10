package handler

import (
	"net/http"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AIHandler handles HTTP requests for AI-powered features.
type AIHandler struct {
	usecase usecase.AIUseCase
	logger  *zap.Logger
}

// NewAIHandler creates a new AIHandler.
func NewAIHandler(uc usecase.AIUseCase, logger *zap.Logger) *AIHandler {
	return &AIHandler{usecase: uc, logger: logger}
}

// GetSpendingForecast handles GET /api/v1/ai/spending-forecast
//
//	@Summary		Previsão de gastos
//	@Description	Análise preditiva de gastos com IA (plano Pro)
//	@Tags			AI
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	AIForecastResponse
//	@Failure		402	{object}	ErrorResponse	"Plano insuficiente"
//	@Router			/ai/spending-forecast [get]
func (h *AIHandler) GetSpendingForecast(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	result, err := h.usecase.GetSpendingForecast(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("GetSpendingForecast", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to generate forecast"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetPortfolioAnalysis handles GET /api/v1/ai/portfolio-analysis
//
//	@Summary		Análise de portfólio com IA
//	@Description	Insights sobre diversificação e performance (plano Pro)
//	@Tags			AI
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	AIForecastResponse
//	@Failure		402	{object}	ErrorResponse	"Plano insuficiente"
//	@Router			/ai/portfolio-analysis [get]
func (h *AIHandler) GetPortfolioAnalysis(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	result, err := h.usecase.GetPortfolioAnalysis(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("GetPortfolioAnalysis", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to generate analysis"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Chat handles POST /api/v1/ai/chat
//
//	@Summary		Assistente financeiro
//	@Description	Conversa com o assistente financeiro alimentado por Claude
//	@Tags			AI
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body	object{message=string}	true	"Mensagem (máx. 2000 chars)"
//	@Success		200	{object}	AIChatResponse
//	@Failure		400	{object}	ErrorResponse
//	@Router			/ai/chat [post]
func (h *AIHandler) Chat(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	var req struct {
		Message string `json:"message" binding:"required,min=1,max=2000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	response, err := h.usecase.Chat(c.Request.Context(), userID, req.Message)
	if err != nil {
		h.logger.Error("AIChat", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to process chat"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"response": response}})
}
