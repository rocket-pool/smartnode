package wallet

import (
	"errors"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func removeWallet(c *cli.Context) (*api.RemoveWalletResponse, error) {

	// Get services
	if err := services.RequireNodePassword(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Check if wallet is already initialized
	if !w.IsInitialized() {
		return nil, errors.New("The wallet is not initialized")
	}

	if err = w.Remove(); err != nil {
		return nil, err
	}

	// Return response
	return &api.RemoveWalletResponse{}, nil

}
