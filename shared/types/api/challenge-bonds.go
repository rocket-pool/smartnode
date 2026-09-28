package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
)

type CanSettleChallengeBondResponse struct {
	APIResponse
	Saturn2Deployed bool            `json:"saturn2Deployed"`
	CanSettle       bool            `json:"canSettle"`
	Reason          string          `json:"reason"`
	Proposer        common.Address  `json:"proposer"`
	Responder       common.Address  `json:"responder"`
	BondAmount      *big.Int        `json:"bondAmount"`
	EstimatedReward *big.Int        `json:"estimatedReward"`
	EstimatedBurn   *big.Int        `json:"estimatedBurn"`
	GasLimits       gaslimit.Limits `json:"gasLimits"`
}

type SettleChallengeBondResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}
