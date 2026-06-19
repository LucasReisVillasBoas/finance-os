package currency

import (
	"testing"
)

const sampleBody = `{
  "USDBRL": {
    "code": "USD",
    "codein": "BRL",
    "name": "Dólar Americano/Real Brasileiro",
    "high": "5.45",
    "low": "5.40",
    "varBid": "0.01",
    "pctChange": "0.2",
    "bid": "5.42",
    "ask": "5.43",
    "timestamp": "1718541000",
    "create_date": "2026-06-16 12:30:00"
  },
  "EURBRL": {
    "code": "EUR",
    "codein": "BRL",
    "name": "Euro/Real Brasileiro",
    "high": "5.90",
    "low": "5.85",
    "varBid": "-0.02",
    "pctChange": "-0.3",
    "bid": "5.88",
    "ask": "5.89",
    "timestamp": "1718541000",
    "create_date": "2026-06-16 12:30:00"
  }
}`

func TestParseQuotes(t *testing.T) {
	pairs := []string{"USD-BRL", "EUR-BRL"}
	quotes, err := parseQuotes([]byte(sampleBody), pairs)
	if err != nil {
		t.Fatalf("parseQuotes returned error: %v", err)
	}

	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}

	// Order must follow the requested pairs.
	if quotes[0].Code != "USD" {
		t.Errorf("expected first quote USD, got %s", quotes[0].Code)
	}
	if quotes[1].Code != "EUR" {
		t.Errorf("expected second quote EUR, got %s", quotes[1].Code)
	}

	usd := quotes[0]
	if usd.Codein != "BRL" {
		t.Errorf("expected codein BRL, got %s", usd.Codein)
	}
	if usd.Bid != 5.42 {
		t.Errorf("expected bid 5.42, got %v", usd.Bid)
	}
	if usd.Ask != 5.43 {
		t.Errorf("expected ask 5.43, got %v", usd.Ask)
	}
	if usd.PctChange != 0.2 {
		t.Errorf("expected pctChange 0.2, got %v", usd.PctChange)
	}
	if usd.UpdatedAt.IsZero() {
		t.Errorf("expected non-zero UpdatedAt")
	}

	eur := quotes[1]
	if eur.PctChange != -0.3 {
		t.Errorf("expected negative pctChange -0.3, got %v", eur.PctChange)
	}
}

func TestParseQuotes_Invalid(t *testing.T) {
	if _, err := parseQuotes([]byte("not json"), DefaultPairs); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"5.42", 5.42},
		{"-0.3", -0.3},
		{"", 0},
		{"abc", 0},
	}
	for _, tt := range tests {
		if got := parseFloat(tt.in); got != tt.want {
			t.Errorf("parseFloat(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
