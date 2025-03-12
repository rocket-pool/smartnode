package wallet

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func masquerade(c *cli.Context, address common.Address) (*api.MasqueradeResponse, error) {

	// Get services
	if err := services.RequireNodePassword(c); err != nil {
		return nil, err
	}

	response := api.MasqueradeResponse{}

	return &response, nil

}
