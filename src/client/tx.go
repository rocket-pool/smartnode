package client

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type TxRequester struct {
	context *client.RequesterContext
}

func NewTxRequester(context *client.RequesterContext) *TxRequester {
	return &TxRequester{
		context: context,
	}
}

func (r *TxRequester) GetName() string {
	return "TX"
}
func (r *TxRequester) GetRoute() string {
	return "tx"
}
func (r *TxRequester) GetContext() *client.RequesterContext {
	return r.context
}

// Use the node private key to sign a transaction without submitting it
func (r *TxRequester) SignTx(txSubmission *eth.TransactionSubmission, nonce *big.Int, maxFee *big.Int, maxPriorityFee *big.Int) (*types.ApiResponse[api.TxSignTxData], error) {
	body := api.SubmitTxBody{
		Submission:     txSubmission,
		Nonce:          nonce,
		MaxFee:         maxFee,
		MaxPriorityFee: maxPriorityFee,
	}
	return client.SendPostRequest[api.TxSignTxData](r, "sign-tx", "SignTx", body)
}

// Submit a transaction
func (r *TxRequester) SubmitTx(txSubmission *eth.TransactionSubmission, nonce *big.Int, maxFee *big.Int, maxPriorityFee *big.Int) (*types.ApiResponse[api.TxData], error) {
	body := api.SubmitTxBody{
		Submission:     txSubmission,
		Nonce:          nonce,
		MaxFee:         maxFee,
		MaxPriorityFee: maxPriorityFee,
	}
	return client.SendPostRequest[api.TxData](r, "submit-tx", "SubmitTx", body)
}

// Use the node private key to sign a batch of transactions without submitting them
func (r *TxRequester) SignTxBatch(txSubmissions []*eth.TransactionSubmission, firstNonce *big.Int, maxFee *big.Int, maxPriorityFee *big.Int) (*types.ApiResponse[api.TxBatchSignTxData], error) {
	body := api.BatchSubmitTxsBody{
		Submissions:    txSubmissions,
		FirstNonce:     firstNonce,
		MaxFee:         maxFee,
		MaxPriorityFee: maxPriorityFee,
	}
	return client.SendPostRequest[api.TxBatchSignTxData](r, "batch-sign-tx", "SignTxBatch", body)
}

// Submit a batch of transactions
func (r *TxRequester) SubmitTxBatch(txSubmissions []*eth.TransactionSubmission, firstNonce *big.Int, maxFee *big.Int, maxPriorityFee *big.Int) (*types.ApiResponse[api.BatchTxData], error) {
	body := api.BatchSubmitTxsBody{
		Submissions:    txSubmissions,
		FirstNonce:     firstNonce,
		MaxFee:         maxFee,
		MaxPriorityFee: maxPriorityFee,
	}
	return client.SendPostRequest[api.BatchTxData](r, "batch-submit-tx", "SubmitTxBatch", body)
}

// Wait for a transaction
func (r *TxRequester) WaitForTransaction(txHash common.Hash) (*types.ApiResponse[types.SuccessData], error) {
	args := map[string]string{
		"hash": txHash.Hex(),
	}
	return client.SendGetRequest[types.SuccessData](r, "wait", "WaitForTransaction", args)
}
