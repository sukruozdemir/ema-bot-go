package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sukruozdemir/ema-bot-go/internal/config"
)

func main() {
	var (
		configPath = flag.String("config", "", "Path to configuration file")
		showConfig = flag.Bool("show-config", false, "Show current configuration")
		validate   = flag.Bool("validate", false, "Validate configuration")
	)
	flag.Parse()

	if *configPath != "" {
		os.Setenv("EMA_BOT_CONFIG", *configPath)
	}

	if *showConfig {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Configuration loaded from: %s\n", config.ConfigPath())
		fmt.Printf("Exchange: %s\n", cfg.Exchange)
		fmt.Printf("Market Type: %s\n", cfg.MarketType)
		fmt.Printf("EMA Lengths: %v\n", cfg.Emas)
		fmt.Printf("Timeframes: %v\n", cfg.Timeframes)
		if cfg.SelectAll {
			fmt.Println("Symbols: All")
		} else {
			fmt.Printf("Symbols: %v\n", cfg.Symbols)
		}
		fmt.Printf("Saved At: %s\n", cfg.SavedAt.Format("2006-01-02 15:04:05"))
		return
	}

	if *validate {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		if err := cfg.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Configuration is valid ✓")
		return
	}

	fmt.Println("EMA Bot Configuration Tool")
	fmt.Println("Usage:")
	fmt.Println("  -config=path     Set custom config file path")
	fmt.Println("  -show-config     Display current configuration")
	fmt.Println("  -validate        Validate current configuration")
	fmt.Println()
	fmt.Println("Run the main application with: go run ./cmd/app")
}
