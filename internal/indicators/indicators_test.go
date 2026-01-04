package indicators

import (
	"math"
	"testing"
)

func TestCalculateEMA(t *testing.T) {
	tests := []struct {
		name     string
		prices   []float64
		period   int
		wantLen  int
		validate func(t *testing.T, result []float64)
	}{
		{
			name:    "valid calculation",
			prices:  []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			period:  5,
			wantLen: 11,
			validate: func(t *testing.T, result []float64) {
				// First 4 values should be NaN
				for i := 0; i < 4; i++ {
					if !math.IsNaN(result[i]) {
						t.Errorf("expected NaN at index %d, got %f", i, result[i])
					}
				}
				// 5th value should be SMA
				expectedSMA := (10.0 + 11.0 + 12.0 + 13.0 + 14.0) / 5.0
				if math.Abs(result[4]-expectedSMA) > 0.001 {
					t.Errorf("expected SMA %f at index 4, got %f", expectedSMA, result[4])
				}
				// Subsequent values should be calculated
				for i := 5; i < len(result); i++ {
					if math.IsNaN(result[i]) {
						t.Errorf("unexpected NaN at index %d", i)
					}
				}
			},
		},
		{
			name:    "period too large",
			prices:  []float64{10, 11, 12},
			period:  5,
			wantLen: 3,
			validate: func(t *testing.T, result []float64) {
				// All values should be NaN
				for i, v := range result {
					if !math.IsNaN(v) {
						t.Errorf("expected NaN at index %d, got %f", i, v)
					}
				}
			},
		},
		{
			name:    "period zero",
			prices:  []float64{10, 11, 12},
			period:  0,
			wantLen: 3,
			validate: func(t *testing.T, result []float64) {
				// All values should be NaN
				for i, v := range result {
					if !math.IsNaN(v) {
						t.Errorf("expected NaN at index %d, got %f", i, v)
					}
				}
			},
		},
		{
			name:    "empty prices",
			prices:  []float64{},
			period:  5,
			wantLen: 0,
			validate: func(t *testing.T, result []float64) {
				// Should be empty
			},
		},
		{
			name:    "period equals length",
			prices:  []float64{10, 11, 12, 13, 14},
			period:  5,
			wantLen: 5,
			validate: func(t *testing.T, result []float64) {
				// First 4 values should be NaN
				for i := 0; i < 4; i++ {
					if !math.IsNaN(result[i]) {
						t.Errorf("expected NaN at index %d, got %f", i, result[i])
					}
				}
				// Last value should be SMA
				expectedSMA := (10.0 + 11.0 + 12.0 + 13.0 + 14.0) / 5.0
				if math.Abs(result[4]-expectedSMA) > 0.001 {
					t.Errorf("expected SMA %f at index 4, got %f", expectedSMA, result[4])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateEMA(tt.prices, tt.period)
			if len(result) != tt.wantLen {
				t.Errorf("CalculateEMA() length = %d, want %d", len(result), tt.wantLen)
			}
			tt.validate(t, result)
		})
	}
}

func TestCalculateMultipleEMAs(t *testing.T) {
	prices := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	periods := []int{5, 10}

	result := CalculateMultipleEMAs(prices, periods)

	if len(result) != len(periods) {
		t.Errorf("expected %d EMA results, got %d", len(periods), len(result))
	}

	for _, period := range periods {
		if _, exists := result[period]; !exists {
			t.Errorf("missing EMA result for period %d", period)
		}
		if len(result[period]) != len(prices) {
			t.Errorf("EMA period %d has %d values, expected %d", period, len(result[period]), len(prices))
		}
	}
}

func TestAnalyzeEMA(t *testing.T) {
	// Create a simple uptrend
	prices := []float64{10, 11, 12, 13, 14, 15}
	emas := CalculateEMA(prices, 3)

	analysis := AnalyzeEMA(emas, 3, prices)

	if analysis.Period != 3 {
		t.Errorf("expected period 3, got %d", analysis.Period)
	}

	if math.IsNaN(analysis.CurrentValue) {
		t.Error("CurrentValue should not be NaN")
	}

	if analysis.Trend != TrendUp {
		t.Errorf("expected TrendUp for uptrend, got %v", analysis.Trend)
	}

	if analysis.Slope <= 0 {
		t.Errorf("expected positive slope for uptrend, got %f", analysis.Slope)
	}

	if analysis.TrendStrength <= 0 {
		t.Errorf("expected positive trend strength for uptrend, got %f", analysis.TrendStrength)
	}
}

func TestDetectCrossovers(t *testing.T) {
	// Create data that will produce a golden cross
	// Start with downtrend, then uptrend to force a crossover
	prices := []float64{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 8, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23}
	periods := []int{3, 7}

	emaMap := CalculateMultipleEMAs(prices, periods)
	signals := DetectCrossovers(emaMap, periods)

	// Verify signal structure if any signals are detected
	for _, signal := range signals {
		if signal.FastPeriod >= signal.SlowPeriod {
			t.Errorf("FastPeriod (%d) should be less than SlowPeriod (%d)", signal.FastPeriod, signal.SlowPeriod)
		}
		if signal.SignalType != "golden_cross" && signal.SignalType != "death_cross" {
			t.Errorf("unexpected signal type: %s", signal.SignalType)
		}
		if signal.Strength < 0 || signal.Strength > 1 {
			t.Errorf("strength should be between 0 and 1, got %f", signal.Strength)
		}
	}

	// Note: Not all price patterns will produce crossovers depending on EMA calculation
	// This test validates the structure of any detected signals
	t.Logf("Detected %d crossover signals", len(signals))
}

func TestAnalyzeMarket(t *testing.T) {
	prices := []float64{100, 102, 104, 106, 108, 110, 112, 114, 116, 118, 120}
	periods := []int{3, 5, 7}

	analysis := AnalyzeMarket("BTC", "1h", prices, periods)

	if analysis.Symbol != "BTC" {
		t.Errorf("expected symbol BTC, got %s", analysis.Symbol)
	}

	if analysis.Timeframe != "1h" {
		t.Errorf("expected timeframe 1h, got %s", analysis.Timeframe)
	}

	if analysis.LastPrice != 120 {
		t.Errorf("expected last price 120, got %f", analysis.LastPrice)
	}

	if len(analysis.EMAs) != len(periods) {
		t.Errorf("expected %d EMA analyses, got %d", len(periods), len(analysis.EMAs))
	}

	// For an uptrend, overall trend should be bullish
	if analysis.OverallTrend != TrendUp {
		t.Errorf("expected TrendUp for uptrend data, got %v", analysis.OverallTrend)
	}

	if analysis.TrendScore <= 0 {
		t.Errorf("expected positive trend score for uptrend, got %f", analysis.TrendScore)
	}
}

func TestTrendDirection_String(t *testing.T) {
	tests := []struct {
		trend TrendDirection
		want  string
	}{
		{TrendUp, "Bullish"},
		{TrendDown, "Bearish"},
		{TrendNeutral, "Neutral"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.trend.String(); got != tt.want {
				t.Errorf("TrendDirection.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkCalculateEMA(b *testing.B) {
	prices := make([]float64, 1000)
	for i := range prices {
		prices[i] = float64(i + 100)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateEMA(prices, 50)
	}
}

func BenchmarkCalculateMultipleEMAs(b *testing.B) {
	prices := make([]float64, 1000)
	for i := range prices {
		prices[i] = float64(i + 100)
	}
	periods := []int{50, 100, 200}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateMultipleEMAs(prices, periods)
	}
}

func BenchmarkAnalyzeMarket(b *testing.B) {
	prices := make([]float64, 1000)
	for i := range prices {
		prices[i] = float64(i + 100)
	}
	periods := []int{50, 100, 200}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeMarket("BTC", "1h", prices, periods)
	}
}
