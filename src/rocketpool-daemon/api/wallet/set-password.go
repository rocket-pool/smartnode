package wallet

import (
	"errors"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
)

// ===============
// === Factory ===
// ===============

type walletSetPasswordContextFactory struct {
	handler *WalletHandler
}

func (f *walletSetPasswordContextFactory) Create(args url.Values) (*walletSetPasswordContext, error) {
	c := &walletSetPasswordContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("password", args, input.ValidateNodePassword, &c.password),
		server.ValidateArg("save", args, input.ValidateBool, &c.save),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletSetPasswordContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletSetPasswordContext, types.SuccessData](
		router, "set-password", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletSetPasswordContext struct {
	handler  *WalletHandler
	password string
	save     bool
}

func (c *walletSetPasswordContext) PrepareData(data *types.SuccessData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	err := w.SetPassword(c.password, c.save)
	if err != nil {
		return types.ResponseStatus_Error, err
	}
	return types.ResponseStatus_Success, err
}
