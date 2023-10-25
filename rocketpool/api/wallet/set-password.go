package wallet

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type walletSetPasswordContextFactory struct {
	handler *WalletHandler
}

func (f *walletSetPasswordContextFactory) Create(vars map[string]string) (*walletSetPasswordContext, error) {
	c := &walletSetPasswordContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("password", vars, input.ValidateNodePassword, &c.password),
		server.ValidateArg("save", vars, input.ValidateBool, &c.save),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletSetPasswordContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessRoute[*walletSetPasswordContext, api.SuccessData](
		router, "set-password", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletSetPasswordContext struct {
	handler  *WalletHandler
	password []byte
	save     bool
}

func (c *walletSetPasswordContext) PrepareData(data *api.SuccessData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	_, hasPassword := w.GetPassword()
	if hasPassword {
		return fmt.Errorf("wallet password has already been set")
	}
	w.RememberPassword(c.password)
	if c.save {
		err := w.SavePassword()
		if err != nil {
			return fmt.Errorf("error saving wallet password to disk: %w", err)
		}
	}

	data.Success = true
	return nil
}
