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

type walletForgetPasswordContextFactory struct {
	handler *WalletHandler
}

func (f *walletForgetPasswordContextFactory) Create(args url.Values) (*walletForgetPasswordContext, error) {
	c := &walletForgetPasswordContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *walletForgetPasswordContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletForgetPasswordContext, api.SuccessData](
		router, "forget-password", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletForgetPasswordContext struct {
	handler  *WalletHandler
	password []byte
	save     bool
}

func (c *walletForgetPasswordContext) PrepareData(data *api.SuccessData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	w.ForgetPassword()
	data.Success = true
	return nil
}
