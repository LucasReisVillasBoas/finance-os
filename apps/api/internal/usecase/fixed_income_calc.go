package usecase

import "math"

// CalculateCDIProjection projects the yield of a post-fixed CDI investment.
// rate: percentage of CDI (e.g., 110 for 110% CDI)
// cdiRate: current annual CDI rate (e.g., 13.75)
// amount: invested amount
// days: days until maturity
func CalculateCDIProjection(amount, rate, cdiRate float64, days int) float64 {
	dailyRate := math.Pow(1+cdiRate/100, 1.0/252) - 1
	effectiveRate := dailyRate * (rate / 100)
	return amount * math.Pow(1+effectiveRate, float64(days))
}

// CalculateIRRegressive returns the regressive IR tax rate for fixed income.
// The rate decreases the longer the investment is held:
// - up to 180 days: 22.5%
// - 181–360 days: 20%
// - 361–720 days: 17.5%
// - above 720 days: 15%
func CalculateIRRegressive(days int) float64 {
	switch {
	case days <= 180:
		return 0.225
	case days <= 360:
		return 0.20
	case days <= 720:
		return 0.175
	default:
		return 0.15
	}
}
