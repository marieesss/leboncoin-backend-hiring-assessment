package http

import (
	"fmt"
	"math"
)

// maxPrice is the highest price the database accepts
const maxPrice = 9_999_999_999.99

// centsToEuros converts a price in cents to the euro
func centsToEuros(cents int64) float64 {
	return float64(cents) / 100
}

// eurosToCents converts a price in euros to cents
func eurosToCents(euros float64) (int64, error) {
	if euros < 0 || euros > maxPrice {
		return 0, fmt.Errorf("price %v must be between 0 and %.2f", euros, maxPrice)
	}

	cents := math.Round(euros * 100)
	if math.Abs(euros*100-cents) > 1e-6 {
		return 0, fmt.Errorf("price %v must have at most 2 decimals", euros)
	}

	return int64(cents), nil
}
