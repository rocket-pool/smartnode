package wallet

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type walletExportEthKeyContextFactory struct {
	handler *WalletHandler
}

func (f *walletExportEthKeyContextFactory) Create(args url.Values) (*walletExportEthKeyContext, error) {
	c := &walletExportEthKeyContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *walletExportEthKeyContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletExportEthKeyContext, api.WalletExportEthKeyData](
		router, "export-eth-key", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletExportEthKeyContext struct {
	handler *WalletHandler
}

func (c *walletExportEthKeyContext) PrepareData(data *api.WalletExportEthKeyData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	// Requirements
	err := sp.RequireWalletReady()
	if err != nil {
		return err
	}

	// Make a new password
	password, err := utils.GenerateRandomPassword()
	if err != nil {
		return fmt.Errorf("error generating random password: %w", err)
	}

	ethkey, err := w.GetEthKeystore(password)
	if err != nil {
		return fmt.Errorf("error getting eth-style keystore: %w", err)
	}
	data.EthKeyJson = ethkey
	data.Password = password
	return nil
}
