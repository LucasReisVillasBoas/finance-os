package entity

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Type        string    `json:"type" db:"type"` // checking, savings, credit_card, investment, wallet, other
	Institution *string   `json:"institution,omitempty" db:"institution"`
	Balance     float64   `json:"balance" db:"balance"`
	CreditLimit *float64  `json:"credit_limit,omitempty" db:"credit_limit"`
	Color       *string   `json:"color,omitempty" db:"color"`
	Icon        *string   `json:"icon,omitempty" db:"icon"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
