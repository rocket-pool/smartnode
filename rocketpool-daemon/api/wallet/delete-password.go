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
	server.RegisterQuerylessGet[*walletDeletePasswordContext, types.SuccessData](
		router, "delete-password", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletDeletePasswordContext struct {
	handler *WalletHandler
}

func (c *walletDeletePasswordContext) PrepareData(data *types.SuccessData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	err := w.DeletePassword()
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error deleting wallet password from disk: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
