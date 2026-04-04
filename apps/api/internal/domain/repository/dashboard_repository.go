package repository

import (
	"context"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/google/uuid"
)

// DashboardOverview holds aggregated financial data for the dashboard.
type DashboardOverview struct {
	NetBalance         float64               `json:"net_balance"`
	TotalIncome        float64               `json:"total_income"`
	TotalExpense       float64               `json:"total_expense"`
	TotalPatrimony     float64               `json:"total_patrimony"`
	TopCategories      []CategorySummary     `json:"top_categories"`
	AlertBudgets       []BudgetProgress      `json:"alert_budgets"`
	RecentTransactions []*entity.Transaction `json:"recent_transactions"`
}

// MonthlyCashflow holds income/expense aggregate for a single month.
type MonthlyCashflow struct {
	Month   int     `json:"month"`
	Year    int     `json:"year"`
	Label   string  `json:"label"` // e.g. "Jan/26"
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Balance float64 `json:"balance"`
}

// DashboardRepository defines data access for dashboard aggregates.
type DashboardRepository interface {
	GetOverview(ctx context.Context, userID uuid.UUID, month, year int) (*DashboardOverview, error)
	GetCashflow(ctx context.Context, userID uuid.UUID, months int) ([]MonthlyCashflow, error)
}
