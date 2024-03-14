package wallet

import (
	"errors"
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
		router, "sign-tx", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletSignTxContext struct {
	handler *WalletHandler
	tx      []byte
}

func (c *walletSignTxContext) PrepareData(data *api.WalletSignTxData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	err := errors.Join(
		sp.RequireWalletReady(),
	)
	if err != nil {
		return err
	}

	signedBytes, err := w.SignTransaction(c.tx)
	if err != nil {
		return fmt.Errorf("error signing transaction: %w", err)
	}
	data.SignedTx = signedBytes
	return nil
}
