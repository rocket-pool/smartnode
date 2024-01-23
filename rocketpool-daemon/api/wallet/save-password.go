package wallet

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
	server.RegisterQuerylessGet[*walletSavePasswordContext, api.SuccessData](
		router, "save-password", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletSavePasswordContext struct {
	handler *WalletHandler
}

func (c *walletSavePasswordContext) PrepareData(data *api.SuccessData, opts *bind.TransactOpts) error {
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
