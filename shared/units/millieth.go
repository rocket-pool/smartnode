package units

// MilliEthFromFloat constructs a MilliEth from a float64 amount, preserving
// the float's shortest round-trip decimal representation.
func MilliEthFromFloat(value float64) MilliEth {
	return MilliEth{Decimal: decimalFromFloat(value)}
}

func (m MilliEth) ToWei() Wei {
	return Wei{m.Decimal.Shift(milliEthExp - weiExp).Round(weiExp)}
}

func (m MilliEth) ToEth() Eth {
	return Eth{m.Decimal.Shift(milliEthExp - ethExp).Round(ethExp)}
}

func (m MilliEth) ToGwei() Gwei {
	return Gwei{m.Decimal.Shift(milliEthExp - gweiExp).Round(gweiExp)}
}

func (m MilliEth) ToMilliEth() MilliEth {
	return MilliEth{m.Decimal.Round(milliEthExp)}
}

func (m *MilliEth) UnmarshalText(text []byte) error {
	if err := m.Decimal.UnmarshalText(text); err != nil {
		return err
	}
	m.Decimal = m.Decimal.Round(milliEthExp)
	return nil
}
