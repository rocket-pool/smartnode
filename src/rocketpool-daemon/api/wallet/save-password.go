package wallet

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
)

// ===============
// === Factory ===
// ===============

type walletSavePasswordContextFactory struct {
	handler *WalletHandler
}

func (f *walletSavePasswordContextFactory) Create(args url.Values) (*walletSavePasswordContext, error) {
	c := &walletSavePasswordContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *walletSavePasswordContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletSavePasswordContext, types.SuccessData](
		router, "save-password", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletSavePasswordContext struct {
	handler *WalletHandler
}

func (c *walletSavePasswordContext) PrepareData(data *types.SuccessData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	_, hasPassword := w.GetPassword()
	if !hasPassword {
		return fmt.Errorf("wallet password has not been set yet")
	}
	err := w.SavePassword()
	if err != nil {
		return fmt.Errorf("error saving wallet password to disk: %w", err)
	}

	data.Success = true
	return nil
}
