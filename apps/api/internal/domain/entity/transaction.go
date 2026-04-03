package entity

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	AccountID      uuid.UUID  `json:"account_id" db:"account_id"`
	CategoryID     *uuid.UUID `json:"category_id,omitempty" db:"category_id"`
	Type           string     `json:"type" db:"type"` // income, expense, transfer
	Amount         float64    `json:"amount" db:"amount"`
	Description    *string    `json:"description,omitempty" db:"description"`
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	Date           time.Time  `json:"date" db:"date"`
	TransferPairID *uuid.UUID `json:"transfer_pair_id,omitempty" db:"transfer_pair_id"`
	RecurrenceID   *uuid.UUID `json:"recurrence_id,omitempty" db:"recurrence_id"`
	ImportID       *string    `json:"import_id,omitempty" db:"import_id"`
	Tags           []string   `json:"tags,omitempty" db:"tags"`
	AICategorized  bool       `json:"ai_categorized" db:"ai_categorized"`
	AIConfidence   *float64   `json:"ai_confidence,omitempty" db:"ai_confidence"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	// Joined fields
	AccountName   *string `json:"account_name,omitempty" db:"account_name"`
	CategoryName  *string `json:"category_name,omitempty" db:"category_name"`
	CategoryColor *string `json:"category_color,omitempty" db:"category_color"`
	CategoryIcon  *string `json:"category_icon,omitempty" db:"category_icon"`
}
