package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sukruozdemir/ema-bot-go/internal/models"
)

func TestNewManager(t *testing.T) {
	manager := NewManager("", 0)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.cacheDir == "" {
		t.Error("cacheDir should not be empty")
	}

	if manager.ttl != 24*time.Hour {
		t.Errorf("default ttl = %v, want 24h", manager.ttl)
	}

	// Test with custom values
	customDir := "/custom/cache"
	customTTL := 12 * time.Hour
	manager = NewManager(customDir, customTTL)

	if manager.cacheDir != customDir {
		t.Errorf("cacheDir = %v, want %v", manager.cacheDir, customDir)
	}

	if manager.ttl != customTTL {
		t.Errorf("ttl = %v, want %v", manager.ttl, customTTL)
	}
}

func TestManager_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir, 24*time.Hour)

	markets := []models.Market{
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

	// Save markets
	err := manager.Save("binance", "spot", markets)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify cache file exists
	cachePath := manager.GetCachePath("binance", "spot")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("cache file was not created")
	}

	// Load markets
	cache, err := manager.Load("binance", "spot")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cache.Markets) != len(markets) {
		t.Errorf("loaded %d markets, want %d", len(cache.Markets), len(markets))
	}

	for i, market := range cache.Markets {
		if market.Symbol != markets[i].Symbol {
			t.Errorf("market[%d] symbol = %v, want %v", i, market.Symbol, markets[i].Symbol)
		}
	}

	if !cache.IsValid() {
		t.Error("cache should be valid")
	}
}

func TestManager_LoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir, 24*time.Hour)

	_, err := manager.Load("binance", "spot")
	if err == nil {
		t.Error("Load() should return error for non-existent cache")
	}
}

func TestManager_LoadExpired(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir, 1*time.Millisecond) // Very short TTL

	markets := []models.Market{
		{ID: "BTC/USDT", Symbol: "BTC/USDT"},
	}

	// Save markets
	err := manager.Save("binance", "spot", markets)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Wait for cache to expire
	time.Sleep(10 * time.Millisecond)

	// Try to load expired cache
	_, err = manager.Load("binance", "spot")
	if err == nil {
		t.Error("Load() should return error for expired cache")
	}
}

func TestManager_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir, 24*time.Hour)

	markets := []models.Market{
		{ID: "BTC/USDT", Symbol: "BTC/USDT"},
	}

	// Save markets
	err := manager.Save("binance", "spot", markets)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Clear cache
	err = manager.Clear("binance", "spot")
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	// Verify cache file is removed
	cachePath := manager.GetCachePath("binance", "spot")
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Error("cache file should be removed")
	}

	// Clear again should not error
	err = manager.Clear("binance", "spot")
	if err != nil {
		t.Errorf("Clear() on non-existent cache should not error, got %v", err)
	}
}

func TestManager_ClearAll(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir, 24*time.Hour)

	markets := []models.Market{
		{ID: "BTC/USDT", Symbol: "BTC/USDT"},
	}

	// Save multiple caches
	manager.Save("binance", "spot", markets)
	manager.Save("binance", "swap", markets)
	manager.Save("kraken", "spot", markets)

	// Clear all
	err := manager.ClearAll()
	if err != nil {
		t.Fatalf("ClearAll() error = %v", err)
	}

	// Verify all cache files are removed
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Error("cache directory should be removed")
	}

	// ClearAll again should not error
	err = manager.ClearAll()
	if err != nil {
		t.Errorf("ClearAll() on non-existent cache should not error, got %v", err)
	}
}

func TestManager_GetCacheInfo(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir, 24*time.Hour)

	// Get info for non-existent cache
	info, err := manager.GetCacheInfo("binance", "spot")
	if err != nil {
		t.Fatalf("GetCacheInfo() error = %v", err)
	}

	if info.Exists {
		t.Error("cache should not exist")
	}

	// Save markets
	markets := []models.Market{
		{ID: "BTC/USDT", Symbol: "BTC/USDT"},
		{ID: "ETH/USDT", Symbol: "ETH/USDT"},
	}

	err = manager.Save("binance", "spot", markets)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Get info for existing cache
	info, err = manager.GetCacheInfo("binance", "spot")
	if err != nil {
		t.Fatalf("GetCacheInfo() error = %v", err)
	}

	if !info.Exists {
		t.Error("cache should exist")
	}

	if !info.IsValid {
		t.Error("cache should be valid")
	}

	if info.MarketCount != 2 {
		t.Errorf("MarketCount = %d, want 2", info.MarketCount)
	}

	if info.Size == 0 {
		t.Error("cache file size should not be 0")
	}

	if info.TimeToExpiry <= 0 {
		t.Error("TimeToExpiry should be positive")
	}
}

func TestManager_GetCachePath(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir, 24*time.Hour)

	path := manager.GetCachePath("binance", "spot")
	expected := filepath.Join(tmpDir, "binance_spot_markets.json")

	if path != expected {
		t.Errorf("GetCachePath() = %v, want %v", path, expected)
	}
}
