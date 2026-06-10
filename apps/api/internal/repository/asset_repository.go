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

type assetRepository struct {
	db *pgxpool.Pool
}

// NewAssetRepository creates a new PostgreSQL-backed AssetRepository.
func NewAssetRepository(db *pgxpool.Pool) domainrepo.AssetRepository {
	return &assetRepository{db: db}
}

func (r *assetRepository) Create(ctx context.Context, a *entity.Asset) error {
	query := `
		INSERT INTO assets (id, ticker, name, type, exchange, currency, current_price, price_updated_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(ctx, query,
		a.ID, a.Ticker, a.Name, a.Type, a.Exchange, a.Currency,
		a.CurrentPrice, a.PriceUpdatedAt, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("assetRepository.Create: %w", err)
	}
	return nil
}

func (r *assetRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Asset, error) {
	query := `
		SELECT id, ticker, name, type, exchange, currency, current_price, price_updated_at, created_at, updated_at
		FROM assets
		WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	a, err := scanAsset(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("assetRepository.FindByID: %w", err)
	}
	return a, nil
}

func (r *assetRepository) FindByTicker(ctx context.Context, ticker, exchange string) (*entity.Asset, error) {
	query := `
		SELECT id, ticker, name, type, exchange, currency, current_price, price_updated_at, created_at, updated_at
		FROM assets
		WHERE ticker = $1 AND exchange = $2`
	row := r.db.QueryRow(ctx, query, ticker, exchange)
	a, err := scanAsset(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("assetRepository.FindByTicker: %w", err)
	}
	return a, nil
}

func (r *assetRepository) Search(ctx context.Context, query string) ([]*entity.Asset, error) {
	q := "%" + query + "%"
	sqlQuery := `
		SELECT id, ticker, name, type, exchange, currency, current_price, price_updated_at, created_at, updated_at
		FROM assets
		WHERE ticker ILIKE $1 OR name ILIKE $1
		ORDER BY ticker ASC
		LIMIT 20`
	rows, err := r.db.Query(ctx, sqlQuery, q)
	if err != nil {
		return nil, fmt.Errorf("assetRepository.Search: %w", err)
	}
	defer rows.Close()

	var assets []*entity.Asset
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, fmt.Errorf("assetRepository.Search scan: %w", err)
		}
		assets = append(assets, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("assetRepository.Search rows: %w", err)
	}
	if assets == nil {
		assets = []*entity.Asset{}
	}
	return assets, nil
}

func (r *assetRepository) UpdatePrice(ctx context.Context, id uuid.UUID, price float64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE assets SET current_price = $2, price_updated_at = $3, updated_at = $3 WHERE id = $1`,
		id, price, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("assetRepository.UpdatePrice: %w", err)
	}
	return nil
}

func (r *assetRepository) FindAll(ctx context.Context) ([]*entity.Asset, error) {
	query := `
		SELECT id, ticker, name, type, exchange, currency, current_price, price_updated_at, created_at, updated_at
		FROM assets
		ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("assetRepository.FindAll: %w", err)
	}
	defer rows.Close()

	var assets []*entity.Asset
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, fmt.Errorf("assetRepository.FindAll scan: %w", err)
		}
		assets = append(assets, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("assetRepository.FindAll rows: %w", err)
	}
	if assets == nil {
		assets = []*entity.Asset{}
	}
	return assets, nil
}

func scanAsset(row pgx.Row) (*entity.Asset, error) {
	a := &entity.Asset{}
	err := row.Scan(
		&a.ID, &a.Ticker, &a.Name, &a.Type, &a.Exchange, &a.Currency,
		&a.CurrentPrice, &a.PriceUpdatedAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}
