package entity

import (
	"time"

	"github.com/google/uuid"
)

// Budget represents a spending limit for a category in a given period.
type Budget struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	CategoryID   *uuid.UUID `json:"category_id,omitempty" db:"category_id"`
	Amount       float64    `json:"amount" db:"amount"`
	Period       string     `json:"period" db:"period"` // weekly, monthly, yearly
	Month        *int       `json:"month,omitempty" db:"month"`
	Year         *int       `json:"year,omitempty" db:"year"`
	ThresholdPct float64    `json:"threshold_pct" db:"threshold_pct"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	// Joined fields
	CategoryName  *string `json:"category_name,omitempty" db:"category_name"`
	CategoryColor *string `json:"category_color,omitempty" db:"category_color"`
	CategoryIcon  *string `json:"category_icon,omitempty" db:"category_icon"`
}
