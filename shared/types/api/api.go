package api

import (
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/smartnode/shared/types"
)

type ApiResponse[Data any] struct {
	WalletStatus types.WalletStatus `json:"walletStatus"`
	Data         *Data              `json:"data"`
}

type SuccessData struct {
	Success bool `json:"success"`
}

type DataBatch[DataType any] struct {
	Batch []DataType `json:"batch"`
}

type TxInfoData struct {
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type BatchTxInfoData struct {
	TxInfos []*core.TransactionInfo `json:"txInfos"`
}
