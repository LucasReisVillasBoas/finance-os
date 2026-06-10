package handler

import (
	"errors"
	"net/http"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AccountHandler handles HTTP requests for account endpoints.
type AccountHandler struct {
	usecase usecase.AccountUseCase
	logger  *zap.Logger
}

// NewAccountHandler creates a new AccountHandler.
func NewAccountHandler(uc usecase.AccountUseCase, logger *zap.Logger) *AccountHandler {
	return &AccountHandler{usecase: uc, logger: logger}
}

// List handles GET /api/v1/accounts
//
//	@Summary		Listar contas
//	@Description	Retorna todas as contas ativas do usuário autenticado
//	@Tags			Accounts
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	AccountListResponse
//	@Failure		401	{object}	ErrorResponse
//	@Router			/accounts [get]
func (h *AccountHandler) List(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	accounts, err := h.usecase.GetAll(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("list accounts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": accounts})
}

// Create handles POST /api/v1/accounts
//
//	@Summary		Criar conta
//	@Description	Cria uma nova conta bancária, carteira ou cartão
//	@Tags			Accounts
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		usecase.CreateAccountRequest	true	"Dados da conta"
//	@Success		201		{object}	AccountResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/accounts [post]
func (h *AccountHandler) Create(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	var req usecase.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	account, err := h.usecase.Create(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("create account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": account})
}

// GetByID handles GET /api/v1/accounts/:id
//
//	@Summary		Buscar conta por ID
//	@Tags			Accounts
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"UUID da conta"
//	@Success		200	{object}	AccountResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/accounts/{id} [get]
func (h *AccountHandler) GetByID(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid account id"}})
		return
	}

	account, err := h.usecase.GetByID(c.Request.Context(), accountID, userID)
	if err != nil {
		if errors.Is(err, usecase.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("get account by id", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": account})
}

// Update handles PUT /api/v1/accounts/:id
//
//	@Summary		Atualizar conta
//	@Tags			Accounts
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"UUID da conta"
//	@Param			body	body		usecase.UpdateAccountRequest	true	"Campos a atualizar"
//	@Success		200		{object}	AccountResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Router			/accounts/{id} [put]
func (h *AccountHandler) Update(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid account id"}})
		return
	}

	var req usecase.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	account, err := h.usecase.Update(c.Request.Context(), accountID, userID, req)
	if err != nil {
		if errors.Is(err, usecase.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("update account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": account})
}

// Delete handles DELETE /api/v1/accounts/:id
//
//	@Summary		Desativar conta
//	@Tags			Accounts
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"UUID da conta"
//	@Success		200	{object}	MessageResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/accounts/{id} [delete]
func (h *AccountHandler) Delete(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid account id"}})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), accountID, userID); err != nil {
		if errors.Is(err, usecase.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		h.logger.Error("delete account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "account deactivated successfully"}})
}

// Summary handles GET /api/v1/accounts/summary
//
//	@Summary		Resumo de saldo
//	@Description	Retorna saldo total e saldo por conta
//	@Tags			Accounts
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	AccountSummaryResponse
//	@Failure		401	{object}	ErrorResponse
//	@Router			/accounts/summary [get]
func (h *AccountHandler) Summary(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	summary, err := h.usecase.GetSummary(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("get account summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// parseUserID extracts and parses the user_id set by the auth middleware.
func parseUserID(c *gin.Context) (uuid.UUID, error) {
	raw, _ := c.Get("user_id")
	str, ok := raw.(string)
	if !ok || str == "" {
		return uuid.Nil, errors.New("user_id not in context")
	}
	return uuid.Parse(str)
}
