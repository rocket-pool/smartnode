package megapool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
	"github.com/wealdtech/go-ens/v3"
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
	response.NodeAccount = nodeAccount.Address
	response.NodeAccountAddressFormatted = formatResolvedAddress(c, response.NodeAccount)

	details, err := GetNodeMegapoolDetails(rp, bc, nodeAccount.Address)
	if err != nil {
		return nil, err
	}
	details.MegapoolAddressFormatted = formatResolvedAddress(c, details.MegapoolAddress)

	response.Megapool = details

	return &response, nil
}

func formatResolvedAddress(c *cli.Context, address common.Address) string {
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return address.Hex()
	}

	name, err := ens.ReverseResolve(rp.Client, address)
	if err != nil {
		return address.Hex()
	}
	return fmt.Sprintf("%s (%s)", name, address.Hex())
}
