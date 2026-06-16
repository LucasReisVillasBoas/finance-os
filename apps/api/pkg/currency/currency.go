// Package currency fetches foreign-exchange quotes (e.g. USD-BRL, EUR-BRL)
// from the AwesomeAPI service (https://docs.awesomeapi.com.br/), which is
// free to use and does not require an API key.
package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// awesomeAPIBaseURL is the base URL for the AwesomeAPI "last quote" endpoint.
const awesomeAPIBaseURL = "https://economia.awesomeapi.com.br/json/last"

// DefaultPairs are the currency pairs fetched when none are specified.
var DefaultPairs = []string{"USD-BRL", "EUR-BRL"}

// Quote holds a single currency pair quote.
type Quote struct {
	// Code is the base currency (e.g. "USD").
	Code string `json:"code"`
	// Codein is the quote currency (e.g. "BRL").
	Codein string `json:"codein"`
	// Name is a human-readable description (e.g. "Dólar Americano/Real Brasileiro").
	Name string `json:"name"`
	// Bid is the current buy price.
	Bid float64 `json:"bid"`
	// Ask is the current sell price.
	Ask float64 `json:"ask"`
	// High is the day's high.
	High float64 `json:"high"`
	// Low is the day's low.
	Low float64 `json:"low"`
	// PctChange is the percentage variation in the day.
	PctChange float64 `json:"pct_change"`
	// UpdatedAt is when the quote was produced.
	UpdatedAt time.Time `json:"updated_at"`
}

// rawQuote matches a single entry of the AwesomeAPI response. AwesomeAPI
// returns all numeric values as strings, so we parse them explicitly.
type rawQuote struct {
	Code      string `json:"code"`
	Codein    string `json:"codein"`
	Name      string `json:"name"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Bid       string `json:"bid"`
	Ask       string `json:"ask"`
	PctChange string `json:"pctChange"`
	Timestamp string `json:"timestamp"`
}

// Service fetches currency quotes from AwesomeAPI.
type Service struct {
	httpClient *http.Client
}

// NewService creates a currency Service with a 10-second timeout.
func NewService() *Service {
	return &Service{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetRates fetches the latest quotes for the given pairs (e.g. "USD-BRL").
// When pairs is empty, DefaultPairs is used.
func (s *Service) GetRates(ctx context.Context, pairs []string) ([]Quote, error) {
	if len(pairs) == 0 {
		pairs = DefaultPairs
	}

	url := awesomeAPIBaseURL + "/" + strings.Join(pairs, ",")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("currency.GetRates: new request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("currency.GetRates: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("currency.GetRates: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("currency.GetRates: read body: %w", err)
	}

	quotes, err := parseQuotes(body, pairs)
	if err != nil {
		return nil, fmt.Errorf("currency.GetRates: %w", err)
	}
	return quotes, nil
}

// parseQuotes decodes an AwesomeAPI response body into a slice of Quote,
// preserving the order of the requested pairs. It is separated from the HTTP
// logic to keep parsing unit-testable.
func parseQuotes(body []byte, pairs []string) ([]Quote, error) {
	var raw map[string]rawQuote
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	quotes := make([]Quote, 0, len(raw))
	seen := make(map[string]bool, len(raw))

	// Emit in the order requested first (AwesomeAPI keys strip the dash,
	// e.g. "USD-BRL" -> "USDBRL").
	for _, pair := range pairs {
		key := strings.ReplaceAll(pair, "-", "")
		if r, ok := raw[key]; ok {
			quotes = append(quotes, toQuote(r))
			seen[key] = true
		}
	}
	// Include any remaining keys not explicitly requested.
	for key, r := range raw {
		if !seen[key] {
			quotes = append(quotes, toQuote(r))
		}
	}

	return quotes, nil
}

// toQuote converts a rawQuote (all-string fields) into a typed Quote.
func toQuote(r rawQuote) Quote {
	q := Quote{
		Code:      r.Code,
		Codein:    r.Codein,
		Name:      r.Name,
		Bid:       parseFloat(r.Bid),
		Ask:       parseFloat(r.Ask),
		High:      parseFloat(r.High),
		Low:       parseFloat(r.Low),
		PctChange: parseFloat(r.PctChange),
	}
	if ts, err := strconv.ParseInt(r.Timestamp, 10, 64); err == nil {
		q.UpdatedAt = time.Unix(ts, 0).UTC()
	}
	return q
}

// parseFloat parses a string into a float64, returning 0 on error.
func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
