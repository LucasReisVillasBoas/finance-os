package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/financeos/api/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Fake DashboardRepository ---

type fakeDashboardRepo struct {
	overview         *domainrepo.DashboardOverview
	cashflow         []domainrepo.MonthlyCashflow
	patrimonyHistory []domainrepo.PatrimonySnapshot
	err              error
}

func newFakeDashboardRepo() *fakeDashboardRepo {
	return &fakeDashboardRepo{
		overview: &domainrepo.DashboardOverview{
			NetBalance:         1000.0,
			TotalIncome:        2000.0,
			TotalExpense:       500.0,
			TotalPatrimony:     5000.0,
			TopCategories:      []domainrepo.CategorySummary{},
			AlertBudgets:       []domainrepo.BudgetProgress{},
			RecentTransactions: []*entity.Transaction{},
		},
		cashflow: []domainrepo.MonthlyCashflow{
			{Month: 1, Year: 2026, Label: "Jan/26", Income: 2000, Expense: 500, Balance: 1500},
			{Month: 2, Year: 2026, Label: "Fev/26", Income: 2200, Expense: 600, Balance: 1600},
		},
	}
}

func (r *fakeDashboardRepo) GetOverview(ctx context.Context, userID uuid.UUID, month, year int) (*domainrepo.DashboardOverview, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.overview, nil
}

func (r *fakeDashboardRepo) GetCashflow(ctx context.Context, userID uuid.UUID, months int) ([]domainrepo.MonthlyCashflow, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.cashflow, nil
}

func (r *fakeDashboardRepo) GetPatrimonyHistory(ctx context.Context, userID uuid.UUID, months int) ([]domainrepo.PatrimonySnapshot, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.patrimonyHistory, nil
}

// --- Tests ---

func TestDashboardUseCase_GetOverview(t *testing.T) {
	tests := []struct {
		name      string
		userID    uuid.UUID
		month     int
		year      int
		wantErr   bool
		checkData func(t *testing.T, overview *domainrepo.DashboardOverview)
	}{
		{
			name:    "success returns overview",
			userID:  uuid.New(),
			month:   4,
			year:    2026,
			wantErr: false,
			checkData: func(t *testing.T, overview *domainrepo.DashboardOverview) {
				assert.Equal(t, 1000.0, overview.NetBalance)
				assert.Equal(t, 2000.0, overview.TotalIncome)
				assert.Equal(t, 500.0, overview.TotalExpense)
				assert.Equal(t, 5000.0, overview.TotalPatrimony)
				assert.NotNil(t, overview.TopCategories)
				assert.NotNil(t, overview.AlertBudgets)
				assert.NotNil(t, overview.RecentTransactions)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeDashboardRepo()
			uc := usecase.NewDashboardUseCase(repo)

			overview, err := uc.GetOverview(context.Background(), tt.userID, tt.month, tt.year)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, overview)
			if tt.checkData != nil {
				tt.checkData(t, overview)
			}
		})
	}
}

func TestDashboardUseCase_GetCashflow(t *testing.T) {
	tests := []struct {
		name      string
		userID    uuid.UUID
		wantErr   bool
		wantCount int
	}{
		{
			name:      "success returns cashflow",
			userID:    uuid.New(),
			wantErr:   false,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeDashboardRepo()
			uc := usecase.NewDashboardUseCase(repo)

			cashflow, err := uc.GetCashflow(context.Background(), tt.userID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, cashflow, tt.wantCount)

			if len(cashflow) > 0 {
				assert.Equal(t, "Jan/26", cashflow[0].Label)
				assert.Equal(t, 2000.0, cashflow[0].Income)
				assert.Equal(t, 500.0, cashflow[0].Expense)
				assert.Equal(t, 1500.0, cashflow[0].Balance)
			}
		})
	}
}

// ------------------------------------------------------------------ D1 / D4 ----
// GetOverview – D1 unified net-worth fields and D4 investment capacity widget

func TestDashboardUseCase_GetOverview_NetWorthAndCapacity(t *testing.T) {
	repo := &fakeDashboardRepo{
		overview: &domainrepo.DashboardOverview{
			NetBalance:            3000.0,
			TotalIncome:           5000.0,
			TotalExpense:          2000.0,
			TotalPatrimony:        8000.0,
			// D1 – unified net worth
			InvestmentValue:       15000.0,
			CustomAssetValue:      5000.0,
			TotalNetWorth:         28000.0, // bank + investments + custom assets
			// D4 – investment capacity
			InvestmentCapacity:    3000.0,
			InvestmentCapacityPct: 60.0,
			TopCategories:         []domainrepo.CategorySummary{},
			AlertBudgets:          []domainrepo.BudgetProgress{},
			RecentTransactions:    []*entity.Transaction{},
		},
	}
	uc := usecase.NewDashboardUseCase(repo)

	overview, err := uc.GetOverview(context.Background(), uuid.New(), 6, 2026)
	require.NoError(t, err)
	require.NotNil(t, overview)

	// D1 assertions
	assert.Equal(t, 15000.0, overview.InvestmentValue, "D1: investment value")
	assert.Equal(t, 5000.0, overview.CustomAssetValue, "D1: custom asset value")
	assert.Equal(t, 28000.0, overview.TotalNetWorth, "D1: total net worth")

	// D4 assertions
	assert.Equal(t, 3000.0, overview.InvestmentCapacity, "D4: investment capacity amount")
	assert.Equal(t, 60.0, overview.InvestmentCapacityPct, "D4: investment capacity %")
}

// ------------------------------------------------------------------ D2 ----
// GetPatrimonyHistory – monthly net-worth snapshots flow through correctly

func TestDashboardUseCase_GetPatrimonyHistory(t *testing.T) {
	snapshots := []domainrepo.PatrimonySnapshot{
		{Month: 1, Year: 2026, Label: "Jan/26", BankSavings: 5000, InvestedTotal: 10000, TotalNetWorth: 15000},
		{Month: 2, Year: 2026, Label: "Fev/26", BankSavings: 5500, InvestedTotal: 11000, TotalNetWorth: 16500},
		{Month: 3, Year: 2026, Label: "Mar/26", BankSavings: 6000, InvestedTotal: 12000, TotalNetWorth: 18000},
	}

	repo := &fakeDashboardRepo{patrimonyHistory: snapshots}
	uc := usecase.NewDashboardUseCase(repo)

	history, err := uc.GetPatrimonyHistory(context.Background(), uuid.New())
	require.NoError(t, err)
	require.Len(t, history, 3)

	assert.Equal(t, "Jan/26", history[0].Label)
	assert.Equal(t, 5000.0, history[0].BankSavings)
	assert.Equal(t, 10000.0, history[0].InvestedTotal)
	assert.Equal(t, 15000.0, history[0].TotalNetWorth)

	assert.Equal(t, "Fev/26", history[1].Label)
	assert.Equal(t, 16500.0, history[1].TotalNetWorth)

	assert.Equal(t, "Mar/26", history[2].Label)
	assert.Equal(t, 18000.0, history[2].TotalNetWorth)
}

func TestDashboardUseCase_GetPatrimonyHistory_Empty(t *testing.T) {
	repo := &fakeDashboardRepo{patrimonyHistory: []domainrepo.PatrimonySnapshot{}}
	uc := usecase.NewDashboardUseCase(repo)

	history, err := uc.GetPatrimonyHistory(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Empty(t, history)
}

// ------------------------------------------------------------------ Error paths ----

func TestDashboardUseCase_GetOverview_Error(t *testing.T) {
	repo := &fakeDashboardRepo{err: errors.New("db failure")}
	uc := usecase.NewDashboardUseCase(repo)

	_, err := uc.GetOverview(context.Background(), uuid.New(), 1, 2026)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db failure")
}

func TestDashboardUseCase_GetCashflow_Error(t *testing.T) {
	repo := &fakeDashboardRepo{err: errors.New("connection timeout")}
	uc := usecase.NewDashboardUseCase(repo)

	_, err := uc.GetCashflow(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection timeout")
}

func TestDashboardUseCase_GetPatrimonyHistory_Error(t *testing.T) {
	repo := &fakeDashboardRepo{err: errors.New("query error")}
	uc := usecase.NewDashboardUseCase(repo)

	_, err := uc.GetPatrimonyHistory(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}
