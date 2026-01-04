package errors

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(ErrTypeValidation, "test error message")

	if err == nil {
		t.Fatal("New() returned nil")
	}

	if err.Type != ErrTypeValidation {
		t.Errorf("Type = %v, want %v", err.Type, ErrTypeValidation)
	}

	if err.Message != "test error message" {
		t.Errorf("Message = %v, want %v", err.Message, "test error message")
	}

	if err.Cause != nil {
		t.Errorf("Cause should be nil, got %v", err.Cause)
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrap(ErrTypeExchange, "wrapped error", originalErr)

	if err == nil {
		t.Fatal("Wrap() returned nil")
	}

	if err.Type != ErrTypeExchange {
		t.Errorf("Type = %v, want %v", err.Type, ErrTypeExchange)
	}

	if err.Message != "wrapped error" {
		t.Errorf("Message = %v, want %v", err.Message, "wrapped error")
	}

	if err.Cause != originalErr {
		t.Errorf("Cause = %v, want %v", err.Cause, originalErr)
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name:     "error without cause",
			err:      New(ErrTypeConfig, "configuration failed"),
			expected: "[config] configuration failed",
		},
		{
			name:     "error with cause",
			err:      Wrap(ErrTypeNetwork, "network request failed", errors.New("connection timeout")),
			expected: "[network] network request failed: connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(ErrTypeCache, "cache error", originalErr)

	unwrapped := wrappedErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	// Test error without cause
	simpleErr := New(ErrTypeValidation, "validation error")
	unwrapped = simpleErr.Unwrap()
	if unwrapped != nil {
		t.Errorf("Unwrap() of simple error should be nil, got %v", unwrapped)
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		typ  ErrType
	}{
		{"ErrInvalidEMALength", ErrInvalidEMALength, ErrTypeValidation},
		{"ErrInvalidTimeframe", ErrInvalidTimeframe, ErrTypeValidation},
		{"ErrInvalidMarketType", ErrInvalidMarketType, ErrTypeValidation},
		{"ErrNoMarketsFound", ErrNoMarketsFound, ErrTypeExchange},
		{"ErrExchangeNotFound", ErrExchangeNotFound, ErrTypeExchange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.err.Type != tt.typ {
				t.Errorf("%s Type = %v, want %v", tt.name, tt.err.Type, tt.typ)
			}
		})
	}
}

func TestErrorChaining(t *testing.T) {
	err1 := errors.New("base error")
	err2 := Wrap(ErrTypeCache, "cache failed", err1)
	err3 := Wrap(ErrTypeExchange, "exchange operation failed", err2)

	// Test errors.Is
	if !errors.Is(err3, err2) {
		t.Error("errors.Is should find wrapped error")
	}

	// Test errors.As
	var appErr *AppError
	if !errors.As(err3, &appErr) {
		t.Error("errors.As should find AppError type")
	}

	if appErr.Type != ErrTypeExchange {
		t.Errorf("AppError Type = %v, want %v", appErr.Type, ErrTypeExchange)
	}
}
