package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
	"github.com/rocket-pool/smartnode/shared/units"
)

type QueueStatusResponse struct {
	APIResponse
	DepositPoolBalance    units.Wei `json:"depositPoolBalance"`
	MinipoolQueueLength   uint64    `json:"minipoolQueueLength"`
	MinipoolQueueCapacity units.Wei `json:"minipoolQueueCapacity"`
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
