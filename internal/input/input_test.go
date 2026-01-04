package input

import (
	"reflect"
	"testing"
)

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		name string
		s    string
		sep  string
		want []string
	}{
		{
			name: "comma separated",
			s:    "BTC,ETH,ADA",
			sep:  ",",
			want: []string{"BTC", "ETH", "ADA"},
		},
		{
			name: "comma separated with spaces",
			s:    "BTC, ETH, ADA",
			sep:  ",",
			want: []string{"BTC", "ETH", "ADA"},
		},
		{
			name: "space separated",
			s:    "BTC ETH ADA",
			sep:  " ",
			want: []string{"BTC", "ETH", "ADA"},
		},
		{
			name: "empty string",
			s:    "",
			sep:  ",",
			want: nil,
		},
		{
			name: "single item",
			s:    "BTC",
			sep:  ",",
			want: []string{"BTC"},
		},
		{
			name: "trailing comma",
			s:    "BTC,ETH,",
			sep:  ",",
			want: []string{"BTC", "ETH"},
		},
		{
			name: "multiple spaces",
			s:    "BTC  ,  ETH  ,  ADA",
			sep:  ",",
			want: []string{"BTC", "ETH", "ADA"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitAndTrim(tt.s, tt.sep)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitAndTrim() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEMALengths(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{
			name:    "valid lengths",
			input:   "50,100,200",
			want:    []int{50, 100, 200},
			wantErr: false,
		},
		{
			name:    "valid with spaces",
			input:   "50, 100, 200",
			want:    []int{50, 100, 200},
			wantErr: false,
		},
		{
			name:    "single length",
			input:   "50",
			want:    []int{50},
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid number",
			input:   "50,abc,200",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "negative number",
			input:   "50,-100,200",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "zero",
			input:   "0,100",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "mixed valid and invalid",
			input:   "50,100,abc",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEMALengths(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEMALengths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEMALengths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTimeframes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:    "valid timeframes",
			input:   "1h,4h,1d",
			want:    []string{"1h", "4h", "1d"},
			wantErr: false,
		},
		{
			name:    "valid with spaces",
			input:   "1h, 4h, 1d",
			want:    []string{"1h", "4h", "1d"},
			wantErr: false,
		},
		{
			name:    "single timeframe",
			input:   "1h",
			want:    []string{"1h"},
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "all valid formats",
			input:   "1m,5m,15m,1h,4h,1d,1w",
			want:    []string{"1m", "5m", "15m", "1h", "4h", "1d", "1w"},
			wantErr: false,
		},
		{
			name:    "invalid timeframe",
			input:   "1h,invalid,1d",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTimeframes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimeframes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTimeframes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidTimeframe(t *testing.T) {
	tests := []struct {
		name      string
		timeframe string
		want      bool
	}{
		{"1 minute", "1m", true},
		{"5 minutes", "5m", true},
		{"15 minutes", "15m", true},
		{"30 minutes", "30m", true},
		{"1 hour", "1h", true},
		{"4 hours", "4h", true},
		{"1 day", "1d", true},
		{"1 week", "1w", true},
		{"1 month", "1M", true},
		{"invalid", "invalid", false},
		{"empty", "", false},
		{"just number", "5", false},
		{"just letter", "h", false},
		{"uppercase H", "1H", false}, // Not in valid list
		{"uppercase D", "1D", false}, // Not in valid list
		{"2 hours", "2h", true},
		{"12 hours", "12h", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidTimeframe(tt.timeframe); got != tt.want {
				t.Errorf("IsValidTimeframe(%s) = %v, want %v", tt.timeframe, got, tt.want)
			}
		})
	}
}

func TestIsValidSymbol(t *testing.T) {
	tests := []struct {
		name   string
		symbol string
		want   bool
	}{
		{"BTC", "BTC", true},
		{"ETH", "ETH", true},
		{"lowercase btc", "btc", false}, // Must be uppercase
		{"mixed case", "Btc", false},    // Must be uppercase
		{"with number", "1INCH", false}, // Numbers not allowed
		{"short symbol", "BNB", true},
		{"long symbol", "BITCOIN", true},
		{"empty", "", false},
		{"with space", "BTC ETH", false},
		{"with comma", "BTC,ETH", false},
		{"with slash", "BTC/USDT", false},
		{"with hyphen", "BTC-USDT", false},
		{"special chars", "BTC@#$", false},
		{"too short", "B", false},
		{"too long", "VERYLONGSYMBOL", false},
		{"USDT", "USDT", true},
		{"ADA", "ADA", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidSymbol(tt.symbol); got != tt.want {
				t.Errorf("IsValidSymbol(%s) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase", "hello", "hello"},
		{"uppercase", "HELLO", "hello"},
		{"mixed case", "HeLLo", "hello"},
		{"with spaces", "  hello  ", "hello"},
		{"empty", "", ""},
		{"spaces only", "   ", ""},
		{"mixed case with spaces", "  HeLLo WoRLd  ", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.input); got != tt.want {
				t.Errorf("Normalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSingleWord(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"single word", "bitcoin", true},
		{"with space", "bitcoin ethereum", false},
		{"with comma", "bitcoin,ethereum", false},
		{"empty", "", false},
		{"uppercase", "BITCOIN", true},
		{"number", "123", true},
		{"mixed", "btc123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSingleWord(tt.s); got != tt.want {
				t.Errorf("IsSingleWord() = %v, want %v", got, tt.want)
			}
		})
	}
}
