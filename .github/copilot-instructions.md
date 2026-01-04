# Copilot Instructions for EMA Bot Go

## Project Overview
EMA Bot Go is a professional-grade cryptocurrency market analysis bot built in Go. It analyzes Exponential Moving Averages (EMAs) across 100+ crypto exchanges using the CCXT library. The application follows clean architecture with clear separation of concerns between CLI, business logic, and external integrations.

## Architecture Principles

### Clean Architecture Layers
1. **cmd/** - Entry points (app runner and config tool). Minimal logic; delegates to internal/app
2. **internal/app/** - Application orchestration. Handles startup flow, config loading, user interaction flow
3. **internal/services/** - Business logic for external interactions (exchange API). Includes CCXT wrapper and caching
4. **internal/config/** - Configuration management with validation and persistence to `~/.config/ema-bot/config.json`
5. **internal/models/** - Data structures (Market, EMAConfig, MarketCache)
6. **internal/errors/** - Custom error type system with typed error codes
7. **internal/ui/** - Terminal UI with ANSI colors. Use existing functions like `PrintSuccess()`, `PrintError()`
8. **internal/input/** - User input parsing and validation helpers

### Critical Integration Points

**Exchange Service → CCXT**: The service wraps `ccxt.ICoreExchange` (from `github.com/ccxt/ccxt/go/v4`). Key methods:
- `GetSpotMarkets()` and `GetSwapMarkets()` - fetch and filter markets via CCXT's `LoadMarkets()`
- Returns `[]models.Market` with fields: Symbol, Type (spot/swap/future), Spot, Swap, Future, Linear, Inverse booleans
- Uses `sync.Map` returned by CCXT to iterate markets

**Cache Strategy**: Markets cached to `~/.cache/ema-bot/` (see `exchange.go` for paths). Cache expires after 24 hours. Always check `MarketCache.IsValid()` before use.

**Config Persistence**: Stores to path from `config.ConfigPath()` (respects `EMA_BOT_CONFIG` env var). Must validate before saving with `config.Validate()`.

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

### Testing
No test files exist yet. When adding tests, follow Go conventions: `*_test.go` files in same package, use `testing.T`.

## Project-Specific Patterns & Conventions

### Configuration Workflow
1. On startup, `app.loadOrCreateConfig()` checks for saved config
2. If found and user confirms, reuse it
3. Otherwise, prompt user interactively via `input.*` functions for:
   - EMA lengths (comma-separated integers)
   - Exchange name (validated against CCXT availability)
   - Timeframes (comma-separated: 1m, 5m, 15m, 1h, 4h, 1d, 1w)
   - Market type (spot or swap)
   - Symbols (comma-separated or "all" for select_all=true)
4. Save via `config.Save(cfg)`

### Market Filtering Logic
- `GetSelectedMarkets(symbols []string, markets []models.Market)` - filter to user-specified symbols
- `shouldIncludeMarket()` - internal filter: only active markets, quote in USDT or BNB, type matches (spot/swap)
- Markets have boolean flags (Spot, Swap, Future, Linear, Inverse) used for filtering

### Input Parsing
Use utilities in `internal/input/`:
- `ParseEMALengths()` → validates integers
- `ParseTimeframes()` → validates against known timeframes
- `SplitAndTrim()` → handles comma/space-separated input safely

### Logging
Uses `go.uber.org/zap`. Logger passed through dependency injection. Use `logger.Info()`, `logger.Error()` for important events, `logger.Debug()` for market iteration details.

### UI Output
Always use package functions from `internal/ui/`:
- `PrintSuccess(msg)` - green checkmark
- `PrintError(msg, err)` - red error with optional error details
- `PrintWarning(msg)` - yellow warning
- `PrintInfo(msg)` - blue info
- `PrintTable()` for structured data (colors and formatting applied)

## Key Dependencies
- `github.com/ccxt/ccxt/go/v4` - cryptocurrency exchange API wrapper (100+ exchanges)
- `go.uber.org/zap` - structured logging
- Go 1.24+ standard library: context, encoding/json, bufio, sync

## Configuration File Example
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

## Common Extension Points
- **New Market Type**: Update `MarketType` validation in config.go, modify filters in exchange.go
- **New EMA Calculation**: Add logic after market fetch in app.processMarkets()
- **Additional Exchanges**: CCXT handles this; just validate exchange name exists
- **Caching Strategy**: Modify `DefaultCacheExpiry` and `getMarketsWithCache()` logic in exchange.go

## Gotchas
1. **CCXT async channels**: `LoadMarkets()` returns a channel. Always handle with select + context.Done() check
2. **Market filtering**: Test with small datasets first; CCXT can return hundreds of markets
3. **Cache paths**: Respect platform-specific config dirs; don't hardcode paths
4. **User input**: Always validate via input.* functions before passing to models
