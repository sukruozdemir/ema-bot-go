package models

import "time"

// Market represents a trading market/pair
type Market struct {
	ID             string `json:"id"`
	Symbol         string `json:"symbol"`
	Base           string `json:"base"`
	Quote          string `json:"quote"`
	Active         bool   `json:"active"`
	Type           string `json:"type"` // spot, swap, future, etc.
	Spot           bool   `json:"spot"`
	Swap           bool   `json:"swap"`
	Future         bool   `json:"future"`
	Option         bool   `json:"option"`
	Contract       bool   `json:"contract"`
	Linear         bool   `json:"linear"`
	Inverse        bool   `json:"inverse"`
	PricePrecision int    `json:"price_precision,omitempty"`
}

// MarketCache represents cached market data
type MarketCache struct {
	Timestamp time.Time `json:"timestamp"`
	Markets   []Market  `json:"markets"`
	ExpiresAt time.Time `json:"expires_at"`
}

// IsValid checks if the cached data is still valid
func (mc *MarketCache) IsValid() bool {
	return time.Now().Before(mc.ExpiresAt)
}

// EMAConfig represents EMA configuration for a market
type EMAConfig struct {
	Symbol     string    `json:"symbol"`
	Timeframe  string    `json:"timeframe"`
	Length     int       `json:"length"`
	Value      float64   `json:"value,omitempty"`
	LastUpdate time.Time `json:"last_update,omitempty"`
}
