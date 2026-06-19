package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/financeos/api/pkg/brapi"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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
	AssetID  *uuid.UUID `json:"asset_id"`
	Name     string     `json:"name" binding:"required"`
	Type     string     `json:"type" binding:"required"`
	Ticker   string     `json:"ticker"`
	Exchange string     `json:"exchange"`
}

type UpdateHoldingRequest struct {
	Name *string `json:"name"`
	Type *string `json:"type"`
}

type CreateInvestmentTransactionRequest struct {
	Type      string     `json:"type" binding:"required,oneof=buy sell dividend split bonus"`
	Quantity  *float64   `json:"quantity"`
	Price     *float64   `json:"price"`
	Fees      float64    `json:"fees"`
	Total     float64    `json:"total" binding:"required"`
	Date      time.Time  `json:"date" binding:"required"`
	Notes     *string    `json:"notes"`
	AccountID *uuid.UUID `json:"account_id"` // if dividend: credit this account as income
	UserID    uuid.UUID  `json:"-"`           // set by handler from JWT context
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

// ---- BrapiSearcher interface (allows mocking in tests) ----

// BrapiSearcher is the interface for fetching asset data from BRAPI.
type BrapiSearcher interface {
	Search(ctx context.Context, query string) ([]brapi.AssetResult, error)
	FetchPrice(ctx context.Context, ticker string) (float64, error)
	FetchAvailableTickers(ctx context.Context) ([]string, error)
	SearchByQuery(ctx context.Context, query string, allTickers []string) ([]brapi.AssetResult, error)
}

// ---- Interfaces ----

type InvestmentUseCase interface {
	// Dependency injection (called once at startup)
	WithTransactionRepo(tr domainrepo.TransactionRepository)
	WithNotificationUC(n NotificationUseCase)

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

	// Performance
	GetPortfolioPerformance(ctx context.Context, userID uuid.UUID) (*PortfolioPerformance, error)

	// Tax report
	GetTaxReport(ctx context.Context, userID uuid.UUID, year int) (*TaxReport, error)

	// Assets
	SearchAssets(ctx context.Context, query string) ([]*entity.Asset, error)

	// Custom assets
	CreateCustomAsset(ctx context.Context, userID uuid.UUID, req CreateCustomAssetRequest) (*entity.CustomAsset, error)
	GetCustomAssets(ctx context.Context, userID uuid.UUID) ([]*entity.CustomAsset, error)
	UpdateCustomAsset(ctx context.Context, id, userID uuid.UUID, req UpdateCustomAssetRequest) (*entity.CustomAsset, error)
	DeleteCustomAsset(ctx context.Context, id, userID uuid.UUID) error
}

// ---- Implementation ----

// ConcentrationWarning is returned when a single holding exceeds 30% of portfolio value.
type ConcentrationWarning struct {
	HoldingName string
	Pct         float64
}

type investmentUseCase struct {
	portfolioRepo   domainrepo.PortfolioRepository
	holdingRepo     domainrepo.HoldingRepository
	investTxRepo    domainrepo.InvestmentTransactionRepository
	assetRepo       domainrepo.AssetRepository
	customAssetRepo domainrepo.CustomAssetRepository
	transactionRepo domainrepo.TransactionRepository
	notificationUC  NotificationUseCase
	brapiSvc        BrapiSearcher
	cache           *redis.Client
}

// NewInvestmentUseCase creates a new InvestmentUseCase implementation.
func NewInvestmentUseCase(
	pr domainrepo.PortfolioRepository,
	hr domainrepo.HoldingRepository,
	itr domainrepo.InvestmentTransactionRepository,
	ar domainrepo.AssetRepository,
	car domainrepo.CustomAssetRepository,
	brapiSvc BrapiSearcher,
	cache *redis.Client,
) InvestmentUseCase {
	return &investmentUseCase{
		portfolioRepo:   pr,
		holdingRepo:     hr,
		investTxRepo:    itr,
		assetRepo:       ar,
		customAssetRepo: car,
		brapiSvc:        brapiSvc,
		cache:           cache,
	}
}

// WithTransactionRepo injects the transaction repository for dividend → income auto-posting.
func (uc *investmentUseCase) WithTransactionRepo(tr domainrepo.TransactionRepository) {
	uc.transactionRepo = tr
}

// WithNotificationUC injects the notification use case for concentration alerts.
func (uc *investmentUseCase) WithNotificationUC(n NotificationUseCase) {
	uc.notificationUC = n
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
	assetID := req.AssetID

	// If ticker and exchange provided but no assetID, upsert asset in the DB
	if req.Ticker != "" && req.Exchange != "" && assetID == nil {
		existing, findErr := uc.assetRepo.FindByTicker(ctx, req.Ticker, req.Exchange)
		if findErr != nil {
			// Non-fatal — proceed without linking asset
			existing = nil
		}
		if existing != nil {
			assetID = &existing.ID
		} else if uc.brapiSvc != nil {
			// Try to fetch from BRAPI and persist
			results, brapiErr := uc.brapiSvc.Search(ctx, req.Ticker)
			if brapiErr == nil && len(results) > 0 {
				r := results[0]
				ticker := r.Ticker
				exchange := r.Exchange
				price := r.CurrentPrice
				newAsset := &entity.Asset{
					ID:           uuid.New(),
					Ticker:       &ticker,
					Name:         r.Name,
					Type:         r.Type,
					Exchange:     &exchange,
					Currency:     r.Currency,
					CurrentPrice: &price,
					CreatedAt:    now,
					UpdatedAt:    now,
				}
				if createErr := uc.assetRepo.Create(ctx, newAsset); createErr == nil {
					assetID = &newAsset.ID
				}
			}
		}
	}

	h := &entity.Holding{
		ID:          uuid.New(),
		PortfolioID: portfolioID,
		AssetID:     assetID,
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

	// D3 — auto-post dividend as income in the financial account
	if req.Type == "dividend" && req.AccountID != nil && uc.transactionRepo != nil {
		portfolio, err := uc.portfolioRepo.FindByID(ctx, h.PortfolioID, req.UserID)
		if err == nil && portfolio != nil {
			desc := fmt.Sprintf("Dividendo: %s", h.Name)
			financialTx := &entity.Transaction{
				ID:        uuid.New(),
				UserID:    portfolio.UserID,
				AccountID: *req.AccountID,
				Type:      "income",
				Amount:    req.Total,
				Date:      req.Date,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			financialTx.Description = &desc
			// Best-effort: don't fail the investment transaction if this fails
			_ = uc.transactionRepo.Create(ctx, financialTx)
		}
	}

	// D6 — check portfolio concentration and notify if any holding > 30%
	if uc.notificationUC != nil {
		go uc.checkConcentration(ctx, h.PortfolioID)
	}

	return t, nil
}

func (uc *investmentUseCase) checkConcentration(ctx context.Context, portfolioID uuid.UUID) {
	holdings, err := uc.holdingRepo.FindByPortfolioID(ctx, portfolioID)
	if err != nil || len(holdings) == 0 {
		return
	}
	portfolio, err := uc.portfolioRepo.FindByID(ctx, portfolioID, uuid.Nil)
	if err != nil || portfolio == nil {
		return
	}

	var total float64
	for _, h := range holdings {
		total += h.CurrentValue
	}
	if total == 0 {
		return
	}

	for _, h := range holdings {
		pct := (h.CurrentValue / total) * 100
		if pct > 30 {
			msg := fmt.Sprintf("%s representa %.1f%% do seu portfólio. Considere rebalancear.", h.Name, pct)
			_, _ = uc.notificationUC.Create(ctx, portfolio.UserID, "concentration_alert",
				"Concentração elevada na carteira", msg, map[string]interface{}{
					"holding_id": h.ID,
					"pct":        pct,
				})
		}
	}
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
	// 1. Try database first (assets already used by this user)
	dbAssets, err := uc.assetRepo.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("investmentUseCase.SearchAssets db: %w", err)
	}
	if dbAssets == nil {
		dbAssets = []*entity.Asset{}
	}
	if len(dbAssets) >= 5 {
		return dbAssets, nil
	}

	if uc.brapiSvc == nil || uc.cache == nil {
		return dbAssets, nil
	}

	// 2. Load full B3 ticker list from Redis (cached 24h)
	const tickersCacheKey = "brapi:tickers:all"
	var allTickers []string

	if raw, cErr := uc.cache.Get(ctx, tickersCacheKey).Bytes(); cErr == nil {
		_ = json.Unmarshal(raw, &allTickers)
	}

	if len(allTickers) == 0 {
		tickers, fetchErr := uc.brapiSvc.FetchAvailableTickers(ctx)
		if fetchErr != nil {
			// BRAPI unavailable — fallback to DB results
			return dbAssets, nil
		}
		allTickers = tickers
		if data, mErr := json.Marshal(tickers); mErr == nil {
			_ = uc.cache.Set(ctx, tickersCacheKey, data, 24*time.Hour).Err()
		}
	}

	// 3. Check per-query cache (1h TTL)
	cacheKey := "asset_search:" + strings.ToLower(query)
	var brapiAssets []*entity.Asset

	if raw, cErr := uc.cache.Get(ctx, cacheKey).Bytes(); cErr == nil {
		_ = json.Unmarshal(raw, &brapiAssets)
	}

	// 4. Cache miss — filter locally + batch-fetch prices from BRAPI
	if len(brapiAssets) == 0 {
		brapiResults, brapiErr := uc.brapiSvc.SearchByQuery(ctx, query, allTickers)
		if brapiErr != nil {
			// BRAPI unavailable — fallback to DB results
			return dbAssets, nil
		}

		now := time.Now()
		brapiAssets = make([]*entity.Asset, 0, len(brapiResults))
		for _, r := range brapiResults {
			ticker := r.Ticker
			exchange := r.Exchange
			price := r.CurrentPrice
			brapiAssets = append(brapiAssets, &entity.Asset{
				ID:           uuid.New(),
				Ticker:       &ticker,
				Name:         r.Name,
				Type:         r.Type,
				Exchange:     &exchange,
				Currency:     r.Currency,
				CurrentPrice: &price,
				CreatedAt:    now,
				UpdatedAt:    now,
			})
		}

		if len(brapiAssets) > 0 {
			if data, mErr := json.Marshal(brapiAssets); mErr == nil {
				_ = uc.cache.Set(ctx, cacheKey, data, time.Hour).Err()
			}
		}
	}

	// 5. Merge DB + BRAPI, dedup by ticker|exchange, limit to 20
	seen := make(map[string]struct{})
	merged := make([]*entity.Asset, 0, len(dbAssets)+len(brapiAssets))

	for _, a := range dbAssets {
		key := ""
		if a.Ticker != nil {
			key = *a.Ticker
		}
		key += "|"
		if a.Exchange != nil {
			key += *a.Exchange
		}
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			merged = append(merged, a)
		}
	}

	for _, a := range brapiAssets {
		key := ""
		if a.Ticker != nil {
			key = *a.Ticker
		}
		key += "|"
		if a.Exchange != nil {
			key += *a.Exchange
		}
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			merged = append(merged, a)
		}
	}

	if len(merged) > 20 {
		merged = merged[:20]
	}
	return merged, nil
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

// ---- Portfolio Performance (D7) ----

// PortfolioPerformance holds the return metrics for a user's portfolio.
type PortfolioPerformance struct {
	TotalInvested    float64 `json:"total_invested"`
	CurrentValue     float64 `json:"current_value"`
	ReturnPct        float64 `json:"return_pct"`
	ReturnAmount     float64 `json:"return_amount"`
	CDIEstimatePct   float64 `json:"cdi_estimate_pct"`
	IBOVEstimatePct  float64 `json:"ibov_estimate_pct"`
	Period           string  `json:"period"`
}

func (uc *investmentUseCase) GetPortfolioPerformance(ctx context.Context, userID uuid.UUID) (*PortfolioPerformance, error) {
	portfolios, err := uc.portfolioRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("investmentUseCase.GetPortfolioPerformance portfolios: %w", err)
	}

	var totalInvested, currentValue float64
	for _, p := range portfolios {
		holdings, err := uc.holdingRepo.FindByPortfolioID(ctx, p.ID)
		if err != nil {
			continue
		}
		for _, h := range holdings {
			totalInvested += h.TotalInvested
			currentValue += h.CurrentValue
		}
	}

	perf := &PortfolioPerformance{
		TotalInvested: totalInvested,
		CurrentValue:  currentValue,
		Period:        "total",
		// CDI and IBOV estimates — representative rates updated manually each quarter.
		// Replace with live BACEN/BRAPI fetch when real-time benchmark is needed.
		CDIEstimatePct:  10.75,
		IBOVEstimatePct: 8.50,
	}
	if totalInvested > 0 {
		perf.ReturnAmount = currentValue - totalInvested
		perf.ReturnPct = (perf.ReturnAmount / totalInvested) * 100
	}
	return perf, nil
}

// ---- Tax Report (D5) ----

type TaxReportEntry struct {
	Month           int     `json:"month"`
	Year            int     `json:"year"`
	Label           string  `json:"label"`
	AssetType       string  `json:"asset_type"`
	GrossProfit     float64 `json:"gross_profit"`
	TaxRate         float64 `json:"tax_rate"`
	TaxDue          float64 `json:"tax_due"`
	SaleTotal       float64 `json:"sale_total"`
	IsExempt        bool    `json:"is_exempt"`
	ExemptionReason string  `json:"exemption_reason,omitempty"`
}

type TaxReport struct {
	Year        int              `json:"year"`
	Entries     []TaxReportEntry `json:"entries"`
	TotalProfit float64          `json:"total_profit"`
	TotalTaxDue float64          `json:"total_tax_due"`
}

// brazilianTaxRate returns the applicable IR rate and exemption info for a given asset type
// and monthly sell total (used for the R$20k stock exemption rule).
func brazilianTaxRate(assetType string, monthlySaleTotal float64) (rate float64, exempt bool, reason string) {
	switch assetType {
	case "stock", "etf":
		if monthlySaleTotal <= 20000 {
			return 0, true, "Vendas totais no mês ≤ R$20.000 (isento para PF)"
		}
		return 0.15, false, ""
	case "fii":
		return 0.20, false, "" // FII: 20%, sem isenção de R$20k
	case "crypto":
		if monthlySaleTotal <= 35000 {
			return 0, true, "Vendas totais no mês ≤ R$35.000 (isento para PF)"
		}
		return 0.15, false, "" // simplificado: alíquota base cripto
	case "fixed_income", "fund":
		return 0.15, false, "" // IR regressivo — tabela simplificada (alíquota mínima)
	default:
		return 0.15, false, ""
	}
}

func (uc *investmentUseCase) GetTaxReport(ctx context.Context, userID uuid.UUID, year int) (*TaxReport, error) {
	portfolios, err := uc.portfolioRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("investmentUseCase.GetTaxReport portfolios: %w", err)
	}

	monthNames := []string{"Jan", "Fev", "Mar", "Abr", "Mai", "Jun", "Jul", "Ago", "Set", "Out", "Nov", "Dez"}

	// month+type → (saleTotal, grossProfit, holdingType)
	type monthTypeKey struct {
		Month int
		Type  string
	}
	type aggregated struct {
		SaleTotal   float64
		GrossProfit float64
	}
	agg := map[monthTypeKey]*aggregated{}

	for _, p := range portfolios {
		holdings, err := uc.holdingRepo.FindByPortfolioID(ctx, p.ID)
		if err != nil {
			continue
		}
		for _, h := range holdings {
			txs, err := uc.investTxRepo.FindByHoldingID(ctx, h.ID)
			if err != nil {
				continue
			}
			for _, tx := range txs {
				if tx.Date.Year() != year || tx.Type != "sell" {
					continue
				}
				// Approximate PnL: (sale_price - avg_price) * qty
				var pnl float64
				if tx.Quantity != nil && tx.Price != nil && h.AvgPrice > 0 {
					pnl = (*tx.Price - h.AvgPrice) * (*tx.Quantity)
				}
				key := monthTypeKey{int(tx.Date.Month()), h.Type}
				if agg[key] == nil {
					agg[key] = &aggregated{}
				}
				agg[key].SaleTotal += tx.Total
				if pnl > 0 {
					agg[key].GrossProfit += pnl
				}
			}
		}
	}

	report := &TaxReport{Year: year}
	for key, data := range agg {
		rate, exempt, reason := brazilianTaxRate(key.Type, data.SaleTotal)
		taxDue := 0.0
		if !exempt && data.GrossProfit > 0 {
			taxDue = data.GrossProfit * rate
		}
		report.TotalProfit += data.GrossProfit
		report.TotalTaxDue += taxDue
		report.Entries = append(report.Entries, TaxReportEntry{
			Month:           key.Month,
			Year:            year,
			Label:           fmt.Sprintf("%s/%d", monthNames[key.Month-1], year),
			AssetType:       key.Type,
			GrossProfit:     data.GrossProfit,
			TaxRate:         rate * 100,
			TaxDue:          taxDue,
			SaleTotal:       data.SaleTotal,
			IsExempt:        exempt,
			ExemptionReason: reason,
		})
	}
	if report.Entries == nil {
		report.Entries = []TaxReportEntry{}
	}
	return report, nil
}
