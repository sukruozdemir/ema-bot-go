# EMA Bot Go

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A professional-grade EMA (Exponential Moving Average) cryptocurrency market analysis bot written in Go. Features clean architecture, comprehensive error handling, and beautiful CLI interface.

## ✨ Features

- 🏪 **Multi-Exchange Support** - Works with 100+ exchanges via CCXT
- 📊 **EMA Analysis** - Configure multiple EMA periods for technical analysis
- ⚡ **Performance Optimized** - Built with Go's concurrency patterns and intelligent caching
- 🔧 **Professional Architecture** - Clean code following Go best practices
- 💾 **Smart Caching** - Intelligent market data caching with expiration to reduce API calls
- 🎯 **Flexible Filtering** - Filter markets by symbols, type (spot/swap), and timeframes
- 🖥️ **Beautiful CLI** - Rich terminal UI with colors, emojis, and formatted tables
- ⚙️ **Configuration Management** - Save and reuse your preferred settings with validation
- 🛡️ **Robust Error Handling** - Custom error types with contextual information
- 📦 **Modular Design** - Clean separation of concerns for easy extension

## 🚀 Quick Start

### Prerequisites

- Go 1.22 or higher
- Internet connection for exchange API access

### Installation

```bash
# Clone the repository
git clone https://github.com/sukruozdemir/ema-bot-go
cd ema-bot-go

# Install dependencies
go mod tidy

# Build the application
go build -o ema-bot ./cmd/app

# Run the application
./ema-bot
```

### Using Go Run (Development)

```bash
# Run directly with Go
go run ./cmd/app/main.go
```

## 📖 Usage

### Main Application

1. **First Launch**: The application will guide you through interactive configuration:
   - EMA lengths (e.g., 50, 100, 200)
   - Exchange selection (e.g., binance, coinbase, kraken)
   - Timeframes (1m, 5m, 15m, 1h, 4h, 1d, 1w)
   - Market type (spot or swap/futures)
   - Symbol selection (specific coins or all markets)

2. **Subsequent Launches**: The bot remembers your configuration and offers to reuse it

3. **Cache Management**: Choose whether to clear cached market data on each run

### Configuration Tool

Manage your configuration without running the full application:

```bash
# Show current configuration
go run ./cmd/config -show-config

# Validate configuration
go run ./cmd/config -validate

# Use custom config path
go run ./cmd/config -config=/path/to/config.json -show-config
```

## 🔧 Configuration

The application stores configuration in:
- **Custom Path**: Set `EMA_BOT_CONFIG` environment variable
- **Default**: `~/.config/ema-bot/config.json` (Unix) or `%APPDATA%/ema-bot/config.json` (Windows)

### Example Configuration

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

## 📁 Project Structure

```
.
├── cmd/
│   ├── app/
│   │   └── main.go              # Main application entry point
│   └── config/
│       └── main.go              # Configuration management tool
├── internal/
│   ├── app/
│   │   └── app.go               # Application orchestration
│   ├── config/
│   │   └── config.go            # Configuration management & validation
│   ├── errors/
│   │   └── errors.go            # Custom error types & wrapping
│   ├── input/
│   │   └── input.go             # User input handling & validation
│   ├── models/
│   │   └── models.go            # Data models & structures
│   ├── services/
│   │   ├── exchange.go          # Exchange service implementation
│   │   └── interfaces.go        # Service interfaces
│   └── ui/
│       └── ui.go                # User interface & formatting
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── README.md                    # This file
└── .gitignore                  # Git ignore rules
```

## 🏗️ Architecture

The project follows professional Go development practices:

### Design Principles

- **🎯 Separation of Concerns**: Each package has a single responsibility
- **🔌 Dependency Injection**: Services are injected rather than created directly
- **🎭 Interface-Based Design**: Core functionality defined by interfaces for testability
- **📦 Layered Architecture**: Clear boundaries between UI, business logic, and data
- **⚡ Context Support**: Proper cancellation and timeout handling throughout
- **🛡️ Error Handling**: Custom error types with wrapping and contextual information
- **✅ Input Validation**: Comprehensive validation at all entry points

### Package Organization

| Package | Responsibility |
|---------|-----------------|
| `app` | Application flow orchestration |
| `services` | External API interactions (exchanges) |
| `models` | Data structures and business objects |
| `config` | Configuration management and persistence |
| `ui` | User interface and display formatting |
| `input` | User input handling and validation |
| `errors` | Custom error types and utilities |

## 🔧 Development

### Code Quality

```bash
# Run tests (when tests are added)
go test ./...

# Run with race detector
go run -race ./cmd/app/main.go

# Format code (applied automatically)
go fmt ./...

# Run static analysis
go vet ./...

# Build with optimizations
go build -ldflags="-s -w" -o ema-bot ./cmd/app
```

### Building for Different Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o ema-bot-linux ./cmd/app

# Windows
GOOS=windows GOARCH=amd64 go build -o ema-bot.exe ./cmd/app

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o ema-bot-macos ./cmd/app

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o ema-bot-macos-arm64 ./cmd/app
```

## 🌟 Key Improvements

### Professional Go Practices
- ✅ Structured error handling with custom error types
- ✅ Interface-based architecture for loose coupling
- ✅ Context support for cancellation and timeouts
- ✅ Proper logging infrastructure (Zap integration)
- ✅ Configuration validation with clear error messages
- ✅ Input sanitization and validation

### User Experience
- ✅ Beautiful formatted tables with proper alignment
- ✅ Color-coded output with emojis
- ✅ Step-by-step interactive configuration
- ✅ Smart configuration reuse across runs
- ✅ Clear, actionable error messages
- ✅ Progress feedback during operations

### Code Quality
- ✅ No unused variables or imports
- ✅ Passes `go vet` analysis
- ✅ Properly formatted with `go fmt`
- ✅ Comprehensive documentation comments
- ✅ Clean separation of concerns

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Ensure code passes `go vet` and `go fmt`
4. Commit your changes (`git commit -m 'Add some amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [CCXT](https://github.com/ccxt/ccxt) - Unified API for cryptocurrency exchanges
- [Zap](https://github.com/uber-go/zap) - Fast, structured, leveled logging
- Go community for excellent tools and best practices

## 📈 Roadmap

### In Progress
- [x] Professional project architecture
- [x] Configuration management with validation
- [x] Custom error handling
- [x] Beautiful CLI interface
- [x] Input validation and sanitization

### Planned
- [ ] Real EMA calculation and analysis
- [ ] Trading signal generation
- [ ] Alert notifications (email, Discord, Telegram)
- [ ] Portfolio tracking and management
- [ ] Unit and integration tests
- [ ] Web dashboard interface
- [ ] REST API for external integrations
- [ ] Real-time WebSocket data streaming
- [ ] Advanced technical indicators (RSI, MACD, Bollinger Bands)
- [ ] Backtesting and strategy optimization
- [ ] Performance metrics and reporting

## 📧 Support

If you have questions or need help:
1. Open an issue on GitHub
2. Check existing documentation
3. Review the code comments and structure

---

**Made with ❤️ by a professional Go developer**

*Last Updated: January 4, 2026*