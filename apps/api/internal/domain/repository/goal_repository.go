package repository

import (
	"context"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/google/uuid"
)

// GoalProjection holds computed projection data for a goal.
type GoalProjection struct {
	GoalID          uuid.UUID  `json:"goal_id"`
	Name            string     `json:"name"`
	TargetAmount    float64    `json:"target_amount"`
	CurrentAmount   float64    `json:"current_amount"`
	RemainingAmount float64    `json:"remaining_amount"`
	MonthsToGoal    *int       `json:"months_to_goal,omitempty"`
	EstimatedDate   *time.Time `json:"estimated_date,omitempty"`
	ProgressPct     float64    `json:"progress_pct"`
}

// GoalRepository defines data access operations for goals.
type GoalRepository interface {
	Create(ctx context.Context, g *entity.Goal) error
	FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Goal, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Goal, error)
	Update(ctx context.Context, g *entity.Goal) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	AddContribution(ctx context.Context, c *entity.GoalContribution) error
	GetContributions(ctx context.Context, goalID uuid.UUID) ([]*entity.GoalContribution, error)
}
