package ui

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/sukruozdemir/ema-bot-go/internal/config"
	"github.com/sukruozdemir/ema-bot-go/internal/input"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

// PrintWelcome displays the application welcome message
func PrintWelcome() {
	welcome := fmt.Sprintf(`%s%s
╔═══════════════════════════════════════════════════════╗
║                                                       ║
║                🤖 EMA BOT v2.0 🤖                     ║
║           Professional Trading Analysis Tool          ║
║                                                       ║
╚═══════════════════════════════════════════════════════╝%s

%s💡 Tip: Use 'all' for symbols to analyze all available markets%s
`, ColorBold, ColorCyan, ColorReset, ColorYellow, ColorReset)

	fmt.Print(welcome)
}

// PrintError displays an error message with formatting
func PrintError(message string, err error) {
	fmt.Printf("%s❌ Error: %s%s\n", ColorRed, message, ColorReset)
	if err != nil {
		fmt.Printf("%s   Details: %v%s\n", ColorRed, err, ColorReset)
	}
}

// PrintSuccess displays a success message with formatting
func PrintSuccess(message string) {
	fmt.Printf("%s✅ %s%s\n", ColorGreen, message, ColorReset)
}

// PrintWarning displays a warning message with formatting
func PrintWarning(message string) {
	fmt.Printf("%s⚠️  %s%s\n", ColorYellow, message, ColorReset)
}

// PrintInfo displays an info message with formatting
func PrintInfo(message string) {
	fmt.Printf("%s💡 %s%s\n", ColorBlue, message, ColorReset)
}

// PrettyPrintConfig displays the configuration in a formatted table and asks for confirmation
func PrettyPrintConfig(cfg *config.Config, r *bufio.Reader) bool {
	fmt.Printf("%s%s\n🔧 Found Saved Configuration%s\n", ColorBold, ColorCyan, ColorReset)

	printConfigTable(cfg)

	fmt.Printf("%s%s\nConfiguration Details:%s\n", ColorBold, ColorBlue, ColorReset)
	fmt.Printf("• Created: %s\n", cfg.SavedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("• Age: %s\n", formatDuration(time.Since(cfg.SavedAt)))

	fmt.Print("\n")
	response := input.ReadLine(r, fmt.Sprintf("%sUse this configuration? (y/n): %s", ColorBold, ColorReset))
	normalized := input.Normalize(response)

	return normalized == "y" || normalized == "yes"
}

// printConfigTable prints the configuration in a nicely formatted table
func printConfigTable(cfg *config.Config) {
	fmt.Println("┌─────────────────┬───────────────────────────────────────┐")
	fmt.Printf("│ %-15s │ %-37s │\n", "Setting", "Value")
	fmt.Println("├─────────────────┼───────────────────────────────────────┤")

	// EMA Lengths
	emaStr := formatIntSlice(cfg.Emas)
	emaDisplay := truncateString(emaStr, 37)
	fmt.Printf("│ %-15s │ %-37s │\n", "EMA Lengths", emaDisplay)

	// Exchange
	exchangeDisplay := truncateString(cfg.Exchange, 37)
	fmt.Printf("│ %-15s │ %-37s │\n", "Exchange", exchangeDisplay)

	// Timeframes
	tfStr := strings.Join(cfg.Timeframes, ", ")
	tfDisplay := truncateString(tfStr, 37)
	fmt.Printf("│ %-15s │ %-37s │\n", "Timeframes", tfDisplay)

	// Market Type
	marketTypeDisplay := fmt.Sprintf("%s %s", getMarketTypeEmoji(cfg.MarketType), cfg.MarketType)
	marketTypeDisplay = truncateString(marketTypeDisplay, 37)
	fmt.Printf("│ %-15s │ %-37s │\n", "Market Type", marketTypeDisplay)

	// Symbols
	var symbolsStr string
	if cfg.SelectAll {
		symbolsStr = "🌐 All markets"
	} else {
		symbolsStr = fmt.Sprintf("📋 %s", strings.Join(cfg.Symbols, ", "))
	}
	symbolsDisplay := truncateString(symbolsStr, 37)
	fmt.Printf("│ %-15s │ %-37s │\n", "Symbols", symbolsDisplay)

	fmt.Println("└─────────────────┴───────────────────────────────────────┘")
}

// PrintMarketSummary displays a summary of fetched markets
func PrintMarketSummary(marketType string, total, filtered int) {
	fmt.Printf("\n%s📊 Market Summary%s\n", ColorBold, ColorReset)
	fmt.Printf("• Market Type: %s %s\n", getMarketTypeEmoji(marketType), marketType)
	fmt.Printf("• Total Markets: %s%d%s\n", ColorCyan, total, ColorReset)
	fmt.Printf("• Filtered Markets: %s%d%s\n", ColorGreen, filtered, ColorReset)

	if filtered == 0 {
		PrintWarning("No markets found matching your criteria")
		fmt.Println("  Suggestions:")
		fmt.Println("  • Check if the exchange supports the selected market type")
		fmt.Println("  • Verify your symbol selection")
		fmt.Println("  • Try clearing the cache and retrying")
	} else {
		PrintSuccess(fmt.Sprintf("Successfully loaded %d markets for analysis", filtered))
	}
}

// Helper functions

func formatIntSlice(ints []int) string {
	if len(ints) == 0 {
		return ""
	}

	strs := make([]string, len(ints))
	for i, v := range ints {
		strs[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(strs, ", ")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getMarketTypeEmoji(marketType string) string {
	switch marketType {
	case "spot":
		return "💰"
	case "swap":
		return "🔄"
	case "future":
		return "📈"
	default:
		return "📊"
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "less than a minute"
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}
