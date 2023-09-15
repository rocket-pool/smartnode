package api

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/core"
)

type FaucetStatusData struct {
	Balance            *big.Int `json:"balance"`
	Allowance          *big.Int `json:"allowance"`
	WithdrawableAmount *big.Int `json:"withdrawableAmount"`
	WithdrawalFee      *big.Int `json:"withdrawalFee"`
	ResetsInBlocks     uint64   `json:"resetsInBlocks"`
}

type FaucetWithdrawRplData struct {
	CanWithdraw               bool                  `json:"canWithdraw"`
	InsufficientFaucetBalance bool                  `json:"insufficientFaucetBalance"`
	InsufficientAllowance     bool                  `json:"insufficientAllowance"`
	InsufficientNodeBalance   bool                  `json:"insufficientNodeBalance"`
	TxInfo                    *core.TransactionInfo `json:"txInfo"`
}
