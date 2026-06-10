package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ImportHandler handles HTTP requests for data imports.
type ImportHandler struct {
	usecase usecase.ImportUseCase
	logger  *zap.Logger
}

// NewImportHandler creates a new ImportHandler.
func NewImportHandler(uc usecase.ImportUseCase, logger *zap.Logger) *ImportHandler {
	return &ImportHandler{usecase: uc, logger: logger}
}

// ImportOFX handles POST /api/v1/imports/ofx
//
//	@Summary		Importar extrato OFX
//	@Description	Importa transações a partir de arquivo OFX/QFX (plano Pro)
//	@Tags			Imports
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file		formData	file		true	"Arquivo OFX"
//	@Param			account_id	formData	string		true	"UUID da conta destino"
//	@Success		200	{object}	MessageResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		402	{object}	ErrorResponse	"Plano insuficiente"
//	@Router			/imports/ofx [post]
func (h *ImportHandler) ImportOFX(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	accountIDStr := c.PostForm("account_id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid account_id"}})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "file is required"}})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("import ofx read file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to read file"}})
		return
	}

	result, err := h.usecase.ImportOFX(c.Request.Context(), userID, accountID, data)
	if err != nil {
		h.logger.Error("import ofx", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "import failed"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ImportCSV handles POST /api/v1/imports/csv
//
//	@Summary		Importar extrato CSV
//	@Description	Importa transações de arquivo CSV com mapeamento de colunas (plano Pro)
//	@Tags			Imports
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file		formData	file		true	"Arquivo CSV"
//	@Param			account_id	formData	string		true	"UUID da conta destino"
//	@Param			mapping		formData	string		false	"JSON de mapeamento de colunas"
//	@Success		200	{object}	MessageResponse
//	@Failure		402	{object}	ErrorResponse	"Plano insuficiente"
//	@Router			/imports/csv [post]
func (h *ImportHandler) ImportCSV(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid user id"}})
		return
	}

	accountIDStr := c.PostForm("account_id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid account_id"}})
		return
	}

	mappingStr := c.PostForm("mapping")
	var mapping usecase.CSVMapping
	if mappingStr != "" {
		if err := json.Unmarshal([]byte(mappingStr), &mapping); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "invalid mapping JSON"}})
			return
		}
	} else {
		// Default mapping: date=0, amount=1, description=2, type=3
		mapping = usecase.CSVMapping{
			DateCol:        0,
			AmountCol:      1,
			DescriptionCol: 2,
			TypeCol:        3,
			DateFormat:     "2006-01-02",
		}
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "file is required"}})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("import csv read file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to read file"}})
		return
	}

	result, err := h.usecase.ImportCSV(c.Request.Context(), userID, accountID, data, mapping)
	if err != nil {
		h.logger.Error("import csv", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "import failed"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// PreviewCSV handles POST /api/v1/imports/csv/preview
//
//	@Summary		Pré-visualizar CSV antes de importar
//	@Description	Retorna as primeiras linhas do CSV para conferência (plano Pro)
//	@Tags			Imports
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file	formData	file	true	"Arquivo CSV"
//	@Success		200	{object}	MessageResponse
//	@Failure		402	{object}	ErrorResponse	"Plano insuficiente"
//	@Router			/imports/csv/preview [post]
func (h *ImportHandler) PreviewCSV(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_INPUT", "message": "file is required"}})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("import csv preview read file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to read file"}})
		return
	}

	rows, err := h.usecase.PreviewCSV(c.Request.Context(), data)
	if err != nil {
		h.logger.Error("import csv preview", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "preview failed"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rows})
}
