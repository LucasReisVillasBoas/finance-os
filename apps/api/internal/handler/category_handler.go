package handler

import (
	"errors"
	"net/http"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CategoryHandler handles HTTP requests for category endpoints.
type CategoryHandler struct {
	usecase usecase.CategoryUseCase
	logger  *zap.Logger
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(uc usecase.CategoryUseCase, logger *zap.Logger) *CategoryHandler {
	return &CategoryHandler{usecase: uc, logger: logger}
}

// List handles GET /api/v1/categories
//
//	@Summary		Listar categorias
//	@Description	Retorna categorias do sistema e categorias personalizadas do usuário
//	@Tags			Categories
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	CategoryListResponse
//	@Failure		401	{object}	ErrorResponse
//	@Router			/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	categories, err := h.usecase.GetAll(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("list categories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// Create handles POST /api/v1/categories
//
//	@Summary		Criar categoria
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body	usecase.CreateCategoryRequest	true	"Dados da categoria"
//	@Success		201	{object}	CategoryResponse
//	@Failure		400	{object}	ErrorResponse
//	@Router			/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	var req usecase.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	category, err := h.usecase.Create(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("create category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": category})
}

// Update handles PUT /api/v1/categories/:id
//
//	@Summary		Atualizar categoria
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string	true	"UUID da categoria"
//	@Param			body	body	usecase.UpdateCategoryRequest	true	"Campos a atualizar"
//	@Success		200	{object}	CategoryResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid category id"}})
		return
	}

	var req usecase.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	category, err := h.usecase.Update(c.Request.Context(), categoryID, userID, req)
	if err != nil {
		if errors.Is(err, usecase.ErrCategoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		if errors.Is(err, usecase.ErrCannotModifySystemCategory) {
			c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "FORBIDDEN", "message": err.Error()}})
			return
		}
		h.logger.Error("update category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": category})
}

// Delete handles DELETE /api/v1/categories/:id
//
//	@Summary		Remover categoria
//	@Tags			Categories
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID da categoria"
//	@Success		200	{object}	MessageResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user context"}})
		return
	}

	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid category id"}})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), categoryID, userID); err != nil {
		if errors.Is(err, usecase.ErrCategoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": err.Error()}})
			return
		}
		if errors.Is(err, usecase.ErrCannotModifySystemCategory) {
			c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "FORBIDDEN", "message": err.Error()}})
			return
		}
		h.logger.Error("delete category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "category deleted successfully"}})
}
