package usecase

import (
	"testing"

	"github.com/financeos/api/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestRecalcHolding_Buy(t *testing.T) {
	tests := []struct {
		name            string
		initial         entity.Holding
		qty             float64
		price           float64
		fees            float64
		expectedQty     float64
		expectedAvg     float64
		expectedInvested float64
	}{
		{
			name:            "first buy",
			initial:         entity.Holding{},
			qty:             10,
			price:           20.0,
			fees:            5.0,
			expectedQty:     10,
			expectedAvg:     20.5, // (10*20 + 5) / 10 = 205/10 = 20.5
			expectedInvested: 205,
		},
		{
			name: "second buy at different price",
			initial: entity.Holding{
				Quantity:      10,
				AvgPrice:      20.5,
				TotalInvested: 205,
			},
			qty:             5,
			price:           25.0,
			fees:            0,
			expectedQty:     15,
			expectedAvg:     22.0, // (205 + 125) / 15 = 330/15 = 22
			expectedInvested: 330,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.initial
			RecalcHolding(&h, "buy", tt.qty, tt.price, tt.fees)
			assert.Equal(t, tt.expectedQty, h.Quantity)
			assert.InDelta(t, tt.expectedAvg, h.AvgPrice, 0.01)
			assert.InDelta(t, tt.expectedInvested, h.TotalInvested, 0.01)
		})
	}
}

func TestRecalcHolding_Sell(t *testing.T) {
	tests := []struct {
		name              string
		initial           entity.Holding
		qty               float64
		price             float64
		fees              float64
		expectedQty       float64
		expectedInvested  float64
		expectedRealizedPnL float64
	}{
		{
			name: "sell at profit",
			initial: entity.Holding{
				Quantity:      10,
				AvgPrice:      20.0,
				TotalInvested: 200,
				RealizedPnL:   0,
			},
			qty:               5,
			price:             30.0,
			fees:              2.0,
			expectedQty:       5,
			expectedInvested:  100, // 20 * 5
			expectedRealizedPnL: 48, // (30-20)*5 - 2 = 48
		},
		{
			name: "sell at loss",
			initial: entity.Holding{
				Quantity:      10,
				AvgPrice:      20.0,
				TotalInvested: 200,
				RealizedPnL:   0,
			},
			qty:               10,
			price:             15.0,
			fees:              0,
			expectedQty:       0,
			expectedInvested:  0,
			expectedRealizedPnL: -50, // (15-20)*10 = -50
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.initial
			RecalcHolding(&h, "sell", tt.qty, tt.price, tt.fees)
			assert.Equal(t, tt.expectedQty, h.Quantity)
			assert.InDelta(t, tt.expectedInvested, h.TotalInvested, 0.01)
			assert.InDelta(t, tt.expectedRealizedPnL, h.RealizedPnL, 0.01)
		})
	}
}

func TestRecalcHolding_Dividend(t *testing.T) {
	h := entity.Holding{
		Quantity:      100,
		AvgPrice:      10.0,
		TotalInvested: 1000,
		RealizedPnL:   0,
	}
	// For dividend: qty=1, price=50 means $50 of dividend
	RecalcHolding(&h, "dividend", 1, 50.0, 0)
	assert.Equal(t, 100.0, h.Quantity)        // quantity unchanged
	assert.Equal(t, 1000.0, h.TotalInvested)  // total invested unchanged
	assert.InDelta(t, 50.0, h.RealizedPnL, 0.01)
}

func TestCalculateIRRegressive(t *testing.T) {
	tests := []struct {
		name     string
		days     int
		expected float64
	}{
		{"up to 180 days", 180, 0.225},
		{"181 days", 181, 0.20},
		{"360 days", 360, 0.20},
		{"361 days", 361, 0.175},
		{"720 days", 720, 0.175},
		{"721 days", 721, 0.15},
		{"1000 days", 1000, 0.15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateIRRegressive(tt.days)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateCDIProjection(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		rate     float64
		cdiRate  float64
		days     int
		minValue float64 // result must be greater than invested
	}{
		{
			name:     "1000 BRL at 100% CDI 13.75% for 252 days",
			amount:   1000,
			rate:     100,
			cdiRate:  13.75,
			days:     252,
			minValue: 1100, // rough expected ~1137.5
		},
		{
			name:     "5000 BRL at 110% CDI 13.75% for 126 days",
			amount:   5000,
			rate:     110,
			cdiRate:  13.75,
			days:     126,
			minValue: 5050, // should yield some positive return
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCDIProjection(tt.amount, tt.rate, tt.cdiRate, tt.days)
			assert.Greater(t, result, tt.minValue, "CDI projection should be higher than invested amount")
		})
	}
}
