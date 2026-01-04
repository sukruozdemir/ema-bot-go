package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid spot config",
			config: Config{
				Emas:       []int{50, 100, 200},
				Exchange:   "binance",
				Timeframes: []string{"1h", "4h"},
				MarketType: "spot",
				SelectAll:  true,
			},
			wantErr: false,
		},
		{
			name: "valid swap config with symbols",
			config: Config{
				Emas:       []int{50, 100},
				Exchange:   "binance",
				Timeframes: []string{"1h"},
				MarketType: "swap",
				SelectAll:  false,
				Symbols:    []string{"BTC", "ETH"},
			},
			wantErr: false,
		},
		{
			name: "empty EMAs",
			config: Config{
				Emas:       []int{},
				Exchange:   "binance",
				Timeframes: []string{"1h"},
				MarketType: "spot",
				SelectAll:  true,
			},
			wantErr: true,
		},
		{
			name: "negative EMA",
			config: Config{
				Emas:       []int{50, -100},
				Exchange:   "binance",
				Timeframes: []string{"1h"},
				MarketType: "spot",
				SelectAll:  true,
			},
			wantErr: true,
		},
		{
			name: "zero EMA",
			config: Config{
				Emas:       []int{0},
				Exchange:   "binance",
				Timeframes: []string{"1h"},
				MarketType: "spot",
				SelectAll:  true,
			},
			wantErr: true,
		},
		{
			name: "empty exchange",
			config: Config{
				Emas:       []int{50},
				Exchange:   "",
				Timeframes: []string{"1h"},
				MarketType: "spot",
				SelectAll:  true,
			},
			wantErr: true,
		},
		{
			name: "empty timeframes",
			config: Config{
				Emas:       []int{50},
				Exchange:   "binance",
				Timeframes: []string{},
				MarketType: "spot",
				SelectAll:  true,
			},
			wantErr: true,
		},
		{
			name: "invalid market type",
			config: Config{
				Emas:       []int{50},
				Exchange:   "binance",
				Timeframes: []string{"1h"},
				MarketType: "invalid",
				SelectAll:  true,
			},
			wantErr: true,
		},
		{
			name: "no symbols when not select all",
			config: Config{
				Emas:       []int{50},
				Exchange:   "binance",
				Timeframes: []string{"1h"},
				MarketType: "spot",
				SelectAll:  false,
				Symbols:    []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Set environment variable to use temp path
	os.Setenv("EMA_BOT_CONFIG", configPath)
	defer os.Unsetenv("EMA_BOT_CONFIG")

	// Create a valid config
	originalConfig := Config{
		Emas:       []int{50, 100, 200},
		Exchange:   "binance",
		Timeframes: []string{"1h", "4h", "1d"},
		MarketType: "spot",
		SelectAll:  false,
		Symbols:    []string{"BTC", "ETH", "ADA"},
	}

	// Save config
	err := Save(originalConfig)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load config
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify loaded config matches original
	if len(loadedConfig.Emas) != len(originalConfig.Emas) {
		t.Errorf("Emas length mismatch: got %d, want %d", len(loadedConfig.Emas), len(originalConfig.Emas))
	}

	for i, ema := range loadedConfig.Emas {
		if ema != originalConfig.Emas[i] {
			t.Errorf("EMA at index %d: got %d, want %d", i, ema, originalConfig.Emas[i])
		}
	}

	if loadedConfig.Exchange != originalConfig.Exchange {
		t.Errorf("Exchange: got %s, want %s", loadedConfig.Exchange, originalConfig.Exchange)
	}

	if loadedConfig.MarketType != originalConfig.MarketType {
		t.Errorf("MarketType: got %s, want %s", loadedConfig.MarketType, originalConfig.MarketType)
	}

	if loadedConfig.SelectAll != originalConfig.SelectAll {
		t.Errorf("SelectAll: got %v, want %v", loadedConfig.SelectAll, originalConfig.SelectAll)
	}

	// Verify SavedAt was set
	if loadedConfig.SavedAt.IsZero() {
		t.Error("SavedAt should be set")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	// Use non-existent path
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.json")

	os.Setenv("EMA_BOT_CONFIG", configPath)
	defer os.Unsetenv("EMA_BOT_CONFIG")

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for non-existent file")
	}
}

func TestSave_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	os.Setenv("EMA_BOT_CONFIG", configPath)
	defer os.Unsetenv("EMA_BOT_CONFIG")

	// Create invalid config
	invalidConfig := Config{
		Emas:       []int{},
		Exchange:   "",
		Timeframes: []string{},
		MarketType: "invalid",
	}

	err := Save(invalidConfig)
	if err == nil {
		t.Error("Save() should return error for invalid config")
	}
}

func TestConfigPath(t *testing.T) {
	// Test with environment variable
	expectedPath := "/custom/path/config.json"
	os.Setenv("EMA_BOT_CONFIG", expectedPath)
	defer os.Unsetenv("EMA_BOT_CONFIG")

	path := ConfigPath()
	if path != expectedPath {
		t.Errorf("ConfigPath() = %s, want %s", path, expectedPath)
	}

	// Test without environment variable
	os.Unsetenv("EMA_BOT_CONFIG")
	path = ConfigPath()
	if path == "" {
		t.Error("ConfigPath() should return non-empty path")
	}
}

func TestConfig_SavedAt(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	os.Setenv("EMA_BOT_CONFIG", configPath)
	defer os.Unsetenv("EMA_BOT_CONFIG")

	config := Config{
		Emas:       []int{50},
		Exchange:   "binance",
		Timeframes: []string{"1h"},
		MarketType: "spot",
		SelectAll:  true,
	}

	before := time.Now()
	err := Save(config)
	after := time.Now()

	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify SavedAt is within expected range
	if loadedConfig.SavedAt.Before(before) || loadedConfig.SavedAt.After(after) {
		t.Errorf("SavedAt timestamp out of expected range: got %v, want between %v and %v",
			loadedConfig.SavedAt, before, after)
	}
}
