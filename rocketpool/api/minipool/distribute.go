package minipool

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func getDistributeBalanceDetails(c *cli.Context) (*api.GetDistributeBalanceDetailsResponse, error) {

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
	response := api.GetDistributeBalanceDetailsResponse{}

	isAtlasDeployed, err := state.IsAtlasDeployed(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking if Atlas has been deployed: %w", err)
	}
	response.IsAtlasDeployed = isAtlasDeployed

	// Prevent distribution prior to Atlas
	if !isAtlasDeployed {
		return &response, nil
	}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	addresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Load details in batches
	details := make([]api.MinipoolBalanceDistributionDetails, len(addresses))
	for bsi := 0; bsi < len(addresses); bsi += MinipoolDetailsBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolDetailsBatchSize
		if mei > len(addresses) {
			mei = len(addresses)
		}

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address := addresses[mi]
				minipoolDetails := api.MinipoolBalanceDistributionDetails{
					Address: address,
				}
				mp, err := minipool.NewMinipool(rp, address, nil)
				if err != nil {
					return fmt.Errorf("error creating binding for minipool %s: %w", address.Hex(), err)
				}

				var wg2 errgroup.Group
				wg2.Go(func() error {
					var err error
					minipoolDetails.Balance, err = rp.Client.BalanceAt(context.Background(), address, nil)
					if err != nil {
						return fmt.Errorf("error getting balance of minipool %s: %w", address.Hex(), err)
					}
					minipoolDetails.NodeShareOfBalance, err = mp.CalculateNodeShare(minipoolDetails.Balance, nil)
					return err
				})
				wg2.Go(func() error {
					status, err := mp.GetStatus(nil)
					minipoolDetails.InvalidStatus = (status != types.Staking)
					return err
				})
				wg2.Go(func() error {
					version := mp.GetVersion()
					minipoolDetails.VersionTooLow = (version < 3)
					return nil
				})

				// Wait for data
				if err := wg2.Wait(); err != nil {
					return err
				}

				details[mi] = minipoolDetails
				return nil
			})
		}
		if err := wg.Wait(); err != nil {
			return nil, err
		}

	}

	// Update & return response
	response.Details = details
	return &response, nil

}

func canDistributeBalance(c *cli.Context, minipoolAddress common.Address) (*api.CanDistributeBalanceResponse, error) {

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
	response := api.CanDistributeBalanceResponse{}

	isAtlasDeployed, err := state.IsAtlasDeployed(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking if Atlas has been deployed: %w", err)
	}
	response.IsAtlasDeployed = isAtlasDeployed

	// Prevent distribution prior to Atlas
	if !isAtlasDeployed {
		return &response, nil
	}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Check minipool status
	status, err := mp.GetStatus(nil)
	if err != nil {
		return nil, err
	}
	response.MinipoolStatus = status

	// Get minipool delegate version
	version := mp.GetVersion()
	response.MinipoolVersion = version

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := mp.EstimateDistributeBalanceGas(opts)
	if err == nil {
		response.GasInfo = gasInfo
	}

	// Update & return response
	response.CanDistribute = !(status == types.Dissolved || !isAtlasDeployed || version < 3)
	return &response, nil

}

func estimateDistributeBalanceGas(c *cli.Context, minipoolAddress common.Address) (*api.EstimateDistributeBalanceGasResponse, error) {

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
	response := api.EstimateDistributeBalanceGasResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := mp.EstimateDistributeBalanceGas(opts)
	if err == nil {
		response.GasInfo = gasInfo
	}

	// Return response
	return &response, nil
}

func distributeBalance(c *cli.Context, minipoolAddress common.Address) (*api.CloseMinipoolResponse, error) {

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

	// Distribute the minipool's balance
	hash, err := mp.DistributeBalance(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
