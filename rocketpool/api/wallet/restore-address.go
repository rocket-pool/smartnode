package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func restoreAddress(c *cli.Context) (*api.RestoreAddressResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	err = w.RestoreAddressToWallet()
	if err != nil {
		return nil, fmt.Errorf("Error restoring address")
	}

	response := api.RestoreAddressResponse{}

	return &response, nil
}
