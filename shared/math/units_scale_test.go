package math

import (
	"math/big"
	"testing"
)

func TestWeiToEthWithDecimalsExactScale(t *testing.T) {
	for _, decimals := range []uint8{0, 6, 18, 23, 24, 77, 255} {
		scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
		for _, units := range []int64{1, 3, 7, 12345, -12345, 9007199254740991} {
			amount := big.NewInt(units)
			want, _ := new(big.Rat).SetFrac(amount, scale).Float64()
			if got := WeiToEthWithDecimals(amount, decimals); got != want {
				t.Errorf("units=%d decimals=%d: got %.17g, want %.17g", units, decimals, got, want)
			}
			if amount.Int64() != units {
				t.Fatal("input amount was modified")
			}
		}
	}
	if WeiToEthWithDecimals(nil, 18) != 0 {
		t.Fatal("nil amount must convert to zero")
	}
}
