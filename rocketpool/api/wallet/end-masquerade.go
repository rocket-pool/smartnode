package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func endMasquerade(c *cli.Context) (*api.EndMasqueradeResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	err = w.EndMasquerade()
	if err != nil {
		return nil, fmt.Errorf("Error ending masquerade mode")
	}

	response := api.EndMasqueradeResponse{}

	return &response, nil
}
