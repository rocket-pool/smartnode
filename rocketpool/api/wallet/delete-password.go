package wallet

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type walletDeletePasswordContextFactory struct {
	handler *WalletHandler
}

func (f *walletDeletePasswordContextFactory) Create(args url.Values) (*walletDeletePasswordContext, error) {
	c := &walletDeletePasswordContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *walletDeletePasswordContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletDeletePasswordContext, api.SuccessData](
		router, "delete-password", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletDeletePasswordContext struct {
	handler  *WalletHandler
	password []byte
	save     bool
}

func (c *walletDeletePasswordContext) PrepareData(data *api.SuccessData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	err := w.DeletePassword()
	if err != nil {
		return fmt.Errorf("error deleting wallet password from disk: %w", err)
	}

	data.Success = true
	return nil
}
