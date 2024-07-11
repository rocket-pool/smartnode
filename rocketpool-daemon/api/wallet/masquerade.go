package wallet

import (
	"errors"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
)

// ===============
// === Factory ===
// ===============

type walletMasqueradeContextFactory struct {
	handler *WalletHandler
}

func (f *walletMasqueradeContextFactory) Create(args url.Values) (*walletMasqueradeContext, error) {
	c := &walletMasqueradeContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletMasqueradeContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletMasqueradeContext, types.SuccessData](
		router, "masquerade", f, f.handler.logger.Logger, f.handler.serviceProvider.IServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletMasqueradeContext struct {
	handler *WalletHandler
	address common.Address
}

func (c *walletMasqueradeContext) PrepareData(data *types.SuccessData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	err := w.MasqueradeAsAddress(c.address)
	if err != nil {
		return types.ResponseStatus_Error, err
	}
	return types.ResponseStatus_Success, nil
}
