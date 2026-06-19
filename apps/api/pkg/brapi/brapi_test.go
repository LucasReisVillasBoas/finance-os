package brapi

import (
	"strings"
	"testing"
)

func TestQuoteURL(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		ticker      string
		wantPrefix  string
		wantToken   bool
		tokenValue  string
	}{
		{
			name:       "without token",
			token:      "",
			ticker:     "PETR4",
			wantPrefix: "https://brapi.dev/api/quote/PETR4?fundamental=false",
			wantToken:  false,
		},
		{
			name:       "with token",
			token:      "abc123",
			ticker:     "VALE3",
			wantPrefix: "https://brapi.dev/api/quote/VALE3?fundamental=false",
			wantToken:  true,
			tokenValue: "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewBrapiService(tt.token)
			url := svc.quoteURL(tt.ticker)

			if !strings.HasPrefix(url, tt.wantPrefix) {
				t.Fatalf("quoteURL = %q, want prefix %q", url, tt.wantPrefix)
			}
			hasToken := strings.Contains(url, "&token=")
			if hasToken != tt.wantToken {
				t.Fatalf("quoteURL token presence = %v, want %v (url=%q)", hasToken, tt.wantToken, url)
			}
			if tt.wantToken && !strings.Contains(url, "&token="+tt.tokenValue) {
				t.Fatalf("quoteURL = %q, want token value %q", url, tt.tokenValue)
			}
		})
	}
}
