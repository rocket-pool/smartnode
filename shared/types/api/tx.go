package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
)

type TxSignTxData struct {
	SignedTx string `json:"signedTx"`
}

type TxBatchSignTxData struct {
	SignedTxs []string `json:"signedTxs"`
}

type TxData struct {
	TxHash common.Hash `json:"txHash"`
}

type BatchTxData struct {
	TxHashes []common.Hash `json:"txHashes"`
}

type SubmitTxBody struct {
	Submission     *eth.TransactionSubmission `json:"submission"`
	Nonce          *big.Int                   `json:"nonce,omitempty"`
	MaxFee         *big.Int                   `json:"maxFee"`
	MaxPriorityFee *big.Int                   `json:"maxPriorityFee"`
}

type BatchSubmitTxsBody struct {
	Submissions    []*eth.TransactionSubmission `json:"submissions"`
	FirstNonce     *big.Int                     `json:"firstNonce,omitempty"`
	MaxFee         *big.Int                     `json:"maxFee"`
	MaxPriorityFee *big.Int                     `json:"maxPriorityFee"`
}
