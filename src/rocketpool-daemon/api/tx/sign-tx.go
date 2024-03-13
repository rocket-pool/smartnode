package tx

import (
	"encoding/hex"
	"errors"
	"fmt"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type txSignTxContextFactory struct {
	handler *TxHandler
}

func (f *txSignTxContextFactory) Create(body api.SubmitTxBody) (*txSignTxContext, error) {
	c := &txSignTxContext{
		handler: f.handler,
		body:    body,
	}
	// Validate the submission
	if body.Submission.TxInfo == nil {
		return nil, fmt.Errorf("submission TX info must be set")
	}
	if body.Submission.GasLimit == 0 {
		return nil, fmt.Errorf("submission gas limit must be set")
	}
	if body.MaxFee == nil {
		return nil, fmt.Errorf("submission max fee must be set")
	}
	if body.MaxPriorityFee == nil {
		return nil, fmt.Errorf("submission max priority fee must be set")
	}
	return c, nil
}

func (f *txSignTxContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessPost[*txSignTxContext, api.SubmitTxBody, api.TxSignTxData](
		router, "sign-tx", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type txSignTxContext struct {
	handler *TxHandler
	body    api.SubmitTxBody
}

func (c *txSignTxContext) PrepareData(data *api.TxSignTxData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()

	err := errors.Join(
		sp.RequireWalletReady(),
	)
	if err != nil {
		return err
	}

	if c.body.Nonce != nil {
		opts.Nonce = c.body.Nonce
	}
	opts.GasLimit = c.body.Submission.GasLimit
	opts.GasFeeCap = c.body.MaxFee
	opts.GasTipCap = c.body.MaxPriorityFee

	tx, err := rp.SignTransaction(c.body.Submission.TxInfo, opts)
	if err != nil {
		return fmt.Errorf("error signing transaction: %w", err)
	}

	bytes, err := tx.MarshalBinary()
	if err != nil {
		return fmt.Errorf("error marshalling transaction: %w", err)
	}
	data.SignedTx = hex.EncodeToString(bytes)
	return nil
}
