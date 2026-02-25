package wallet

import (
	"errors"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func initWallet(c *cli.Context) (*api.InitWalletResponse, error) {
	return initWalletWithPath(c, c.String("derivation-path"))
}

func initWalletWithPath(c *cli.Context, derivationPath string) (*api.InitWalletResponse, error) {

	// Get services
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
	path := derivationPath
	switch path {
	case "":
		path = wallet.DefaultNodeKeyPath
	case "ledgerLive":
		path = wallet.LedgerLiveNodeKeyPath
	case "mew":
		path = wallet.MyEtherWalletNodeKeyPath
	}

	// Initialize wallet but don't save it
	mnemonic, err := w.Initialize(path, 0)
	if err != nil {
		return nil, err
	}
	response.Mnemonic = mnemonic

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address

	// Return response
	return &response, nil

}
