package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
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

type CanFaucetWithdrawRplResponse struct {
	Status                    string             `json:"status"`
	Error                     string             `json:"error"`
	CanWithdraw               bool               `json:"canWithdraw"`
	InsufficientFaucetBalance bool               `json:"insufficientFaucetBalance"`
	InsufficientAllowance     bool               `json:"insufficientAllowance"`
	InsufficientNodeBalance   bool               `json:"insufficientNodeBalance"`
	GasInfo                   rocketpool.GasInfo `json:"gasInfo"`
}
type FaucetWithdrawRplResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	Amount *big.Int    `json:"amount"`
	TxHash common.Hash `json:"txHash"`
}
