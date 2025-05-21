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

	// Check if the megapool is already deployed
	response.MegapoolDeployed, err = megapool.GetMegapoolDeployed(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	if !response.MegapoolDeployed {
		response.CanDistribute = false
		return &response, nil
	}

	response.MegapoolAddress, err = megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, response.MegapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// LastDistributionBlock is 0 if the megapool has never had a staking validator
	response.LastDistributionBlock, err = mp.GetLastDistributionBlock(nil)
	if err != nil {
		return nil, err
	}

	if response.LastDistributionBlock == 0 {
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
