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

type customAssetRepository struct {
	db *pgxpool.Pool
}

// NewCustomAssetRepository creates a new PostgreSQL-backed CustomAssetRepository.
func NewCustomAssetRepository(db *pgxpool.Pool) domainrepo.CustomAssetRepository {
	return &customAssetRepository{db: db}
}

func (r *customAssetRepository) Create(ctx context.Context, a *entity.CustomAsset) error {
	query := `
		INSERT INTO custom_assets (id, user_id, name, type, current_value, purchase_value, purchase_date, monthly_income, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.Exec(ctx, query,
		a.ID, a.UserID, a.Name, a.Type, a.CurrentValue, a.PurchaseValue,
		a.PurchaseDate, a.MonthlyIncome, a.Description, a.IsActive, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("customAssetRepository.Create: %w", err)
	}
	return nil
}

func (r *customAssetRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.CustomAsset, error) {
	query := `
		SELECT id, user_id, name, type, current_value, purchase_value, purchase_date, monthly_income, description, is_active, created_at, updated_at
		FROM custom_assets
		WHERE id = $1 AND user_id = $2 AND is_active = true`
	row := r.db.QueryRow(ctx, query, id, userID)
	a, err := scanCustomAsset(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("customAssetRepository.FindByID: %w", err)
	}
	return a, nil
}

func (r *customAssetRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.CustomAsset, error) {
	query := `
		SELECT id, user_id, name, type, current_value, purchase_value, purchase_date, monthly_income, description, is_active, created_at, updated_at
		FROM custom_assets
		WHERE user_id = $1 AND is_active = true
		ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("customAssetRepository.FindByUserID: %w", err)
	}
	defer rows.Close()

	var assets []*entity.CustomAsset
	for rows.Next() {
		a, err := scanCustomAsset(rows)
		if err != nil {
			return nil, fmt.Errorf("customAssetRepository.FindByUserID scan: %w", err)
		}
		assets = append(assets, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("customAssetRepository.FindByUserID rows: %w", err)
	}
	return assets, nil
}

func (r *customAssetRepository) Update(ctx context.Context, a *entity.CustomAsset) error {
	query := `
		UPDATE custom_assets SET
			name = $3, type = $4, current_value = $5, purchase_value = $6,
			purchase_date = $7, monthly_income = $8, description = $9, updated_at = $10
		WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query,
		a.ID, a.UserID, a.Name, a.Type, a.CurrentValue, a.PurchaseValue,
		a.PurchaseDate, a.MonthlyIncome, a.Description, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("customAssetRepository.Update: %w", err)
	}
	return nil
}

func (r *customAssetRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE custom_assets SET is_active = false, updated_at = $3 WHERE id = $1 AND user_id = $2`,
		id, userID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("customAssetRepository.Delete: %w", err)
	}
	return nil
}

func scanCustomAsset(row pgx.Row) (*entity.CustomAsset, error) {
	a := &entity.CustomAsset{}
	err := row.Scan(
		&a.ID, &a.UserID, &a.Name, &a.Type, &a.CurrentValue, &a.PurchaseValue,
		&a.PurchaseDate, &a.MonthlyIncome, &a.Description, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}
