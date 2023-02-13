package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func canBeginReduceBondAmount(c *cli.Context, minipoolAddress common.Address, newBondAmountWei *big.Int) (*api.CanBeginReduceBondAmountResponse, error) {
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
	response := api.CanBeginReduceBondAmountResponse{}

	// Check if bond reduction is enabled
	bondReductionEnabled, err := protocol.GetBondReductionEnabled(rp, nil)
	if err != nil {
		return nil, err
	}
	response.BondReductionDisabled = !bondReductionEnabled

	// Check the minipool version
	version, err := rocketpool.GetContractVersion(rp, minipoolAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool %s contract version: %w", minipoolAddress.Hex(), err)
	}
	response.MinipoolVersionTooLow = (version < 3)

	response.CanReduce = !(response.BondReductionDisabled || response.MinipoolVersionTooLow)

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := minipool.EstimateBeginReduceBondAmountGas(rp, minipoolAddress, newBondAmountWei, opts)
	if err == nil {
		response.GasInfo = gasInfo
	}

	// Update & return response
	return &response, nil
}

func beginReduceBondAmount(c *cli.Context, minipoolAddress common.Address, newBondAmountWei *big.Int) (*api.BeginReduceBondAmountResponse, error) {
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
	response := api.BeginReduceBondAmountResponse{}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Start bond reduction
	hash, err := minipool.BeginReduceBondAmount(rp, minipoolAddress, newBondAmountWei, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil
}

func canReduceBondAmount(c *cli.Context, minipoolAddress common.Address) (*api.CanReduceBondAmountResponse, error) {
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
	response := api.CanReduceBondAmountResponse{}

	// Make the minipool binding
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool binding for %s: %w", minipoolAddress.Hex(), err)
	}
	response.MinipoolVersion = mp.GetVersion()
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if success {
		// Get gas estimate
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return nil, err
		}
		gasInfo, err := mpv3.EstimateReduceBondAmountGas(opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
	}

	response.CanReduce = success

	// Update & return response
	return &response, nil
}

func reduceBondAmount(c *cli.Context, minipoolAddress common.Address) (*api.ReduceBondAmountResponse, error) {
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
	response := api.ReduceBondAmountResponse{}

	// Make the minipool binding
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool binding for %s: %w", minipoolAddress.Hex(), err)
	}
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return nil, fmt.Errorf("bond reduction is not supported for minipool version %d; please upgrade the delegate for minipool %s to reduce its bond", mp.GetVersion(), minipoolAddress.Hex())
	}

	// Get the node transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Start bond reduction
	hash, err := mpv3.ReduceBondAmount(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil
}
