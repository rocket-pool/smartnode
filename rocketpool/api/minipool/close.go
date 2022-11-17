package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	v110_network "github.com/rocket-pool/rocketpool-go/legacy/v1.1.0/network"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

func canCloseMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CanCloseMinipoolResponse, error) {

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
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanCloseMinipoolResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Validate minipool owner
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	if err := validateMinipoolOwner(mp, nodeAccount.Address); err != nil {
		return nil, err
	}

	isAtlasDeployed, err := rputils.IsAtlasDeployed(rp)
	if err != nil {
		return nil, fmt.Errorf("error checking if Atlas has been deployed: %w", err)
	}
	response.IsAtlasDeployed = isAtlasDeployed

	// Check minipool status
	status, err := mp.GetStatus(nil)
	if err != nil {
		return nil, err
	}
	response.InvalidStatus = (status != types.Dissolved)

	if !isAtlasDeployed {
		networkPricesAddress := cfg.Smartnode.GetV110NetworkPricesAddress()

		// Check consensus status
		inConsensus, err := v110_network.InConsensus(rp, nil, &networkPricesAddress)
		if err != nil {
			return nil, err
		}
		response.InConsensus = inConsensus
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := mp.EstimateCloseGas(opts)
	if err == nil {
		response.GasInfo = gasInfo
	}

	// Update & return response
	response.CanClose = !(response.InvalidStatus || !response.InConsensus)
	return &response, nil

}

func closeMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CloseMinipoolResponse, error) {

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
	response := api.CloseMinipoolResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
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

	// Close
	hash, err := mp.Close(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
