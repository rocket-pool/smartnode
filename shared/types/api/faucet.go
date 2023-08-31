package api

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/core"
)

type FaucetStatusResponse struct {
	Status             string   `json:"status"`
	Error              string   `json:"error"`
	Balance            *big.Int `json:"balance"`
	Allowance          *big.Int `json:"allowance"`
	WithdrawableAmount *big.Int `json:"withdrawableAmount"`
	WithdrawalFee      *big.Int `json:"withdrawalFee"`
	ResetsInBlocks     uint64   `json:"resetsInBlocks"`
}

type FaucetWithdrawRplResponse struct {
	Status                    string                `json:"status"`
	Error                     string                `json:"error"`
	CanWithdraw               bool                  `json:"canWithdraw"`
	InsufficientFaucetBalance bool                  `json:"insufficientFaucetBalance"`
	InsufficientAllowance     bool                  `json:"insufficientAllowance"`
	InsufficientNodeBalance   bool                  `json:"insufficientNodeBalance"`
	TxInfo                    *core.TransactionInfo `json:"txInfo"`
}
