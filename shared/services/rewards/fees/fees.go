package fees

import (
	"math/big"
)

var oneEth = big.NewInt(1000000000000000000)
var tenEth = big.NewInt(0).Mul(oneEth, big.NewInt(10))
var pointOhFourEth = big.NewInt(40000000000000000)
var pointOneEth = big.NewInt(0).Div(oneEth, big.NewInt(10))
var sixteenEth = big.NewInt(0).Mul(oneEth, big.NewInt(16))

func GetMinipoolFeeWithBonus(bond, fee, percentOfBorrowedEth *big.Int) *big.Int {
	if bond.Cmp(sixteenEth) >= 0 {
		return fee
	}
	// fee = max(fee, 0.10 Eth + (0.04 Eth * min(10 Eth, percentOfBorrowedETH) / 10 Eth))
	_min := big.NewInt(0).Set(tenEth)
	if _min.Cmp(percentOfBorrowedEth) > 0 {
		_min.Set(percentOfBorrowedEth)
	}
	dividend := _min.Mul(_min, pointOhFourEth)
	divResult := dividend.Div(dividend, tenEth)
	feeWithBonus := divResult.Add(divResult, pointOneEth)
	if fee.Cmp(feeWithBonus) >= 0 {
		return fee
	}
	return feeWithBonus
}
