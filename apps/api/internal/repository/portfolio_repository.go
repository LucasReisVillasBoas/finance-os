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

type portfolioRepository struct {
	db *pgxpool.Pool
}

// NewPortfolioRepository creates a new PostgreSQL-backed PortfolioRepository.
func NewPortfolioRepository(db *pgxpool.Pool) domainrepo.PortfolioRepository {
	return &portfolioRepository{db: db}
}

func (r *portfolioRepository) Create(ctx context.Context, p *entity.Portfolio) error {
	query := `
		INSERT INTO portfolios (id, user_id, name, description, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query,
		p.ID, p.UserID, p.Name, p.Description, p.IsDefault, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("portfolioRepository.Create: %w", err)
	}
	return nil
}

func (r *portfolioRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Portfolio, error) {
	query := `
		SELECT id, user_id, name, description, is_default, created_at, updated_at
		FROM portfolios
		WHERE id = $1 AND user_id = $2`
	row := r.db.QueryRow(ctx, query, id, userID)
	p, err := scanPortfolio(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("portfolioRepository.FindByID: %w", err)
	}
	return p, nil
}

func (r *portfolioRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Portfolio, error) {
	query := `
		SELECT id, user_id, name, description, is_default, created_at, updated_at
		FROM portfolios
		WHERE user_id = $1
		ORDER BY is_default DESC, name ASC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("portfolioRepository.FindByUserID: %w", err)
	}
	defer rows.Close()

	var portfolios []*entity.Portfolio
	for rows.Next() {
		p, err := scanPortfolio(rows)
		if err != nil {
			return nil, fmt.Errorf("portfolioRepository.FindByUserID scan: %w", err)
		}
		portfolios = append(portfolios, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("portfolioRepository.FindByUserID rows: %w", err)
	}
	if portfolios == nil {
		portfolios = []*entity.Portfolio{}
	}
	return portfolios, nil
}

func (r *portfolioRepository) Update(ctx context.Context, p *entity.Portfolio) error {
	query := `
		UPDATE portfolios SET
			name = $3, description = $4, is_default = $5, updated_at = $6
		WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, p.ID, p.UserID, p.Name, p.Description, p.IsDefault, time.Now())
	if err != nil {
		return fmt.Errorf("portfolioRepository.Update: %w", err)
	}
	return nil
}

func (r *portfolioRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM portfolios WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("portfolioRepository.Delete: %w", err)
	}
	return nil
}

func scanPortfolio(row pgx.Row) (*entity.Portfolio, error) {
	p := &entity.Portfolio{}
	err := row.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.IsDefault, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}
