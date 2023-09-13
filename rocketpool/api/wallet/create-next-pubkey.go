package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func createNextPubkey(c *cli.Context) (*api.ApiResponse, error) {
	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, fmt.Errorf("error getting wallet: %w", err)
	}

	// Response
	response := api.ApiResponse{}

	// Create and save a new validator key
	_, err = w.CreateValidatorKey()
	if err != nil {
		return nil, fmt.Errorf("error creating next validator key: %w", err)
	}

	// Save wallet
	if err := w.Save(); err != nil {
		return nil, fmt.Errorf("error saving wallet: %w", err)
	}

	return &response, nil
}
