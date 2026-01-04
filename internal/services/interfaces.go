package services

import (
	"context"

	"github.com/sukruozdemir/ema-bot-go/internal/models"
)

// ExchangeInterface defines the contract for exchange operations
type ExchangeInterface interface {
	// Markets
	GetSpotMarkets(ctx context.Context) ([]models.Market, error)
	GetSwapMarkets(ctx context.Context) ([]models.Market, error)
	GetSelectedMarkets(symbols []string, markets []models.Market) []models.Market

	// Historical data
	FetchOhlcvWithDataCount(ctx context.Context, market models.Market, timeframe string, requestedCount int) ([][]float64, error)

	// Cache management
	ClearCache() error

	// Validation
	IsValidExchange() bool
}

// CacheInterface defines the contract for caching operations
type CacheInterface interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}) error
	Delete(key string) error
	Clear() error
}
