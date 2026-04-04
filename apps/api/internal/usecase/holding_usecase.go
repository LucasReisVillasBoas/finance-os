package usecase

import (
	"github.com/financeos/api/internal/domain/entity"
)

// RecalcHolding recalculates the holding position after an investment operation.
func RecalcHolding(h *entity.Holding, txType string, qty, price, fees float64) {
	switch txType {
	case "buy":
		totalCost := qty*price + fees
		newQty := h.Quantity + qty
		if newQty > 0 {
			h.AvgPrice = (h.TotalInvested + totalCost) / newQty
		}
		h.Quantity = newQty
		h.TotalInvested += totalCost
	case "sell":
		h.RealizedPnL += (price-h.AvgPrice)*qty - fees
		h.Quantity -= qty
		h.TotalInvested = h.AvgPrice * h.Quantity
		if h.Quantity <= 0 {
			h.Quantity = 0
			h.TotalInvested = 0
		}
	case "dividend":
		// qty is used as amount for dividends
		h.RealizedPnL += qty * price
	}

	// UpdateUnrealizedPnL if current_price is available
	if h.AssetCurrentPrice != nil && *h.AssetCurrentPrice > 0 {
		h.CurrentValue = h.Quantity * *h.AssetCurrentPrice
		h.UnrealizedPnL = h.CurrentValue - h.TotalInvested
		if h.TotalInvested > 0 {
			h.UnrealizedPnLPct = (h.UnrealizedPnL / h.TotalInvested) * 100
		}
	}
}
