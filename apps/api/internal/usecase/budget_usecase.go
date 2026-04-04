package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
)

// Sentinel errors for budgets.
var (
	ErrBudgetNotFound = errors.New("budget not found")
)

// CreateBudgetRequest holds data needed to create a budget.
type CreateBudgetRequest struct {
	CategoryID   *uuid.UUID `json:"category_id"`
	Amount       float64    `json:"amount" binding:"required,gt=0"`
	Period       string     `json:"period" binding:"required,oneof=weekly monthly yearly"`
	Month        *int       `json:"month"`
	Year         *int       `json:"year"`
	ThresholdPct float64    `json:"threshold_pct"`
}

// UpdateBudgetRequest holds fields that can be updated on a budget.
type UpdateBudgetRequest struct {
	CategoryID   *uuid.UUID `json:"category_id"`
	Amount       *float64   `json:"amount" binding:"omitempty,gt=0"`
	Period       *string    `json:"period" binding:"omitempty,oneof=weekly monthly yearly"`
	Month        *int       `json:"month"`
	Year         *int       `json:"year"`
	ThresholdPct *float64   `json:"threshold_pct"`
}

// BudgetUseCase defines business logic for budgets.
type BudgetUseCase interface {
	Create(ctx context.Context, userID uuid.UUID, req CreateBudgetRequest) (*entity.Budget, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Budget, error)
	List(ctx context.Context, userID uuid.UUID, month, year int) ([]*entity.Budget, error)
	Update(ctx context.Context, id, userID uuid.UUID, req UpdateBudgetRequest) (*entity.Budget, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	GetProgress(ctx context.Context, userID uuid.UUID, month, year int) ([]*domainrepo.BudgetProgress, error)
}

type budgetUseCase struct {
	repo domainrepo.BudgetRepository
}

// NewBudgetUseCase creates a new BudgetUseCase implementation.
func NewBudgetUseCase(repo domainrepo.BudgetRepository) BudgetUseCase {
	return &budgetUseCase{repo: repo}
}

func (uc *budgetUseCase) Create(ctx context.Context, userID uuid.UUID, req CreateBudgetRequest) (*entity.Budget, error) {
	threshold := req.ThresholdPct
	if threshold == 0 {
		threshold = 80
	}

	now := time.Now()
	b := &entity.Budget{
		ID:           uuid.New(),
		UserID:       userID,
		CategoryID:   req.CategoryID,
		Amount:       req.Amount,
		Period:       req.Period,
		Month:        req.Month,
		Year:         req.Year,
		ThresholdPct: threshold,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.repo.Create(ctx, b); err != nil {
		return nil, fmt.Errorf("budgetUseCase.Create: %w", err)
	}
	return b, nil
}

func (uc *budgetUseCase) GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Budget, error) {
	b, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("budgetUseCase.GetByID: %w", err)
	}
	if b == nil {
		return nil, ErrBudgetNotFound
	}
	return b, nil
}

func (uc *budgetUseCase) List(ctx context.Context, userID uuid.UUID, month, year int) ([]*entity.Budget, error) {
	budgets, err := uc.repo.FindByUserAndPeriod(ctx, userID, month, year)
	if err != nil {
		return nil, fmt.Errorf("budgetUseCase.List: %w", err)
	}
	return budgets, nil
}

func (uc *budgetUseCase) Update(ctx context.Context, id, userID uuid.UUID, req UpdateBudgetRequest) (*entity.Budget, error) {
	b, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("budgetUseCase.Update find: %w", err)
	}
	if b == nil {
		return nil, ErrBudgetNotFound
	}

	if req.CategoryID != nil {
		b.CategoryID = req.CategoryID
	}
	if req.Amount != nil {
		b.Amount = *req.Amount
	}
	if req.Period != nil {
		b.Period = *req.Period
	}
	if req.Month != nil {
		b.Month = req.Month
	}
	if req.Year != nil {
		b.Year = req.Year
	}
	if req.ThresholdPct != nil {
		b.ThresholdPct = *req.ThresholdPct
	}

	if err := uc.repo.Update(ctx, b); err != nil {
		return nil, fmt.Errorf("budgetUseCase.Update save: %w", err)
	}
	return b, nil
}

func (uc *budgetUseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	b, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("budgetUseCase.Delete find: %w", err)
	}
	if b == nil {
		return ErrBudgetNotFound
	}
	if err := uc.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("budgetUseCase.Delete: %w", err)
	}
	return nil
}

func (uc *budgetUseCase) GetProgress(ctx context.Context, userID uuid.UUID, month, year int) ([]*domainrepo.BudgetProgress, error) {
	progress, err := uc.repo.GetProgress(ctx, userID, month, year)
	if err != nil {
		return nil, fmt.Errorf("budgetUseCase.GetProgress: %w", err)
	}
	return progress, nil
}
