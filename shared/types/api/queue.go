package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
)

type QueueStatusResponse struct {
	APIResponse
	DepositPoolBalance    *big.Int `json:"depositPoolBalance"`
	MinipoolQueueLength   uint64   `json:"minipoolQueueLength"`
	MinipoolQueueCapacity *big.Int `json:"minipoolQueueCapacity"`
}

type CanProcessQueueResponse struct {
	APIResponse
	CanProcess                 bool            `json:"canProcess"`
	AssignDepositsDisabled     bool            `json:"assignDepositsDisabled"`
	NoMinipoolsAvailable       bool            `json:"noMinipoolsAvailable"`
	InsufficientDepositBalance bool            `json:"insufficientDepositBalance"`
	GasLimits                  gaslimit.Limits `json:"gasLimits"`
}
type ProcessQueueResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type GetQueueDetailsResponse struct {
	APIResponse
	TotalLength    uint32 `json:"totalLength"`
	ExpressLength  uint32 `json:"expressLength"`
	StandardLength uint32 `json:"standardLength"`
	ExpressRate    uint64 `json:"expressRate"`
	QueueIndex     uint32 `json:"queueIndex"`
}

type CanAssignDepositsResponse struct {
	APIResponse
	CanAssign              bool            `json:"canAssign"`
	AssignDepositsDisabled bool            `json:"assignDepositsDisabled"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
}

type AssignDepositsResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}
