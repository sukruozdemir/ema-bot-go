package indicators

import "math"

// CalculateEMA computes the Exponential Moving Average (EMA) for a slice of float64 prices.
//
// Parameters:
//   - prices: slice of price values (oldest to newest). Must not be empty.
//   - period: EMA period (must be > 0 and <= len(prices))
//
// Returns a slice of EMA values (same length as prices):
//   - First period-1 values are NaN (EMA not yet calculated)
//   - From index period-1 onwards, contains valid EMA values
//
// The EMA is calculated using the standard formula:
// EMA(today) = (Price(today) - EMA(yesterday)) * multiplier + EMA(yesterday)
// where multiplier = 2 / (period + 1)
func CalculateEMA(prices []float64, period int) []float64 {
	n := len(prices)
	emas := make([]float64, n)

	// Initialize with NaN for indices that won't have valid EMA
	for i := range n {
		emas[i] = math.NaN()
	}

	// Validate period
	if period <= 0 || period > n || n == 0 {
		return emas // return all NaN for invalid input
	}

	multiplier := 2.0 / float64(period+1)

	// Calculate SMA for the first EMA value
	var sum float64
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	emas[period-1] = sum / float64(period)

	// Calculate EMA for remaining values
	for i := period; i < n; i++ {
		emas[i] = (prices[i]-emas[i-1])*multiplier + emas[i-1]
	}

	return emas
}

// CalculateMultipleEMAs efficiently computes multiple EMA periods in a single pass.
//
// Parameters:
//   - prices: slice of price values (oldest to newest)
//   - periods: slice of EMA periods (e.g., []int{50, 100, 200})
//
// Returns a map[int][]float64 where keys are periods and values are EMA slices
func CalculateMultipleEMAs(prices []float64, periods []int) map[int][]float64 {
	result := make(map[int][]float64)
	for _, period := range periods {
		result[period] = CalculateEMA(prices, period)
	}
	return result
}
