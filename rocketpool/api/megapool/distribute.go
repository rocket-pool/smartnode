package megapool

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func canDistributeMegapool(c *cli.Context) (*api.CanDistributeMegapoolResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	response := api.CanDistributeMegapoolResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Load the megapool details
	details, err := GetNodeMegapoolDetails(rp, bc, nodeAccount.Address)
	if err != nil {
		return nil, err
	}

	if !details.Deployed {
		response.CanDistribute = false
		response.MegapoolNotDeployed = true
		return &response, nil
	}

	response.MegapoolAddress = details.Address
	response.Details = details

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, response.MegapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	response.LastDistributionBlock = details.LastDistributionBlock

	if response.LastDistributionBlock == 0 {
		response.CanDistribute = false
		return &response, nil
	}

	if details.LockedValidatorCount > 0 {
		response.CanDistribute = false
		response.LockedValidatorCount = details.LockedValidatorCount
		return &response, nil
	}

	if details.ExitingValidatorCount > 0 {
		response.ExitingValidatorCount = details.ExitingValidatorCount
		response.CanDistribute = false
		return &response, nil
	}

	gasInfo, err := mp.EstimateDistributeGas(opts)
	if err != nil {
		return nil, err
	}
	// Return response
	response.CanDistribute = true
	response.GasInfo = gasInfo
	return &response, nil
}

func distributeMegapool(c *cli.Context) (*api.DistributeMegapoolResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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

	// Get the node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Response
	response := api.DistributeMegapoolResponse{}

	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Distribute
	hash, err := mp.Distribute(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil
}
