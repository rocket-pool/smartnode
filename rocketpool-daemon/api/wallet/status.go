package wallet

import (
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type walletStatusFactory struct {
	handler *WalletHandler
}

func (f *walletStatusFactory) Create(args url.Values) (*walletStatusContext, error) {
	c := &walletStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *walletStatusFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletStatusContext, api.WalletStatusData](
		router, "status", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletStatusContext struct {
	handler *WalletHandler
}

func (c *walletStatusContext) PrepareData(data *api.WalletStatusData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	data.WalletStatus = w.GetStatus()
	data.AccountAddress, _ = w.GetAddress()
	return nil
}
