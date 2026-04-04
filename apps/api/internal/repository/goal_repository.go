package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type goalRepository struct {
	db *pgxpool.Pool
}

// NewGoalRepository creates a new PostgreSQL-backed GoalRepository.
func NewGoalRepository(db *pgxpool.Pool) domainrepo.GoalRepository {
	return &goalRepository{db: db}
}

func (r *goalRepository) Create(ctx context.Context, g *entity.Goal) error {
	query := `
		INSERT INTO goals (
			id, user_id, name, target_amount, current_amount, target_date,
			monthly_contribution, icon, color, is_achieved, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err := r.db.Exec(ctx, query,
		g.ID, g.UserID, g.Name, g.TargetAmount, g.CurrentAmount, g.TargetDate,
		g.MonthlyContribution, g.Icon, g.Color, g.IsAchieved, g.CreatedAt, g.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("goalRepository.Create: %w", err)
	}
	return nil
}

func (r *goalRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Goal, error) {
	query := `
		SELECT id, user_id, name, target_amount, current_amount, target_date,
			monthly_contribution, icon, color, is_achieved, created_at, updated_at
		FROM goals
		WHERE id = $1 AND user_id = $2`
	row := r.db.QueryRow(ctx, query, id, userID)
	g, err := scanGoal(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("goalRepository.FindByID: %w", err)
	}
	return g, nil
}

func (r *goalRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Goal, error) {
	query := `
		SELECT id, user_id, name, target_amount, current_amount, target_date,
			monthly_contribution, icon, color, is_achieved, created_at, updated_at
		FROM goals
		WHERE user_id = $1
		ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("goalRepository.FindByUserID query: %w", err)
	}
	defer rows.Close()

	var result []*entity.Goal
	for rows.Next() {
		g, err := scanGoal(rows)
		if err != nil {
			return nil, fmt.Errorf("goalRepository.FindByUserID scan: %w", err)
		}
		result = append(result, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("goalRepository.FindByUserID rows: %w", err)
	}
	if result == nil {
		result = []*entity.Goal{}
	}
	return result, nil
}

func (r *goalRepository) Update(ctx context.Context, g *entity.Goal) error {
	query := `
		UPDATE goals SET
			name = $3, target_amount = $4, current_amount = $5, target_date = $6,
			monthly_contribution = $7, icon = $8, color = $9, is_achieved = $10,
			updated_at = $11
		WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query,
		g.ID, g.UserID, g.Name, g.TargetAmount, g.CurrentAmount, g.TargetDate,
		g.MonthlyContribution, g.Icon, g.Color, g.IsAchieved, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("goalRepository.Update: %w", err)
	}
	return nil
}

func (r *goalRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM goals WHERE id = $1 AND user_id = $2`, id, userID,
	)
	if err != nil {
		return fmt.Errorf("goalRepository.Delete: %w", err)
	}
	return nil
}

func (r *goalRepository) AddContribution(ctx context.Context, c *entity.GoalContribution) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("goalRepository.AddContribution begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	insertQuery := `
		INSERT INTO goal_contributions (id, goal_id, amount, notes, date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := tx.Exec(ctx, insertQuery,
		c.ID, c.GoalID, c.Amount, c.Notes, c.Date, c.CreatedAt,
	); err != nil {
		return fmt.Errorf("goalRepository.AddContribution insert: %w", err)
	}

	updateQuery := `
		UPDATE goals SET
			current_amount = current_amount + $1,
			is_achieved = (current_amount + $1 >= target_amount),
			updated_at = NOW()
		WHERE id = $2`
	if _, err := tx.Exec(ctx, updateQuery, c.Amount, c.GoalID); err != nil {
		return fmt.Errorf("goalRepository.AddContribution update goal: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("goalRepository.AddContribution commit: %w", err)
	}
	return nil
}

func (r *goalRepository) GetContributions(ctx context.Context, goalID uuid.UUID) ([]*entity.GoalContribution, error) {
	query := `
		SELECT id, goal_id, amount, notes, date, created_at
		FROM goal_contributions
		WHERE goal_id = $1
		ORDER BY date DESC`
	rows, err := r.db.Query(ctx, query, goalID)
	if err != nil {
		return nil, fmt.Errorf("goalRepository.GetContributions query: %w", err)
	}
	defer rows.Close()

	var result []*entity.GoalContribution
	for rows.Next() {
		var c entity.GoalContribution
		if err := rows.Scan(&c.ID, &c.GoalID, &c.Amount, &c.Notes, &c.Date, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("goalRepository.GetContributions scan: %w", err)
		}
		result = append(result, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("goalRepository.GetContributions rows: %w", err)
	}
	if result == nil {
		result = []*entity.GoalContribution{}
	}
	return result, nil
}

func scanGoal(row pgx.Row) (*entity.Goal, error) {
	g := &entity.Goal{}
	err := row.Scan(
		&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.TargetDate,
		&g.MonthlyContribution, &g.Icon, &g.Color, &g.IsAchieved, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return g, nil
}
