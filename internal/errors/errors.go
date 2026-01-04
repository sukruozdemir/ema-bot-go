package errors

import "fmt"

// ErrType represents the type of error
type ErrType string

const (
	ErrTypeConfig     ErrType = "config"
	ErrTypeExchange   ErrType = "exchange"
	ErrTypeCache      ErrType = "cache"
	ErrTypeValidation ErrType = "validation"
	ErrTypeNetwork    ErrType = "network"
)

// AppError represents an application-specific error
type AppError struct {
	Type    ErrType
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError
func New(errType ErrType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
	}
}

// Wrap creates a new AppError that wraps another error
func Wrap(errType ErrType, message string, cause error) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Cause:   cause,
	}
}

// Predefined errors
var (
	ErrInvalidEMALength  = New(ErrTypeValidation, "invalid EMA length")
	ErrInvalidTimeframe  = New(ErrTypeValidation, "invalid timeframe")
	ErrInvalidMarketType = New(ErrTypeValidation, "invalid market type")
	ErrNoMarketsFound    = New(ErrTypeExchange, "no markets found")
	ErrExchangeNotFound  = New(ErrTypeExchange, "exchange not found")
)
