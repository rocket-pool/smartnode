package units

import (
	"testing"

	"github.com/shopspring/decimal"
)

// Canonical fixed-point precisions relative to 1 wei:
//   Wei: 0, Gwei: 9, MilliEth: 15, Eth: 18
func TestUnitRoundingPrecisions(t *testing.T) {
	t.Run("wei is integral", func(t *testing.T) {
		// 0.5 wei rounds to 1 wei (half away from zero / shopspring Round)
		half := Wei{Decimal: decimal.RequireFromString("0.5")}
		if got := half.ToWei(); !got.Equal(decimal.NewFromInt(1)) {
			t.Fatalf("0.5 wei -> ToWei = %s, want 1", got)
		}

		// Eth with an extra half-wei of precision rounds to a whole wei
		eth := Eth{Decimal: decimal.NewFromInt(1).Add(decimal.NewFromInt(5).Shift(-19))} // 1 eth + 0.5 wei
		if got := eth.ToWei(); !got.Equal(decimal.RequireFromString("1000000000000000001")) {
			t.Fatalf("1 eth + 0.5 wei -> ToWei = %s, want 1000000000000000001", got)
		}
	})

	t.Run("eth keeps 18 decimal places", func(t *testing.T) {
		oneWei := weiFromString("1")
		eth := oneWei.ToEth()
		if !eth.Equal(decimal.RequireFromString("0.000000000000000001")) {
			t.Fatalf("1 wei -> eth = %s, want 1e-18", eth)
		}

		// Dust below 1 wei in eth space rounds away
		dust := Eth{Decimal: decimal.NewFromInt(1).Shift(-19)} // 0.1 wei
		if got := dust.ToEth(); !got.Equal(decimal.Zero) {
			t.Fatalf("0.1 wei as eth -> ToEth = %s, want 0", got)
		}
	})

	t.Run("gwei keeps 9 decimal places", func(t *testing.T) {
		oneWei := weiFromString("1")
		gwei := oneWei.ToGwei()
		if !gwei.Equal(decimal.RequireFromString("0.000000001")) {
			t.Fatalf("1 wei -> gwei = %s, want 1e-9", gwei)
		}

		dust := Gwei{Decimal: decimal.NewFromInt(1).Shift(-10)} // 0.1 wei
		if got := dust.ToGwei(); !got.Equal(decimal.Zero) {
			t.Fatalf("0.1 wei as gwei -> ToGwei = %s, want 0", got)
		}
	})

	t.Run("milliEth keeps 15 decimal places", func(t *testing.T) {
		oneWei := weiFromString("1")
		milli := oneWei.ToMilliEth()
		if !milli.Equal(decimal.RequireFromString("0.000000000000001")) {
			t.Fatalf("1 wei -> milliEth = %s, want 1e-15", milli)
		}

		dust := MilliEth{Decimal: decimal.NewFromInt(1).Shift(-16)} // 0.1 wei
		if got := dust.ToMilliEth(); !got.Equal(decimal.Zero) {
			t.Fatalf("0.1 wei as milliEth -> ToMilliEth = %s, want 0", got)
		}
	})

	t.Run("milliEth to eth preserves 1 wei", func(t *testing.T) {
		// Regression: Round(15) on eth would zero out 1e-18
		milli := MilliEth{Decimal: decimal.NewFromInt(1).Shift(-15)} // 1 wei
		eth := milli.ToEth()
		if !eth.Equal(decimal.RequireFromString("0.000000000000000001")) {
			t.Fatalf("1 wei as milliEth -> eth = %s, want 1e-18", eth)
		}
		if !eth.ToWei().Equal(decimal.NewFromInt(1)) {
			t.Fatalf("round-trip wei = %s, want 1", eth.ToWei())
		}
	})

	t.Run("cross-unit conversions stay wei-aligned", func(t *testing.T) {
		cases := []Unit{
			weiFromString("123456789"),
			Gwei{Decimal: decimal.RequireFromString("1.234567891")},
			MilliEth{Decimal: decimal.RequireFromString("0.000123456789012345")},
			Eth{Decimal: decimal.RequireFromString("0.000000000123456789")},
		}
		for _, u := range cases {
			want := u.ToWei()
			if !want.Equal(want.Round(0)) {
				t.Fatalf("%T ToWei not integral: %s", u, want)
			}
			if got := u.ToEth().ToWei(); !got.Equal(want.Decimal) {
				t.Fatalf("%T via eth: got %s want %s", u, got, want)
			}
			if got := u.ToGwei().ToWei(); !got.Equal(want.Decimal) {
				t.Fatalf("%T via gwei: got %s want %s", u, got, want)
			}
			if got := u.ToMilliEth().ToWei(); !got.Equal(want.Decimal) {
				t.Fatalf("%T via milliEth: got %s want %s", u, got, want)
			}
		}
	})
}
