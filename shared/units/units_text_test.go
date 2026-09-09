package units

import (
	"encoding"
	"testing"

	"github.com/shopspring/decimal"
)

func TestUnitTextRoundTrip(t *testing.T) {
	preciseEth := "1234567890.123456789012345678"
	preciseGwei := "9876543210.987654321"
	preciseMilli := "111222333444.555666777"
	preciseWei := "123456789012345678901234567890" // > 2^53

	t.Run("Eth", func(t *testing.T) {
		in := Eth{Decimal: decimal.RequireFromString(preciseEth)}
		assertTextMarshaler(t, in)

		raw, err := in.MarshalText()
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if string(raw) != preciseEth {
			t.Fatalf("marshal = %q, want %q", raw, preciseEth)
		}

		var out Eth
		assertTextUnmarshaler(t, &out)
		if err := out.UnmarshalText(raw); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !in.Equal(out.Decimal) {
			t.Fatalf("round-trip = %s, want %s", out, in)
		}
	})

	t.Run("Gwei", func(t *testing.T) {
		in := Gwei{Decimal: decimal.RequireFromString(preciseGwei)}
		assertTextMarshaler(t, in)

		raw, err := in.MarshalText()
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if string(raw) != preciseGwei {
			t.Fatalf("marshal = %q, want %q", raw, preciseGwei)
		}

		var out Gwei
		assertTextUnmarshaler(t, &out)
		if err := out.UnmarshalText(raw); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !in.Equal(out.Decimal) {
			t.Fatalf("round-trip = %s, want %s", out, in)
		}
	})

	t.Run("MilliEth", func(t *testing.T) {
		in := MilliEth{Decimal: decimal.RequireFromString(preciseMilli)}
		assertTextMarshaler(t, in)

		raw, err := in.MarshalText()
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if string(raw) != preciseMilli {
			t.Fatalf("marshal = %q, want %q", raw, preciseMilli)
		}

		var out MilliEth
		assertTextUnmarshaler(t, &out)
		if err := out.UnmarshalText(raw); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !in.Equal(out.Decimal) {
			t.Fatalf("round-trip = %s, want %s", out, in)
		}
	})

	t.Run("Wei", func(t *testing.T) {
		in := weiFromString(preciseWei)
		assertTextMarshaler(t, in)

		raw, err := in.MarshalText()
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if string(raw) != preciseWei {
			t.Fatalf("marshal = %q, want %q", raw, preciseWei)
		}

		var out Wei
		assertTextUnmarshaler(t, &out)
		if err := out.UnmarshalText(raw); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !in.Equal(out.Decimal) {
			t.Fatalf("round-trip = %s, want %s", out, in)
		}
	})
}

func TestUnitTextDoesNotUseFloat(t *testing.T) {
	const amount = "0.123456789012345678"
	in := Eth{Decimal: decimal.RequireFromString(amount)}

	raw, err := in.MarshalText()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(raw) != amount {
		t.Fatalf("marshal = %q, want %q", raw, amount)
	}

	asFloat, err := decimal.NewFromString(amount)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if decimal.NewFromFloat(asFloat.InexactFloat64()).Equal(in.Decimal) {
		t.Fatal("expected float64 path to lose precision; test value is not discriminating enough")
	}

	var out Eth
	if err := out.UnmarshalText(raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !out.Equal(in.Decimal) {
		t.Fatalf("text path lost precision: got %s, want %s", out, in)
	}
}

func TestWeiTextPreservesBeyondFloatMantissa(t *testing.T) {
	in := weiFromString("9007199254740993") // 2^53 + 1

	raw, err := in.MarshalText()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(raw) != in.String() {
		t.Fatalf("marshal = %q, want %q", raw, in.String())
	}

	if decimal.NewFromFloat(in.InexactFloat64()).Equal(in.Decimal) {
		t.Fatal("expected float64 path to lose precision for 2^53+1")
	}

	var out Wei
	if err := out.UnmarshalText(raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !in.Equal(out.Decimal) {
		t.Fatalf("text path lost precision: got %s, want %s", out, in)
	}
}

func assertTextMarshaler(t *testing.T, v any) {
	t.Helper()
	if _, ok := v.(encoding.TextMarshaler); !ok {
		t.Fatalf("%T does not implement encoding.TextMarshaler by value", v)
	}
}

func assertTextUnmarshaler(t *testing.T, v any) {
	t.Helper()
	if _, ok := v.(encoding.TextUnmarshaler); !ok {
		t.Fatalf("%T does not implement encoding.TextUnmarshaler", v)
	}
}
