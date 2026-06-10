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

type holdingRepository struct {
	db *pgxpool.Pool
}

// NewHoldingRepository creates a new PostgreSQL-backed HoldingRepository.
func NewHoldingRepository(db *pgxpool.Pool) domainrepo.HoldingRepository {
	return &holdingRepository{db: db}
}

func (r *holdingRepository) Create(ctx context.Context, h *entity.Holding) error {
	query := `
		INSERT INTO holdings (id, portfolio_id, asset_id, name, type, quantity, avg_price, total_invested, current_value, unrealized_pnl, unrealized_pnl_pct, realized_pnl, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err := r.db.Exec(ctx, query,
		h.ID, h.PortfolioID, h.AssetID, h.Name, h.Type,
		h.Quantity, h.AvgPrice, h.TotalInvested, h.CurrentValue,
		h.UnrealizedPnL, h.UnrealizedPnLPct, h.RealizedPnL,
		h.CreatedAt, h.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("holdingRepository.Create: %w", err)
	}
	return nil
}

func (r *holdingRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Holding, error) {
	query := `
		SELECT h.id, h.portfolio_id, h.asset_id, h.name, h.type, h.quantity, h.avg_price,
		       h.total_invested, h.current_value, h.unrealized_pnl, h.unrealized_pnl_pct,
		       h.realized_pnl, h.created_at, h.updated_at,
		       a.ticker AS asset_ticker, a.current_price AS asset_current_price
		FROM holdings h
		LEFT JOIN assets a ON a.id = h.asset_id
		WHERE h.id = $1`
	row := r.db.QueryRow(ctx, query, id)
	h, err := scanHolding(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("holdingRepository.FindByID: %w", err)
	}
	return h, nil
}

func (r *holdingRepository) FindByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]*entity.Holding, error) {
	query := `
		SELECT h.id, h.portfolio_id, h.asset_id, h.name, h.type, h.quantity, h.avg_price,
		       h.total_invested, h.current_value, h.unrealized_pnl, h.unrealized_pnl_pct,
		       h.realized_pnl, h.created_at, h.updated_at,
		       a.ticker AS asset_ticker, a.current_price AS asset_current_price
		FROM holdings h
		LEFT JOIN assets a ON a.id = h.asset_id
		WHERE h.portfolio_id = $1
		ORDER BY h.name ASC`
	rows, err := r.db.Query(ctx, query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("holdingRepository.FindByPortfolioID: %w", err)
	}
	defer rows.Close()

	var holdings []*entity.Holding
	for rows.Next() {
		h, err := scanHolding(rows)
		if err != nil {
			return nil, fmt.Errorf("holdingRepository.FindByPortfolioID scan: %w", err)
		}
		holdings = append(holdings, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("holdingRepository.FindByPortfolioID rows: %w", err)
	}
	if holdings == nil {
		holdings = []*entity.Holding{}
	}
	return holdings, nil
}

func (r *holdingRepository) FindAll(ctx context.Context) ([]*entity.Holding, error) {
	query := `
		SELECT h.id, h.portfolio_id, h.asset_id, h.name, h.type, h.quantity, h.avg_price,
		       h.total_invested, h.current_value, h.unrealized_pnl, h.unrealized_pnl_pct,
		       h.realized_pnl, h.created_at, h.updated_at,
		       a.ticker AS asset_ticker, a.current_price AS asset_current_price
		FROM holdings h
		LEFT JOIN assets a ON a.id = h.asset_id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("holdingRepository.FindAll: %w", err)
	}
	defer rows.Close()

	var holdings []*entity.Holding
	for rows.Next() {
		h, err := scanHolding(rows)
		if err != nil {
			return nil, fmt.Errorf("holdingRepository.FindAll scan: %w", err)
		}
		holdings = append(holdings, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("holdingRepository.FindAll rows: %w", err)
	}
	if holdings == nil {
		holdings = []*entity.Holding{}
	}
	return holdings, nil
}

func (r *holdingRepository) Update(ctx context.Context, h *entity.Holding) error {
	query := `
		UPDATE holdings SET
			name = $2, type = $3, quantity = $4, avg_price = $5,
			total_invested = $6, current_value = $7, unrealized_pnl = $8,
			unrealized_pnl_pct = $9, realized_pnl = $10, updated_at = $11
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		h.ID, h.Name, h.Type, h.Quantity, h.AvgPrice,
		h.TotalInvested, h.CurrentValue, h.UnrealizedPnL,
		h.UnrealizedPnLPct, h.RealizedPnL, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("holdingRepository.Update: %w", err)
	}
	return nil
}

func (r *holdingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM holdings WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("holdingRepository.Delete: %w", err)
	}
	return nil
}

func scanHolding(row pgx.Row) (*entity.Holding, error) {
	h := &entity.Holding{}
	err := row.Scan(
		&h.ID, &h.PortfolioID, &h.AssetID, &h.Name, &h.Type, &h.Quantity, &h.AvgPrice,
		&h.TotalInvested, &h.CurrentValue, &h.UnrealizedPnL, &h.UnrealizedPnLPct,
		&h.RealizedPnL, &h.CreatedAt, &h.UpdatedAt,
		&h.AssetTicker, &h.AssetCurrentPrice,
	)
	if err != nil {
		return nil, err
	}
	return h, nil
}
