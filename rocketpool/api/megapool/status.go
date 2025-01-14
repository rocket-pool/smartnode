package megapool

import (
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func getStatus(c *cli.Context) (*api.MegapoolStatusResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolStatusResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	details, err := GetNodeMegapoolDetails(rp, bc, nodeAccount.Address)
	if err != nil {
		return nil, err
	}

	response.Megapool = details

	return &response, nil
}
