package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sukruozdemir/ema-bot-go/internal/errors"
	"github.com/sukruozdemir/ema-bot-go/internal/indicators"
)

// ExportFormat represents the export file format
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
)

// AnalysisExport contains analysis data ready for export
type AnalysisExport struct {
	Timestamp  time.Time                   `json:"timestamp"`
	Exchange   string                      `json:"exchange"`
	MarketType string                      `json:"market_type"`
	Config     ExportConfig                `json:"config"`
	Results    []indicators.MarketAnalysis `json:"results"`
	Summary    ExportSummary               `json:"summary"`
}

// ExportConfig contains the configuration used for analysis
type ExportConfig struct {
	EMAs       []int    `json:"emas"`
	Timeframes []string `json:"timeframes"`
	Symbols    []string `json:"symbols,omitempty"`
	SelectAll  bool     `json:"select_all"`
}

// ExportSummary contains high-level analysis summary
type ExportSummary struct {
	TotalMarkets   int                          `json:"total_markets"`
	BullishSignals int                          `json:"bullish_signals"`
	BearishSignals int                          `json:"bearish_signals"`
	StrongTrends   int                          `json:"strong_trends"`
	WeakTrends     int                          `json:"weak_trends"`
	TopPerformers  []PerformanceMetric          `json:"top_performers"`
	RecentSignals  []indicators.CrossoverSignal `json:"recent_signals"`
}

// PerformanceMetric represents market performance data
type PerformanceMetric struct {
	Symbol       string  `json:"symbol"`
	Timeframe    string  `json:"timeframe"`
	TrendScore   float64 `json:"trend_score"`
	OverallTrend string  `json:"overall_trend"`
	SignalCount  int     `json:"signal_count"`
	StrongEMAs   int     `json:"strong_emas"`
}

// Exporter handles data export operations
type Exporter struct {
	outputDir string
}

// NewExporter creates a new exporter with the specified output directory
func NewExporter(outputDir string) *Exporter {
	if outputDir == "" {
		// Default to user's documents or temp directory
		if homeDir, err := os.UserHomeDir(); err == nil {
			outputDir = filepath.Join(homeDir, "Documents", "ema-bot-exports")
		} else {
			outputDir = filepath.Join(os.TempDir(), "ema-bot-exports")
		}
	}
	return &Exporter{outputDir: outputDir}
}

// ExportAnalysis exports analysis results in the specified format
func (e *Exporter) ExportAnalysis(analyses []indicators.MarketAnalysis, exchange, marketType string, config ExportConfig, format ExportFormat) (string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(e.outputDir, 0755); err != nil {
		return "", errors.Wrap(errors.ErrTypeConfig, "failed to create export directory", err)
	}

	// Prepare export data
	exportData := e.prepareExportData(analyses, exchange, marketType, config)

	// Generate filename
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("ema_analysis_%s_%s_%s.%s", exchange, marketType, timestamp, format)
	filepath := filepath.Join(e.outputDir, filename)

	// Export based on format
	switch format {
	case FormatJSON:
		return filepath, e.exportJSON(exportData, filepath)
	case FormatCSV:
		return filepath, e.exportCSV(exportData, filepath)
	default:
		return "", errors.New(errors.ErrTypeValidation, "unsupported export format")
	}
}

// prepareExportData creates a comprehensive export data structure
func (e *Exporter) prepareExportData(analyses []indicators.MarketAnalysis, exchange, marketType string, config ExportConfig) AnalysisExport {
	summary := e.calculateSummary(analyses)

	return AnalysisExport{
		Timestamp:  time.Now(),
		Exchange:   exchange,
		MarketType: marketType,
		Config:     config,
		Results:    analyses,
		Summary:    summary,
	}
}

// calculateSummary computes summary statistics from analysis results
func (e *Exporter) calculateSummary(analyses []indicators.MarketAnalysis) ExportSummary {
	summary := ExportSummary{
		TotalMarkets:  len(analyses),
		TopPerformers: make([]PerformanceMetric, 0),
		RecentSignals: make([]indicators.CrossoverSignal, 0),
	}

	var performanceMetrics []PerformanceMetric

	for _, analysis := range analyses {
		// Count signals
		for _, signal := range analysis.Signals {
			if signal.CrossoverBar <= 5 { // Recent signals (last 5 bars)
				summary.RecentSignals = append(summary.RecentSignals, signal)
			}

			switch signal.SignalType {
			case "golden_cross":
				summary.BullishSignals++
			case "death_cross":
				summary.BearishSignals++
			}
		}

		// Count trend strength
		if analysis.TrendScore > 0.5 {
			summary.StrongTrends++
		} else if analysis.TrendScore > 0.1 {
			summary.WeakTrends++
		}

		// Calculate performance metrics
		strongEMAs := 0
		for _, ema := range analysis.EMAs {
			if ema.TrendStrength > 0.3 {
				strongEMAs++
			}
		}

		metric := PerformanceMetric{
			Symbol:       analysis.Symbol,
			Timeframe:    analysis.Timeframe,
			TrendScore:   analysis.TrendScore,
			OverallTrend: analysis.OverallTrend.String(),
			SignalCount:  len(analysis.Signals),
			StrongEMAs:   strongEMAs,
		}
		performanceMetrics = append(performanceMetrics, metric)
	}

	// Sort by trend score and take top performers
	sort.Slice(performanceMetrics, func(i, j int) bool {
		return performanceMetrics[i].TrendScore > performanceMetrics[j].TrendScore
	})

	maxTopPerformers := 10
	if len(performanceMetrics) < maxTopPerformers {
		maxTopPerformers = len(performanceMetrics)
	}
	summary.TopPerformers = performanceMetrics[:maxTopPerformers]

	return summary
}

// exportJSON exports data as JSON file
func (e *Exporter) exportJSON(data AnalysisExport, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return errors.Wrap(errors.ErrTypeConfig, "failed to create JSON file", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return errors.Wrap(errors.ErrTypeConfig, "failed to encode JSON data", err)
	}

	return nil
}

// exportCSV exports data as CSV file
func (e *Exporter) exportCSV(data AnalysisExport, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return errors.Wrap(errors.ErrTypeConfig, "failed to create CSV file", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV headers
	headers := []string{
		"Symbol", "Timeframe", "LastPrice", "OverallTrend", "TrendScore",
		"EMA_Periods", "EMA_Values", "EMA_Trends", "Signals", "SignalTypes",
	}
	if err := writer.Write(headers); err != nil {
		return errors.Wrap(errors.ErrTypeConfig, "failed to write CSV headers", err)
	}

	// Write data rows
	for _, analysis := range data.Results {
		// Prepare EMA data
		var emaPeriods, emaValues, emaTrends []string
		for _, ema := range analysis.EMAs {
			emaPeriods = append(emaPeriods, fmt.Sprintf("%d", ema.Period))
			emaValues = append(emaValues, fmt.Sprintf("%.4f", ema.CurrentValue))
			emaTrends = append(emaTrends, ema.Trend.String())
		}

		// Prepare signal data
		var signals, signalTypes []string
		for _, signal := range analysis.Signals {
			signals = append(signals, fmt.Sprintf("EMA%d/EMA%d", signal.FastPeriod, signal.SlowPeriod))
			signalTypes = append(signalTypes, signal.SignalType)
		}

		row := []string{
			analysis.Symbol,
			analysis.Timeframe,
			fmt.Sprintf("%.4f", analysis.LastPrice),
			analysis.OverallTrend.String(),
			fmt.Sprintf("%.4f", analysis.TrendScore),
			strings.Join(emaPeriods, ";"),
			strings.Join(emaValues, ";"),
			strings.Join(emaTrends, ";"),
			strings.Join(signals, ";"),
			strings.Join(signalTypes, ";"),
		}

		if err := writer.Write(row); err != nil {
			return errors.Wrap(errors.ErrTypeConfig, "failed to write CSV row", err)
		}
	}

	return nil
}

// GetOutputDirectory returns the configured output directory
func (e *Exporter) GetOutputDirectory() string {
	return e.outputDir
}
