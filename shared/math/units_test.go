package math

import "testing"

func TestEthToWeiWithDecimalsPrecision(t *testing.T) {
	tests := []struct {
		amount   float64
		decimals uint8
		want     string
	}{
		{1, 24, "1000000000000000000000000"},
		{1, 36, "1000000000000000000000000000000000000"},
		{0.000001, 6, "1"},
		{0.29, 18, "290000000000000000"},
		{-0.29, 18, "-290000000000000000"},
		{1.999, 2, "199"},
		{-1.999, 2, "-199"},
	}
	for _, test := range tests {
		if got := EthToWeiWithDecimals(test.amount, test.decimals).String(); got != test.want {
			t.Errorf("EthToWeiWithDecimals(%v, %d) = %s, want %s", test.amount, test.decimals, got, test.want)
		}
	}
}
