package wallet

import (
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
)

// ===============
// === Factory ===
// ===============

type walletRestoreAddressContextFactory struct {
	handler *WalletHandler
}

func (f *walletRestoreAddressContextFactory) Create(args url.Values) (*walletRestoreAddressContext, error) {
	c := &walletRestoreAddressContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *walletRestoreAddressContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletRestoreAddressContext, types.SuccessData](
		router, "restore-address", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletRestoreAddressContext struct {
	handler *WalletHandler
}

func (c *walletRestoreAddressContext) PrepareData(data *types.SuccessData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	err := w.RestoreAddressToWallet()
	if err != nil {
		return types.ResponseStatus_Error, err
	}
	return types.ResponseStatus_Success, nil
}
