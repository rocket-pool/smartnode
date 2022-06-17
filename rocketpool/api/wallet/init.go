package wallet

import (
	"errors"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func initWallet(c *cli.Context) (*api.InitWalletResponse, error) {

	// Get services
	if err := services.RequireNodePassword(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.InitWalletResponse{}

	// Check if wallet is already initialized
	if w.IsInitialized() {
		return nil, errors.New("The wallet is already initialized")
	}

	// Get the derivation path
	path := c.String("derivation-path")
	switch path {
	case "":
		path = wallet.DefaultNodeKeyPath
	case "ledgerLive":
		path = wallet.LedgerLiveNodeKeyPath
	case "mew":
		path = wallet.MyEtherWalletNodeKeyPath
	}

	// Initialize wallet
	mnemonic, err := w.Initialize(path, 0)
	if err != nil {
		return nil, err
	}
	response.Mnemonic = mnemonic

	// Save wallet
	if err := w.Save(); err != nil {
		return nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address

	// Return response
	return &response, nil

}
