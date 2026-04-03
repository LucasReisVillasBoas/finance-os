package entity

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty" db:"user_id"` // null = system category
	Name      string     `json:"name" db:"name"`
	Type      string     `json:"type" db:"type"` // income, expense, transfer
	Icon      *string    `json:"icon,omitempty" db:"icon"`
	Color     *string    `json:"color,omitempty" db:"color"`
	IsSystem  bool       `json:"is_system" db:"is_system"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	IsActive  bool       `json:"is_active" db:"is_active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}
