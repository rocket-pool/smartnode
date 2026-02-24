package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Claim struct
type Claim struct {
	RewardIndex            *big.Int
	AmountRPL              *big.Int
	AmountSmoothingPoolETH *big.Int
	AmountVoterETH         *big.Int
	MerkleProof            []common.Hash
}

type Claims []Claim
