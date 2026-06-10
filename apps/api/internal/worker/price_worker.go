package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/financeos/api/pkg/brapi"
	"go.uber.org/zap"
)

// PriceWorker updates asset prices periodically.
type PriceWorker struct {
	assetRepo   domainrepo.AssetRepository
	holdingRepo domainrepo.HoldingRepository
	logger      *zap.Logger
	httpClient  *http.Client
	brapiSvc    *brapi.BrapiService
}

// NewPriceWorker creates a new PriceWorker.
func NewPriceWorker(ar domainrepo.AssetRepository, hr domainrepo.HoldingRepository, l *zap.Logger, brapiSvc *brapi.BrapiService) *PriceWorker {
	return &PriceWorker{
		assetRepo:   ar,
		holdingRepo: hr,
		logger:      l,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		brapiSvc:    brapiSvc,
	}
}

// Run starts the worker loop, updating prices every 15 minutes.
func (w *PriceWorker) Run(ctx context.Context) {
	w.updatePrices(ctx)
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.updatePrices(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// coinGeckoResponse matches CoinGecko simple/price API response.
type coinGeckoResponse map[string]map[string]float64

// updatePrices fetches and updates prices for all assets with a defined ticker.
func (w *PriceWorker) updatePrices(ctx context.Context) {
	assets, err := w.assetRepo.FindAll(ctx)
	if err != nil {
		w.logger.Error("price_worker: find all assets", zap.Error(err))
		return
	}

	updated := 0
	for _, asset := range assets {
		if asset.Ticker == nil || *asset.Ticker == "" {
			continue
		}

		var newPrice float64
		var fetchErr error

		exchange := ""
		if asset.Exchange != nil {
			exchange = *asset.Exchange
		}

		switch {
		case asset.Type == "crypto":
			newPrice, fetchErr = w.fetchCryptoPrice(ctx, *asset.Ticker)
		case exchange == "B3":
			if w.brapiSvc != nil {
				newPrice, fetchErr = w.brapiSvc.FetchPrice(ctx, *asset.Ticker)
			} else {
				continue
			}
		default:
			// Skip assets with no known price source
			continue
		}

		if fetchErr != nil {
			w.logger.Warn("price_worker: failed to fetch price",
				zap.String("ticker", *asset.Ticker),
				zap.Error(fetchErr),
			)
			continue
		}

		if newPrice <= 0 {
			continue
		}

		if err := w.assetRepo.UpdatePrice(ctx, asset.ID, newPrice); err != nil {
			w.logger.Error("price_worker: update price", zap.String("asset_id", asset.ID.String()), zap.Error(err))
			continue
		}

		// Recalculate holdings for this asset
		w.recalcHoldings(ctx, asset.ID, newPrice)

		updated++
	}

	if updated > 0 {
		w.logger.Info("price_worker: updated prices", zap.Int("count", updated))
	}
}

// fetchCryptoPrice fetches a cryptocurrency price from CoinGecko.
// The ticker is expected to be the CoinGecko coin ID (e.g., "bitcoin", "ethereum").
func (w *PriceWorker) fetchCryptoPrice(ctx context.Context, coinID string) (float64, error) {
	// Normalize to lowercase for CoinGecko
	id := strings.ToLower(coinID)
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=brl", id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("fetchCryptoPrice: new request: %w", err)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetchCryptoPrice: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("fetchCryptoPrice: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("fetchCryptoPrice: read body: %w", err)
	}

	var cgResp coinGeckoResponse
	if err := json.Unmarshal(body, &cgResp); err != nil {
		return 0, fmt.Errorf("fetchCryptoPrice: unmarshal: %w", err)
	}

	coinData, ok := cgResp[id]
	if !ok {
		return 0, fmt.Errorf("fetchCryptoPrice: no data for coin %s", id)
	}

	price, ok := coinData["brl"]
	if !ok {
		return 0, fmt.Errorf("fetchCryptoPrice: no BRL price for coin %s", id)
	}

	return price, nil
}

// recalcHoldings recalculates current_value and unrealized P&L for holdings linked to an asset.
func (w *PriceWorker) recalcHoldings(ctx context.Context, assetID interface{ String() string }, price float64) {
	holdings, err := w.holdingRepo.FindAll(ctx)
	if err != nil {
		w.logger.Error("price_worker: find all holdings", zap.Error(err))
		return
	}

	assetIDStr := assetID.String()
	for _, h := range holdings {
		if h.AssetID == nil || h.AssetID.String() != assetIDStr {
			continue
		}
		h.AssetCurrentPrice = &price
		h.CurrentValue = h.Quantity * price
		h.UnrealizedPnL = h.CurrentValue - h.TotalInvested
		if h.TotalInvested > 0 {
			h.UnrealizedPnLPct = (h.UnrealizedPnL / h.TotalInvested) * 100
		}
		if err := w.holdingRepo.Update(ctx, h); err != nil {
			w.logger.Error("price_worker: update holding",
				zap.String("holding_id", h.ID.String()),
				zap.Error(err),
			)
		}
	}
}
