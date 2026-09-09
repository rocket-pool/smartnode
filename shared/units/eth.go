package units

import (
	"github.com/shopspring/decimal"
)

func NewEth(value uint64) Eth {
	return Eth{Decimal: decimal.NewFromUint64(value)}
}

func EthFromInt(value int) Eth {
	return Eth{Decimal: decimal.NewFromInt(int64(value))}
}

// EthFromFloat constructs an Eth from a float64 amount, preserving the
// float's shortest round-trip decimal representation.
func EthFromFloat(value float64) Eth {
	return Eth{Decimal: decimalFromFloat(value)}
}

func (e Eth) ToWei() Wei {
	return Wei{e.Decimal.Shift(ethExp - weiExp).Round(weiExp)}
}

func (e Eth) ToEth() Eth {
	return Eth{e.Decimal.Round(ethExp)}
}

func (e Eth) ToGwei() Gwei {
	return Gwei{e.Decimal.Shift(ethExp - gweiExp).Round(gweiExp)}
}

func (e Eth) ToMilliEth() MilliEth {
	return MilliEth{e.Decimal.Shift(ethExp - milliEthExp).Round(milliEthExp)}
}

func (e Eth) Add(other Eth) Eth {
	return Eth{e.Decimal.Add(other.Decimal).Round(ethExp)}
}

func (e Eth) Div(other Eth) Eth {
	return Eth{e.Decimal.Div(other.Decimal).Round(ethExp)}
}

func (e Eth) Mul(other Eth) Eth {
	return Eth{e.Decimal.Mul(other.Decimal).Round(ethExp)}
}

func (e Eth) Pow(power float64) Eth {
	return Eth{e.Decimal.Pow(decimal.NewFromFloat(power)).Round(ethExp)}
}

func (e Eth) Round(precision int32) Eth {
	return Eth{e.Decimal.Round(precision)}
}

func (e Eth) Sub(other Eth) Eth {
	return Eth{e.Decimal.Sub(other.Decimal).Round(ethExp)}
}

func (e *Eth) UnmarshalText(text []byte) error {
	if err := e.Decimal.UnmarshalText(text); err != nil {
		return err
	}
	e.Decimal = e.Decimal.Round(ethExp)
	return nil
}
