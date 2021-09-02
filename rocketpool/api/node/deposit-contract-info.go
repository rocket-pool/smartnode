package node

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)


func getDepositContractInfo(c *cli.Context) (*api.DepositContractInfoResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { 
        response := api.DepositContractInfoResponse{}
        response.SufficientSync = false
        return &response, nil 
    }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.DepositContractInfoResponse{}
    response.SufficientSync = true

    // Get the ETH1 network ID that Rocket Pool is on
    config, err := services.GetConfig(c)
    if err != nil {
        return nil, fmt.Errorf("Error getting configuration: %w", err)
    }
    rpNetwork, err := strconv.ParseUint(config.Chains.Eth1.ChainID, 0, 64)
    if err != nil {
        return nil, fmt.Errorf("%s is not a valid ETH1 chain ID (in the config file): %w",
            config.Chains.Eth1.ChainID, err)
    }
    response.RPNetwork = rpNetwork

    // Get the deposit contract address Rocket Pool will deposit to
    rpDepositContract, err := rp.GetContract("casperDeposit")
    if err != nil {
        return nil, fmt.Errorf("Error getting Casper deposit contract: %w", err)
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

