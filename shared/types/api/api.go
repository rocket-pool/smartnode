package api

import "github.com/rocket-pool/rocketpool-go/core"

type ApiResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type TxResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type BatchTxResponse struct {
	Status  string                  `json:"status"`
	Error   string                  `json:"error"`
	TxInfos []*core.TransactionInfo `json:"txInfos"`
}
