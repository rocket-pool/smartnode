package wallet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func masquerade(c *cli.Context, address common.Address) (*api.MasqueradeResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	err = w.MasqueradeAsAddress(address)
	if err != nil {
		return nil, fmt.Errorf("Error masquerading as address %s", address.Hex())
	}

	response := api.MasqueradeResponse{}

	return &response, nil

}
