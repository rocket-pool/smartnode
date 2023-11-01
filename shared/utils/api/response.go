package api

import (
	"math/big"
)

func ZeroIfNil(in **big.Int) {
	if *in == nil {
		*in = big.NewInt(0)
	}
}
