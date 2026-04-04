package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
)

// Sentinel errors for investments
var (
	ErrPortfolioNotFound          = errors.New("portfolio not found")
	ErrHoldingNotFound            = errors.New("holding not found")
	ErrInvestmentTransactionNotFound = errors.New("investment transaction not found")
	ErrAssetNotFound              = errors.New("asset not found")
	ErrCustomAssetNotFound        = errors.New("custom asset not found")
)

// ---- Request types ----

type CreatePortfolioRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=255"`
	Description *string `json:"description"`
	IsDefault   bool    `json:"is_default"`
}

type UpdatePortfolioRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsDefault   *bool   `json:"is_default"`
}

type CreateHoldingRequest struct {
	AssetID *uuid.UUID `json:"asset_id"`
	Name    string     `json:"name" binding:"required"`
	Type    string     `json:"type" binding:"required"`
}

type UpdateHoldingRequest struct {
	Name *string `json:"name"`
	Type *string `json:"type"`
}

type CreateInvestmentTransactionRequest struct {
	Type     string    `json:"type" binding:"required,oneof=buy sell dividend split bonus"`
	Quantity *float64  `json:"quantity"`
	Price    *float64  `json:"price"`
	Fees     float64   `json:"fees"`
	Total    float64   `json:"total" binding:"required"`
	Date     time.Time `json:"date" binding:"required"`
	Notes    *string   `json:"notes"`
}

type CreateCustomAssetRequest struct {
	Name          string     `json:"name" binding:"required,min=2,max=255"`
	Type          string     `json:"type" binding:"required"`
	CurrentValue  float64    `json:"current_value" binding:"required"`
	PurchaseValue *float64   `json:"purchase_value"`
	PurchaseDate  *time.Time `json:"purchase_date"`
	MonthlyIncome float64    `json:"monthly_income"`
	Description   *string    `json:"description"`
}

type UpdateCustomAssetRequest struct {
	Name          *string    `json:"name"`
	Type          *string    `json:"type"`
	CurrentValue  *float64   `json:"current_value"`
	PurchaseValue *float64   `json:"purchase_value"`
	PurchaseDate  *time.Time `json:"purchase_date"`
	MonthlyIncome *float64   `json:"monthly_income"`
	Description   *string    `json:"description"`
}

// ---- Interfaces ----

type InvestmentUseCase interface {
	// Portfolio
	CreatePortfolio(ctx context.Context, userID uuid.UUID, req CreatePortfolioRequest) (*entity.Portfolio, error)
	GetPortfolios(ctx context.Context, userID uuid.UUID) ([]*entity.Portfolio, error)
	UpdatePortfolio(ctx context.Context, id, userID uuid.UUID, req UpdatePortfolioRequest) (*entity.Portfolio, error)
	DeletePortfolio(ctx context.Context, id, userID uuid.UUID) error

	// Holdings
	CreateHolding(ctx context.Context, portfolioID uuid.UUID, req CreateHoldingRequest) (*entity.Holding, error)
	GetHoldings(ctx context.Context, portfolioID uuid.UUID) ([]*entity.Holding, error)
	UpdateHolding(ctx context.Context, id uuid.UUID, req UpdateHoldingRequest) (*entity.Holding, error)
	DeleteHolding(ctx context.Context, id uuid.UUID) error

	// Investment transactions
	CreateInvestmentTransaction(ctx context.Context, holdingID uuid.UUID, req CreateInvestmentTransactionRequest) (*entity.InvestmentTransaction, error)
	GetInvestmentTransactions(ctx context.Context, holdingID uuid.UUID) ([]*entity.InvestmentTransaction, error)
	DeleteInvestmentTransaction(ctx context.Context, id uuid.UUID) error

	// Assets
	SearchAssets(ctx context.Context, query string) ([]*entity.Asset, error)

	// Custom assets
	CreateCustomAsset(ctx context.Context, userID uuid.UUID, req CreateCustomAssetRequest) (*entity.CustomAsset, error)
	GetCustomAssets(ctx context.Context, userID uuid.UUID) ([]*entity.CustomAsset, error)
	UpdateCustomAsset(ctx context.Context, id, userID uuid.UUID, req UpdateCustomAssetRequest) (*entity.CustomAsset, error)
	DeleteCustomAsset(ctx context.Context, id, userID uuid.UUID) error
}

// ---- Implementation ----

type investmentUseCase struct {
	portfolioRepo    domainrepo.PortfolioRepository
	holdingRepo      domainrepo.HoldingRepository
	investTxRepo     domainrepo.InvestmentTransactionRepository
	assetRepo        domainrepo.AssetRepository
	customAssetRepo  domainrepo.CustomAssetRepository
}

// NewInvestmentUseCase creates a new InvestmentUseCase implementation.
func NewInvestmentUseCase(
	pr domainrepo.PortfolioRepository,
	hr domainrepo.HoldingRepository,
	itr domainrepo.InvestmentTransactionRepository,
	ar domainrepo.AssetRepository,
	car domainrepo.CustomAssetRepository,
) InvestmentUseCase {
	return &investmentUseCase{
		portfolioRepo:   pr,
		holdingRepo:     hr,
		investTxRepo:    itr,
		assetRepo:       ar,
		customAssetRepo: car,
	}
}

// ---- Portfolio ----

func (uc *investmentUseCase) CreatePortfolio(ctx context.Context, userID uuid.UUID, req CreatePortfolioRequest) (*entity.Portfolio, error) {
	now := time.Now()
	p := &entity.Portfolio{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		IsDefault:   req.IsDefault,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := uc.portfolioRepo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("investmentUseCase.CreatePortfolio: %w", err)
	}
	return p, nil
}

func (uc *investmentUseCase) GetPortfolios(ctx context.Context, userID uuid.UUID) ([]*entity.Portfolio, error) {
	portfolios, err := uc.portfolioRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("investmentUseCase.GetPortfolios: %w", err)
	}
	return portfolios, nil
}

func (uc *investmentUseCase) UpdatePortfolio(ctx context.Context, id, userID uuid.UUID, req UpdatePortfolioRequest) (*entity.Portfolio, error) {
	p, err := uc.portfolioRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("investmentUseCase.UpdatePortfolio find: %w", err)
	}
	if p == nil {
		return nil, ErrPortfolioNotFound
	}

	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = req.Description
	}
	if req.IsDefault != nil {
		p.IsDefault = *req.IsDefault
	}

	if err := uc.portfolioRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("investmentUseCase.UpdatePortfolio save: %w", err)
	}
	return p, nil
}

func (uc *investmentUseCase) DeletePortfolio(ctx context.Context, id, userID uuid.UUID) error {
	p, err := uc.portfolioRepo.FindByID(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("investmentUseCase.DeletePortfolio find: %w", err)
	}
	if p == nil {
		return ErrPortfolioNotFound
	}
	if err := uc.portfolioRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("investmentUseCase.DeletePortfolio delete: %w", err)
	}
	return nil
}

// ---- Holdings ----

func (uc *investmentUseCase) CreateHolding(ctx context.Context, portfolioID uuid.UUID, req CreateHoldingRequest) (*entity.Holding, error) {
	now := time.Now()
	h := &entity.Holding{
		ID:          uuid.New(),
		PortfolioID: portfolioID,
		AssetID:     req.AssetID,
		Name:        req.Name,
		Type:        req.Type,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := uc.holdingRepo.Create(ctx, h); err != nil {
		return nil, fmt.Errorf("investmentUseCase.CreateHolding: %w", err)
	}
	return h, nil
}

func (uc *investmentUseCase) GetHoldings(ctx context.Context, portfolioID uuid.UUID) ([]*entity.Holding, error) {
	holdings, err := uc.holdingRepo.FindByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("investmentUseCase.GetHoldings: %w", err)
	}
	return holdings, nil
}

func (uc *investmentUseCase) UpdateHolding(ctx context.Context, id uuid.UUID, req UpdateHoldingRequest) (*entity.Holding, error) {
	h, err := uc.holdingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("investmentUseCase.UpdateHolding find: %w", err)
	}
	if h == nil {
		return nil, ErrHoldingNotFound
	}
	if req.Name != nil {
		h.Name = *req.Name
	}
	if req.Type != nil {
		h.Type = *req.Type
	}
	if err := uc.holdingRepo.Update(ctx, h); err != nil {
		return nil, fmt.Errorf("investmentUseCase.UpdateHolding save: %w", err)
	}
	return h, nil
}

func (uc *investmentUseCase) DeleteHolding(ctx context.Context, id uuid.UUID) error {
	h, err := uc.holdingRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("investmentUseCase.DeleteHolding find: %w", err)
	}
	if h == nil {
		return ErrHoldingNotFound
	}
	if err := uc.holdingRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("investmentUseCase.DeleteHolding delete: %w", err)
	}
	return nil
}

// ---- Investment Transactions ----

func (uc *investmentUseCase) CreateInvestmentTransaction(ctx context.Context, holdingID uuid.UUID, req CreateInvestmentTransactionRequest) (*entity.InvestmentTransaction, error) {
	h, err := uc.holdingRepo.FindByID(ctx, holdingID)
	if err != nil {
		return nil, fmt.Errorf("investmentUseCase.CreateInvestmentTransaction find holding: %w", err)
	}
	if h == nil {
		return nil, ErrHoldingNotFound
	}

	t := &entity.InvestmentTransaction{
		ID:        uuid.New(),
		HoldingID: holdingID,
		Type:      req.Type,
		Quantity:  req.Quantity,
		Price:     req.Price,
		Fees:      req.Fees,
		Total:     req.Total,
		Date:      req.Date,
		Notes:     req.Notes,
		CreatedAt: time.Now(),
	}

	if err := uc.investTxRepo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("investmentUseCase.CreateInvestmentTransaction create tx: %w", err)
	}

	// Recalculate holding position
	qty := 0.0
	if req.Quantity != nil {
		qty = *req.Quantity
	}
	price := 0.0
	if req.Price != nil {
		price = *req.Price
	}
	RecalcHolding(h, req.Type, qty, price, req.Fees)

	if err := uc.holdingRepo.Update(ctx, h); err != nil {
		return nil, fmt.Errorf("investmentUseCase.CreateInvestmentTransaction update holding: %w", err)
	}

	return t, nil
}

func (uc *investmentUseCase) GetInvestmentTransactions(ctx context.Context, holdingID uuid.UUID) ([]*entity.InvestmentTransaction, error) {
	txs, err := uc.investTxRepo.FindByHoldingID(ctx, holdingID)
	if err != nil {
		return nil, fmt.Errorf("investmentUseCase.GetInvestmentTransactions: %w", err)
	}
	return txs, nil
}

func (uc *investmentUseCase) DeleteInvestmentTransaction(ctx context.Context, id uuid.UUID) error {
	if err := uc.investTxRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("investmentUseCase.DeleteInvestmentTransaction: %w", err)
	}
	return nil
}

// ---- Assets ----

func (uc *investmentUseCase) SearchAssets(ctx context.Context, query string) ([]*entity.Asset, error) {
	assets, err := uc.assetRepo.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("investmentUseCase.SearchAssets: %w", err)
	}
	if assets == nil {
		assets = []*entity.Asset{}
	}
	return assets, nil
}

// ---- Custom Assets ----

func (uc *investmentUseCase) CreateCustomAsset(ctx context.Context, userID uuid.UUID, req CreateCustomAssetRequest) (*entity.CustomAsset, error) {
	now := time.Now()
	a := &entity.CustomAsset{
		ID:            uuid.New(),
		UserID:        userID,
		Name:          req.Name,
		Type:          req.Type,
		CurrentValue:  req.CurrentValue,
		PurchaseValue: req.PurchaseValue,
		PurchaseDate:  req.PurchaseDate,
		MonthlyIncome: req.MonthlyIncome,
		Description:   req.Description,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := uc.customAssetRepo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("investmentUseCase.CreateCustomAsset: %w", err)
	}
	return a, nil
}

func (uc *investmentUseCase) GetCustomAssets(ctx context.Context, userID uuid.UUID) ([]*entity.CustomAsset, error) {
	assets, err := uc.customAssetRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("investmentUseCase.GetCustomAssets: %w", err)
	}
	return assets, nil
}

func (uc *investmentUseCase) UpdateCustomAsset(ctx context.Context, id, userID uuid.UUID, req UpdateCustomAssetRequest) (*entity.CustomAsset, error) {
	a, err := uc.customAssetRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("investmentUseCase.UpdateCustomAsset find: %w", err)
	}
	if a == nil {
		return nil, ErrCustomAssetNotFound
	}

	if req.Name != nil {
		a.Name = *req.Name
	}
	if req.Type != nil {
		a.Type = *req.Type
	}
	if req.CurrentValue != nil {
		a.CurrentValue = *req.CurrentValue
	}
	if req.PurchaseValue != nil {
		a.PurchaseValue = req.PurchaseValue
	}
	if req.PurchaseDate != nil {
		a.PurchaseDate = req.PurchaseDate
	}
	if req.MonthlyIncome != nil {
		a.MonthlyIncome = *req.MonthlyIncome
	}
	if req.Description != nil {
		a.Description = req.Description
	}

	if err := uc.customAssetRepo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("investmentUseCase.UpdateCustomAsset save: %w", err)
	}
	return a, nil
}

func (uc *investmentUseCase) DeleteCustomAsset(ctx context.Context, id, userID uuid.UUID) error {
	a, err := uc.customAssetRepo.FindByID(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("investmentUseCase.DeleteCustomAsset find: %w", err)
	}
	if a == nil {
		return ErrCustomAssetNotFound
	}
	if err := uc.customAssetRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("investmentUseCase.DeleteCustomAsset delete: %w", err)
	}
	return nil
}
