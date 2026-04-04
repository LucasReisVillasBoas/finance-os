package entity

import (
	"time"

	"github.com/google/uuid"
)

type Portfolio struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Asset struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Ticker         *string    `json:"ticker,omitempty" db:"ticker"`
	Name           string     `json:"name" db:"name"`
	Type           string     `json:"type" db:"type"` // stock, fii, etf, crypto, fixed_income, fund, other
	Exchange       *string    `json:"exchange,omitempty" db:"exchange"`
	Currency       string     `json:"currency" db:"currency"`
	CurrentPrice   *float64   `json:"current_price,omitempty" db:"current_price"`
	PriceUpdatedAt *time.Time `json:"price_updated_at,omitempty" db:"price_updated_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type Holding struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	PortfolioID      uuid.UUID  `json:"portfolio_id" db:"portfolio_id"`
	AssetID          *uuid.UUID `json:"asset_id,omitempty" db:"asset_id"`
	Name             string     `json:"name" db:"name"`
	Type             string     `json:"type" db:"type"`
	Quantity         float64    `json:"quantity" db:"quantity"`
	AvgPrice         float64    `json:"avg_price" db:"avg_price"`
	TotalInvested    float64    `json:"total_invested" db:"total_invested"`
	CurrentValue     float64    `json:"current_value" db:"current_value"`
	UnrealizedPnL    float64    `json:"unrealized_pnl" db:"unrealized_pnl"`
	UnrealizedPnLPct float64    `json:"unrealized_pnl_pct" db:"unrealized_pnl_pct"`
	RealizedPnL      float64    `json:"realized_pnl" db:"realized_pnl"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	// Joined
	AssetTicker       *string  `json:"asset_ticker,omitempty" db:"asset_ticker"`
	AssetCurrentPrice *float64 `json:"asset_current_price,omitempty" db:"asset_current_price"`
}

type InvestmentTransaction struct {
	ID        uuid.UUID `json:"id" db:"id"`
	HoldingID uuid.UUID `json:"holding_id" db:"holding_id"`
	Type      string    `json:"type" db:"type"` // buy, sell, dividend, split, bonus
	Quantity  *float64  `json:"quantity,omitempty" db:"quantity"`
	Price     *float64  `json:"price,omitempty" db:"price"`
	Fees      float64   `json:"fees" db:"fees"`
	Total     float64   `json:"total" db:"total"`
	Date      time.Time `json:"date" db:"date"`
	Notes     *string   `json:"notes,omitempty" db:"notes"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CustomAsset struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	Name          string     `json:"name" db:"name"`
	Type          string     `json:"type" db:"type"`
	CurrentValue  float64    `json:"current_value" db:"current_value"`
	PurchaseValue *float64   `json:"purchase_value,omitempty" db:"purchase_value"`
	PurchaseDate  *time.Time `json:"purchase_date,omitempty" db:"purchase_date"`
	MonthlyIncome float64    `json:"monthly_income" db:"monthly_income"`
	Description   *string    `json:"description,omitempty" db:"description"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}
