package brapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AssetResult holds the result of a BRAPI asset search.
type AssetResult struct {
	Ticker       string
	Name         string
	Type         string
	Exchange     string
	CurrentPrice float64
	Currency     string
}

// brapiQuoteResponse matches the BRAPI /quote/:ticker response.
type brapiQuoteResponse struct {
	Results []struct {
		Symbol             string  `json:"symbol"`
		ShortName          string  `json:"shortName"`
		LongName           string  `json:"longName"`
		RegularMarketPrice float64 `json:"regularMarketPrice"`
		Currency           string  `json:"currency"`
	} `json:"results"`
	Error string `json:"error"`
}

// BrapiService calls the BRAPI API for asset data.
type BrapiService struct {
	httpClient *http.Client
}

// NewBrapiService creates a BrapiService with a 10-second timeout.
func NewBrapiService() *BrapiService {
	return &BrapiService{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Search fetches asset information for the given ticker/query from BRAPI.
// Returns nil, nil when BRAPI reports an error (e.g. ticker not found).
func (s *BrapiService) Search(ctx context.Context, query string) ([]AssetResult, error) {
	url := fmt.Sprintf("https://brapi.dev/api/quote/%s?fundamental=false", query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("brapi.Search: new request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("brapi.Search: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("brapi.Search: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("brapi.Search: read body: %w", err)
	}

	var brapiResp brapiQuoteResponse
	if err := json.Unmarshal(body, &brapiResp); err != nil {
		return nil, fmt.Errorf("brapi.Search: unmarshal: %w", err)
	}

	// BRAPI returns a non-empty "error" field when the ticker is not found.
	if brapiResp.Error != "" {
		return nil, nil
	}

	results := make([]AssetResult, 0, len(brapiResp.Results))
	for _, r := range brapiResp.Results {
		name := r.LongName
		if name == "" {
			name = r.ShortName
		}
		if name == "" {
			name = r.Symbol
		}
		currency := r.Currency
		if currency == "" {
			currency = "BRL"
		}
		results = append(results, AssetResult{
			Ticker:       r.Symbol,
			Name:         name,
			Type:         "stock",
			Exchange:     "B3",
			CurrentPrice: r.RegularMarketPrice,
			Currency:     currency,
		})
	}

	return results, nil
}

// brapiAvailableResponse matches GET /available response.
type brapiAvailableResponse struct {
	Stocks  []string `json:"stocks"`
	Indexes []string `json:"indexes"`
}

// FetchAvailableTickers returns all tickers available on B3 via GET /available.
func (s *BrapiService) FetchAvailableTickers(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://brapi.dev/api/available", nil)
	if err != nil {
		return nil, fmt.Errorf("brapi.FetchAvailableTickers: new request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("brapi.FetchAvailableTickers: do: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("brapi.FetchAvailableTickers: read: %w", err)
	}

	var result brapiAvailableResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("brapi.FetchAvailableTickers: unmarshal: %w", err)
	}
	return result.Stocks, nil
}

// SearchByQuery filters tickers locally then fetches prices concurrently.
// BRAPI free tier only allows single-ticker requests, so we fetch each in parallel.
// allTickers is the full B3 ticker list (from FetchAvailableTickers, cached).
func (s *BrapiService) SearchByQuery(ctx context.Context, query string, allTickers []string) ([]AssetResult, error) {
	upperQuery := strings.ToUpper(query)
	matches := make([]string, 0, 5)
	for _, t := range allTickers {
		if strings.Contains(t, upperQuery) {
			matches = append(matches, t)
			if len(matches) >= 5 {
				break
			}
		}
	}
	if len(matches) == 0 {
		return nil, nil
	}

	// Fetch each ticker individually in parallel (free tier limitation).
	type result struct {
		asset AssetResult
		err   error
	}
	ch := make(chan result, len(matches))
	var wg sync.WaitGroup

	for _, ticker := range matches {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			url := fmt.Sprintf("https://brapi.dev/api/quote/%s?fundamental=false", t)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				ch <- result{asset: AssetResult{Ticker: t, Name: t, Type: "stock", Exchange: "B3", Currency: "BRL"}}
				return
			}
			resp, err := s.httpClient.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				if resp != nil {
					resp.Body.Close()
				}
				ch <- result{asset: AssetResult{Ticker: t, Name: t, Type: "stock", Exchange: "B3", Currency: "BRL"}}
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				ch <- result{asset: AssetResult{Ticker: t, Name: t, Type: "stock", Exchange: "B3", Currency: "BRL"}}
				return
			}
			var brapiResp brapiQuoteResponse
			if err := json.Unmarshal(body, &brapiResp); err != nil || len(brapiResp.Results) == 0 {
				ch <- result{asset: AssetResult{Ticker: t, Name: t, Type: "stock", Exchange: "B3", Currency: "BRL"}}
				return
			}
			r := brapiResp.Results[0]
			name := r.LongName
			if name == "" {
				name = r.ShortName
			}
			if name == "" {
				name = r.Symbol
			}
			ch <- result{asset: AssetResult{
				Ticker:       r.Symbol,
				Name:         name,
				Type:         "stock",
				Exchange:     "B3",
				CurrentPrice: r.RegularMarketPrice,
				Currency:     "BRL",
			}}
		}(ticker)
	}

	wg.Wait()
	close(ch)

	results := make([]AssetResult, 0, len(matches))
	for res := range ch {
		results = append(results, res.asset)
	}
	return results, nil
}

// FetchPrice fetches the current price for a single ticker from BRAPI.
func (s *BrapiService) FetchPrice(ctx context.Context, ticker string) (float64, error) {
	results, err := s.Search(ctx, ticker)
	if err != nil {
		return 0, fmt.Errorf("brapi.FetchPrice: %w", err)
	}
	if len(results) == 0 {
		return 0, fmt.Errorf("brapi.FetchPrice: no results for ticker %s", ticker)
	}
	return results[0].CurrentPrice, nil
}
