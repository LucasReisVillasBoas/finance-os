package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/financeos/api/pkg/currency"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// currencyQuotesCacheKey is the Redis key for cached currency quotes.
const currencyQuotesCacheKey = "quotes:currencies"

// currencyQuotesTTL is how long currency quotes are cached. FX rates do not
// need to be real-time here, and AwesomeAPI's free tier is rate-limited.
const currencyQuotesTTL = 10 * time.Minute

// QuoteHandler serves market quote endpoints (currencies, etc.).
type QuoteHandler struct {
	currencySvc *currency.Service
	cache       *redis.Client
	logger      *zap.Logger
}

// NewQuoteHandler creates a QuoteHandler.
func NewQuoteHandler(currencySvc *currency.Service, cache *redis.Client, logger *zap.Logger) *QuoteHandler {
	return &QuoteHandler{
		currencySvc: currencySvc,
		cache:       cache,
		logger:      logger,
	}
}

// GetCurrencyQuotes handles GET /api/v1/quotes/currencies.
// Returns the latest USD-BRL and EUR-BRL quotes (default pairs), cached in
// Redis for a few minutes to respect AwesomeAPI's free rate limits.
//
//	@Summary		Get currency quotes (USD, EUR)
//	@Tags			Quotes
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/quotes/currencies [get]
func (h *QuoteHandler) GetCurrencyQuotes(c *gin.Context) {
	ctx := c.Request.Context()

	// Serve from cache when available.
	if h.cache != nil {
		if cached, err := h.cache.Get(ctx, currencyQuotesCacheKey).Bytes(); err == nil {
			var quotes []currency.Quote
			if jsonErr := json.Unmarshal(cached, &quotes); jsonErr == nil {
				c.JSON(http.StatusOK, gin.H{"data": quotes})
				return
			}
		}
	}

	quotes, err := h.currencySvc.GetRates(ctx, currency.DefaultPairs)
	if err != nil {
		h.logger.Error("get currency quotes", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"code":    "QUOTE_PROVIDER_UNAVAILABLE",
				"message": "could not fetch currency quotes",
			},
		})
		return
	}

	if quotes == nil {
		quotes = []currency.Quote{}
	}

	// Best-effort cache write.
	if h.cache != nil {
		if data, jsonErr := json.Marshal(quotes); jsonErr == nil {
			_ = h.cache.Set(ctx, currencyQuotesCacheKey, data, currencyQuotesTTL).Err()
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": quotes})
}
