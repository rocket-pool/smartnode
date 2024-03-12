package wallet

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/beacon"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type walletCreateValidatorKeyContextFactory struct {
	handler *WalletHandler
}

func (f *walletCreateValidatorKeyContextFactory) Create(args url.Values) (*walletCreateValidatorKeyContext, error) {
	c := &walletCreateValidatorKeyContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("pubkey", args, input.ValidatePubkey, &c.pubkey),
		server.ValidateArg("start-index", args, input.ValidateUint, &c.index),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletCreateValidatorKeyContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletCreateValidatorKeyContext, api.SuccessData](
		router, "create-validator-key", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletCreateValidatorKeyContext struct {
	handler *WalletHandler
	pubkey  beacon.ValidatorPubkey
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
