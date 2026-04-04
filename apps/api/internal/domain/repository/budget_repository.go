package repository

import (
	"context"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/google/uuid"
)

// BudgetProgress holds computed progress data for a budget.
type BudgetProgress struct {
	BudgetID      uuid.UUID  `json:"budget_id"`
	CategoryID    *uuid.UUID `json:"category_id"`
	CategoryName  string     `json:"category_name"`
	CategoryColor *string    `json:"category_color"`
	CategoryIcon  *string    `json:"category_icon"`
	Planned       float64    `json:"planned"`
	Actual        float64    `json:"actual"`
	Percentage    float64    `json:"percentage"`
	IsAlert       bool       `json:"is_alert"` // actual/planned >= threshold_pct/100
}

// BudgetRepository defines data access operations for budgets.
type BudgetRepository interface {
	Create(ctx context.Context, b *entity.Budget) error
	FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Budget, error)
	FindByUserAndPeriod(ctx context.Context, userID uuid.UUID, month, year int) ([]*entity.Budget, error)
	Update(ctx context.Context, b *entity.Budget) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	GetProgress(ctx context.Context, userID uuid.UUID, month, year int) ([]*BudgetProgress, error)
}
