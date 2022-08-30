package node

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func GetDepositContractInfo(c *cli.Context) (*api.DepositContractInfoResponse, error) {

	// Get services
	if err := services.RequireRocketStorage(c); err != nil {
		response := api.DepositContractInfoResponse{}
		response.SufficientSync = false
		return &response, nil
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.DepositContractInfoResponse{}
	response.SufficientSync = true

	// Get the ETH1 network ID that Rocket Pool is on
	config, err := services.GetConfig(c)
	if err != nil {
		return nil, fmt.Errorf("Error getting configuration: %w", err)
	}
	response.RPNetwork = uint64(config.Smartnode.GetChainID())

	// Get the deposit contract address Rocket Pool will deposit to
	rpDepositContract, err := rp.GetContract("casperDeposit")
	if err != nil {
		return nil, fmt.Errorf("Error getting Casper deposit contract: %w", err)
	}
	if rpDepositContract == nil {
		return nil, fmt.Errorf("Deposit contract was undefined.")
	}
	response.RPDepositContract = *rpDepositContract.Address

	// Get the Beacon Client info
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, fmt.Errorf("Error getting beacon client: %w", err)
	}
	eth2DepositContract, err := bc.GetEth2DepositContract()
	if err != nil {
		return nil, fmt.Errorf("Error getting beacon client deposit contract: %w", err)
	}

	response.BeaconNetwork = eth2DepositContract.ChainID
	response.BeaconDepositContract = eth2DepositContract.Address

	// Return response
	return &response, nil

}
