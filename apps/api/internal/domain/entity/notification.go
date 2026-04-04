package entity

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents an in-app notification for a user.
type Notification struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	UserID    uuid.UUID              `json:"user_id" db:"user_id"`
	Type      string                 `json:"type" db:"type"`
	Title     string                 `json:"title" db:"title"`
	Message   *string                `json:"message,omitempty" db:"message"`
	Data      map[string]interface{} `json:"data,omitempty" db:"data"`
	IsRead    bool                   `json:"is_read" db:"is_read"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}
