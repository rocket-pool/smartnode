package tx

import (
	"context"
	"fmt"
	"math/big"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type txBatchSubmitTxsContextFactory struct {
	handler *TxHandler
}

func (f *txBatchSubmitTxsContextFactory) Create(body api.BatchSubmitTxsBody) (*txBatchSubmitTxsContext, error) {
	c := &txBatchSubmitTxsContext{
		handler: f.handler,
		body:    body,
	}
	// Validate the submissions
	for i, submission := range body.Submissions {
		if submission.TxInfo == nil {
			return nil, fmt.Errorf("submission %d TX info must be set", i)
		}
		if submission.GasLimit == 0 {
			return nil, fmt.Errorf("submission %d gas limit must be set", i)
		}
	}
	if body.MaxFee == nil {
		return nil, fmt.Errorf("submission max fee must be set")
	}
	if body.MaxPriorityFee == nil {
		return nil, fmt.Errorf("submission max priority fee must be set")
	}
	return c, nil
}

func (f *txBatchSubmitTxsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessPost[*txBatchSubmitTxsContext, api.BatchSubmitTxsBody, api.BatchTxData](
		router, "batch-submit-txs", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type txBatchSubmitTxsContext struct {
	handler *TxHandler
	body    api.BatchSubmitTxsBody
}

func (c *txBatchSubmitTxsContext) PrepareData(data *api.BatchTxData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	ec := sp.GetEthClient()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireWalletReady()
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	// Get the first nonce
	var currentNonce *big.Int
	if c.body.FirstNonce != nil {
		currentNonce = c.body.FirstNonce
	} else {
		nonce, err := ec.NonceAt(context.Background(), nodeAddress, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting latest nonce for node: %w", err)
		}
		currentNonce = big.NewInt(0).SetUint64(nonce)
	}

	txHashes := make([]common.Hash, len(c.body.Submissions))
	opts.GasFeeCap = c.body.MaxFee
	opts.GasTipCap = c.body.MaxPriorityFee
	for i, submission := range c.body.Submissions {
		opts.Nonce = currentNonce
		opts.GasLimit = submission.GasLimit

		tx, err := rp.ExecuteTransaction(submission.TxInfo, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error submitting transaction %d: %w", i, err)
		}
		txHashes[i] = tx.Hash()

		// Update the nonce to the next one
		currentNonce.Add(currentNonce, common.Big1)
	}

	data.TxHashes = txHashes
	return types.ResponseStatus_Success, nil
}
