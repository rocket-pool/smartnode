package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanBeginReduceBondAmountResponse{}

	// Data
	var wg errgroup.Group
	var nodeDepositAmount *big.Int

	// Check if bond reduction is enabled
	wg.Go(func() error {
		bondReductionEnabled, err := protocol.GetBondReductionEnabled(rp, nil)
		if err != nil {
			return fmt.Errorf("error checking if bond reduction is enabled: %w", err)
		}
		response.BondReductionDisabled = !bondReductionEnabled
		return nil
	})

	// Check the minipool version
	wg.Go(func() error {
		version, err := rocketpool.GetContractVersion(rp, minipoolAddress, nil)
		if err != nil {
			return fmt.Errorf("error getting minipool %s contract version: %w", minipoolAddress.Hex(), err)
		}
		response.MinipoolVersionTooLow = (version < 3)
		return nil
	})

	// Check the balance and status on Beacon
	wg.Go(func() error {
		var err error
		pubkey, err := minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
		if err != nil {
			return fmt.Errorf("error retrieving pubkey for minipool %s: %w", minipoolAddress.Hex(), err)
		}
		status, err := bc.GetValidatorStatus(pubkey, nil)
		if err != nil {
			return fmt.Errorf("error getting validator status for minipool %s (pubkey %s): %w", minipoolAddress.Hex(), pubkey.Hex(), err)
		}
		response.Balance = status.Balance
		response.BeaconState = status.Status
		return nil
	})

	// Get match request info
	wg.Go(func() error {
		mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
		if err != nil {
			return fmt.Errorf("error creating binding for minipool %s: %w", minipoolAddress.Hex(), err)
		}
		nodeDepositAmount, err = mp.GetNodeDepositBalance(nil)
		if err != nil {
			return fmt.Errorf("error getting node deposit balance for minipool %s: %w", minipoolAddress.Hex(), err)
		}
		// How much more ETH they're requesting from the staking pool
		response.BorrowRequest = big.NewInt(0).Sub(nodeDepositAmount, newBondAmountWei)
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Check the beacon state
	response.InvalidBeaconState = !(response.BeaconState == beacon.ValidatorState_PendingInitialized ||
		response.BeaconState == beacon.ValidatorState_PendingQueued ||
		response.BeaconState == beacon.ValidatorState_ActiveOngoing)

	// Make sure the balance is high enough
	threshold := uint64(32000000000)
	response.BalanceTooLow = response.Balance < threshold

	response.CanReduce = !(response.BondReductionDisabled || response.MinipoolVersionTooLow || response.BalanceTooLow || response.InvalidBeaconState)

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
