package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
)

type PerformanceChallengeDetails struct {
	Challenge megapool.PerformanceChallenge       `json:"challenge"`
	State     megapool.PerformanceChallengeStatus `json:"state"`
}

type PerformanceChallengesResponse struct {
	APIResponse
	Saturn2Deployed bool                          `json:"saturn2Deployed"`
	Challenges      []PerformanceChallengeDetails `json:"challenges"`
}

type CanActOnPerformanceChallengeResponse struct {
	APIResponse
	CanAct    bool            `json:"canAct"`
	Reason    string          `json:"reason"`
	GasLimits gaslimit.Limits `json:"gasLimits"`
}

type PerformanceChallengeResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}
