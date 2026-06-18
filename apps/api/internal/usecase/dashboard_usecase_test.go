package usecase_test

import (
	"context"
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
	overview *domainrepo.DashboardOverview
	cashflow []domainrepo.MonthlyCashflow
	err      error
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
	return []domainrepo.PatrimonySnapshot{}, nil
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
