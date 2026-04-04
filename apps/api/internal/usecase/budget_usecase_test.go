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

// --- Fake BudgetRepository ---

type fakeBudgetRepo struct {
	items map[uuid.UUID]*entity.Budget
}

func newFakeBudgetRepo() *fakeBudgetRepo {
	return &fakeBudgetRepo{items: make(map[uuid.UUID]*entity.Budget)}
}

func (r *fakeBudgetRepo) Create(ctx context.Context, b *entity.Budget) error {
	r.items[b.ID] = b
	return nil
}

func (r *fakeBudgetRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Budget, error) {
	b, ok := r.items[id]
	if !ok || b.UserID != userID {
		return nil, nil
	}
	return b, nil
}

func (r *fakeBudgetRepo) FindByUserAndPeriod(ctx context.Context, userID uuid.UUID, month, year int) ([]*entity.Budget, error) {
	var result []*entity.Budget
	for _, b := range r.items {
		if b.UserID == userID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (r *fakeBudgetRepo) Update(ctx context.Context, b *entity.Budget) error {
	r.items[b.ID] = b
	return nil
}

func (r *fakeBudgetRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	delete(r.items, id)
	return nil
}

func (r *fakeBudgetRepo) GetProgress(ctx context.Context, userID uuid.UUID, month, year int) ([]*domainrepo.BudgetProgress, error) {
	return []*domainrepo.BudgetProgress{}, nil
}

// --- Tests ---

func TestCreateBudget_DefaultThreshold(t *testing.T) {
	repo := newFakeBudgetRepo()
	uc := usecase.NewBudgetUseCase(repo)

	userID := uuid.New()
	req := usecase.CreateBudgetRequest{
		Amount: 2000.0,
		Period: "monthly",
		// ThresholdPct not set → should default to 80
	}

	budget, err := uc.Create(context.Background(), userID, req)
	require.NoError(t, err)
	require.NotNil(t, budget)

	assert.Equal(t, 80.0, budget.ThresholdPct, "default threshold should be 80")
	assert.Equal(t, userID, budget.UserID)
	assert.Equal(t, 2000.0, budget.Amount)
	assert.Equal(t, "monthly", budget.Period)
	assert.NotEqual(t, uuid.Nil, budget.ID)
}

func TestCreateBudget_Success(t *testing.T) {
	repo := newFakeBudgetRepo()
	uc := usecase.NewBudgetUseCase(repo)

	userID := uuid.New()
	catID := uuid.New()
	month := 4
	year := 2026

	req := usecase.CreateBudgetRequest{
		CategoryID:   &catID,
		Amount:       500.0,
		Period:       "monthly",
		Month:        &month,
		Year:         &year,
		ThresholdPct: 90.0,
	}

	budget, err := uc.Create(context.Background(), userID, req)
	require.NoError(t, err)
	require.NotNil(t, budget)

	assert.Equal(t, &catID, budget.CategoryID)
	assert.Equal(t, 500.0, budget.Amount)
	assert.Equal(t, "monthly", budget.Period)
	assert.Equal(t, &month, budget.Month)
	assert.Equal(t, &year, budget.Year)
	assert.Equal(t, 90.0, budget.ThresholdPct)
}

func TestCreateBudget_CustomThresholdNotOverridden(t *testing.T) {
	repo := newFakeBudgetRepo()
	uc := usecase.NewBudgetUseCase(repo)

	userID := uuid.New()
	req := usecase.CreateBudgetRequest{
		Amount:       1000.0,
		Period:       "yearly",
		ThresholdPct: 50.0,
	}

	budget, err := uc.Create(context.Background(), userID, req)
	require.NoError(t, err)
	assert.Equal(t, 50.0, budget.ThresholdPct, "custom threshold should not be overridden")
}

func TestDeleteBudget_NotFound(t *testing.T) {
	repo := newFakeBudgetRepo()
	uc := usecase.NewBudgetUseCase(repo)

	err := uc.Delete(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, usecase.ErrBudgetNotFound)
}

func TestUpdateBudget_NotFound(t *testing.T) {
	repo := newFakeBudgetRepo()
	uc := usecase.NewBudgetUseCase(repo)

	_, err := uc.Update(context.Background(), uuid.New(), uuid.New(), usecase.UpdateBudgetRequest{})
	assert.ErrorIs(t, err, usecase.ErrBudgetNotFound)
}

func TestListBudgets(t *testing.T) {
	repo := newFakeBudgetRepo()
	uc := usecase.NewBudgetUseCase(repo)

	userID := uuid.New()

	// Create two budgets
	for i := 0; i < 2; i++ {
		_, err := uc.Create(context.Background(), userID, usecase.CreateBudgetRequest{
			Amount: float64(100 * (i + 1)),
			Period: "monthly",
		})
		require.NoError(t, err)
	}

	budgets, err := uc.List(context.Background(), userID, 4, 2026)
	require.NoError(t, err)
	assert.Len(t, budgets, 2)
}
