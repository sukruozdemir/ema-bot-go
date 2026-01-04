package indicators

import "math"

// TrendDirection represents the direction of a trend
type TrendDirection int

const (
	TrendNeutral TrendDirection = iota
	TrendUp
	TrendDown
)

func (t TrendDirection) String() string {
	switch t {
	case TrendUp:
		return "Bullish"
	case TrendDown:
		return "Bearish"
	default:
		return "Neutral"
	}
}

// EMAAnalysis contains detailed EMA analysis results
type EMAAnalysis struct {
	Period        int            `json:"period"`
	CurrentValue  float64        `json:"current_value"`
	PreviousValue float64        `json:"previous_value"`
	Change        float64        `json:"change"`
	ChangePercent float64        `json:"change_percent"`
	Trend         TrendDirection `json:"trend"`
	TrendStrength float64        `json:"trend_strength"` // 0-1 scale
	Slope         float64        `json:"slope"`          // Rate of change
}

// CrossoverSignal represents an EMA crossover event
type CrossoverSignal struct {
	FastPeriod   int     `json:"fast_period"`
	SlowPeriod   int     `json:"slow_period"`
	SignalType   string  `json:"signal_type"`   // "golden_cross", "death_cross"
	CrossoverBar int     `json:"crossover_bar"` // Bars ago when crossover occurred
	Strength     float64 `json:"strength"`      // Signal strength 0-1
}

// MarketAnalysis contains comprehensive analysis for a market
type MarketAnalysis struct {
	Symbol       string            `json:"symbol"`
	Timeframe    string            `json:"timeframe"`
	LastPrice    float64           `json:"last_price"`
	EMAs         []EMAAnalysis     `json:"emas"`
	Signals      []CrossoverSignal `json:"signals"`
	OverallTrend TrendDirection    `json:"overall_trend"`
	TrendScore   float64           `json:"trend_score"` // -1 to 1
}

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

// AnalyzeEMA performs detailed analysis on a single EMA series
func AnalyzeEMA(emas []float64, period int, prices []float64) EMAAnalysis {
	analysis := EMAAnalysis{
		Period:       period,
		CurrentValue: math.NaN(),
		Change:       0,
		Trend:        TrendNeutral,
		Slope:        0,
	}

	if len(emas) == 0 {
		return analysis
	}

	// Find the latest valid EMA value
	latestIdx := -1
	for i := len(emas) - 1; i >= 0; i-- {
		if !math.IsNaN(emas[i]) {
			analysis.CurrentValue = emas[i]
			latestIdx = i
			break
		}
	}

	if latestIdx == -1 {
		return analysis
	}

	// Find the previous valid EMA value
	if latestIdx > 0 {
		for i := latestIdx - 1; i >= 0; i-- {
			if !math.IsNaN(emas[i]) {
				analysis.PreviousValue = emas[i]
				analysis.Change = analysis.CurrentValue - emas[i]
				if emas[i] != 0 {
					analysis.ChangePercent = (analysis.Change / emas[i]) * 100
				}
				break
			}
		}
	}

	// Calculate trend and slope over the last 5 bars (or available bars)
	lookback := 5
	if lookback > latestIdx {
		lookback = latestIdx
	}

	if lookback > 0 {
		startIdx := latestIdx - lookback
		var slopeSum float64
		var validSlopes int

		for i := startIdx + 1; i <= latestIdx; i++ {
			if !math.IsNaN(emas[i-1]) && !math.IsNaN(emas[i]) {
				slopeSum += emas[i] - emas[i-1]
				validSlopes++
			}
		}

		if validSlopes > 0 {
			analysis.Slope = slopeSum / float64(validSlopes)
		}
	}

	// Determine trend direction and strength
	if analysis.Slope > 0 {
		analysis.Trend = TrendUp
		analysis.TrendStrength = math.Min(math.Abs(analysis.Slope)/analysis.CurrentValue*100, 1.0)
	} else if analysis.Slope < 0 {
		analysis.Trend = TrendDown
		analysis.TrendStrength = math.Min(math.Abs(analysis.Slope)/analysis.CurrentValue*100, 1.0)
	}

	return analysis
}

// DetectCrossovers identifies EMA crossover signals
func DetectCrossovers(emaMap map[int][]float64, periods []int) []CrossoverSignal {
	var signals []CrossoverSignal

	if len(periods) < 2 {
		return signals
	}

	// Check each pair of EMAs for crossovers
	for i := 0; i < len(periods)-1; i++ {
		for j := i + 1; j < len(periods); j++ {
			fastPeriod := periods[i]
			slowPeriod := periods[j]

			if fastPeriod > slowPeriod {
				fastPeriod, slowPeriod = slowPeriod, fastPeriod
			}

			fastEMA := emaMap[fastPeriod]
			slowEMA := emaMap[slowPeriod]

			if len(fastEMA) != len(slowEMA) || len(fastEMA) < 2 {
				continue
			}

			// Check for crossover in the last few bars
			for barIdx := len(fastEMA) - 1; barIdx >= max(0, len(fastEMA)-20); barIdx-- {
				if barIdx == 0 {
					continue
				}

				current := fastEMA[barIdx]
				currentSlow := slowEMA[barIdx]
				previous := fastEMA[barIdx-1]
				previousSlow := slowEMA[barIdx-1]

				// Skip if any values are NaN
				if math.IsNaN(current) || math.IsNaN(currentSlow) ||
					math.IsNaN(previous) || math.IsNaN(previousSlow) {
					continue
				}

				// Check for golden cross (fast EMA crosses above slow EMA)
				if previous <= previousSlow && current > currentSlow {
					strength := calculateCrossoverStrength(fastEMA, slowEMA, barIdx)
					signals = append(signals, CrossoverSignal{
						FastPeriod:   fastPeriod,
						SlowPeriod:   slowPeriod,
						SignalType:   "golden_cross",
						CrossoverBar: len(fastEMA) - 1 - barIdx,
						Strength:     strength,
					})
				}

				// Check for death cross (fast EMA crosses below slow EMA)
				if previous >= previousSlow && current < currentSlow {
					strength := calculateCrossoverStrength(fastEMA, slowEMA, barIdx)
					signals = append(signals, CrossoverSignal{
						FastPeriod:   fastPeriod,
						SlowPeriod:   slowPeriod,
						SignalType:   "death_cross",
						CrossoverBar: len(fastEMA) - 1 - barIdx,
						Strength:     strength,
					})
				}
			}
		}
	}

	return signals
}

// AnalyzeMarket performs comprehensive EMA analysis for a market
func AnalyzeMarket(symbol, timeframe string, prices []float64, periods []int) MarketAnalysis {
	analysis := MarketAnalysis{
		Symbol:       symbol,
		Timeframe:    timeframe,
		EMAs:         make([]EMAAnalysis, 0, len(periods)),
		OverallTrend: TrendNeutral,
	}

	if len(prices) > 0 {
		analysis.LastPrice = prices[len(prices)-1]
	}

	// Calculate EMAs
	emaMap := CalculateMultipleEMAs(prices, periods)

	// Analyze each EMA
	var trendScore float64
	var validTrends int

	for _, period := range periods {
		emaAnalysis := AnalyzeEMA(emaMap[period], period, prices)
		analysis.EMAs = append(analysis.EMAs, emaAnalysis)

		// Contribute to overall trend score
		if !math.IsNaN(emaAnalysis.CurrentValue) {
			switch emaAnalysis.Trend {
			case TrendUp:
				trendScore += emaAnalysis.TrendStrength
				validTrends++
			case TrendDown:
				trendScore -= emaAnalysis.TrendStrength
				validTrends++
			}
		}
	}

	// Calculate overall trend
	if validTrends > 0 {
		analysis.TrendScore = trendScore / float64(validTrends)
		if analysis.TrendScore > 0.1 {
			analysis.OverallTrend = TrendUp
		} else if analysis.TrendScore < -0.1 {
			analysis.OverallTrend = TrendDown
		}
	}

	// Detect crossover signals
	analysis.Signals = DetectCrossovers(emaMap, periods)

	return analysis
}

// Helper functions

func calculateCrossoverStrength(fastEMA, slowEMA []float64, crossoverIdx int) float64 {
	if crossoverIdx >= len(fastEMA) || crossoverIdx < 1 {
		return 0.5
	}

	// Look at the momentum before and after the crossover
	beforeDiff := math.Abs(fastEMA[crossoverIdx-1] - slowEMA[crossoverIdx-1])
	afterDiff := math.Abs(fastEMA[crossoverIdx] - slowEMA[crossoverIdx])

	// Higher difference after crossover indicates stronger signal
	if beforeDiff == 0 {
		return 0.5
	}

	strength := afterDiff / (beforeDiff + afterDiff)
	return math.Max(0.1, math.Min(1.0, strength))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
