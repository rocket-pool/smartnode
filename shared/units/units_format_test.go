package units

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
)

func TestEthFormatFloat(t *testing.T) {
	eth := Eth{Decimal: decimal.RequireFromString("1.23456789")}

	tests := []struct {
		format string
		want   string
	}{
		{"%f", "1.234568"}, // default 6 places, rounded
		{"%.2f", "1.23"},
		{"%.0f", "1"},
		{"%.8f", "1.23456789"},
		{"%10.2f", "      1.23"},
		{"%-10.2f", "1.23      "},
		{"%v", "1.23456789"},
		{"%s", "1.23456789"},
		{"%#v", "units.Eth(1.23456789)"},
	}

	for _, tt := range tests {
		got := fmt.Sprintf(tt.format, eth)
		if got != tt.want {
			t.Errorf("Sprintf(%q, eth) = %q, want %q", tt.format, got, tt.want)
		}
	}
}

func TestGweiMilliFormatFloat(t *testing.T) {
	gwei := Gwei{Decimal: decimal.RequireFromString("12.345678901")}
	if got, want := fmt.Sprintf("%.3f", gwei), "12.346"; got != want {
		t.Errorf("gwei = %q, want %q", got, want)
	}

	milli := MilliEth{Decimal: decimal.RequireFromString("1000.5")}
	if got, want := fmt.Sprintf("%.1f", milli), "1000.5"; got != want {
		t.Errorf("milliEth = %q, want %q", got, want)
	}
}

func TestWeiFormat(t *testing.T) {
	wei := weiFromString("1500000")

	if got, want := fmt.Sprintf("%d", wei), "1500000"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := fmt.Sprintf("%s", wei), "1500000"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := fmt.Sprintf("%.2f", wei), "1500000.00"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := fmt.Sprintf("%#v", wei), "units.Wei(1500000)"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatPreservesDigitsBeyondFloat(t *testing.T) {
	// More digits than float64 can represent; %.18f must keep them via decimal rounding.
	eth := Eth{Decimal: decimal.RequireFromString("0.123456789012345678")}
	got := fmt.Sprintf("%.18f", eth)
	want := "0.123456789012345678"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	asFloat := eth.InexactFloat64()
	floatGot := fmt.Sprintf("%.18f", asFloat)
	if floatGot == want {
		t.Fatal("expected float64 formatting to differ; test value is not discriminating enough")
	}
}
