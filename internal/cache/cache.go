package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/sukruozdemir/ema-bot-go/internal/errors"
	"github.com/sukruozdemir/ema-bot-go/internal/models"
)

// Manager handles caching operations for market data
type Manager struct {
	cacheDir string
	ttl      time.Duration
}

// NewManager creates a new cache manager
func NewManager(cacheDir string, ttl time.Duration) *Manager {
	if cacheDir == "" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			cacheDir = filepath.Join(homeDir, ".cache", "ema-bot")
		} else {
			cacheDir = filepath.Join(os.TempDir(), "ema-bot-cache")
		}
	}

	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	return &Manager{
		cacheDir: cacheDir,
		ttl:      ttl,
	}
}

// GetCachePath returns the cache file path for a given exchange and market type
func (m *Manager) GetCachePath(exchange, marketType string) string {
	filename := exchange + "_" + marketType + "_markets.json"
	return filepath.Join(m.cacheDir, filename)
}

// Load loads cached market data
func (m *Manager) Load(exchange, marketType string) (*models.MarketCache, error) {
	cachePath := m.GetCachePath(exchange, marketType)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.ErrTypeCache, "cache file not found")
		}
		return nil, errors.Wrap(errors.ErrTypeCache, "failed to read cache file", err)
	}

	var cache models.MarketCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, errors.Wrap(errors.ErrTypeCache, "failed to parse cache file", err)
	}

	if !cache.IsValid() {
		return nil, errors.New(errors.ErrTypeCache, "cache has expired")
	}

	return &cache, nil
}

// Save saves market data to cache
func (m *Manager) Save(exchange, marketType string, markets []models.Market) error {
	if err := os.MkdirAll(m.cacheDir, 0755); err != nil {
		return errors.Wrap(errors.ErrTypeCache, "failed to create cache directory", err)
	}

	cache := models.MarketCache{
		Timestamp: time.Now(),
		Markets:   markets,
		ExpiresAt: time.Now().Add(m.ttl),
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return errors.Wrap(errors.ErrTypeCache, "failed to marshal cache data", err)
	}

	cachePath := m.GetCachePath(exchange, marketType)
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return errors.Wrap(errors.ErrTypeCache, "failed to write cache file", err)
	}

	return nil
}

// Clear removes cached data for a specific exchange and market type
func (m *Manager) Clear(exchange, marketType string) error {
	cachePath := m.GetCachePath(exchange, marketType)

	if err := os.Remove(cachePath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already cleared
		}
		return errors.Wrap(errors.ErrTypeCache, "failed to remove cache file", err)
	}

	return nil
}

// ClearAll removes all cached data
func (m *Manager) ClearAll() error {
	if err := os.RemoveAll(m.cacheDir); err != nil {
		if os.IsNotExist(err) {
			return nil // Already cleared
		}
		return errors.Wrap(errors.ErrTypeCache, "failed to clear all cache", err)
	}
	return nil
}

// GetCacheInfo returns information about cached data
func (m *Manager) GetCacheInfo(exchange, marketType string) (*CacheInfo, error) {
	cachePath := m.GetCachePath(exchange, marketType)

	info, err := os.Stat(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &CacheInfo{Exists: false}, nil
		}
		return nil, errors.Wrap(errors.ErrTypeCache, "failed to get cache info", err)
	}

	cache, err := m.Load(exchange, marketType)
	if err != nil {
		return &CacheInfo{
			Exists:    true,
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			IsValid:   false,
			ExpiresAt: time.Time{},
		}, nil
	}

	return &CacheInfo{
		Exists:       true,
		Size:         info.Size(),
		ModTime:      info.ModTime(),
		IsValid:      cache.IsValid(),
		MarketCount:  len(cache.Markets),
		ExpiresAt:    cache.ExpiresAt,
		TimeToExpiry: time.Until(cache.ExpiresAt),
	}, nil
}

// CacheInfo contains information about a cache file
type CacheInfo struct {
	Exists       bool
	Size         int64
	ModTime      time.Time
	IsValid      bool
	MarketCount  int
	ExpiresAt    time.Time
	TimeToExpiry time.Duration
}
