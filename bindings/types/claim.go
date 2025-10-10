package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Claim struct
type Claim struct {
	Index               *big.Int
	AmountRPL           *big.Int
	AmountSmoothingETH  *big.Int
	AmountVoterShareETH *big.Int
	Proof               []common.Hash
}

type Claims []Claim
