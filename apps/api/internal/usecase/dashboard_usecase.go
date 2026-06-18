package usecase

import (
	"context"
	"fmt"

	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
)

// DashboardUseCase defines business logic for dashboard data.
type DashboardUseCase interface {
	GetOverview(ctx context.Context, userID uuid.UUID, month, year int) (*domainrepo.DashboardOverview, error)
	GetCashflow(ctx context.Context, userID uuid.UUID) ([]domainrepo.MonthlyCashflow, error)
	GetPatrimonyHistory(ctx context.Context, userID uuid.UUID) ([]domainrepo.PatrimonySnapshot, error)
}

type dashboardUseCase struct {
	repo domainrepo.DashboardRepository
}

// NewDashboardUseCase creates a new DashboardUseCase implementation.
func NewDashboardUseCase(repo domainrepo.DashboardRepository) DashboardUseCase {
	return &dashboardUseCase{repo: repo}
}

func (uc *dashboardUseCase) GetOverview(ctx context.Context, userID uuid.UUID, month, year int) (*domainrepo.DashboardOverview, error) {
	overview, err := uc.repo.GetOverview(ctx, userID, month, year)
	if err != nil {
		return nil, fmt.Errorf("dashboardUseCase.GetOverview: %w", err)
	}
	return overview, nil
}

// GetCashflow returns the last 12 months of cashflow data.
func (uc *dashboardUseCase) GetCashflow(ctx context.Context, userID uuid.UUID) ([]domainrepo.MonthlyCashflow, error) {
	cashflow, err := uc.repo.GetCashflow(ctx, userID, 12)
	if err != nil {
		return nil, fmt.Errorf("dashboardUseCase.GetCashflow: %w", err)
	}
	return cashflow, nil
}

// GetPatrimonyHistory returns 12 months of cumulative net worth snapshots.
func (uc *dashboardUseCase) GetPatrimonyHistory(ctx context.Context, userID uuid.UUID) ([]domainrepo.PatrimonySnapshot, error) {
	history, err := uc.repo.GetPatrimonyHistory(ctx, userID, 12)
	if err != nil {
		return nil, fmt.Errorf("dashboardUseCase.GetPatrimonyHistory: %w", err)
	}
	return history, nil
}
