package wallet

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getStatus(c *cli.Context) (*api.WalletStatusResponse, error) {

	// Get services
	pm, err := services.GetPasswordManager(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.WalletStatusResponse{}

	// Get wallet type
	response.IsMasquerading = w.IsNodeMasquerading()

	// Get wallet status
	if response.IsMasquerading {
		response.PasswordSet = true
		response.WalletInitialized = true
	} else {
		response.PasswordSet = pm.IsPasswordSet()
		response.WalletInitialized = w.IsInitialized()
	}

	// Get accounts if initialized
	if response.WalletInitialized {
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return nil, err
		}
		nodeAddress, err := w.GetAddress()
		if err != nil {
			return nil, err
		}
		response.AccountAddress = nodeAccount.Address
		response.NodeAddress = nodeAddress
	}

	// Return response
	return &response, nil

}
