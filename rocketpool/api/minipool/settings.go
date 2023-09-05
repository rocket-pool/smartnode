package minipool

import "math/big"

const (
	minipoolBatchSize              int = 100
	minipoolCompleteShareBatchSize int = 500
)

var zeroVar *big.Int

func zero() *big.Int {
	if zeroVar == nil {
		zeroVar = big.NewInt(0)
	}
	return zeroVar
}
