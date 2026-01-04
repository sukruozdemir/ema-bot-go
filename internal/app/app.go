package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/sukruozdemir/ema-bot-go/internal/config"
	"github.com/sukruozdemir/ema-bot-go/internal/errors"
	"github.com/sukruozdemir/ema-bot-go/internal/export"
	"github.com/sukruozdemir/ema-bot-go/internal/indicators"
	"github.com/sukruozdemir/ema-bot-go/internal/input"
	"github.com/sukruozdemir/ema-bot-go/internal/models"
	"github.com/sukruozdemir/ema-bot-go/internal/services"
	"github.com/sukruozdemir/ema-bot-go/internal/ui"
)

// App represents the main application
type App struct {
	logger          *zap.Logger
	reader          *bufio.Reader
	config          *config.Config
	exchangeService services.ExchangeInterface
}

// NewApp creates a new application instance
func NewApp(logger *zap.Logger) *App {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &App{
		logger: logger,
		reader: bufio.NewReader(os.Stdin),
	}
}

// Run starts the application
func Run(ctx context.Context, logger *zap.Logger) error {
	app := NewApp(logger)
	return app.Start(ctx)
}

// Start initializes and runs the application
func (a *App) Start(ctx context.Context) error {
	a.logger.Info("Starting EMA Bot application")

	ui.PrintWelcome()

	// Load or create configuration
	if err := a.loadOrCreateConfig(ctx); err != nil {
		ui.PrintError("Failed to load configuration", err)
		return err
	}

	// Initialize exchange service
	if err := a.initializeExchange(); err != nil {
		ui.PrintError("Failed to initialize exchange", err)
		return err
	}

	// Ask user about cache clearing
	if a.shouldClearCache() {
		if err := a.exchangeService.ClearCache(); err != nil {
			ui.PrintWarning(fmt.Sprintf("Failed to clear cache: %v", err))
		} else {
			ui.PrintSuccess("Cache cleared successfully")
		}
	}

	// Fetch and process markets
	if err := a.processMarkets(ctx); err != nil {
		ui.PrintError("Failed to process markets", err)
		return err
	}

	ui.PrintSuccess("Application completed successfully")
	return nil
}

// loadOrCreateConfig attempts to load saved config or creates a new one interactively
func (a *App) loadOrCreateConfig(ctx context.Context) error {
	cfg, err := config.Load()
	if err == nil && a.confirmUseExistingConfig(cfg) {
		a.config = cfg
		ui.PrintInfo("Using saved configuration")
		return nil
	}

	// Create new configuration interactively
	newConfig, err := a.createConfigInteractively()
	if err != nil {
		return errors.Wrap(errors.ErrTypeConfig, "failed to create configuration", err)
	}

	a.config = newConfig
	return nil
}

// confirmUseExistingConfig shows existing config and asks user for confirmation
func (a *App) confirmUseExistingConfig(cfg *config.Config) bool {
	return ui.PrettyPrintConfig(cfg, a.reader)
}

// createConfigInteractively prompts user for all configuration values
func (a *App) createConfigInteractively() (*config.Config, error) {
	a.logger.Info("Creating new configuration interactively")

	// EMA Lengths
	emas, err := a.readEMALengths()
	if err != nil {
		return nil, err
	}

	// Exchange
	exchange, err := a.readExchange()
	if err != nil {
		return nil, err
	}

	// Timeframes
	timeframes, err := a.readTimeframes()
	if err != nil {
		return nil, err
	}

	// Market Type
	marketType := a.readMarketType()

	// Symbols
	selectAll, symbols := a.readSymbols()

	newConfig := &config.Config{
		Emas:       emas,
		Exchange:   exchange,
		Timeframes: timeframes,
		MarketType: marketType,
		SelectAll:  selectAll,
		Symbols:    symbols,
	}

	// Validate configuration
	if err := newConfig.Validate(); err != nil {
		return nil, errors.Wrap(errors.ErrTypeConfig, "configuration validation failed", err)
	}

	// Save configuration
	if err := config.Save(*newConfig); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to save configuration: %v", err))
	} else {
		ui.PrintSuccess("Configuration saved successfully")
	}

	return newConfig, nil
}

// readEMALengths prompts user for EMA lengths with validation
func (a *App) readEMALengths() ([]int, error) {
	for {
		fmt.Println("\n📊 EMA Configuration")
		fmt.Println("Please enter the EMA lengths you want to use.")
		fmt.Println("Example: 50, 100, 200")

		inputStr := input.ReadLine(a.reader, "EMA Lengths (default: 200): ")
		if inputStr == "" {
			inputStr = "200"
		}

		emas, err := input.ParseEMALengths(inputStr)
		if err != nil {
			ui.PrintError("Invalid EMA lengths", err)
			continue
		}

		if len(emas) == 0 {
			ui.PrintError("No valid EMA lengths entered", nil)
			continue
		}

		return emas, nil
	}
}

// readExchange prompts user for exchange name with validation
func (a *App) readExchange() (string, error) {
	fmt.Println("\n🏪 Exchange Configuration")
	fmt.Println("Enter the name of the exchange you want to use (e.g., binance)")
	fmt.Println("Note: Only one exchange is supported per configuration")

	for {
		exchangeName := input.ReadLine(a.reader, "Exchange Name (default: binance): ")
		if exchangeName == "" {
			exchangeName = "binance"
		}

		if !input.IsSingleWord(exchangeName) {
			ui.PrintError("Please enter only one exchange name (no commas or spaces)", nil)
			continue
		}

		return strings.ToLower(exchangeName), nil
	}
}

// readTimeframes prompts user for timeframes with validation
func (a *App) readTimeframes() ([]string, error) {
	fmt.Println("\n⏱️  Timeframe Configuration")
	fmt.Println("Enter the timeframes you want to analyze (comma-separated)")
	fmt.Println("Common timeframes: 1m, 5m, 15m, 1h, 4h, 1d, 1w")

	for {
		tfInput := input.ReadLine(a.reader, "Timeframes (default: 1d): ")
		if tfInput == "" {
			tfInput = "1d"
		}

		timeframes, err := input.ParseTimeframes(tfInput)
		if err != nil {
			ui.PrintError("Invalid timeframes", err)
			continue
		}

		return timeframes, nil
	}
}

// readMarketType prompts user to select between spot and swap markets
func (a *App) readMarketType() string {
	fmt.Println("\n💱 Market Type Selection")
	fmt.Println("1. Spot markets (💰 direct trading)")
	fmt.Println("2. Swap/Futures markets (🔄 leveraged trading)")

	for {
		mt := input.Normalize(input.ReadLine(a.reader, "Market Type (1/2 or spot/swap, default: swap): "))
		if mt == "" {
			return "swap"
		}

		switch mt {
		case "1", "spot":
			return "spot"
		case "2", "swap":
			return "swap"
		default:
			ui.PrintError("Please enter '1', '2', 'spot', or 'swap'", nil)
		}
	}
}

// readSymbols prompts user for symbols or 'all'
func (a *App) readSymbols() (bool, []string) {
	fmt.Println("\n🎯 Symbol Selection")
	fmt.Println("Enter specific symbols (e.g., BTC, ETH, ADA) or 'all' for all markets")

	for {
		symInput := input.ReadLine(a.reader, "Symbols (comma-separated or 'all', default: all): ")
		if symInput == "" {
			symInput = "all"
		}

		if input.Normalize(symInput) == "all" {
			return true, nil
		}

		symbols := input.SplitAndTrim(symInput, ",")
		if len(symbols) == 0 {
			ui.PrintError("No valid symbols entered", nil)
			continue
		}

		// Convert to uppercase and validate
		validSymbols := make([]string, 0, len(symbols))
		for _, symbol := range symbols {
			upperSymbol := strings.ToUpper(symbol)
			if input.IsValidSymbol(upperSymbol) {
				validSymbols = append(validSymbols, upperSymbol)
			} else {
				ui.PrintWarning(fmt.Sprintf("Skipping invalid symbol: %s", symbol))
			}
		}

		if len(validSymbols) == 0 {
			ui.PrintError("No valid symbols found", nil)
			continue
		}

		return false, validSymbols
	}
}

// initializeExchange creates and configures the exchange service
func (a *App) initializeExchange() error {
	exchangeService, err := services.NewExchangeService(a.config.Exchange, a.logger)
	if err != nil {
		return errors.Wrap(errors.ErrTypeExchange, "failed to create exchange service", err)
	}

	if !exchangeService.IsValidExchange() {
		return errors.New(errors.ErrTypeExchange, "exchange validation failed")
	}

	a.exchangeService = exchangeService
	a.logger.Info("Exchange service initialized", zap.String("exchange", a.config.Exchange))

	return nil
}

// shouldClearCache asks user if they want to clear the cache
func (a *App) shouldClearCache() bool {
	fmt.Println("\n🗄️  Cache Management")
	response := input.ReadLine(a.reader, "Clear market cache before fetching? (y/n, default: n): ")
	if response == "" {
		return false
	}
	normalized := input.Normalize(response)
	return normalized == "y" || normalized == "yes"
}

// processMarkets fetches and processes markets according to configuration
func (a *App) processMarkets(ctx context.Context) error {
	a.logger.Info("Processing markets",
		zap.String("market_type", a.config.MarketType),
		zap.Bool("select_all", a.config.SelectAll))

	// Fetch markets based on type
	var markets []interface{}

	if a.config.MarketType == "spot" {
		spotMarkets, fetchErr := a.exchangeService.GetSpotMarkets(ctx)
		if fetchErr != nil {
			return errors.Wrap(errors.ErrTypeExchange, "failed to fetch spot markets", fetchErr)
		}
		markets = make([]interface{}, len(spotMarkets))
		for i, m := range spotMarkets {
			markets[i] = m
		}
	} else {
		swapMarkets, fetchErr := a.exchangeService.GetSwapMarkets(ctx)
		if fetchErr != nil {
			return errors.Wrap(errors.ErrTypeExchange, "failed to fetch swap markets", fetchErr)
		}
		markets = make([]interface{}, len(swapMarkets))
		for i, m := range swapMarkets {
			markets[i] = m
		}
	}

	totalMarkets := len(markets)

	// Filter markets if specific symbols are selected
	var filteredMarkets []interface{}
	if !a.config.SelectAll && len(a.config.Symbols) > 0 {
		// Convert back to models.Market for filtering
		modelMarkets := make([]models.Market, len(markets))
		for i, m := range markets {
			if market, ok := m.(models.Market); ok {
				modelMarkets[i] = market
			}
		}

		filtered := a.exchangeService.GetSelectedMarkets(a.config.Symbols, modelMarkets)
		filteredMarkets = make([]interface{}, len(filtered))
		for i, m := range filtered {
			filteredMarkets[i] = m
		}
	} else {
		filteredMarkets = markets
	}

	// Display summary
	ui.PrintMarketSummary(a.config.MarketType, totalMarkets, len(filteredMarkets))

	if len(filteredMarkets) == 0 {
		return errors.Wrap(errors.ErrTypeExchange,
			"no markets found matching selection criteria",
			errors.ErrNoMarketsFound)
	}

	a.logger.Info("Markets processed successfully",
		zap.Int("total_markets", totalMarkets),
		zap.Int("filtered_markets", len(filteredMarkets)))

	// Add a small buffer so we have some lookback beyond the period
	requestedCount := 1000

	ui.PrintInfo(fmt.Sprintf("🔄 Starting analysis for %d markets across %d timeframes", len(filteredMarkets), len(a.config.Timeframes)))

	var processedMarkets int
	var allAnalyses []indicators.MarketAnalysis
	currentOperation := 0

	// For each filtered market, fetch OHLCV for each timeframe and compute EMAs
	for _, im := range filteredMarkets {
		market, ok := im.(models.Market)
		if !ok {
			continue
		}

		processedMarkets++
		ui.PrintMarketProgress(market.Symbol, processedMarkets, len(filteredMarkets))

		for _, tf := range a.config.Timeframes {
			currentOperation++

			// Respect context cancellation
			if ctx.Err() != nil {
				return ctx.Err()
			}

			a.logger.Info("Fetching market OHLCV",
				zap.String("symbol", market.Symbol),
				zap.String("timeframe", tf),
				zap.Int("count", requestedCount))

			ohlcv, err := a.exchangeService.FetchOhlcvWithDataCount(ctx, market, tf, requestedCount)
			if err != nil {
				a.logger.Error("Failed to fetch OHLCV", zap.String("symbol", market.Symbol), zap.String("timeframe", tf), zap.Error(err))
				ui.PrintWarning(fmt.Sprintf("Skipping %s %s due to fetch error", market.Symbol, tf))
				continue
			}

			// Extract close prices
			closes := make([]float64, 0, len(ohlcv))
			for _, row := range ohlcv {
				if len(row) >= 5 {
					closes = append(closes, row[4])
				}
			}

			if len(closes) == 0 {
				ui.PrintWarning(fmt.Sprintf("No close prices for %s %s", market.Symbol, tf))
				continue
			}

			// Perform comprehensive EMA analysis
			analysis := indicators.AnalyzeMarket(market.Base, tf, closes, a.config.Emas)

			// Store analysis for export
			allAnalyses = append(allAnalyses, analysis)

			// Display analysis results
			ui.PrintAnalysisResults(analysis)

			// Small pause between markets to be nice to exchanges
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
			}
		}
	}

	// Final summary
	ui.PrintSuccess(fmt.Sprintf("✨ Analysis completed! Processed %d markets across %d timeframes", processedMarkets, len(a.config.Timeframes)))

	// Ask user about export
	if len(allAnalyses) > 0 && a.askForExport() {
		if err := a.exportResults(allAnalyses); err != nil {
			ui.PrintError("Failed to export results", err)
		}
	}

	return nil
}

// askForExport asks user if they want to export results
func (a *App) askForExport() bool {
	fmt.Println("\n💾 Export Results")
	response := input.ReadLine(a.reader, "Export analysis results to file? (y/n, default: y): ")
	if response == "" {
		return true
	}
	normalized := input.Normalize(response)
	return normalized == "y" || normalized == "yes"
}

// exportResults exports analysis results to files
func (a *App) exportResults(analyses []indicators.MarketAnalysis) error {
	exporter := export.NewExporter("")

	exportConfig := export.ExportConfig{
		EMAs:       a.config.Emas,
		Timeframes: a.config.Timeframes,
		Symbols:    a.config.Symbols,
		SelectAll:  a.config.SelectAll,
	}

	// Export as JSON
	jsonPath, err := exporter.ExportAnalysis(analyses, a.config.Exchange, a.config.MarketType, exportConfig, export.FormatJSON)
	if err != nil {
		return err
	}
	ui.PrintSuccess(fmt.Sprintf("JSON export saved: %s", jsonPath))

	// Export as CSV
	csvPath, err := exporter.ExportAnalysis(analyses, a.config.Exchange, a.config.MarketType, exportConfig, export.FormatCSV)
	if err != nil {
		ui.PrintWarning(fmt.Sprintf("CSV export failed: %v", err))
	} else {
		ui.PrintSuccess(fmt.Sprintf("CSV export saved: %s", csvPath))
	}

	ui.PrintInfo(fmt.Sprintf("📁 Export directory: %s", exporter.GetOutputDirectory()))

	return nil
}
