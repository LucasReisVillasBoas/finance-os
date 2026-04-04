package repository

import (
	"context"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/google/uuid"
)

type PortfolioRepository interface {
	Create(ctx context.Context, p *entity.Portfolio) error
	FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Portfolio, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Portfolio, error)
	Update(ctx context.Context, p *entity.Portfolio) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type HoldingRepository interface {
	Create(ctx context.Context, h *entity.Holding) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Holding, error)
	FindByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]*entity.Holding, error)
	Update(ctx context.Context, h *entity.Holding) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]*entity.Holding, error)
}

type InvestmentTransactionRepository interface {
	Create(ctx context.Context, t *entity.InvestmentTransaction) error
	FindByHoldingID(ctx context.Context, holdingID uuid.UUID) ([]*entity.InvestmentTransaction, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AssetRepository interface {
	Create(ctx context.Context, a *entity.Asset) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Asset, error)
	FindByTicker(ctx context.Context, ticker, exchange string) (*entity.Asset, error)
	Search(ctx context.Context, query string) ([]*entity.Asset, error)
	UpdatePrice(ctx context.Context, id uuid.UUID, price float64) error
	FindAll(ctx context.Context) ([]*entity.Asset, error)
}

type CustomAssetRepository interface {
	Create(ctx context.Context, a *entity.CustomAsset) error
	FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.CustomAsset, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.CustomAsset, error)
	Update(ctx context.Context, a *entity.CustomAsset) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}
