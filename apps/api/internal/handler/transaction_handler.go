package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TransactionHandler handles HTTP requests for transaction endpoints.
type TransactionHandler struct {
	usecase usecase.TransactionUseCase
	logger  *zap.Logger
}

// NewTransactionHandler creates a new TransactionHandler.
func NewTransactionHandler(uc usecase.TransactionUseCase, logger *zap.Logger) *TransactionHandler {
	return &TransactionHandler{usecase: uc, logger: logger}
}

// List handles GET /api/v1/transactions
//
//	@Summary		Listar transações
//	@Description	Retorna transações paginadas com filtros opcionais
//	@Tags			Transactions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			start_date	query	string	false	"Data inicial (YYYY-MM-DD)"
//	@Param			end_date	query	string	false	"Data final (YYYY-MM-DD)"
//	@Param			type		query	string	false	"income | expense"
//	@Param			category_id	query	string	false	"UUID da categoria"
//	@Param			account_id	query	string	false	"UUID da conta"
//	@Param			search		query	string	false	"Busca por descrição"
//	@Param			page		query	int		false	"Página (default: 1)"
//	@Param			page_size	query	int		false	"Itens por página (default: 20)"
//	@Success		200	{object}	TransactionListResponse
//	@Failure		401	{object}	ErrorResponse
//	@Router			/transactions [get]
func (h *TransactionHandler) List(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	var req usecase.ListTransactionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	transactions, total, err := h.usecase.List(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("list transactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	c.JSON(http.StatusOK, gin.H{
		"data": transactions,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// Create handles POST /api/v1/transactions
//
//	@Summary		Criar transação
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body	usecase.CreateTransactionRequest	true	"Dados da transação"
//	@Success		201	{object}	TransactionResponse
//	@Failure		400	{object}	ErrorResponse
//	@Router			/transactions [post]
func (h *TransactionHandler) Create(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	var req usecase.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	tx, err := h.usecase.Create(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("create transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": tx})
}

// GetSummary handles GET /api/v1/transactions/summary
//
//	@Summary		Resumo de transações
//	@Description	Agrega receitas, despesas e saldo do período
//	@Tags			Transactions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			start_date	query	string	false	"Data inicial (YYYY-MM-DD)"
//	@Param			end_date	query	string	false	"Data final (YYYY-MM-DD)"
//	@Success		200	{object}	TransactionSummaryResponse
//	@Router			/transactions/summary [get]
func (h *TransactionHandler) GetSummary(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var startDate, endDate time.Time
	if startStr != "" {
		if startDate, err = time.Parse("2006-01-02", startStr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid start_date format, use YYYY-MM-DD"}})
			return
		}
	} else {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if endStr != "" {
		if endDate, err = time.Parse("2006-01-02", endStr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid end_date format, use YYYY-MM-DD"}})
			return
		}
		// End of day
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
	} else {
		now := time.Now()
		endDate = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
	}

	summary, err := h.usecase.GetSummary(c.Request.Context(), userID, startDate, endDate)
	if err != nil {
		h.logger.Error("get transaction summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// CreateTransfer handles POST /api/v1/transactions/transfer
//
//	@Summary		Criar transferência
//	@Description	Cria par de transações (saída + entrada) entre contas
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body	usecase.CreateTransferRequest	true	"Dados da transferência"
//	@Success		201	{object}	TransferResponse
//	@Failure		400	{object}	ErrorResponse
//	@Router			/transactions/transfer [post]
func (h *TransactionHandler) CreateTransfer(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	var req usecase.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	txs, err := h.usecase.CreateTransfer(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, usecase.ErrSameAccountTransfer) {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
			return
		}
		h.logger.Error("create transfer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": txs})
}

// GetByID handles GET /api/v1/transactions/:id
//
//	@Summary		Buscar transação por ID
//	@Tags			Transactions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID da transação"
//	@Success		200	{object}	TransactionResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/transactions/{id} [get]
func (h *TransactionHandler) GetByID(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid transaction id"}})
		return
	}

	tx, err := h.usecase.GetByID(c.Request.Context(), txID, userID)
	if err != nil {
		if errors.Is(err, usecase.ErrTransactionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("get transaction by id", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tx})
}

// Update handles PUT /api/v1/transactions/:id
//
//	@Summary		Atualizar transação
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string							true	"UUID da transação"
//	@Param			body	body	usecase.UpdateTransactionRequest	true	"Campos a atualizar"
//	@Success		200	{object}	TransactionResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/transactions/{id} [put]
func (h *TransactionHandler) Update(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid transaction id"}})
		return
	}

	var req usecase.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	tx, err := h.usecase.Update(c.Request.Context(), txID, userID, req)
	if err != nil {
		if errors.Is(err, usecase.ErrTransactionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("update transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tx})
}

// Delete handles DELETE /api/v1/transactions/:id
//
//	@Summary		Remover transação
//	@Tags			Transactions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID da transação"
//	@Success		200	{object}	MessageResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/transactions/{id} [delete]
func (h *TransactionHandler) Delete(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid transaction id"}})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), txID, userID); err != nil {
		if errors.Is(err, usecase.ErrTransactionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("delete transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "transaction deleted successfully"}})
}
