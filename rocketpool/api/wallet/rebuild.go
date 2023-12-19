package wallet

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type walletRebuildContextFactory struct {
	handler *WalletHandler
}

func (f *walletRebuildContextFactory) Create(args url.Values) (*walletRebuildContext, error) {
	c := &walletRebuildContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *walletRebuildContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletRebuildContext, api.WalletRebuildData](
		router, "rebuild", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletRebuildContext struct {
	handler *WalletHandler
}

func (c *walletRebuildContext) PrepareData(data *api.WalletRebuildData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	rp := sp.GetRocketPool()
	w := sp.GetWallet()

	// Requirements
	err := errors.Join(
		sp.RequireWalletReady(),
		sp.RequireEthClientSynced(),
	)
	if err != nil {
		return err
	}

	// Recover validator keys
	data.ValidatorKeys, err = wallet.RecoverMinipoolKeys(cfg, rp, w, false)
	if err != nil {
		return fmt.Errorf("error recovering minipool keys: %w", err)
	}
	return nil
}
