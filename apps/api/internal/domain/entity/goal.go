package entity

import (
	"time"

	"github.com/google/uuid"
)

// Goal represents a financial goal for a user.
type Goal struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	UserID              uuid.UUID  `json:"user_id" db:"user_id"`
	Name                string     `json:"name" db:"name"`
	TargetAmount        float64    `json:"target_amount" db:"target_amount"`
	CurrentAmount       float64    `json:"current_amount" db:"current_amount"`
	TargetDate          *time.Time `json:"target_date,omitempty" db:"target_date"`
	MonthlyContribution *float64   `json:"monthly_contribution,omitempty" db:"monthly_contribution"`
	Icon                *string    `json:"icon,omitempty" db:"icon"`
	Color               *string    `json:"color,omitempty" db:"color"`
	IsAchieved          bool       `json:"is_achieved" db:"is_achieved"`
	PortfolioID         *uuid.UUID `json:"portfolio_id,omitempty" db:"portfolio_id"`
	PortfolioValue      float64    `json:"portfolio_value,omitempty" db:"-"` // computed at read time
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// GoalContribution represents a contribution made toward a goal.
type GoalContribution struct {
	ID        uuid.UUID `json:"id" db:"id"`
	GoalID    uuid.UUID `json:"goal_id" db:"goal_id"`
	Amount    float64   `json:"amount" db:"amount"`
	Notes     *string   `json:"notes,omitempty" db:"notes"`
	Date      time.Time `json:"date" db:"date"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
