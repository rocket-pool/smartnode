package api

import (
	"math/big"

	"github.com/rocket-pool/node-manager-core/eth"
)

type QueueProcessData struct {
	CanProcess                 bool                 `json:"canProcess"`
	AssignDepositsDisabled     bool                 `json:"assignDepositsDisabled"`
	NoMinipoolsAvailable       bool                 `json:"noMinipoolsAvailable"`
	InsufficientDepositBalance bool                 `json:"insufficientDepositBalance"`
	TxInfo                     *eth.TransactionInfo `json:"txInfo"`
}

type QueueStatusData struct {
	DepositPoolBalance    *big.Int `json:"depositPoolBalance"`
	MinipoolQueueLength   uint64   `json:"minipoolQueueLength"`
	MinipoolQueueCapacity *big.Int `json:"minipoolQueueCapacity"`
}
