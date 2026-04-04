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

type budgetRepository struct {
	db *pgxpool.Pool
}

// NewBudgetRepository creates a new PostgreSQL-backed BudgetRepository.
func NewBudgetRepository(db *pgxpool.Pool) domainrepo.BudgetRepository {
	return &budgetRepository{db: db}
}

func (r *budgetRepository) Create(ctx context.Context, b *entity.Budget) error {
	query := `
		INSERT INTO budgets (
			id, user_id, category_id, amount, period, month, year,
			threshold_pct, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.db.Exec(ctx, query,
		b.ID, b.UserID, b.CategoryID, b.Amount, b.Period, b.Month, b.Year,
		b.ThresholdPct, b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("budgetRepository.Create: %w", err)
	}
	return nil
}

func (r *budgetRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Budget, error) {
	query := `
		SELECT b.id, b.user_id, b.category_id, b.amount, b.period, b.month, b.year,
			b.threshold_pct, b.created_at, b.updated_at,
			c.name AS category_name, c.color AS category_color, c.icon AS category_icon
		FROM budgets b
		LEFT JOIN categories c ON c.id = b.category_id
		WHERE b.id = $1 AND b.user_id = $2`
	row := r.db.QueryRow(ctx, query, id, userID)
	bud, err := scanBudget(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("budgetRepository.FindByID: %w", err)
	}
	return bud, nil
}

func (r *budgetRepository) FindByUserAndPeriod(ctx context.Context, userID uuid.UUID, month, year int) ([]*entity.Budget, error) {
	query := `
		SELECT b.id, b.user_id, b.category_id, b.amount, b.period, b.month, b.year,
			b.threshold_pct, b.created_at, b.updated_at,
			c.name AS category_name, c.color AS category_color, c.icon AS category_icon
		FROM budgets b
		LEFT JOIN categories c ON c.id = b.category_id
		WHERE b.user_id = $1 AND (b.month = $2 OR b.month IS NULL) AND (b.year = $3 OR b.year IS NULL)
		ORDER BY b.created_at DESC`
	rows, err := r.db.Query(ctx, query, userID, month, year)
	if err != nil {
		return nil, fmt.Errorf("budgetRepository.FindByUserAndPeriod query: %w", err)
	}
	defer rows.Close()

	var result []*entity.Budget
	for rows.Next() {
		bud, err := scanBudget(rows)
		if err != nil {
			return nil, fmt.Errorf("budgetRepository.FindByUserAndPeriod scan: %w", err)
		}
		result = append(result, bud)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("budgetRepository.FindByUserAndPeriod rows: %w", err)
	}
	if result == nil {
		result = []*entity.Budget{}
	}
	return result, nil
}

func (r *budgetRepository) Update(ctx context.Context, b *entity.Budget) error {
	query := `
		UPDATE budgets SET
			category_id = $3, amount = $4, period = $5, month = $6, year = $7,
			threshold_pct = $8, updated_at = $9
		WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query,
		b.ID, b.UserID, b.CategoryID, b.Amount, b.Period, b.Month, b.Year,
		b.ThresholdPct, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("budgetRepository.Update: %w", err)
	}
	return nil
}

func (r *budgetRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM budgets WHERE id = $1 AND user_id = $2`, id, userID,
	)
	if err != nil {
		return fmt.Errorf("budgetRepository.Delete: %w", err)
	}
	return nil
}

func (r *budgetRepository) GetProgress(ctx context.Context, userID uuid.UUID, month, year int) ([]*domainrepo.BudgetProgress, error) {
	// Build start/end date for the period
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	query := `
		SELECT
			b.id AS budget_id,
			b.category_id,
			COALESCE(c.name, 'Geral') AS category_name,
			c.color AS category_color,
			c.icon AS category_icon,
			b.amount AS planned,
			b.threshold_pct,
			COALESCE(SUM(t.amount), 0) AS actual
		FROM budgets b
		LEFT JOIN categories c ON c.id = b.category_id
		LEFT JOIN transactions t ON (
			t.user_id = b.user_id
			AND t.type = 'expense'
			AND t.date >= $3
			AND t.date <= $4
			AND (b.category_id IS NULL OR t.category_id = b.category_id)
		)
		WHERE b.user_id = $1
		  AND (b.month = $2 OR b.month IS NULL)
		  AND (b.year = $5 OR b.year IS NULL)
		GROUP BY b.id, b.category_id, b.amount, b.threshold_pct, c.name, c.color, c.icon
		ORDER BY b.created_at DESC`

	rows, err := r.db.Query(ctx, query, userID, month, startDate, endDate, year)
	if err != nil {
		return nil, fmt.Errorf("budgetRepository.GetProgress query: %w", err)
	}
	defer rows.Close()

	var result []*domainrepo.BudgetProgress
	for rows.Next() {
		var p domainrepo.BudgetProgress
		var thresholdPct float64
		if err := rows.Scan(
			&p.BudgetID, &p.CategoryID, &p.CategoryName,
			&p.CategoryColor, &p.CategoryIcon,
			&p.Planned, &thresholdPct, &p.Actual,
		); err != nil {
			return nil, fmt.Errorf("budgetRepository.GetProgress scan: %w", err)
		}
		if p.Planned > 0 {
			p.Percentage = (p.Actual / p.Planned) * 100
		}
		p.IsAlert = p.Percentage >= thresholdPct
		result = append(result, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("budgetRepository.GetProgress rows: %w", err)
	}
	if result == nil {
		result = []*domainrepo.BudgetProgress{}
	}
	return result, nil
}

func scanBudget(row pgx.Row) (*entity.Budget, error) {
	b := &entity.Budget{}
	err := row.Scan(
		&b.ID, &b.UserID, &b.CategoryID, &b.Amount, &b.Period, &b.Month, &b.Year,
		&b.ThresholdPct, &b.CreatedAt, &b.UpdatedAt,
		&b.CategoryName, &b.CategoryColor, &b.CategoryIcon,
	)
	if err != nil {
		return nil, err
	}
	return b, nil
}
