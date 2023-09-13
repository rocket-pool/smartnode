package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/core"
)

type ApiResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type TxInfoResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type BatchTxInfoResponse struct {
	Status  string                  `json:"status"`
	Error   string                  `json:"error"`
	TxInfos []*core.TransactionInfo `json:"txInfos"`
}

type TxResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type BatchTxResponse struct {
	Status   string        `json:"status"`
	Error    string        `json:"error"`
	TxHashes []common.Hash `json:"txHashes"`
}

type SignedTxResponse struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	TxBytes []byte `json:"txBytes"`
}
