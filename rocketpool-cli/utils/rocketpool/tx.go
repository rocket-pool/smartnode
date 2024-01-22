package rocketpool

import (
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type TxRequester struct {
	client *http.Client
}

func NewTxRequester(client *http.Client) *TxRequester {
	return &TxRequester{
		client: client,
	}
}

func (r *TxRequester) GetName() string {
	return "TX"
}
func (r *TxRequester) GetRoute() string {
	return "tx"
}
func (r *TxRequester) GetClient() *http.Client {
	return r.client
}

// Use the node private key to sign a transaction without submitting it
func (r *TxRequester) SignTx(txSubmission *core.TransactionSubmission, nonce *big.Int, maxFee *big.Int, maxPriorityFee *big.Int) (*api.ApiResponse[api.TxSignTxData], error) {
	body := api.SubmitTxBody{
		Submission:     txSubmission,
		Nonce:          nonce,
		MaxFee:         maxFee,
		MaxPriorityFee: maxPriorityFee,
	}
	return sendPostRequest[api.TxSignTxData](r, "sign-tx", "SignTx", body)
}

// Submit a transaction
func (r *TxRequester) SubmitTx(txSubmission *core.TransactionSubmission, nonce *big.Int, maxFee *big.Int, maxPriorityFee *big.Int) (*api.ApiResponse[api.TxData], error) {
	body := api.SubmitTxBody{
		Submission:     txSubmission,
		Nonce:          nonce,
		MaxFee:         maxFee,
		MaxPriorityFee: maxPriorityFee,
	}
	return sendPostRequest[api.TxData](r, "submit-tx", "SubmitTx", body)
}

// Use the node private key to sign a batch of transactions without submitting them
func (r *TxRequester) SignTxBatch(txSubmissions []*core.TransactionSubmission, firstNonce *big.Int, maxFee *big.Int, maxPriorityFee *big.Int) (*api.ApiResponse[api.TxBatchSignTxData], error) {
	body := api.BatchSubmitTxsBody{
		Submissions:    txSubmissions,
		FirstNonce:     firstNonce,
		MaxFee:         maxFee,
		MaxPriorityFee: maxPriorityFee,
	}
	return sendPostRequest[api.TxBatchSignTxData](r, "batch-sign-tx", "SignTxBatch", body)
}

// Submit a batch of transactions
func (r *TxRequester) SubmitTxBatch(txSubmissions []*core.TransactionSubmission, firstNonce *big.Int, maxFee *big.Int, maxPriorityFee *big.Int) (*api.ApiResponse[api.BatchTxData], error) {
	body := api.BatchSubmitTxsBody{
		Submissions:    txSubmissions,
		FirstNonce:     firstNonce,
		MaxFee:         maxFee,
		MaxPriorityFee: maxPriorityFee,
	}
	return sendPostRequest[api.BatchTxData](r, "batch-submit-tx", "SubmitTxBatch", body)
}

// Wait for a transaction
func (r *TxRequester) WaitForTransaction(txHash common.Hash) (*api.ApiResponse[api.SuccessData], error) {
	args := map[string]string{
		"hash": txHash.Hex(),
	}
	return sendGetRequest[api.SuccessData](r, "wait", "WaitForTransaction", args)
}
