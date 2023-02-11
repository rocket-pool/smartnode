package state

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

const (
	threadLimit int = 6
)

// Global constants
var zero = big.NewInt(0)
var two = big.NewInt(2)
var oneInWei = eth.EthToWei(1)
