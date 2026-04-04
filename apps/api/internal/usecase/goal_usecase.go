package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
)

// Sentinel errors for goals.
var (
	ErrGoalNotFound = errors.New("goal not found")
)

// CreateGoalRequest holds data needed to create a goal.
type CreateGoalRequest struct {
	Name                string     `json:"name" binding:"required,min=2,max=255"`
	TargetAmount        float64    `json:"target_amount" binding:"required,gt=0"`
	TargetDate          *time.Time `json:"target_date"`
	MonthlyContribution *float64   `json:"monthly_contribution"`
	Icon                *string    `json:"icon"`
	Color               *string    `json:"color"`
}

// UpdateGoalRequest holds fields that can be updated on a goal.
type UpdateGoalRequest struct {
	Name                *string    `json:"name" binding:"omitempty,min=2,max=255"`
	TargetAmount        *float64   `json:"target_amount" binding:"omitempty,gt=0"`
	TargetDate          *time.Time `json:"target_date"`
	MonthlyContribution *float64   `json:"monthly_contribution"`
	Icon                *string    `json:"icon"`
	Color               *string    `json:"color"`
}

// ContributeRequest holds data for a contribution to a goal.
type ContributeRequest struct {
	Amount float64   `json:"amount" binding:"required,gt=0"`
	Notes  *string   `json:"notes"`
	Date   time.Time `json:"date" binding:"required"`
}

// GoalUseCase defines business logic for goals.
type GoalUseCase interface {
	Create(ctx context.Context, userID uuid.UUID, req CreateGoalRequest) (*entity.Goal, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Goal, error)
	List(ctx context.Context, userID uuid.UUID) ([]*entity.Goal, error)
	Update(ctx context.Context, id, userID uuid.UUID, req UpdateGoalRequest) (*entity.Goal, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	Contribute(ctx context.Context, goalID, userID uuid.UUID, req ContributeRequest) (*entity.GoalContribution, error)
	GetProjections(ctx context.Context, userID uuid.UUID) ([]*domainrepo.GoalProjection, error)
}

type goalUseCase struct {
	repo domainrepo.GoalRepository
}

// NewGoalUseCase creates a new GoalUseCase implementation.
func NewGoalUseCase(repo domainrepo.GoalRepository) GoalUseCase {
	return &goalUseCase{repo: repo}
}

func (uc *goalUseCase) Create(ctx context.Context, userID uuid.UUID, req CreateGoalRequest) (*entity.Goal, error) {
	now := time.Now()
	g := &entity.Goal{
		ID:                  uuid.New(),
		UserID:              userID,
		Name:                req.Name,
		TargetAmount:        req.TargetAmount,
		CurrentAmount:       0,
		TargetDate:          req.TargetDate,
		MonthlyContribution: req.MonthlyContribution,
		Icon:                req.Icon,
		Color:               req.Color,
		IsAchieved:          false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := uc.repo.Create(ctx, g); err != nil {
		return nil, fmt.Errorf("goalUseCase.Create: %w", err)
	}
	return g, nil
}

func (uc *goalUseCase) GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Goal, error) {
	g, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("goalUseCase.GetByID: %w", err)
	}
	if g == nil {
		return nil, ErrGoalNotFound
	}
	return g, nil
}

func (uc *goalUseCase) List(ctx context.Context, userID uuid.UUID) ([]*entity.Goal, error) {
	goals, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("goalUseCase.List: %w", err)
	}
	return goals, nil
}

func (uc *goalUseCase) Update(ctx context.Context, id, userID uuid.UUID, req UpdateGoalRequest) (*entity.Goal, error) {
	g, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("goalUseCase.Update find: %w", err)
	}
	if g == nil {
		return nil, ErrGoalNotFound
	}

	if req.Name != nil {
		g.Name = *req.Name
	}
	if req.TargetAmount != nil {
		g.TargetAmount = *req.TargetAmount
		// Recheck achievement with new target
		g.IsAchieved = g.CurrentAmount >= g.TargetAmount
	}
	if req.TargetDate != nil {
		g.TargetDate = req.TargetDate
	}
	if req.MonthlyContribution != nil {
		g.MonthlyContribution = req.MonthlyContribution
	}
	if req.Icon != nil {
		g.Icon = req.Icon
	}
	if req.Color != nil {
		g.Color = req.Color
	}

	if err := uc.repo.Update(ctx, g); err != nil {
		return nil, fmt.Errorf("goalUseCase.Update save: %w", err)
	}
	return g, nil
}

func (uc *goalUseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	g, err := uc.repo.FindByID(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("goalUseCase.Delete find: %w", err)
	}
	if g == nil {
		return ErrGoalNotFound
	}
	if err := uc.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("goalUseCase.Delete: %w", err)
	}
	return nil
}

func (uc *goalUseCase) Contribute(ctx context.Context, goalID, userID uuid.UUID, req ContributeRequest) (*entity.GoalContribution, error) {
	// Verify goal exists and belongs to user
	g, err := uc.repo.FindByID(ctx, goalID, userID)
	if err != nil {
		return nil, fmt.Errorf("goalUseCase.Contribute find: %w", err)
	}
	if g == nil {
		return nil, ErrGoalNotFound
	}

	c := &entity.GoalContribution{
		ID:        uuid.New(),
		GoalID:    goalID,
		Amount:    req.Amount,
		Notes:     req.Notes,
		Date:      req.Date,
		CreatedAt: time.Now(),
	}

	if err := uc.repo.AddContribution(ctx, c); err != nil {
		return nil, fmt.Errorf("goalUseCase.Contribute: %w", err)
	}
	return c, nil
}

func (uc *goalUseCase) GetProjections(ctx context.Context, userID uuid.UUID) ([]*domainrepo.GoalProjection, error) {
	goals, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("goalUseCase.GetProjections: %w", err)
	}

	now := time.Now()
	projections := make([]*domainrepo.GoalProjection, 0, len(goals))

	for _, g := range goals {
		remaining := g.TargetAmount - g.CurrentAmount
		if remaining < 0 {
			remaining = 0
		}

		var progressPct float64
		if g.TargetAmount > 0 {
			progressPct = (g.CurrentAmount / g.TargetAmount) * 100
			if progressPct > 100 {
				progressPct = 100
			}
		}

		proj := &domainrepo.GoalProjection{
			GoalID:          g.ID,
			Name:            g.Name,
			TargetAmount:    g.TargetAmount,
			CurrentAmount:   g.CurrentAmount,
			RemainingAmount: remaining,
			ProgressPct:     progressPct,
		}

		if g.MonthlyContribution != nil && *g.MonthlyContribution > 0 && remaining > 0 {
			monthsFloat := remaining / *g.MonthlyContribution
			months := int(math.Ceil(monthsFloat))
			proj.MonthsToGoal = &months

			estimatedDate := now.AddDate(0, months, 0)
			proj.EstimatedDate = &estimatedDate
		}

		projections = append(projections, proj)
	}

	return projections, nil
}
