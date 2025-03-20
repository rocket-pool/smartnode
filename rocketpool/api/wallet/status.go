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

	// Get wallet status
	response.PasswordSet = pm.IsPasswordSet()
	response.WalletInitialized = w.IsInitialized()

	// Get accounts if initialized
	if response.WalletInitialized {

		// Get node account
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return nil, err
		}
		//TODO refactor AccountAddress to WalletAddress
		response.AccountAddress = nodeAccount.Address
		response.NodeAddress, response.HasAddress = w.GetAddress()

	}

	// Return response
	return &response, nil

}
