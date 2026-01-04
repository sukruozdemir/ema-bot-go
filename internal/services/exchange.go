package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	ccxt "github.com/ccxt/ccxt/go/v4"
	"go.uber.org/zap"

	"github.com/sukruozdemir/ema-bot-go/internal/errors"
	"github.com/sukruozdemir/ema-bot-go/internal/models"
)

const (
	DefaultCacheExpiry = 24 * time.Hour
	DefaultQuoteFilter = "USDT"
)

// ExchangeService provides exchange-related operations
type ExchangeService struct {
	exchange     ccxt.ICoreExchange
	exchangeName string
	logger       *zap.Logger
	cacheExpiry  time.Duration
}

// NewExchangeService creates a new exchange service with the given exchange name
func NewExchangeService(exchangeName string, logger *zap.Logger) (*ExchangeService, error) {
	if exchangeName == "" {
		return nil, errors.New(errors.ErrTypeValidation, "exchange name cannot be empty")
	}

	exchange, ok := ccxt.DynamicallyCreateInstance(exchangeName, map[string]any{})
	if !ok {
		return nil, errors.Wrap(errors.ErrTypeExchange,
			fmt.Sprintf("failed to create exchange instance: %s", exchangeName),
			errors.ErrExchangeNotFound)
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &ExchangeService{
		exchange:     exchange,
		exchangeName: exchangeName,
		logger:       logger,
		cacheExpiry:  DefaultCacheExpiry,
	}, nil
}

// FetchOhlcvWithDataCount fetches historical OHLCV data for a market symbol with a lookback count.
// It will iterate requests from earliest to latest until requestedCount candles are fetched
// or until the current time is reached. Respects exchange rate limits and context cancellation.
func (es *ExchangeService) FetchOhlcvWithDataCount(ctx context.Context, market models.Market, timeframe string, requestedCount int) ([][]float64, error) {
	if es.exchange == nil {
		return nil, errors.New(errors.ErrTypeExchange, "exchange not initialized")
	}

	if requestedCount <= 0 {
		return nil, errors.New(errors.ErrTypeValidation, "requested data count must be > 0")
	}

	// parse timeframe via ccxt helper (may return interface{}) - expect numeric seconds
	tfRaw := es.exchange.ParseTimeframe(timeframe)
	var tfSecondsInt int64
	switch v := tfRaw.(type) {
	case int:
		tfSecondsInt = int64(v)
	case int64:
		tfSecondsInt = v
	case float64:
		tfSecondsInt = int64(v)
	default:
		return nil, errors.New(errors.ErrTypeValidation, "invalid timeframe type returned by exchange")
	}
	if tfSecondsInt <= 0 {
		return nil, errors.New(errors.ErrTypeValidation, "invalid timeframe")
	}
	timeframeMs := int64(tfSecondsInt) * 1000

	symbolID := market.Symbol

	nowMs := time.Now().UnixMilli()
	startDate := nowMs - int64(requestedCount)*timeframeMs
	endDate := nowMs

	var allData [][]float64

	es.logger.Info("Fetching OHLCV with lookback",
		zap.String("symbol", symbolID),
		zap.String("timeframe", timeframe),
		zap.Int("requested_count", requestedCount))

	for startDate < endDate {
		select {
		case <-ctx.Done():
			return allData, ctx.Err()
		default:
		}

		// ccxt fetch_ohlcv signature: FetchOHLCV(symbol, timeframe, since, limit)
		// The binding returns a channel of interface{} - receive and convert
		respCh := es.exchange.FetchOHLCV(symbolID, timeframe, startDate, nil)
		var rawResp any
		select {
		case rawResp = <-respCh:
		case <-ctx.Done():
			return allData, ctx.Err()
		}

		ohlcvData, err := convertToFloatMatrix(rawResp)
		if err != nil {
			es.logger.Debug("Failed to convert OHLCV response", zap.Error(err))
			return allData, errors.Wrap(errors.ErrTypeExchange, "failed to parse ohlcv response", err)
		}

		if len(ohlcvData) > 0 {
			allData = append(allData, ohlcvData...)
			// advance startDate to last timestamp + 1ms
			last := ohlcvData[len(ohlcvData)-1]
			if len(last) > 0 {
				ts := int64(last[0])
				startDate = ts + 1
			} else {
				startDate += timeframeMs
			}

			es.logger.Debug("Fetched OHLCV batch",
				zap.Int("batch_count", len(ohlcvData)),
				zap.Int("total_count", len(allData)))
		} else {
			// No data returned for this window, advance by one timeframe
			startDate += timeframeMs
		}

		// apply a conservative default rate-limit sleep (ms)
		sleepMs := 250 * time.Millisecond
		select {
		case <-ctx.Done():
			return allData, ctx.Err()
		case <-time.After(sleepMs):
		}
	}

	es.logger.Info("Completed OHLCV fetch",
		zap.String("symbol", symbolID),
		zap.Int("total_candles", len(allData)))

	return allData, nil
}

// GetSpotMarkets fetches and filters spot markets, using cache if available
func (es *ExchangeService) GetSpotMarkets(ctx context.Context) ([]models.Market, error) {
	return es.getMarketsWithCache(ctx, "spot")
}

// GetSwapMarkets fetches and filters swap/futures markets, using cache if available
func (es *ExchangeService) GetSwapMarkets(ctx context.Context) ([]models.Market, error) {
	return es.getMarketsWithCache(ctx, "swap")
}

// getMarketsWithCache loads markets from cache or fetches and caches them
func (es *ExchangeService) getMarketsWithCache(ctx context.Context, marketType string) ([]models.Market, error) {
	// Try to load from cache first
	cached, err := es.loadMarketCache(marketType)
	if err == nil && cached != nil && cached.IsValid() {
		es.logger.Info("Using cached markets",
			zap.String("market_type", marketType),
			zap.Int("count", len(cached.Markets)))
		return cached.Markets, nil
	}

	// Cache miss or expired - fetch from exchange
	es.logger.Info("Fetching markets from exchange",
		zap.String("exchange", es.exchangeName),
		zap.String("market_type", marketType))

	markets, err := es.fetchMarkets(ctx, marketType)
	if err != nil {
		return nil, err
	}

	// Save to cache
	if err := es.saveMarketCache(marketType, markets); err != nil {
		es.logger.Warn("Failed to cache markets",
			zap.String("market_type", marketType),
			zap.Error(err))
	}

	return markets, nil
}

// fetchMarkets fetches markets directly from the exchange
func (es *ExchangeService) fetchMarkets(ctx context.Context, marketType string) ([]models.Market, error) {
	marketsChan := es.exchange.LoadMarkets()
	var filtered []models.Market

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case syncMapItem := <-marketsChan:
		syncMap, ok := syncMapItem.(*sync.Map)
		if !ok {
			return nil, errors.New(errors.ErrTypeExchange,
				fmt.Sprintf("unexpected market data type: %T", syncMapItem))
		}

		// Iterate through all entries in the sync.Map
		syncMap.Range(func(key, value any) bool {
			select {
			case <-ctx.Done():
				return false // Stop iteration
			default:
			}

			rawMarket, ok := value.(map[string]any)
			if !ok {
				return true // Continue iteration
			}

			market, err := es.convertToMarket(rawMarket)
			if err != nil {
				es.logger.Debug("Failed to convert market",
					zap.Any("key", key),
					zap.Error(err))
				return true
			}

			if es.shouldIncludeMarket(market, marketType) {
				filtered = append(filtered, market)
			}

			return true
		})
	}

	if len(filtered) == 0 {
		return nil, errors.Wrap(errors.ErrTypeExchange,
			fmt.Sprintf("no %s markets found matching criteria", marketType),
			errors.ErrNoMarketsFound)
	}

	es.logger.Info("Markets fetched successfully",
		zap.String("market_type", marketType),
		zap.Int("count", len(filtered)))

	return filtered, nil
}

// convertToMarket converts a raw market map to our Market model
func (es *ExchangeService) convertToMarket(rawMarket map[string]any) (models.Market, error) {
	market := models.Market{}

	// Helper function to safely get string values
	getString := func(key string) string {
		if val, exists := rawMarket[key]; exists {
			if str, ok := val.(string); ok {
				return str
			}
		}
		return ""
	}

	// Helper function to safely get bool values
	getBool := func(key string) bool {
		if val, exists := rawMarket[key]; exists {
			if b, ok := val.(bool); ok {
				return b
			}
		}
		return false
	}

	market.ID = getString("id")
	market.Symbol = getString("symbol")
	market.Base = getString("base")
	market.Quote = getString("quote")
	market.Active = getBool("active")
	market.Type = getString("type")
	market.Spot = getBool("spot")
	market.Swap = getBool("swap")
	market.Future = getBool("future")
	market.Option = getBool("option")
	market.Contract = getBool("contract")
	market.Linear = getBool("linear")
	market.Inverse = getBool("inverse")

	// Try to extract price precision from common CCXT fields
	if val, exists := rawMarket["precision"]; exists {
		if pm, ok := val.(map[string]any); ok {
			if pval, ok := pm["price"].(int); ok {
				market.PricePrecision = pval
			} else if pvalf, ok := pm["price"].(float64); ok {
				market.PricePrecision = int(pvalf)
			}
		}
	}

	// Validate required fields
	if market.Symbol == "" || market.Base == "" || market.Quote == "" {
		return market, errors.New(errors.ErrTypeValidation, "invalid market data: missing required fields")
	}

	return market, nil
}

// shouldIncludeMarket determines if a market should be included based on criteria
func (es *ExchangeService) shouldIncludeMarket(market models.Market, marketType string) bool {
	// Check if market is active
	if !market.Active {
		return false
	}

	// Check if quote currency matches our filter
	if market.Quote != DefaultQuoteFilter {
		return false
	}

	// Check market type
	switch marketType {
	case "spot":
		return market.Spot || market.Type == "spot"
	case "swap":
		return market.Swap || market.Type == "swap"
	default:
		return false
	}
}

// GetSelectedMarkets filters markets to only include those whose base currency is in the selection
func (es *ExchangeService) GetSelectedMarkets(symbols []string, markets []models.Market) []models.Market {
	if len(symbols) == 0 {
		return markets
	}

	symbolSet := make(map[string]bool, len(symbols))
	for _, symbol := range symbols {
		symbolSet[symbol] = true
	}

	filtered := make([]models.Market, 0, len(markets))
	for _, market := range markets {
		if symbolSet[market.Base] {
			filtered = append(filtered, market)
		}
	}

	es.logger.Info("Markets filtered by symbol selection",
		zap.Strings("symbols", symbols),
		zap.Int("original_count", len(markets)),
		zap.Int("filtered_count", len(filtered)))

	return filtered
}

// cacheDir returns the directory for market cache files
func (es *ExchangeService) cacheDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	cacheDir := filepath.Join(configDir, "ema-bot", "cache", es.exchangeName)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", errors.Wrap(errors.ErrTypeCache, "failed to create cache directory", err)
	}
	return cacheDir, nil
}

// saveMarketCache saves markets to JSON cache file
func (es *ExchangeService) saveMarketCache(marketType string, markets []models.Market) error {
	dir, err := es.cacheDir()
	if err != nil {
		return err
	}

	cache := &models.MarketCache{
		Timestamp: time.Now(),
		Markets:   markets,
		ExpiresAt: time.Now().Add(es.cacheExpiry),
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return errors.Wrap(errors.ErrTypeCache, "failed to marshal cache data", err)
	}

	filename := filepath.Join(dir, fmt.Sprintf("%s-markets.json", marketType))
	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return errors.Wrap(errors.ErrTypeCache, "failed to write cache file", err)
	}

	return nil
}

// loadMarketCache loads markets from JSON cache file
func (es *ExchangeService) loadMarketCache(marketType string) (*models.MarketCache, error) {
	dir, err := es.cacheDir()
	if err != nil {
		return nil, err
	}

	filename := filepath.Join(dir, fmt.Sprintf("%s-markets.json", marketType))
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.ErrTypeCache, "cache file not found")
		}
		return nil, errors.Wrap(errors.ErrTypeCache, "failed to read cache file", err)
	}

	var cache models.MarketCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, errors.Wrap(errors.ErrTypeCache, "failed to unmarshal cache data", err)
	}

	return &cache, nil
}

// ClearCache removes all cached markets for this exchange
func (es *ExchangeService) ClearCache() error {
	dir, err := es.cacheDir()
	if err != nil {
		return err
	}

	if err := os.RemoveAll(dir); err != nil {
		return errors.Wrap(errors.ErrTypeCache, "failed to clear cache", err)
	}

	es.logger.Info("Cache cleared successfully", zap.String("exchange", es.exchangeName))
	return nil
}

// IsValidExchange checks if the exchange is valid and accessible
func (es *ExchangeService) IsValidExchange() bool {
	return es.exchange != nil
}

// convertToFloatMatrix attempts to convert various ccxt OHLCV responses to [][]float64
func convertToFloatMatrix(raw interface{}) ([][]float64, error) {
	// Common shapes: []interface{} where each entry is []interface{} of numbers
	var out [][]float64

	switch v := raw.(type) {
	case [][]float64:
		return v, nil
	case []interface{}:
		for _, row := range v {
			switch r := row.(type) {
			case []interface{}:
				var nums []float64
				for _, cell := range r {
					switch c := cell.(type) {
					case float64:
						nums = append(nums, c)
					case int:
						nums = append(nums, float64(c))
					case int64:
						nums = append(nums, float64(c))
					case json.Number:
						f, err := r[0].(json.Number).Float64()
						if err != nil {
							return nil, err
						}
						nums = append(nums, f)
					default:
						// try to convert using fmt
						return nil, fmt.Errorf("unsupported ohlcv cell type: %T", c)
					}
				}
				out = append(out, nums)
			default:
				return nil, fmt.Errorf("unsupported ohlcv row type: %T", r)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported ohlcv response type: %T", v)
	}
}
