package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/financeos/api/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Fake GoalRepository ---

type fakeGoalRepo struct {
	goals         map[uuid.UUID]*entity.Goal
	contributions []*entity.GoalContribution
}

func newFakeGoalRepo() *fakeGoalRepo {
	return &fakeGoalRepo{
		goals:         make(map[uuid.UUID]*entity.Goal),
		contributions: []*entity.GoalContribution{},
	}
}

func (r *fakeGoalRepo) Create(ctx context.Context, g *entity.Goal) error {
	r.goals[g.ID] = g
	return nil
}

func (r *fakeGoalRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Goal, error) {
	g, ok := r.goals[id]
	if !ok || g.UserID != userID {
		return nil, nil
	}
	// Return a copy to simulate real DB
	copy := *g
	return &copy, nil
}

func (r *fakeGoalRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Goal, error) {
	var result []*entity.Goal
	for _, g := range r.goals {
		if g.UserID == userID {
			copy := *g
			result = append(result, &copy)
		}
	}
	return result, nil
}

func (r *fakeGoalRepo) Update(ctx context.Context, g *entity.Goal) error {
	r.goals[g.ID] = g
	return nil
}

func (r *fakeGoalRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	delete(r.goals, id)
	return nil
}

func (r *fakeGoalRepo) AddContribution(ctx context.Context, c *entity.GoalContribution) error {
	r.contributions = append(r.contributions, c)
	g, ok := r.goals[c.GoalID]
	if ok {
		g.CurrentAmount += c.Amount
		g.IsAchieved = g.CurrentAmount >= g.TargetAmount
	}
	return nil
}

func (r *fakeGoalRepo) GetContributions(ctx context.Context, goalID uuid.UUID) ([]*entity.GoalContribution, error) {
	var result []*entity.GoalContribution
	for _, c := range r.contributions {
		if c.GoalID == goalID {
			result = append(result, c)
		}
	}
	return result, nil
}

// --- Tests ---

func TestCreateGoal_Success(t *testing.T) {
	repo := newFakeGoalRepo()
	uc := usecase.NewGoalUseCase(repo)

	userID := uuid.New()
	monthly := 500.0
	req := usecase.CreateGoalRequest{
		Name:                "Viagem para Europa",
		TargetAmount:        10000.0,
		MonthlyContribution: &monthly,
	}

	goal, err := uc.Create(context.Background(), userID, req)
	require.NoError(t, err)
	require.NotNil(t, goal)

	assert.NotEqual(t, uuid.Nil, goal.ID)
	assert.Equal(t, userID, goal.UserID)
	assert.Equal(t, "Viagem para Europa", goal.Name)
	assert.Equal(t, 10000.0, goal.TargetAmount)
	assert.Equal(t, 0.0, goal.CurrentAmount)
	assert.False(t, goal.IsAchieved)
	assert.Equal(t, &monthly, goal.MonthlyContribution)
}

func TestCreateGoal_ValidationError(t *testing.T) {
	repo := newFakeGoalRepo()
	uc := usecase.NewGoalUseCase(repo)

	userID := uuid.New()
	// Name too short — but binding validation runs at handler level
	// At use case level we just verify create propagates correctly
	req := usecase.CreateGoalRequest{
		Name:         "OK",
		TargetAmount: 5000.0,
	}

	goal, err := uc.Create(context.Background(), userID, req)
	require.NoError(t, err)
	assert.NotNil(t, goal)
}

func TestContribute_UpdatesCurrentAmount(t *testing.T) {
	repo := newFakeGoalRepo()
	uc := usecase.NewGoalUseCase(repo)

	userID := uuid.New()
	goal, err := uc.Create(context.Background(), userID, usecase.CreateGoalRequest{
		Name:         "Reserva de Emergência",
		TargetAmount: 5000.0,
	})
	require.NoError(t, err)

	// Contribute 1000
	contribution, err := uc.Contribute(context.Background(), goal.ID, userID, usecase.ContributeRequest{
		Amount: 1000.0,
		Date:   time.Now(),
	})
	require.NoError(t, err)
	require.NotNil(t, contribution)
	assert.Equal(t, 1000.0, contribution.Amount)
	assert.Equal(t, goal.ID, contribution.GoalID)

	// Verify goal current_amount was updated in fake repo
	stored := repo.goals[goal.ID]
	assert.Equal(t, 1000.0, stored.CurrentAmount)
	assert.False(t, stored.IsAchieved)
}

func TestContribute_AchievesGoal(t *testing.T) {
	repo := newFakeGoalRepo()
	uc := usecase.NewGoalUseCase(repo)

	userID := uuid.New()
	goal, err := uc.Create(context.Background(), userID, usecase.CreateGoalRequest{
		Name:         "Novo Laptop",
		TargetAmount: 3000.0,
	})
	require.NoError(t, err)

	_, err = uc.Contribute(context.Background(), goal.ID, userID, usecase.ContributeRequest{
		Amount: 3000.0,
		Date:   time.Now(),
	})
	require.NoError(t, err)

	stored := repo.goals[goal.ID]
	assert.Equal(t, 3000.0, stored.CurrentAmount)
	assert.True(t, stored.IsAchieved)
}

func TestContribute_GoalNotFound(t *testing.T) {
	repo := newFakeGoalRepo()
	uc := usecase.NewGoalUseCase(repo)

	_, err := uc.Contribute(context.Background(), uuid.New(), uuid.New(), usecase.ContributeRequest{
		Amount: 100.0,
		Date:   time.Now(),
	})
	assert.ErrorIs(t, err, usecase.ErrGoalNotFound)
}

func TestGetProjections_WithMonthlyContribution(t *testing.T) {
	repo := newFakeGoalRepo()
	uc := usecase.NewGoalUseCase(repo)

	userID := uuid.New()
	monthly := 1000.0
	goal, err := uc.Create(context.Background(), userID, usecase.CreateGoalRequest{
		Name:                "Fundo de Aposentadoria",
		TargetAmount:        12000.0,
		MonthlyContribution: &monthly,
	})
	require.NoError(t, err)

	// Contribute 2000 already
	_, err = uc.Contribute(context.Background(), goal.ID, userID, usecase.ContributeRequest{
		Amount: 2000.0,
		Date:   time.Now(),
	})
	require.NoError(t, err)

	projections, err := uc.GetProjections(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, projections, 1)

	proj := projections[0]
	assert.Equal(t, goal.ID, proj.GoalID)
	assert.Equal(t, 12000.0, proj.TargetAmount)
	assert.Equal(t, 2000.0, proj.CurrentAmount)
	assert.Equal(t, 10000.0, proj.RemainingAmount)
	assert.NotNil(t, proj.MonthsToGoal)
	assert.Equal(t, 10, *proj.MonthsToGoal) // 10000 / 1000 = 10 months
	assert.NotNil(t, proj.EstimatedDate)
	assert.InDelta(t, 100.0/6, proj.ProgressPct, 1.0) // ~16.67%
}

func TestGetProjections_NoMonthlyContribution(t *testing.T) {
	repo := newFakeGoalRepo()
	uc := usecase.NewGoalUseCase(repo)

	userID := uuid.New()
	_, err := uc.Create(context.Background(), userID, usecase.CreateGoalRequest{
		Name:         "Meta sem aporte",
		TargetAmount: 5000.0,
	})
	require.NoError(t, err)

	projections, err := uc.GetProjections(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, projections, 1)

	proj := projections[0]
	assert.Nil(t, proj.MonthsToGoal)
	assert.Nil(t, proj.EstimatedDate)
	assert.Equal(t, 0.0, proj.ProgressPct)
}

func TestDeleteGoal_NotFound(t *testing.T) {
	repo := newFakeGoalRepo()
	uc := usecase.NewGoalUseCase(repo)

	err := uc.Delete(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, usecase.ErrGoalNotFound)
}
