package minipool

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canPromoteMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CanPromoteMinipoolResponse, error) {

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
	response := api.CanPromoteMinipoolResponse{
		CanPromote: false,
	}

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

	// Check the minipool's status
	status, err := mp.GetStatusDetails(nil)
	if err != nil {
		return nil, err
	}

	if status.IsVacant {

		// Get the scrub period
		scrubPeriodSeconds, err := trustednode.GetPromotionScrubPeriod(rp, nil)
		if err != nil {
			return nil, err
		}
		scrubPeriod := time.Duration(scrubPeriodSeconds) * time.Second

		// Get the time of the latest block
		latestEth1Block, err := rp.Client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			return nil, fmt.Errorf("Can't get the latest block time: %w", err)
		}
		latestBlockTime := time.Unix(int64(latestEth1Block.Time), 0)

		creationTime := status.StatusTime
		remainingTime := creationTime.Add(scrubPeriod).Sub(latestBlockTime)
		if remainingTime < 0 {
			response.CanPromote = true
		}
	}

	if response.CanPromote {

		// Get the updated minipool interface
		mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
		if err != nil {
			return nil, err
		}
		mpv3, success := minipool.GetMinipoolAsV3(mp)
		if !success {
			return nil, fmt.Errorf("cannot check if minipool %s can be promoted because its delegate version is too low (v%d); please update the delegate to promote it", mp.GetAddress().Hex(), mp.GetVersion())
		}

		// Get transactor
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return nil, err
		}

		// Get the gas limit
		gasInfo, err := mpv3.EstimatePromoteGas(opts)
		if err != nil {
			return nil, fmt.Errorf("Could not estimate the gas required to promote the minipool: %w", err)
		}
		response.GasInfo = gasInfo

	}

	// Return response
	return &response, nil

}

func promoteMinipool(c *cli.Context, minipoolAddress common.Address) (*api.StakeMinipoolResponse, error) {

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
	response := api.StakeMinipoolResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return nil, fmt.Errorf("cannot promte minipool %s because its delegate version is too low (v%d); please update the delegate to promote it", mp.GetAddress().Hex(), mp.GetVersion())
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

	// Promote
	hash, err := mpv3.Promote(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
