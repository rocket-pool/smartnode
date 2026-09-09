package units

import (
	"math/big"

	"github.com/shopspring/decimal"
)

var (
	OneEth Wei
)

func init() {
	OneEth = Wei{Decimal: decimal.New(1, ethExp)}
}

func NewWei(value *big.Int) Wei {
	if value == nil {
		return Wei{Decimal: decimal.Zero}
	}
	return Wei{Decimal: decimal.NewFromBigInt(value, 0)}
}

func WeiFromUint64(value uint64) Wei {
	return Wei{Decimal: decimal.NewFromUint64(value)}
}

// BigInt returns the integer wei representation. Prefer this over relying on
// the promoted decimal.Decimal.BigInt method at call sites.
func (w Wei) BigInt() *big.Int {
	return w.Decimal.BigInt()
}

func (w Wei) ToWei() Wei {
	return Wei{w.Decimal.Round(weiExp)}
}

func (w Wei) ToEth() Eth {
	return Eth{w.Decimal.Shift(weiExp - ethExp).Round(ethExp)}
}

func (w Wei) ToGwei() Gwei {
	return Gwei{w.Decimal.Shift(weiExp - gweiExp).Round(gweiExp)}
}

func (w Wei) ToMilliEth() MilliEth {
	return MilliEth{w.Decimal.Shift(weiExp - milliEthExp).Round(milliEthExp)}
}

func (w Wei) Add(other Wei) Wei {
	return Wei{w.Decimal.Add(other.Decimal)}
}

func (w Wei) Div(other Wei) Wei {
	return Wei{w.Decimal.DivRound(other.Decimal, weiExp)}
}

func (w Wei) Cmp(other Wei) int {
	return w.Decimal.Cmp(other.Decimal)
}

func (w Wei) Mul(other Wei) Wei {
	return Wei{w.Decimal.Mul(other.Decimal)}
}

func (w Wei) Sub(other Wei) Wei {
	return Wei{w.Decimal.Sub(other.Decimal)}
}

func (w *Wei) UnmarshalText(text []byte) error {
	if err := w.Decimal.UnmarshalText(text); err != nil {
		return err
	}
	w.Decimal = w.Decimal.Round(weiExp)
	return nil
}
