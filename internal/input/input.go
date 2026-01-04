package input

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sukruozdemir/ema-bot-go/internal/errors"
)

// ReadLine reads a line from the given reader with a prompt
func ReadLine(r *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	s, _ := r.ReadString('\n')
	return strings.TrimSpace(s)
}

// IsSingleWord checks if the string contains only a single word (no commas or spaces)
func IsSingleWord(s string) bool {
	return s != "" && !strings.ContainsAny(s, ", ")
}

// SplitAndTrim splits a string by separator and trims each part
func SplitAndTrim(s, sep string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// ParseEMALengths parses a comma-separated string of EMA lengths
func ParseEMALengths(input string) ([]int, error) {
	if input == "" {
		return nil, errors.New(errors.ErrTypeValidation, "EMA lengths input cannot be empty")
	}

	parts := SplitAndTrim(input, ",")
	if len(parts) == 0 {
		return nil, errors.New(errors.ErrTypeValidation, "no valid EMA lengths found")
	}

	var emas []int
	for _, part := range parts {
		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, errors.Wrap(errors.ErrTypeValidation,
				fmt.Sprintf("invalid EMA length '%s'", part), err)
		}

		if val <= 0 {
			return nil, errors.Wrap(errors.ErrTypeValidation,
				fmt.Sprintf("EMA length must be positive, got %d", val),
				errors.ErrInvalidEMALength)
		}

		emas = append(emas, val)
	}

	return emas, nil
}

// ParseTimeframes parses a comma-separated string of timeframes
func ParseTimeframes(input string) ([]string, error) {
	if input == "" {
		return nil, errors.New(errors.ErrTypeValidation, "timeframes input cannot be empty")
	}

	timeframes := SplitAndTrim(input, ",")
	if len(timeframes) == 0 {
		return nil, errors.New(errors.ErrTypeValidation, "no valid timeframes found")
	}

	// Validate timeframe format (basic validation)
	validTimeframes := make([]string, 0, len(timeframes))
	for _, tf := range timeframes {
		if IsValidTimeframe(tf) {
			validTimeframes = append(validTimeframes, tf)
		} else {
			return nil, errors.Wrap(errors.ErrTypeValidation,
				fmt.Sprintf("invalid timeframe format '%s'", tf),
				errors.ErrInvalidTimeframe)
		}
	}

	return validTimeframes, nil
}

// IsValidTimeframe performs basic validation on timeframe format
func IsValidTimeframe(tf string) bool {
	if tf == "" {
		return false
	}

	// Common timeframe patterns: 1m, 5m, 15m, 1h, 4h, 1d, 1w, etc.
	validFormats := []string{
		"1m", "5m", "15m", "30m",
		"1h", "2h", "4h", "6h", "8h", "12h",
		"1d", "3d", "1w", "1M",
	}

	for _, valid := range validFormats {
		if tf == valid {
			return true
		}
	}

	return false
}

// Normalize converts a string to lowercase and trims whitespace
func Normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// NowISO returns the current time in ISO 8601 format
func NowISO() string {
	return time.Now().Format(time.RFC3339)
}

// ValidateSymbols validates a slice of trading symbols
func ValidateSymbols(symbols []string) error {
	if len(symbols) == 0 {
		return errors.New(errors.ErrTypeValidation, "at least one symbol must be specified")
	}

	for _, symbol := range symbols {
		if symbol == "" {
			return errors.New(errors.ErrTypeValidation, "symbol cannot be empty")
		}

		if !IsValidSymbol(symbol) {
			return errors.New(errors.ErrTypeValidation,
				fmt.Sprintf("invalid symbol format: %s", symbol))
		}
	}

	return nil
}

// IsValidSymbol performs basic validation on symbol format
func IsValidSymbol(symbol string) bool {
	if symbol == "" {
		return false
	}

	// Basic validation: should be uppercase letters only, reasonable length
	if len(symbol) < 2 || len(symbol) > 10 {
		return false
	}

	for _, char := range symbol {
		if char < 'A' || char > 'Z' {
			return false
		}
	}

	return true
}
