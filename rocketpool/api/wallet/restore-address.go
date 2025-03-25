package wallet

import (
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func restoreAddress(c *cli.Context) (*api.RestoreAddressResponse, error) {

	// Get services
	if err := services.RequireNodePassword(c); err != nil {
		return nil, err
	}

	// TODO leaving this out for now
	// w, err := services.GetWallet(c)
	// if err != nil {
	// 	return nil, err
	// }

	// err = w.RestoreAddressToWallet()
	// if err != nil {
	// 	return nil, fmt.Errorf("Error restoring address")
	// }

	response := api.RestoreAddressResponse{}

	return &response, nil
}
