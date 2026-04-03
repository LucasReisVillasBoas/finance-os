package repository

import (
	"context"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/google/uuid"
)

// AccountRepository defines data access operations for accounts.
type AccountRepository interface {
	Create(ctx context.Context, account *entity.Account) error
	FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Account, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Account, error)
	Update(ctx context.Context, account *entity.Account) error
	SoftDelete(ctx context.Context, id, userID uuid.UUID) error
	GetSummary(ctx context.Context, userID uuid.UUID) (*AccountSummary, error)
}

// AccountSummary holds aggregated balance data for a user's accounts.
type AccountSummary struct {
	TotalBalance    float64          `json:"total_balance"`
	NetBalance      float64          `json:"net_balance"` // excluding credit
	TotalPatrimony  float64          `json:"total_patrimony"`
	AccountBalances []AccountBalance `json:"account_balances"`
}

// AccountBalance holds the balance for a single account.
type AccountBalance struct {
	AccountID   uuid.UUID `json:"account_id"`
	AccountName string    `json:"account_name"`
	Balance     float64   `json:"balance"`
}
