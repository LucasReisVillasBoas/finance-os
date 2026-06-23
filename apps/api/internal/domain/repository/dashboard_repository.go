package repository

import (
	"context"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/google/uuid"
)

// DashboardOverview holds aggregated financial data for the dashboard.
type DashboardOverview struct {
	NetBalance            float64               `json:"net_balance"`
	TotalIncome           float64               `json:"total_income"`
	TotalExpense          float64               `json:"total_expense"`
	TotalPatrimony        float64               `json:"total_patrimony"`
	InvestmentValue       float64               `json:"investment_value"`
	CustomAssetValue      float64               `json:"custom_asset_value"`
	TotalNetWorth         float64               `json:"total_net_worth"`
	InvestmentCapacity    float64               `json:"investment_capacity"`
	InvestmentCapacityPct float64               `json:"investment_capacity_pct"`
	TopCategories         []CategorySummary     `json:"top_categories"`
	AlertBudgets          []BudgetProgress      `json:"alert_budgets"`
	RecentTransactions    []*entity.Transaction `json:"recent_transactions"`
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

// PatrimonySnapshot holds the net worth breakdown for a single month.
type PatrimonySnapshot struct {
	Month         int     `json:"month"`
	Year          int     `json:"year"`
	Label         string  `json:"label"`
	BankSavings   float64 `json:"bank_savings"`
	InvestedTotal float64 `json:"invested_total"`
	TotalNetWorth float64 `json:"total_net_worth"`
}

// DashboardRepository defines data access for dashboard aggregates.
type DashboardRepository interface {
	GetOverview(ctx context.Context, userID uuid.UUID, month, year int) (*DashboardOverview, error)
	GetCashflow(ctx context.Context, userID uuid.UUID, months int) ([]MonthlyCashflow, error)
	GetPatrimonyHistory(ctx context.Context, userID uuid.UUID, months int) ([]PatrimonySnapshot, error)
}
