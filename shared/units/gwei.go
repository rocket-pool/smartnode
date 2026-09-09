package units

import (
	"github.com/shopspring/decimal"
)

func NewGwei(value uint64) Gwei {
	return Gwei{
		Decimal: decimal.NewFromUint64(value),
	}
}

// GweiFromFloat constructs a Gwei from a float64 amount, preserving the
// float's shortest round-trip decimal representation.
func GweiFromFloat(value float64) Gwei {
	return Gwei{Decimal: decimalFromFloat(value)}
}

func (g Gwei) ToWei() Wei {
	return Wei{g.Decimal.Shift(gweiExp - weiExp).Round(weiExp)}
}

func (g Gwei) ToEth() Eth {
	return Eth{g.Decimal.Shift(gweiExp - ethExp).Round(ethExp)}
}

func (g Gwei) ToGwei() Gwei {
	return Gwei{g.Decimal.Round(gweiExp)}
}

func (g Gwei) ToMilliEth() MilliEth {
	return MilliEth{g.Decimal.Shift(gweiExp - milliEthExp).Round(milliEthExp)}
}

func (g *Gwei) UnmarshalText(text []byte) error {
	if err := g.Decimal.UnmarshalText(text); err != nil {
		return err
	}
	g.Decimal = g.Decimal.Round(gweiExp)
	return nil
}
