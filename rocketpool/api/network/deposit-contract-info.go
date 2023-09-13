package network

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

func getDepositContractInfo(c *cli.Context) (*api.DepositContractInfoResponse, error) {
	// Get services
	if err := services.RequireEthClientSynced(c); err != nil { // Needs to be synced so RP has been deployed
		response := api.DepositContractInfoResponse{}
		response.SufficientSync = false
		return &response, nil
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, fmt.Errorf("error getting Rocket Pool binding: %w", err)
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, fmt.Errorf("error getting configuration: %w", err)
	}
	bc, err := services.GetBeaconClient(c) // Doesn't need to be synced because the deposit contract is in the genesis info
	if err != nil {
		return nil, fmt.Errorf("error getting beacon client: %w", err)
	}

	// Response
	response := api.DepositContractInfoResponse{}
	response.SufficientSync = true

	// Get the deposit contract info
	info, err := rputils.GetDepositContractInfoImpl(rp, cfg, bc)
	if err != nil {
		return nil, fmt.Errorf("error getting deposit contract info: %w", err)
	}
	response.RPNetwork = info.RPNetwork
	response.RPDepositContract = info.RPDepositContract
	response.BeaconNetwork = info.BeaconNetwork
	response.BeaconDepositContract = info.BeaconDepositContract
	return &response, nil
}
