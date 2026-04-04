package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/financeos/api/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// WebhookHandler handles incoming webhooks from external services.
type WebhookHandler struct {
	whatsappUC usecase.WhatsAppUseCase
	evolutionURL string
	evolutionKey string
	logger       *zap.Logger
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(
	whatsappUC usecase.WhatsAppUseCase,
	evolutionURL, evolutionKey string,
	logger *zap.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		whatsappUC:   whatsappUC,
		evolutionURL: evolutionURL,
		evolutionKey: evolutionKey,
		logger:       logger,
	}
}

// WhatsApp handles POST /webhooks/whatsapp
func (h *WebhookHandler) WhatsApp(c *gin.Context) {
	var payload entity.WhatsAppWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.logger.Warn("webhook whatsapp invalid payload", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"status": "ok"}) // always return 200 to webhooks
		return
	}

	// Only process messages events
	if payload.Event != "messages.upsert" {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	phone := payload.Data.Key.RemoteJid
	message := payload.Data.Message.Conversation

	if phone == "" || message == "" {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	// Remove @s.whatsapp.net suffix if present
	phone = strings.Split(phone, "@")[0]

	response, err := h.whatsappUC.HandleMessage(c.Request.Context(), phone, message)
	if err != nil {
		h.logger.Error("webhook whatsapp handle message", zap.Error(err), zap.String("phone", phone))
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	// Send response back via Evolution API
	if response != "" && h.evolutionURL != "" {
		if err := sendEvolutionMessage(h.evolutionURL, h.evolutionKey, payload.Instance, phone, response); err != nil {
			h.logger.Warn("webhook whatsapp send reply failed", zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "response": response})
}

// sendEvolutionMessage sends a text message via Evolution API.
func sendEvolutionMessage(apiURL, apiKey, instance, phone, message string) error {
	payload := map[string]interface{}{
		"number": phone,
		"text":   message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("sendEvolutionMessage marshal: %w", err)
	}

	url := fmt.Sprintf("%s/message/sendText/%s", strings.TrimRight(apiURL, "/"), instance)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("sendEvolutionMessage new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sendEvolutionMessage do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendEvolutionMessage: status %d", resp.StatusCode)
	}

	return nil
}
