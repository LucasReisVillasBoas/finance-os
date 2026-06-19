package usecase

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Fake repositories ----

type fakePortfolioRepo struct {
	mu         sync.Mutex
	portfolios map[uuid.UUID]*entity.Portfolio
}

func newFakePortfolioRepo() *fakePortfolioRepo {
	return &fakePortfolioRepo{portfolios: make(map[uuid.UUID]*entity.Portfolio)}
}

func (r *fakePortfolioRepo) Create(_ context.Context, p *entity.Portfolio) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.portfolios[p.ID] = p
	return nil
}

func (r *fakePortfolioRepo) FindByID(_ context.Context, id, _ uuid.UUID) (*entity.Portfolio, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.portfolios[id]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (r *fakePortfolioRepo) FindByUserID(_ context.Context, userID uuid.UUID) ([]*entity.Portfolio, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*entity.Portfolio
	for _, p := range r.portfolios {
		if p.UserID == userID {
			cp := *p
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *fakePortfolioRepo) Update(_ context.Context, p *entity.Portfolio) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.portfolios[p.ID] = p
	return nil
}

func (r *fakePortfolioRepo) Delete(_ context.Context, id, _ uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.portfolios, id)
	return nil
}

type fakeHoldingRepoInv struct {
	mu   sync.Mutex
	byID map[uuid.UUID]*entity.Holding
}

func newFakeHoldingRepoInv() *fakeHoldingRepoInv {
	return &fakeHoldingRepoInv{byID: make(map[uuid.UUID]*entity.Holding)}
}

func (r *fakeHoldingRepoInv) Create(_ context.Context, h *entity.Holding) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[h.ID] = h
	return nil
}

func (r *fakeHoldingRepoInv) FindByID(_ context.Context, id uuid.UUID) (*entity.Holding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	h, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	cp := *h
	return &cp, nil
}

func (r *fakeHoldingRepoInv) FindByPortfolioID(_ context.Context, portfolioID uuid.UUID) ([]*entity.Holding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*entity.Holding
	for _, h := range r.byID {
		if h.PortfolioID == portfolioID {
			cp := *h
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *fakeHoldingRepoInv) Update(_ context.Context, h *entity.Holding) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[h.ID] = h
	return nil
}

func (r *fakeHoldingRepoInv) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byID, id)
	return nil
}

func (r *fakeHoldingRepoInv) FindAll(_ context.Context) ([]*entity.Holding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*entity.Holding
	for _, h := range r.byID {
		cp := *h
		result = append(result, &cp)
	}
	return result, nil
}

type fakeInvestTxRepoInv struct {
	mu        sync.Mutex
	txs       []*entity.InvestmentTransaction
	byHolding map[uuid.UUID][]*entity.InvestmentTransaction
}

func newFakeInvestTxRepoInv() *fakeInvestTxRepoInv {
	return &fakeInvestTxRepoInv{byHolding: make(map[uuid.UUID][]*entity.InvestmentTransaction)}
}

func (r *fakeInvestTxRepoInv) Create(_ context.Context, t *entity.InvestmentTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.txs = append(r.txs, t)
	r.byHolding[t.HoldingID] = append(r.byHolding[t.HoldingID], t)
	return nil
}

func (r *fakeInvestTxRepoInv) FindByHoldingID(_ context.Context, holdingID uuid.UUID) ([]*entity.InvestmentTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.byHolding[holdingID], nil
}

func (r *fakeInvestTxRepoInv) Delete(_ context.Context, _ uuid.UUID) error { return nil }

type fakeAssetRepoInv struct{}

func (r *fakeAssetRepoInv) Create(_ context.Context, _ *entity.Asset) error               { return nil }
func (r *fakeAssetRepoInv) FindByID(_ context.Context, _ uuid.UUID) (*entity.Asset, error) { return nil, nil }
func (r *fakeAssetRepoInv) FindByTicker(_ context.Context, _, _ string) (*entity.Asset, error) {
	return nil, nil
}
func (r *fakeAssetRepoInv) Search(_ context.Context, _ string) ([]*entity.Asset, error) {
	return []*entity.Asset{}, nil
}
func (r *fakeAssetRepoInv) UpdatePrice(_ context.Context, _ uuid.UUID, _ float64) error { return nil }
func (r *fakeAssetRepoInv) FindAll(_ context.Context) ([]*entity.Asset, error) {
	return []*entity.Asset{}, nil
}

type fakeCustomAssetRepoInv struct{}

func (r *fakeCustomAssetRepoInv) Create(_ context.Context, _ *entity.CustomAsset) error { return nil }
func (r *fakeCustomAssetRepoInv) FindByID(_ context.Context, _, _ uuid.UUID) (*entity.CustomAsset, error) {
	return nil, nil
}
func (r *fakeCustomAssetRepoInv) FindByUserID(_ context.Context, _ uuid.UUID) ([]*entity.CustomAsset, error) {
	return []*entity.CustomAsset{}, nil
}
func (r *fakeCustomAssetRepoInv) Update(_ context.Context, _ *entity.CustomAsset) error { return nil }
func (r *fakeCustomAssetRepoInv) Delete(_ context.Context, _, _ uuid.UUID) error        { return nil }

type fakeTransactionRepoInv struct {
	mu      sync.Mutex
	created []*entity.Transaction
}

func (r *fakeTransactionRepoInv) Create(_ context.Context, tx *entity.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.created = append(r.created, tx)
	return nil
}

func (r *fakeTransactionRepoInv) FindByID(_ context.Context, _, _ uuid.UUID) (*entity.Transaction, error) {
	return nil, nil
}
func (r *fakeTransactionRepoInv) FindByUserID(_ context.Context, _ uuid.UUID, _ domainrepo.TransactionFilter) ([]*entity.Transaction, int, error) {
	return nil, 0, nil
}
func (r *fakeTransactionRepoInv) Update(_ context.Context, _ *entity.Transaction) error { return nil }
func (r *fakeTransactionRepoInv) Delete(_ context.Context, _, _ uuid.UUID) error        { return nil }
func (r *fakeTransactionRepoInv) GetSummary(_ context.Context, _ uuid.UUID, _, _ time.Time) (*domainrepo.TransactionSummary, error) {
	return nil, nil
}
func (r *fakeTransactionRepoInv) CreateTransfer(_ context.Context, _, _ *entity.Transaction) error {
	return nil
}
func (r *fakeTransactionRepoInv) UpdateAccountBalance(_ context.Context, _ uuid.UUID, _ float64) error {
	return nil
}

type fakeNotificationUCInv struct {
	mu      sync.Mutex
	created []*entity.Notification
}

func (f *fakeNotificationUCInv) Create(_ context.Context, userID uuid.UUID, notifType, title, _ string, _ map[string]interface{}) (*entity.Notification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := &entity.Notification{ID: uuid.New(), UserID: userID, Type: notifType, Title: title}
	f.created = append(f.created, n)
	return n, nil
}
func (f *fakeNotificationUCInv) List(_ context.Context, _ uuid.UUID, _ bool) ([]*entity.Notification, error) {
	return nil, nil
}
func (f *fakeNotificationUCInv) MarkAsRead(_ context.Context, _, _ uuid.UUID) error  { return nil }
func (f *fakeNotificationUCInv) MarkAllAsRead(_ context.Context, _ uuid.UUID) error  { return nil }
func (f *fakeNotificationUCInv) DeleteAll(_ context.Context, _ uuid.UUID) error      { return nil }
func (f *fakeNotificationUCInv) CountUnread(_ context.Context, _ uuid.UUID) (int, error) { return 0, nil }

// helper to build a bare-bones InvestmentUseCase with no Redis/BRAPI
func newTestInvestmentUseCase(
	portfolioRepo domainrepo.PortfolioRepository,
	holdingRepo domainrepo.HoldingRepository,
	investTxRepo domainrepo.InvestmentTransactionRepository,
) InvestmentUseCase {
	return NewInvestmentUseCase(portfolioRepo, holdingRepo, investTxRepo, &fakeAssetRepoInv{}, &fakeCustomAssetRepoInv{}, nil, nil)
}

// ------------------------------------------------------------------ D5 ----
// brazilianTaxRate – table-driven coverage of all asset types and thresholds

func TestBrazilianTaxRate(t *testing.T) {
	tests := []struct {
		name       string
		assetType  string
		saleTotal  float64
		wantRate   float64
		wantExempt bool
	}{
		// stock / etf – exempt up to R$20 000, 15 % above
		{"stock exempt at boundary", "stock", 20000, 0, true},
		{"stock exempt below boundary", "stock", 15000, 0, true},
		{"stock taxable above boundary", "stock", 20000.01, 0.15, false},
		{"etf exempt", "etf", 10000, 0, true},
		{"etf taxable", "etf", 25000, 0.15, false},

		// fii – always 20 %, no exemption regardless of sale size
		{"fii small sale", "fii", 1000, 0.20, false},
		{"fii large sale", "fii", 500000, 0.20, false},

		// crypto – exempt up to R$35 000, 15 % above (simplified)
		{"crypto exempt at boundary", "crypto", 35000, 0, true},
		{"crypto exempt below boundary", "crypto", 20000, 0, true},
		{"crypto taxable above boundary", "crypto", 35000.01, 0.15, false},

		// fixed_income / fund – 15 % base rate (regressivo simplified)
		{"fixed_income", "fixed_income", 5000, 0.15, false},
		{"fund", "fund", 5000, 0.15, false},

		// unknown defaults to 15 %
		{"unknown type", "debenture", 1000, 0.15, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, exempt, _ := brazilianTaxRate(tt.assetType, tt.saleTotal)
			assert.Equal(t, tt.wantRate, rate, "rate mismatch")
			assert.Equal(t, tt.wantExempt, exempt, "exempt mismatch")
		})
	}
}

// GetTaxReport – stock sale > R$20k → taxable at 15 %
func TestGetTaxReport_StockSaleTaxable(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	holdingID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	qty := 200.0
	salePrice := 150.0
	holding := &entity.Holding{
		ID:          holdingID,
		PortfolioID: portfolioID,
		Type:        "stock",
		AvgPrice:    10.0,
	}
	tx := &entity.InvestmentTransaction{
		ID:        uuid.New(),
		HoldingID: holdingID,
		Type:      "sell",
		Quantity:  &qty,
		Price:     &salePrice,
		Total:     30000.0, // > 20 k → taxable
		Date:      time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
	}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio

	holdingRepo := newFakeHoldingRepoInv()
	holdingRepo.byID[holdingID] = holding

	investTxRepo := newFakeInvestTxRepoInv()
	investTxRepo.byHolding[holdingID] = []*entity.InvestmentTransaction{tx}

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, investTxRepo)

	report, err := uc.GetTaxReport(context.Background(), userID, 2026)
	require.NoError(t, err)
	assert.Equal(t, 2026, report.Year)
	require.Len(t, report.Entries, 1)

	e := report.Entries[0]
	assert.Equal(t, "stock", e.AssetType)
	assert.Equal(t, 3, e.Month)
	assert.Equal(t, 30000.0, e.SaleTotal)
	assert.False(t, e.IsExempt)
	assert.InDelta(t, 15.0, e.TaxRate, 0.001)
	// profit = (150 – 10) * 200 = 28 000
	assert.InDelta(t, 28000.0, e.GrossProfit, 0.01)
	// tax   = 28 000 * 0.15 = 4 200
	assert.InDelta(t, 4200.0, e.TaxDue, 0.01)
	assert.InDelta(t, 28000.0, report.TotalProfit, 0.01)
	assert.InDelta(t, 4200.0, report.TotalTaxDue, 0.01)
}

// GetTaxReport – stock sale ≤ R$20k → exempt, zero tax due
func TestGetTaxReport_StockSaleExempt(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	holdingID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	qty := 10.0
	salePrice := 30.0
	holding := &entity.Holding{
		ID:          holdingID,
		PortfolioID: portfolioID,
		Type:        "stock",
		AvgPrice:    20.0,
	}
	tx := &entity.InvestmentTransaction{
		ID:        uuid.New(),
		HoldingID: holdingID,
		Type:      "sell",
		Quantity:  &qty,
		Price:     &salePrice,
		Total:     300.0, // well below 20 k
		Date:      time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
	}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio
	holdingRepo := newFakeHoldingRepoInv()
	holdingRepo.byID[holdingID] = holding
	investTxRepo := newFakeInvestTxRepoInv()
	investTxRepo.byHolding[holdingID] = []*entity.InvestmentTransaction{tx}

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, investTxRepo)

	report, err := uc.GetTaxReport(context.Background(), userID, 2026)
	require.NoError(t, err)
	require.Len(t, report.Entries, 1)

	e := report.Entries[0]
	assert.True(t, e.IsExempt)
	assert.Equal(t, 0.0, e.TaxDue)
	assert.Equal(t, 0.0, report.TotalTaxDue)
}

// GetTaxReport – FII: always 20 %, no exemption
func TestGetTaxReport_FIIAlways20Pct(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	holdingID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	qty := 5.0
	salePrice := 200.0
	holding := &entity.Holding{
		ID:          holdingID,
		PortfolioID: portfolioID,
		Type:        "fii",
		AvgPrice:    100.0,
	}
	tx := &entity.InvestmentTransaction{
		ID:        uuid.New(),
		HoldingID: holdingID,
		Type:      "sell",
		Quantity:  &qty,
		Price:     &salePrice,
		Total:     1000.0,
		Date:      time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
	}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio
	holdingRepo := newFakeHoldingRepoInv()
	holdingRepo.byID[holdingID] = holding
	investTxRepo := newFakeInvestTxRepoInv()
	investTxRepo.byHolding[holdingID] = []*entity.InvestmentTransaction{tx}

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, investTxRepo)

	report, err := uc.GetTaxReport(context.Background(), userID, 2026)
	require.NoError(t, err)
	require.Len(t, report.Entries, 1)

	e := report.Entries[0]
	assert.False(t, e.IsExempt)
	assert.InDelta(t, 20.0, e.TaxRate, 0.001)
	// profit = (200 – 100) * 5 = 500
	assert.InDelta(t, 500.0, e.GrossProfit, 0.01)
	// tax    = 500 * 0.20 = 100
	assert.InDelta(t, 100.0, e.TaxDue, 0.01)
}

// GetTaxReport – buy transactions are ignored (only sell events contribute)
func TestGetTaxReport_BuyTransactionsIgnored(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	holdingID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	qty := 10.0
	price := 50.0
	holding := &entity.Holding{ID: holdingID, PortfolioID: portfolioID, Type: "stock", AvgPrice: 30.0}
	tx := &entity.InvestmentTransaction{
		ID:        uuid.New(),
		HoldingID: holdingID,
		Type:      "buy",
		Quantity:  &qty,
		Price:     &price,
		Total:     500.0,
		Date:      time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
	}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio
	holdingRepo := newFakeHoldingRepoInv()
	holdingRepo.byID[holdingID] = holding
	investTxRepo := newFakeInvestTxRepoInv()
	investTxRepo.byHolding[holdingID] = []*entity.InvestmentTransaction{tx}

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, investTxRepo)

	report, err := uc.GetTaxReport(context.Background(), userID, 2026)
	require.NoError(t, err)
	assert.Empty(t, report.Entries)
	assert.Equal(t, 0.0, report.TotalProfit)
	assert.Equal(t, 0.0, report.TotalTaxDue)
}

// GetTaxReport – sells in a different year are excluded
func TestGetTaxReport_WrongYearExcluded(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	holdingID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	qty := 10.0
	price := 50.0
	holding := &entity.Holding{ID: holdingID, PortfolioID: portfolioID, Type: "stock", AvgPrice: 10.0}
	tx := &entity.InvestmentTransaction{
		ID:        uuid.New(),
		HoldingID: holdingID,
		Type:      "sell",
		Quantity:  &qty,
		Price:     &price,
		Total:     500.0,
		Date:      time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), // 2025 – not in 2026
	}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio
	holdingRepo := newFakeHoldingRepoInv()
	holdingRepo.byID[holdingID] = holding
	investTxRepo := newFakeInvestTxRepoInv()
	investTxRepo.byHolding[holdingID] = []*entity.InvestmentTransaction{tx}

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, investTxRepo)

	report, err := uc.GetTaxReport(context.Background(), userID, 2026)
	require.NoError(t, err)
	assert.Empty(t, report.Entries)
}

// ------------------------------------------------------------------ D7 ----
// GetPortfolioPerformance – benchmark rates always present, return % is correct

func TestGetPortfolioPerformance_WithGain(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	holdingID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	holding := &entity.Holding{
		ID:            holdingID,
		PortfolioID:   portfolioID,
		TotalInvested: 1000.0,
		CurrentValue:  1250.0,
	}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio
	holdingRepo := newFakeHoldingRepoInv()
	holdingRepo.byID[holdingID] = holding

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, newFakeInvestTxRepoInv())

	perf, err := uc.GetPortfolioPerformance(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 1000.0, perf.TotalInvested)
	assert.Equal(t, 1250.0, perf.CurrentValue)
	assert.InDelta(t, 250.0, perf.ReturnAmount, 0.01)
	assert.InDelta(t, 25.0, perf.ReturnPct, 0.001) // 250/1000*100
	// D7: benchmark estimates must be populated
	assert.Equal(t, 10.75, perf.CDIEstimatePct)
	assert.Equal(t, 8.50, perf.IBOVEstimatePct)
	assert.Equal(t, "total", perf.Period)
}

func TestGetPortfolioPerformance_WithLoss(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	holdingID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	holding := &entity.Holding{
		ID:            holdingID,
		PortfolioID:   portfolioID,
		TotalInvested: 2000.0,
		CurrentValue:  1600.0,
	}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio
	holdingRepo := newFakeHoldingRepoInv()
	holdingRepo.byID[holdingID] = holding

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, newFakeInvestTxRepoInv())

	perf, err := uc.GetPortfolioPerformance(context.Background(), userID)
	require.NoError(t, err)
	assert.InDelta(t, -400.0, perf.ReturnAmount, 0.01)
	assert.InDelta(t, -20.0, perf.ReturnPct, 0.001)
}

func TestGetPortfolioPerformance_MultipleHoldings(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	h1 := &entity.Holding{ID: uuid.New(), PortfolioID: portfolioID, TotalInvested: 500, CurrentValue: 600}
	h2 := &entity.Holding{ID: uuid.New(), PortfolioID: portfolioID, TotalInvested: 500, CurrentValue: 400}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio
	holdingRepo := newFakeHoldingRepoInv()
	holdingRepo.byID[h1.ID] = h1
	holdingRepo.byID[h2.ID] = h2

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, newFakeInvestTxRepoInv())

	perf, err := uc.GetPortfolioPerformance(context.Background(), userID)
	require.NoError(t, err)
	// total invested = 1000, current = 1000, return = 0 %
	assert.Equal(t, 1000.0, perf.TotalInvested)
	assert.Equal(t, 1000.0, perf.CurrentValue)
	assert.InDelta(t, 0.0, perf.ReturnPct, 0.001)
}

func TestGetPortfolioPerformance_NoHoldings(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio

	uc := newTestInvestmentUseCase(portfolioRepo, newFakeHoldingRepoInv(), newFakeInvestTxRepoInv())

	perf, err := uc.GetPortfolioPerformance(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 0.0, perf.TotalInvested)
	assert.Equal(t, 0.0, perf.CurrentValue)
	assert.Equal(t, 0.0, perf.ReturnAmount)
	assert.Equal(t, 0.0, perf.ReturnPct)
}

// ------------------------------------------------------------------ D3 ----
// CreateInvestmentTransaction – dividend auto-posts income to the financial account

func TestCreateInvestmentTransaction_Dividend_AutoPostsIncome(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	holdingID := uuid.New()
	accountID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	holding := &entity.Holding{
		ID:          holdingID,
		PortfolioID: portfolioID,
		Name:        "ITUB4",
		Type:        "stock",
	}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio
	holdingRepo := newFakeHoldingRepoInv()
	holdingRepo.byID[holdingID] = holding
	txRepo := &fakeTransactionRepoInv{}

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, newFakeInvestTxRepoInv())
	uc.WithTransactionRepo(txRepo)

	req := CreateInvestmentTransactionRequest{
		Type:      "dividend",
		Total:     150.0,
		Date:      time.Now(),
		AccountID: &accountID,
		UserID:    userID,
	}

	invTx, err := uc.CreateInvestmentTransaction(context.Background(), holdingID, req)
	require.NoError(t, err)
	require.NotNil(t, invTx)
	assert.Equal(t, "dividend", invTx.Type)

	txRepo.mu.Lock()
	created := txRepo.created
	txRepo.mu.Unlock()

	require.Len(t, created, 1, "one income transaction should be auto-created")
	income := created[0]
	assert.Equal(t, "income", income.Type)
	assert.Equal(t, 150.0, income.Amount)
	assert.Equal(t, accountID, income.AccountID)
	assert.Equal(t, userID, income.UserID)
	require.NotNil(t, income.Description)
	assert.Contains(t, *income.Description, "ITUB4")
}

func TestCreateInvestmentTransaction_Dividend_NoAccount_SkipsAutoIncome(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	holdingID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	holding := &entity.Holding{ID: holdingID, PortfolioID: portfolioID, Name: "PETR4", Type: "stock"}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio
	holdingRepo := newFakeHoldingRepoInv()
	holdingRepo.byID[holdingID] = holding
	txRepo := &fakeTransactionRepoInv{}

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, newFakeInvestTxRepoInv())
	uc.WithTransactionRepo(txRepo)

	req := CreateInvestmentTransactionRequest{
		Type:  "dividend",
		Total: 200.0,
		Date:  time.Now(),
		// AccountID intentionally omitted
	}

	_, err := uc.CreateInvestmentTransaction(context.Background(), holdingID, req)
	require.NoError(t, err)

	txRepo.mu.Lock()
	count := len(txRepo.created)
	txRepo.mu.Unlock()
	assert.Equal(t, 0, count, "no income tx without accountID")
}

func TestCreateInvestmentTransaction_Buy_NoAutoIncome(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	holdingID := uuid.New()
	accountID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	holding := &entity.Holding{ID: holdingID, PortfolioID: portfolioID, Name: "VALE3", Type: "stock"}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio
	holdingRepo := newFakeHoldingRepoInv()
	holdingRepo.byID[holdingID] = holding
	txRepo := &fakeTransactionRepoInv{}

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, newFakeInvestTxRepoInv())
	uc.WithTransactionRepo(txRepo)

	qty := 10.0
	price := 80.0
	req := CreateInvestmentTransactionRequest{
		Type:      "buy", // not a dividend
		Quantity:  &qty,
		Price:     &price,
		Total:     800.0,
		Date:      time.Now(),
		AccountID: &accountID,
		UserID:    userID,
	}

	_, err := uc.CreateInvestmentTransaction(context.Background(), holdingID, req)
	require.NoError(t, err)

	txRepo.mu.Lock()
	count := len(txRepo.created)
	txRepo.mu.Unlock()
	assert.Equal(t, 0, count, "buy should not create a financial income transaction")
}

// ------------------------------------------------------------------ D6 ----
// Concentration alert – notification triggered when a holding exceeds 30 % of portfolio

func TestConcentrationAlert_TriggeredAbove30Pct(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	holdingID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	// Single holding = 100 % of portfolio → clearly above 30 %
	holding := &entity.Holding{
		ID:           holdingID,
		PortfolioID:  portfolioID,
		Name:         "VALE3",
		Type:         "stock",
		CurrentValue: 1000.0,
	}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio
	holdingRepo := newFakeHoldingRepoInv()
	holdingRepo.byID[holdingID] = holding
	notifUC := &fakeNotificationUCInv{}

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, newFakeInvestTxRepoInv())
	uc.WithNotificationUC(notifUC)

	qty := 10.0
	price := 100.0
	req := CreateInvestmentTransactionRequest{
		Type:     "buy",
		Quantity: &qty,
		Price:    &price,
		Total:    1000.0,
		Date:     time.Now(),
	}

	_, err := uc.CreateInvestmentTransaction(context.Background(), holdingID, req)
	require.NoError(t, err)

	// Allow the goroutine to complete
	time.Sleep(50 * time.Millisecond)

	notifUC.mu.Lock()
	count := len(notifUC.created)
	notifUC.mu.Unlock()
	assert.Greater(t, count, 0, "expected a concentration_alert notification")

	if count > 0 {
		notifUC.mu.Lock()
		n := notifUC.created[0]
		notifUC.mu.Unlock()
		assert.Equal(t, "concentration_alert", n.Type)
	}
}

func TestConcentrationAlert_NotTriggeredAt25Pct(t *testing.T) {
	// Four equal holdings at 25 % each – none should trigger the alert.
	userID := uuid.New()
	portfolioID := uuid.New()

	portfolio := &entity.Portfolio{ID: portfolioID, UserID: userID}
	holdings := make([]*entity.Holding, 4)
	for i := range holdings {
		holdings[i] = &entity.Holding{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Name:         "H" + string(rune('1'+i)),
			Type:         "stock",
			CurrentValue: 250.0, // 25 % of 1 000 total
		}
	}

	portfolioRepo := newFakePortfolioRepo()
	portfolioRepo.portfolios[portfolioID] = portfolio
	holdingRepo := newFakeHoldingRepoInv()
	for _, h := range holdings {
		holdingRepo.byID[h.ID] = h
	}
	notifUC := &fakeNotificationUCInv{}

	uc := newTestInvestmentUseCase(portfolioRepo, holdingRepo, newFakeInvestTxRepoInv())
	uc.WithNotificationUC(notifUC)

	// Trigger via a buy on the first holding
	qty := 1.0
	price := 1.0
	req := CreateInvestmentTransactionRequest{
		Type:     "buy",
		Quantity: &qty,
		Price:    &price,
		Total:    1.0,
		Date:     time.Now(),
	}

	_, err := uc.CreateInvestmentTransaction(context.Background(), holdings[0].ID, req)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	notifUC.mu.Lock()
	count := len(notifUC.created)
	notifUC.mu.Unlock()
	assert.Equal(t, 0, count, "no alert expected when all holdings are at 25%%")
}
