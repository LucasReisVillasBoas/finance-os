package entity

import (
	"time"

	"github.com/google/uuid"
)

// Recurrence represents a recurring transaction template.
type Recurrence struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	AccountID   uuid.UUID  `json:"account_id" db:"account_id"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty" db:"category_id"`
	Type        string     `json:"type" db:"type"`        // income, expense
	Amount      float64    `json:"amount" db:"amount"`
	Description *string    `json:"description,omitempty" db:"description"`
	Frequency   string     `json:"frequency" db:"frequency"` // daily, weekly, biweekly, monthly, yearly
	StartDate   time.Time  `json:"start_date" db:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
	NextDueDate time.Time  `json:"next_due_date" db:"next_due_date"`
	AutoLaunch  bool       `json:"auto_launch" db:"auto_launch"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	// Joined fields
	AccountName  *string `json:"account_name,omitempty" db:"account_name"`
	CategoryName *string `json:"category_name,omitempty" db:"category_name"`
}
