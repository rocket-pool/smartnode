package rewards

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type RollingRecord struct {
	StartSlot                 uint64
	CurrentSlot               uint64
	SuccessfulAttestations    uint64
	TotalAttestationScore     *big.Int
	MinipoolAttestationScores map[common.Address]*big.Int
}
