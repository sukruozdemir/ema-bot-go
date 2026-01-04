package models

import (
	"testing"
	"time"
)

func TestMarketCache_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		cache MarketCache
		want  bool
	}{
		{
			name: "valid cache",
			cache: MarketCache{
				Timestamp: time.Now(),
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			want: true,
		},
		{
			name: "expired cache",
			cache: MarketCache{
				Timestamp: time.Now().Add(-48 * time.Hour),
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			want: false,
		},
		{
			name: "cache expires now",
			cache: MarketCache{
				Timestamp: time.Now().Add(-24 * time.Hour),
				ExpiresAt: time.Now(),
			},
			want: false,
		},
		{
			name: "cache expires in 1 second",
			cache: MarketCache{
				Timestamp: time.Now(),
				ExpiresAt: time.Now().Add(1 * time.Second),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cache.IsValid()
			if got != tt.want {
				t.Errorf("MarketCache.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarket_Structure(t *testing.T) {
	market := Market{
		ID:             "BTC/USDT",
		Symbol:         "BTC/USDT",
		Base:           "BTC",
		Quote:          "USDT",
		Active:         true,
		Type:           "spot",
		Spot:           true,
		Swap:           false,
		Future:         false,
		Option:         false,
		Contract:       false,
		Linear:         false,
		Inverse:        false,
		PricePrecision: 2,
	}

	if market.ID != "BTC/USDT" {
		t.Errorf("ID = %v, want BTC/USDT", market.ID)
	}

	if market.Symbol != "BTC/USDT" {
		t.Errorf("Symbol = %v, want BTC/USDT", market.Symbol)
	}

	if market.Base != "BTC" {
		t.Errorf("Base = %v, want BTC", market.Base)
	}

	if market.Quote != "USDT" {
		t.Errorf("Quote = %v, want USDT", market.Quote)
	}

	if !market.Active {
		t.Error("Active should be true")
	}

	if market.Type != "spot" {
		t.Errorf("Type = %v, want spot", market.Type)
	}

	if !market.Spot {
		t.Error("Spot should be true")
	}

	if market.Swap {
		t.Error("Swap should be false")
	}

	if market.PricePrecision != 2 {
		t.Errorf("PricePrecision = %v, want 2", market.PricePrecision)
	}
}

func TestEMAConfig_Structure(t *testing.T) {
	now := time.Now()
	config := EMAConfig{
		Symbol:     "BTC",
		Timeframe:  "1h",
		Length:     50,
		Value:      43250.45,
		LastUpdate: now,
	}

	if config.Symbol != "BTC" {
		t.Errorf("Symbol = %v, want BTC", config.Symbol)
	}

	if config.Timeframe != "1h" {
		t.Errorf("Timeframe = %v, want 1h", config.Timeframe)
	}

	if config.Length != 50 {
		t.Errorf("Length = %v, want 50", config.Length)
	}

	if config.Value != 43250.45 {
		t.Errorf("Value = %v, want 43250.45", config.Value)
	}

	if !config.LastUpdate.Equal(now) {
		t.Errorf("LastUpdate = %v, want %v", config.LastUpdate, now)
	}
}

func TestMarketCache_WithMarkets(t *testing.T) {
	markets := []Market{
		{
			ID:     "BTC/USDT",
			Symbol: "BTC/USDT",
			Base:   "BTC",
			Quote:  "USDT",
			Active: true,
			Spot:   true,
		},
		{
			ID:     "ETH/USDT",
			Symbol: "ETH/USDT",
			Base:   "ETH",
			Quote:  "USDT",
			Active: true,
			Spot:   true,
		},
	}

	cache := MarketCache{
		Timestamp: time.Now(),
		Markets:   markets,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if len(cache.Markets) != 2 {
		t.Errorf("Markets length = %d, want 2", len(cache.Markets))
	}

	if !cache.IsValid() {
		t.Error("Cache should be valid")
	}

	if cache.Markets[0].Symbol != "BTC/USDT" {
		t.Errorf("First market symbol = %v, want BTC/USDT", cache.Markets[0].Symbol)
	}

	if cache.Markets[1].Symbol != "ETH/USDT" {
		t.Errorf("Second market symbol = %v, want ETH/USDT", cache.Markets[1].Symbol)
	}
}
