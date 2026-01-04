# Copilot Instructions for EMA Bot Go v3.0

## Project Overview
EMA Bot Go is a professional-grade cryptocurrency market analysis bot built in Go. It provides advanced EMA (Exponential Moving Average) analysis with comprehensive trend detection, signal generation, and data export capabilities across 100+ crypto exchanges using the CCXT library. The application follows clean architecture with clear separation of concerns between CLI, business logic, and external integrations.

## Architecture Principles

### Clean Architecture Layers
1. **cmd/** - Entry points (app runner and config tool). Minimal logic; delegates to internal/app
2. **internal/app/** - Application orchestration. Handles startup flow, config loading, user interaction flow
3. **internal/services/** - Business logic for external interactions (exchange API). Includes CCXT wrapper and caching
4. **internal/indicators/** - Advanced technical analysis. EMA calculations, trend detection, signal generation
5. **internal/export/** - Data export functionality. JSON/CSV export with comprehensive summaries
6. **internal/config/** - Configuration management with validation and persistence to `~/.config/ema-bot/config.json`
7. **internal/models/** - Data structures (Market, EMAConfig, MarketCache, analysis results)
8. **internal/errors/** - Custom error type system with typed error codes
9. **internal/ui/** - Advanced terminal UI with ANSI colors, tables, progress bars. Professional data presentation
10. **internal/input/** - User input parsing and validation helpers
11. **internal/cache/** - Intelligent cache management with TTL and expiration handling

### Critical Integration Points

**Exchange Service → CCXT**: The service wraps `ccxt.ICoreExchange` (from `github.com/ccxt/ccxt/go/v4`). Key methods:
- `GetSpotMarkets()` and `GetSwapMarkets()` - fetch and filter markets via CCXT's `LoadMarkets()`
- `FetchOhlcvWithDataCount()` - intelligent OHLCV fetching with rate limiting and pagination
- Returns `[]models.Market` with fields: Symbol, Type (spot/swap/future), Spot, Swap, Future, Linear, Inverse booleans
- Uses `sync.Map` returned by CCXT to iterate markets

**Advanced Analysis Pipeline**:
- `indicators.AnalyzeMarket()` - comprehensive market analysis with trend detection
- `indicators.DetectCrossovers()` - golden/death cross signal identification
- `indicators.EMAAnalysis` - detailed EMA analysis with trend strength and direction
- `indicators.MarketAnalysis` - complete market analysis results

**Enhanced UI System**:
- `ui.Table` - professional data tables with proper alignment and ANSI color support
- `ui.PrintAnalysisResults()` - comprehensive analysis display with formatted tables
- `ui.PrintSignals()` - trading signal visualization with strength indicators
- `ui.ProgressBar` - real-time progress indicators during processing

**Export System**:
- `export.ExportAnalysis()` - comprehensive data export in JSON/CSV formats
- `export.AnalysisExport` - structured export data with summaries and statistics
- Automatic top performer identification and signal history tracking

**Cache Strategy**: Markets cached to `~/.cache/ema-bot/` via `cache.Manager`. Cache expires after 24 hours (configurable). Always check `cache.IsValid()` before use. Use `cache.GetCacheInfo()` for cache metadata.

**Config Persistence**: Stores to path from `config.ConfigPath()` (respects `EMA_BOT_CONFIG` env var). Must validate before saving with `config.Validate()`.

**Test Coverage**: Comprehensive test coverage across all packages:
- `internal/cache/` - 100% coverage with cache_test.go
- `internal/config/` - 100% coverage with config_test.go
- `internal/errors/` - 100% coverage with errors_test.go
- `internal/indicators/` - 95% coverage with indicators_test.go and benchmarks
- `internal/input/` - 90% coverage with input_test.go
- `internal/models/` - 100% coverage with models_test.go

## Error Handling Pattern
Use typed errors from `internal/errors/`:
```go
errors.New(ErrType, "message")           // Simple error
errors.Wrap(ErrType, "message", cause)   // Wrapped error
```
Error types: `ErrTypeConfig`, `ErrTypeExchange`, `ErrTypeCache`, `ErrTypeValidation`, `ErrTypeNetwork`

## Developer Workflows

### Build & Run
```bash
# Using Makefile (recommended)
make build          # Build binaries
make run           # Run application
make test          # Run tests
make coverage      # Generate coverage report

# Traditional Go commands
go build -o ema-bot ./cmd/app
./ema-bot
# OR development:
go run ./cmd/app/main.go
```

### Configuration Management Tool
```bash
go run ./cmd/config -show-config    # Display current config
go run ./cmd/config -validate       # Validate config file
go run ./cmd/config -config=PATH    # Use custom config path
```

### Testing & Quality
```bash
make test          # Run all tests
make test-short    # Run short tests only
make coverage      # Generate HTML coverage report
make bench         # Run benchmarks
make check         # Run all quality checks (fmt, vet, test)
make lint          # Run all linting tools

# Traditional commands
go test ./...                       # Run tests
go test -v -race ./...             # Run with race detector
go test -bench=. ./...             # Run benchmarks
go vet ./...                       # Static analysis
go fmt ./...                       # Format code
```

### Build Automation
```bash
make build          # Build all binaries
make build-dev      # Build without optimization
make cross-compile  # Build for all platforms
make clean          # Clean build artifacts
make install        # Install to GOPATH/bin
make help           # Show all available targets
```

## Project-Specific Patterns & Conventions

### Enhanced Analysis Workflow
1. On startup, `app.loadOrCreateConfig()` checks for saved config
2. If found and user confirms, reuse it; otherwise create interactively
3. Interactive prompts via `input.*` functions for:
   - EMA lengths (comma-separated integers, validated)
   - Exchange name (validated against CCXT availability)
   - Timeframes (comma-separated: 1m, 5m, 15m, 1h, 4h, 1d, 1w)
   - Market type (spot or swap)
   - Symbols (comma-separated or "all" for select_all=true)
4. Save via `config.Save(cfg)`
5. Market processing with real-time progress indicators
6. Comprehensive analysis using `indicators.AnalyzeMarket()`
7. Professional result display using enhanced UI tables
8. Optional export to JSON/CSV with detailed summaries

### Advanced Technical Analysis
- **EMA Analysis**: `indicators.AnalyzeEMA()` provides trend direction, strength, and slope analysis
- **Signal Detection**: `indicators.DetectCrossovers()` identifies golden/death crosses with strength scoring
- **Trend Analysis**: Comprehensive trend scoring from -1 (strong bearish) to +1 (strong bullish)
- **Performance Metrics**: Automatic ranking and top performer identification

### Professional UI System
- **Formatted Tables**: `ui.Table` with proper column alignment and ANSI color support
- **Progress Indicators**: Real-time progress bars with percentage and status updates
- **Rich Formatting**: Colors, emojis, and visual indicators for trends and signals
- **Responsive Design**: Adapts to terminal width with intelligent truncation

### Market Filtering Logic
- `GetSelectedMarkets(symbols []string, markets []models.Market)` - filter to user-specified symbols
- `shouldIncludeMarket()` - internal filter: only active markets, quote in USDT or BNB, type matches (spot/swap)
- Markets have boolean flags (Spot, Swap, Future, Linear, Inverse) used for filtering

### Input Parsing & Validation
Use utilities in `internal/input/`:
- `ParseEMALengths()` → validates integers with range checking
- `ParseTimeframes()` → validates against known timeframes
- `SplitAndTrim()` → handles comma/space-separated input safely
- `IsValidSymbol()` → validates cryptocurrency symbols

### Enhanced Logging
Uses `go.uber.org/zap` with development configuration for detailed debugging:
- `logger.Info()` for major operations and progress
- `logger.Error()` for errors with context
- `logger.Debug()` for detailed market iteration and analysis steps
- Structured logging with key-value pairs for better debugging

### Cache Management
Centralized cache system in `internal/cache/`:
- `cache.NewManager(dir, ttl)` - creates cache manager with custom directory and TTL
- `cache.Load(exchange, marketType)` - loads cached market data
- `cache.Save(exchange, marketType, markets)` - saves market data to cache
- `cache.Clear(exchange, marketType)` - removes specific cache
- `cache.ClearAll()` - removes all cached data
- `cache.GetCacheInfo(exchange, marketType)` - returns cache metadata

### Export & Data Management
Export system in `internal/export/`:
- `NewExporter(outputDir)` - creates exporter with configurable output directory
- `ExportAnalysis()` - exports analysis results in JSON/CSV with comprehensive summaries
- Automatic generation of performance statistics, trend counts, and signal summaries
- Top performer identification and recent signal tracking

### UI Output Standards
Always use package functions from `internal/ui/`:
- `PrintSuccess(msg)` - green checkmark with message
- `PrintError(msg, err)` - red error with optional error details
- `PrintWarning(msg)` - yellow warning with appropriate icon
- `PrintInfo(msg)` - blue info with lightbulb icon
- `PrintAnalysisResults(analysis)` - comprehensive analysis table display
- `PrintSignals(signals)` - formatted signal table with strength indicators
- `PrintMarketProgress(symbol, current, total)` - real-time progress bars

## Key Dependencies
- `github.com/ccxt/ccxt/go/v4` - cryptocurrency exchange API wrapper (100+ exchanges)
- `go.uber.org/zap` - structured logging with development configuration
- Go 1.24+ standard library: context, encoding/json, encoding/csv, bufio, sync

## Configuration File Structure
```json
{
  "emas": [50, 100, 200],
  "exchange": "binance",
  "timeframes": ["1h", "4h", "1d"],
  "market_type": "swap",
  "select_all": false,
  "symbols": ["BTC", "ETH", "ADA"],
  "saved_at": "2026-01-04T10:30:00Z"
}
```

## Analysis Data Structures

### EMAAnalysis
```go
type EMAAnalysis struct {
    Period        int             // EMA period (e.g., 50, 100, 200)
    CurrentValue  float64         // Latest EMA value
    PreviousValue float64         // Previous EMA value for comparison
    Change        float64         // Absolute change
    ChangePercent float64         // Percentage change
    Trend         TrendDirection  // TrendUp, TrendDown, TrendNeutral
    TrendStrength float64         // 0-1 scale trend strength
    Slope         float64         // Rate of change (slope)
}
```

### MarketAnalysis
```go
type MarketAnalysis struct {
    Symbol       string           // Market symbol (e.g., "BTC")
    Timeframe    string           // Analysis timeframe (e.g., "1h")
    LastPrice    float64          // Current market price
    EMAs         []EMAAnalysis    // Analysis for each EMA period
    Signals      []CrossoverSignal// Detected crossover signals
    OverallTrend TrendDirection   // Combined trend assessment
    TrendScore   float64          // Overall trend score (-1 to +1)
}
```

## Common Extension Points
- **New Technical Indicators**: Add to `internal/indicators/` following EMA analysis patterns with tests
- **Additional Export Formats**: Extend `export.ExportAnalysis()` with new format types
- **Enhanced UI Components**: Add new table types and progress indicators to `internal/ui/`
- **Market Type Support**: Update `MarketType` validation in config.go, modify filters in exchange.go
- **Signal Types**: Extend `CrossoverSignal` and `DetectCrossovers()` for new signal patterns
- **Advanced Analysis**: Add new analysis types to `MarketAnalysis` structure
- **Cache Backends**: Extend `cache.Manager` to support Redis, Memcached, or other backends
- **Testing Tools**: Add test helpers and fixtures in `internal/testutils/` package

## Project Tools & Automation

### Makefile Targets
The project includes a comprehensive Makefile for build automation:
- `make build` - Build binaries with optimization
- `make build-dev` - Build without optimization (debugging)
- `make test` - Run all tests with race detector
- `make coverage` - Generate HTML coverage report
- `make bench` - Run performance benchmarks
- `make cross-compile` - Build for Linux, Windows, macOS (Intel & ARM)
- `make clean` - Remove build artifacts
- `make fmt` - Format all Go files
- `make vet` - Run static analysis
- `make lint` - Run all linting tools
- `make check` - Run fmt, vet, and tests
- `make profile-cpu` - Run with CPU profiling
- `make profile-mem` - Run with memory profiling
- `make race` - Run with race detector
- `make help` - Show all available targets

## Gotchas & Best Practices
1. **CCXT async channels**: `LoadMarkets()` returns a channel. Always handle with select + context.Done() check
2. **Market filtering**: Test with small datasets first; CCXT can return hundreds of markets
3. **Cache paths**: Respect platform-specific config dirs; don't hardcode paths
4. **User input**: Always validate via input.* functions before passing to models
5. **ANSI colors**: Use `ui.stripAnsi()` for length calculations when formatting tables
6. **Progress indicators**: Update progress bars frequently for responsive user experience
7. **Export paths**: Use `os.UserHomeDir()` and `filepath.Join()` for cross-platform compatibility
8. **Error context**: Always provide meaningful context when wrapping errors
9. **Signal detection**: Consider false positives; implement strength scoring for reliability
10. **Memory management**: Process markets in batches for large datasets to avoid memory issues

## Performance Considerations
- **Concurrent processing**: Markets processed sequentially with rate limiting to respect exchange APIs
- **Cache efficiency**: 24-hour cache TTL with intelligent invalidation
- **Memory usage**: ~50MB typical for 10 markets across 3 timeframes
- **Analysis speed**: ~200ms per market/timeframe combination
- **Export optimization**: Stream large datasets to files rather than loading entirely in memory

## Testing Strategy
When adding tests, follow Go conventions:
- `*_test.go` files in same package as tested code
- Use `testing.T` for unit tests
- Test all error conditions and edge cases
- Mock external dependencies (CCXT exchange interfaces)
- Benchmark performance-critical functions (EMA calculations, signal detection)
- Integration tests for end-to-end analysis workflows
- Aim for 85%+ code coverage across all packages
- Use table-driven tests for comprehensive scenario coverage

### Current Test Coverage
- **cache**: 100% - Full coverage of TTL, expiration, and info queries
- **config**: 100% - All validation and persistence paths tested
- **errors**: 100% - Error wrapping, unwrapping, and chaining tested
- **indicators**: 95% - EMA calculations, trend detection, signal detection with benchmarks
- **input**: 90% - Parsing and validation for all input types
- **models**: 100% - All model structures and methods tested
- **services**: Manual testing required (CCXT integration)
- **ui**: Manual testing required (terminal display)
- **export**: Pending automated tests
- **app**: Manual integration testing

### Test Execution
```bash
make test          # Run all tests
make test-short    # Skip long-running tests
make coverage      # Generate HTML coverage report
make bench         # Run performance benchmarks
```

## Version 3.0 New Features
- **Advanced Technical Analysis**: Comprehensive trend detection and signal generation
- **Professional UI**: Formatted tables, progress bars, and rich terminal output
- **Data Export**: JSON/CSV export with detailed summaries and statistics
- **Enhanced Performance**: Intelligent caching and optimized processing pipelines
- **Signal Detection**: Golden/death cross identification with strength scoring
- **Top Performer Tracking**: Automatic market ranking and performance metrics
- **Cache Management**: Dedicated cache package with TTL and metadata queries
- **Comprehensive Testing**: 85%+ test coverage with benchmarks and table-driven tests
- **Build Automation**: Professional Makefile with 20+ targets for development workflow
- **Error Handling**: Enhanced error types with wrapping, unwrapping, and chaining support
