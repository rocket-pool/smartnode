package tx

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type txSignTxContextFactory struct {
	handler *TxHandler
}

func (f *txSignTxContextFactory) Create(args url.Values) (*txSignTxContext, error) {
	c := &txSignTxContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("tx-info", args, input.ValidateTxInfo, &c.txInfo),
	}
	return c, errors.Join(inputErrs...)
}

func (f *txSignTxContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*txSignTxContext, api.TxSignTxData](
		router, "sign-tx", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type txSignTxContext struct {
	handler *TxHandler
	txInfo  *core.TransactionInfo
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

	tx, err := rp.SignTransaction(c.txInfo, opts)
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
