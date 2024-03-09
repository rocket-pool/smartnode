package api

import (
	"math/big"

	"github.com/rocket-pool/node-manager-core/eth"
)

type FaucetStatusData struct {
	Balance            *big.Int `json:"balance"`
	Allowance          *big.Int `json:"allowance"`
	WithdrawableAmount *big.Int `json:"withdrawableAmount"`
	WithdrawalFee      *big.Int `json:"withdrawalFee"`
	ResetsInBlocks     uint64   `json:"resetsInBlocks"`
}

type FaucetWithdrawRplData struct {
	Amount                    *big.Int             `json:"amount"`
	CanWithdraw               bool                 `json:"canWithdraw"`
	InsufficientFaucetBalance bool                 `json:"insufficientFaucetBalance"`
	InsufficientAllowance     bool                 `json:"insufficientAllowance"`
	InsufficientNodeBalance   bool                 `json:"insufficientNodeBalance"`
	TxInfo                    *eth.TransactionInfo `json:"txInfo"`
}
