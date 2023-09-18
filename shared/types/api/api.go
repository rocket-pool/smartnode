package api

import (
	"github.com/ethereum/go-ethereum/common"
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

type TxInfoData struct {
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type BatchTxInfoData struct {
	TxInfos []*core.TransactionInfo `json:"txInfos"`
}

type TxData struct {
	TxHash common.Hash `json:"txHash"`
}

type BatchTxData struct {
	TxHashes []common.Hash `json:"txHashes"`
}

type SignedTxData struct {
	TxBytes []byte `json:"txBytes"`
}
