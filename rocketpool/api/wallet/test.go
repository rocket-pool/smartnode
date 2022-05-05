package wallet

import (
	"errors"
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func testMnemonic(c *cli.Context, mnemonic string) (*api.TestMnemonicResponse, error) {

	// Get services
	if err := services.RequireNodePassword(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.TestMnemonicResponse{}

	// Check if wallet is initialized
	if !w.IsInitialized() {
		return nil, errors.New("The wallet has not been initialized yet.")
	}

	// Get current account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting current account: %w", err)
	}
	response.CurrentAddress = nodeAccount.Address

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

	// Create a new wallet and recover from the given info
	recoveredWallet, err := wallet.NewWallet("", uint(w.GetChainID().Uint64()), nil, nil, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("error generating new wallet: %w", err)
	}
	err = recoveredWallet.TestRecovery(path, mnemonic)
	if err != nil {
		return nil, fmt.Errorf("error recovering wallet: %w", err)
	}

	// Get recovered account
	recoveredAccount, err := recoveredWallet.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting recovered account: %w", err)
	}
	response.RecoveredAddress = recoveredAccount.Address

	// Return response
	return &response, nil

}
