package wallet

import (
	"errors"
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type walletSignTxContextFactory struct {
	handler *WalletHandler
}

func (f *walletSignTxContextFactory) Create(args url.Values) (*walletSignTxContext, error) {
	c := &walletSignTxContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("tx", args, input.ValidateByteArray, &c.tx),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletSignTxContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletSignTxContext, api.WalletSignTxData](
		router, "sign-tx", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletSignTxContext struct {
	handler *WalletHandler
	tx      []byte
}

func (c *walletSignTxContext) PrepareData(data *api.WalletSignTxData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	// Requirements
	err := sp.RequireWalletReady()
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	signedBytes, err := w.SignTransaction(c.tx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error signing transaction: %w", err)
	}
	data.SignedTx = signedBytes
	return types.ResponseStatus_Success, nil
}
