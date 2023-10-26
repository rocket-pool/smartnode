package api

import (
	"math/big"

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

type TxSubmission struct {
	TxInfo         *core.TransactionInfo `json:"txInfo"`
	GasLimit       uint64                `json:"gasLimit"`
	MaxFee         *big.Int              `json:"maxFee"`
	MaxPriorityFee *big.Int              `json:"maxPriorityFee"`
}

type SubmitTxBody struct {
	Submission TxSubmission `json:"submission"`
	Nonce      *big.Int     `json:"nonce,omitempty"`
}

type BatchSubmitTxsBody struct {
	Submissions []TxSubmission `json:"submissions"`
	FirstNonce  *big.Int       `json:"firstNonce,omitempty"`
}
