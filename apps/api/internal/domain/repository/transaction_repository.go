package repository

import (
	"context"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/google/uuid"
)

// TransactionFilter holds filter params for listing transactions.
type TransactionFilter struct {
	StartDate  *time.Time
	EndDate    *time.Time
	CategoryID *uuid.UUID
	AccountID  *uuid.UUID
	Type       *string
	Tags       []string
	Search     *string
	Page       int
	PageSize   int
}

// TransactionSummary holds aggregate data for a period.
type TransactionSummary struct {
	TotalIncome  float64           `json:"total_income"`
	TotalExpense float64           `json:"total_expense"`
	Balance      float64           `json:"balance"`
	ByCategory   []CategorySummary `json:"by_category"`
	PrevIncome   float64           `json:"prev_income"`
	PrevExpense  float64           `json:"prev_expense"`
}

// CategorySummary holds aggregated totals per category.
type CategorySummary struct {
	CategoryID   *uuid.UUID `json:"category_id"`
	CategoryName string     `json:"category_name"`
	Total        float64    `json:"total"`
	Count        int        `json:"count"`
	Color        *string    `json:"color"`
}

// TransactionRepository defines data access operations for transactions.
type TransactionRepository interface {
	Create(ctx context.Context, tx *entity.Transaction) error
	FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Transaction, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, filter TransactionFilter) ([]*entity.Transaction, int, error)
	Update(ctx context.Context, tx *entity.Transaction) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	GetSummary(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*TransactionSummary, error)
	CreateTransfer(ctx context.Context, debit, credit *entity.Transaction) error
	UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, delta float64) error
}
