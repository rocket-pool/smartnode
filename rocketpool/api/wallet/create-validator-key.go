package wallet

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type walletCreateValidatorKeyContextFactory struct {
	handler *WalletHandler
}

func (f *walletCreateValidatorKeyContextFactory) Create(vars map[string]string) (*walletCreateValidatorKeyContext, error) {
	c := &walletCreateValidatorKeyContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("pubkey", vars, input.ValidatePubkey, &c.pubkey),
		server.ValidateArg("start-index", vars, input.ValidateUint, &c.index),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletCreateValidatorKeyContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessRoute[*walletCreateValidatorKeyContext, api.SuccessData](
		router, "create-validator-key", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletCreateValidatorKeyContext struct {
	handler *WalletHandler
	pubkey  types.ValidatorPubkey
	index   uint64
}

func (c *walletCreateValidatorKeyContext) PrepareData(data *api.SuccessData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	// Requirements
	err := sp.RequireWalletReady()
	if err != nil {
		return err
	}

	_, err = w.RecoverValidatorKey(c.pubkey, uint(c.index))
	if err != nil {
		return fmt.Errorf("error creating validator key: %w", err)
	}

	data.Success = true
	return nil
}
