package ui

import (
	"bufio"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/sukruozdemir/ema-bot-go/internal/config"
	"github.com/sukruozdemir/ema-bot-go/internal/indicators"
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
	ColorDim    = "\033[2m"
)

// Table represents a data table with formatting options
type Table struct {
	Headers []string
	Rows    [][]string
	Title   string
}

// ProgressBar represents a progress indicator
type ProgressBar struct {
	Total   int
	Current int
	Title   string
}

// PrintWelcome displays the application welcome message
func PrintWelcome() {
	welcome := fmt.Sprintf(`%s%s
╔═══════════════════════════════════════════════════════╗
║                                                       ║
║                🤖 EMA BOT v3.0 🤖                     ║
║           Professional Trading Analysis Tool          ║
║              with Advanced Signal Detection           ║
║                                                       ║
╚═══════════════════════════════════════════════════════╝%s

%s💡 New Features: Trend Analysis, Signal Detection, Export%s
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

// PrintTable renders a formatted table with proper alignment
func (t *Table) Print() {
	if len(t.Rows) == 0 {
		PrintWarning("No data to display in table")
		return
	}

	// Calculate column widths
	widths := make([]int, len(t.Headers))
	for i, header := range t.Headers {
		widths[i] = len(header)
	}

	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(widths) && len(stripAnsi(cell)) > widths[i] {
				widths[i] = len(stripAnsi(cell))
			}
		}
	}

	// Print title if provided
	if t.Title != "" {
		fmt.Printf("\n%s%s%s%s\n", ColorBold, ColorCyan, t.Title, ColorReset)
	}

	// Print header separator
	fmt.Print("┌")
	for i, width := range widths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			fmt.Print("┬")
		}
	}
	fmt.Println("┐")

	// Print headers
	fmt.Print("│")
	for i, header := range t.Headers {
		fmt.Printf(" %s%-*s%s ", ColorBold, widths[i], header, ColorReset)
		if i < len(widths)-1 {
			fmt.Print("│")
		}
	}
	fmt.Println("│")

	// Print separator
	fmt.Print("├")
	for i, width := range widths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			fmt.Print("┼")
		}
	}
	fmt.Println("┤")

	// Print rows
	for _, row := range t.Rows {
		fmt.Print("│")
		for i, cell := range row {
			if i < len(widths) {
				padding := widths[i] - len(stripAnsi(cell))
				fmt.Printf(" %s%s ", cell, strings.Repeat(" ", padding))
				if i < len(widths)-1 {
					fmt.Print("│")
				}
			}
		}
		fmt.Println("│")
	}

	// Print bottom separator
	fmt.Print("└")
	for i, width := range widths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			fmt.Print("┴")
		}
	}
	fmt.Println("┘")
}

// PrintAnalysisResults displays comprehensive EMA analysis in a formatted table
func PrintAnalysisResults(analysis indicators.MarketAnalysis) {
	// Main analysis table
	table := &Table{
		Title:   fmt.Sprintf("📊 EMA Analysis: %s (%s)", analysis.Symbol, analysis.Timeframe),
		Headers: []string{"EMA", "Current", "Previous", "Change %", "Trend", "Strength"},
	}

	// Sort EMAs by period for consistent display
	emas := make([]indicators.EMAAnalysis, len(analysis.EMAs))
	copy(emas, analysis.EMAs)
	sort.Slice(emas, func(i, j int) bool {
		return emas[i].Period < emas[j].Period
	})

	for _, ema := range emas {
		var currentStr, previousStr, changeStr, trendStr, strengthStr string

		// Format current value
		if !math.IsNaN(ema.CurrentValue) {
			currentStr = fmt.Sprintf("%.4f", ema.CurrentValue)
		} else {
			currentStr = "N/A"
		}

		// Format previous value
		if !math.IsNaN(ema.PreviousValue) {
			previousStr = fmt.Sprintf("%.4f", ema.PreviousValue)
		} else {
			previousStr = "N/A"
		}

		// Format change with color
		if ema.Change != 0 {
			color := ColorGreen
			symbol := "📈"
			if ema.Change < 0 {
				color = ColorRed
				symbol = "📉"
			}
			changeStr = fmt.Sprintf("%s%s %.2f%%%s", color, symbol, ema.ChangePercent, ColorReset)
		} else {
			changeStr = "0.00%"
		}

		// Format trend with color and emoji
		switch ema.Trend {
		case indicators.TrendUp:
			trendStr = fmt.Sprintf("%s🚀 %s%s", ColorGreen, ema.Trend.String(), ColorReset)
		case indicators.TrendDown:
			trendStr = fmt.Sprintf("%s🔻 %s%s", ColorRed, ema.Trend.String(), ColorReset)
		default:
			trendStr = fmt.Sprintf("%s⚡ %s%s", ColorYellow, ema.Trend.String(), ColorReset)
		}

		// Format strength as percentage
		strengthStr = fmt.Sprintf("%.1f%%", ema.TrendStrength*100)

		table.Rows = append(table.Rows, []string{
			fmt.Sprintf("EMA %d", ema.Period),
			currentStr,
			previousStr,
			changeStr,
			trendStr,
			strengthStr,
		})
	}

	table.Print()

	// Overall trend summary
	trendEmoji := "⚡"
	trendColor := ColorYellow
	switch analysis.OverallTrend {
	case indicators.TrendUp:
		trendEmoji = "🚀"
		trendColor = ColorGreen
	case indicators.TrendDown:
		trendEmoji = "🔻"
		trendColor = ColorRed
	}

	fmt.Printf("\n%s%s Overall Trend: %s %s (Score: %.2f)%s\n",
		ColorBold, trendColor, trendEmoji, analysis.OverallTrend.String(),
		analysis.TrendScore, ColorReset)

	// Print signals if any
	if len(analysis.Signals) > 0 {
		PrintSignals(analysis.Signals)
	}
}

// PrintSignals displays crossover signals in a formatted table
func PrintSignals(signals []indicators.CrossoverSignal) {
	if len(signals) == 0 {
		return
	}

	table := &Table{
		Title:   "🎯 Trading Signals (Recent Crossovers)",
		Headers: []string{"Type", "EMAs", "Bars Ago", "Strength", "Signal"},
	}

	for _, signal := range signals {
		var signalColor, signalEmoji, signalText string

		switch signal.SignalType {
		case "golden_cross":
			signalColor = ColorGreen
			signalEmoji = "🌟"
			signalText = "BUY"
		case "death_cross":
			signalColor = ColorRed
			signalEmoji = "💀"
			signalText = "SELL"
		default:
			signalColor = ColorYellow
			signalEmoji = "⚡"
			signalText = "WATCH"
		}

		strengthBar := createStrengthBar(signal.Strength)

		table.Rows = append(table.Rows, []string{
			fmt.Sprintf("%s%s%s", signalColor, strings.Title(strings.ReplaceAll(signal.SignalType, "_", " ")), ColorReset),
			fmt.Sprintf("EMA %d/%d", signal.FastPeriod, signal.SlowPeriod),
			fmt.Sprintf("%d", signal.CrossoverBar),
			strengthBar,
			fmt.Sprintf("%s%s %s%s", signalColor, signalEmoji, signalText, ColorReset),
		})
	}

	table.Print()
}

// PrintProgressBar displays a progress bar
func (p *ProgressBar) Print() {
	if p.Total <= 0 {
		return
	}

	percentage := float64(p.Current) / float64(p.Total) * 100
	barWidth := 30
	filledWidth := int(float64(barWidth) * float64(p.Current) / float64(p.Total))

	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)

	fmt.Printf("\r%s%s [%s%s%s] %d/%d (%.1f%%)%s",
		ColorBlue, p.Title, ColorCyan, bar, ColorBlue, p.Current, p.Total, percentage, ColorReset)

	if p.Current >= p.Total {
		fmt.Println() // New line when complete
	}
}

// PrintMarketProgress displays progress for market processing
func PrintMarketProgress(symbol string, current, total int) {
	progress := &ProgressBar{
		Title:   fmt.Sprintf("📊 Processing %s", symbol),
		Current: current,
		Total:   total,
	}
	progress.Print()
}

// Helper functions

// stripAnsi removes ANSI escape codes for length calculation
func stripAnsi(s string) string {
	// Simple ANSI escape sequence removal
	result := strings.ReplaceAll(s, "\033[0m", "")
	result = strings.ReplaceAll(result, "\033[1m", "")
	result = strings.ReplaceAll(result, "\033[2m", "")

	// Remove color codes (30-37, 40-47, 90-97, 100-107)
	for i := 30; i <= 37; i++ {
		result = strings.ReplaceAll(result, fmt.Sprintf("\033[%dm", i), "")
	}
	for i := 40; i <= 47; i++ {
		result = strings.ReplaceAll(result, fmt.Sprintf("\033[%dm", i), "")
	}
	for i := 90; i <= 97; i++ {
		result = strings.ReplaceAll(result, fmt.Sprintf("\033[%dm", i), "")
	}
	for i := 100; i <= 107; i++ {
		result = strings.ReplaceAll(result, fmt.Sprintf("\033[%dm", i), "")
	}

	return result
}

// createStrengthBar creates a visual strength indicator
func createStrengthBar(strength float64) string {
	barLength := 10
	filledLength := int(strength * float64(barLength))

	var color string
	if strength >= 0.7 {
		color = ColorGreen
	} else if strength >= 0.4 {
		color = ColorYellow
	} else {
		color = ColorRed
	}

	bar := strings.Repeat("█", filledLength) + strings.Repeat("░", barLength-filledLength)
	return fmt.Sprintf("%s%s%s %.0f%%", color, bar, ColorReset, strength*100)
}

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
