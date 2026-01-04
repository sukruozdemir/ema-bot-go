package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sukruozdemir/ema-bot-go/internal/errors"
)

// Config represents the application configuration
type Config struct {
	Emas       []int     `json:"emas" validate:"required,min=1"`
	Exchange   string    `json:"exchange" validate:"required"`
	Timeframes []string  `json:"timeframes" validate:"required,min=1"`
	MarketType string    `json:"market_type" validate:"required,oneof=spot swap"`
	SelectAll  bool      `json:"select_all"`
	Symbols    []string  `json:"symbols,omitempty"`
	SavedAt    time.Time `json:"saved_at"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Emas) == 0 {
		return errors.New(errors.ErrTypeValidation, "at least one EMA length must be specified")
	}

	for _, ema := range c.Emas {
		if ema <= 0 {
			return errors.Wrap(errors.ErrTypeValidation, "EMA length must be positive", errors.ErrInvalidEMALength)
		}
	}

	if c.Exchange == "" {
		return errors.New(errors.ErrTypeValidation, "exchange must be specified")
	}

	if len(c.Timeframes) == 0 {
		return errors.New(errors.ErrTypeValidation, "at least one timeframe must be specified")
	}

	if c.MarketType != "spot" && c.MarketType != "swap" {
		return errors.Wrap(errors.ErrTypeValidation, "market type must be 'spot' or 'swap'", errors.ErrInvalidMarketType)
	}

	if !c.SelectAll && len(c.Symbols) == 0 {
		return errors.New(errors.ErrTypeValidation, "symbols must be specified when not selecting all markets")
	}

	return nil
}

// ConfigPath returns the path where the configuration file should be stored
func ConfigPath() string {
	if p := os.Getenv("EMA_BOT_CONFIG"); p != "" {
		return p
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "ema-bot", "config.json")
	}
	return filepath.Join(os.TempDir(), "ema-bot-config.json")
}

// Save persists the configuration to disk
func Save(cfg Config) error {
	cfg.SavedAt = time.Now()

	if err := cfg.Validate(); err != nil {
		return errors.Wrap(errors.ErrTypeConfig, "invalid configuration", err)
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return errors.Wrap(errors.ErrTypeConfig, "failed to marshal configuration", err)
	}

	path := ConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.Wrap(errors.ErrTypeConfig, "failed to create config directory", err)
	}

	if err := os.WriteFile(path, b, 0o644); err != nil {
		return errors.Wrap(errors.ErrTypeConfig, "failed to write config file", err)
	}

	fmt.Printf("Configuration saved to %s\n", path)
	return nil
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	path := ConfigPath()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.ErrTypeConfig, "configuration file not found")
		}
		return nil, errors.Wrap(errors.ErrTypeConfig, "failed to read config file", err)
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, errors.Wrap(errors.ErrTypeConfig, "failed to parse config file", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(errors.ErrTypeConfig, "invalid configuration loaded", err)
	}

	return &cfg, nil
}
