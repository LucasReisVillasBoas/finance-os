package entity

import (
	"time"

	"github.com/google/uuid"
)

// WhatsAppSession holds the state of a WhatsApp bot conversation session.
type WhatsAppSession struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	UserID       uuid.UUID              `json:"user_id" db:"user_id"`
	PhoneNumber  string                 `json:"phone_number" db:"phone_number"`
	State        string                 `json:"state" db:"state"` // idle, awaiting_confirmation
	SessionData  map[string]interface{} `json:"session_data" db:"session_data"`
	LastActivity time.Time              `json:"last_activity" db:"last_activity"`
	IsActive     bool                   `json:"is_active" db:"is_active"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// WhatsAppWebhookPayload represents an incoming webhook event from Evolution API.
type WhatsAppWebhookPayload struct {
	Event    string `json:"event"`
	Instance string `json:"instance"`
	Data     struct {
		Key struct {
			RemoteJid string `json:"remoteJid"`
		} `json:"key"`
		Message struct {
			Conversation string `json:"conversation"`
		} `json:"message"`
	} `json:"data"`
}
