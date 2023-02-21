package minipool

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
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
	zero := big.NewInt(0)
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
				minipoolDetails := &details[mi]
				minipoolDetails.Address = address
				minipoolDetails.Balance = big.NewInt(0)
				minipoolDetails.Refund = big.NewInt(0)
				minipoolDetails.NodeShareOfBalance = big.NewInt(0)
				mp, err := minipool.NewMinipool(rp, address, nil)
				if err != nil {
					return fmt.Errorf("error creating binding for minipool %s: %w", address.Hex(), err)
				}
				minipoolDetails.MinipoolVersion = mp.GetVersion()

				// Ignore minipools that are too old
				if minipoolDetails.MinipoolVersion < 3 {
					minipoolDetails.CanDistribute = false
					return nil
				}

				var wg2 errgroup.Group
				wg2.Go(func() error {
					var err error
					minipoolDetails.Balance, err = rp.Client.BalanceAt(context.Background(), address, nil)
					if err != nil {
						return fmt.Errorf("error getting balance of minipool %s: %w", address.Hex(), err)
					}
					return nil
				})
				wg2.Go(func() error {
					var err error
					minipoolDetails.Refund, err = mp.GetNodeRefundBalance(nil)
					if err != nil {
						return fmt.Errorf("error getting refund balance of minipool %s: %w", address.Hex(), err)
					}
					return nil
				})
				wg2.Go(func() error {
					var err error
					minipoolDetails.Status, err = mp.GetStatus(nil)
					if err != nil {
						return fmt.Errorf("error getting status of minipool %s: %w", address.Hex(), err)
					}
					return nil
				})
				wg2.Go(func() error {
					var err error
					minipoolDetails.IsFinalized, err = mp.GetFinalised(nil)
					if err != nil {
						return fmt.Errorf("error getting finalized status of minipool %s: %w", address.Hex(), err)
					}
					return nil
				})

				// Wait for data
				if err := wg2.Wait(); err != nil {
					return err
				}

				// Ignore minipools with 0 balance
				if minipoolDetails.Balance.Cmp(zero) == 0 {
					minipoolDetails.CanDistribute = false
					return nil
				}

				// Ignore minipools with a balance lower than the refund
				if minipoolDetails.Balance.Cmp(minipoolDetails.Refund) == -1 {
					minipoolDetails.CanDistribute = false
					return nil
				}

				// Ignore minipools with an effective balance higher than v3 rewards-vs-exit cap
				effectiveBalance := big.NewInt(0).Sub(minipoolDetails.Balance, minipoolDetails.Refund)
				eight := eth.EthToWei(8)
				if effectiveBalance.Cmp(eight) >= 0 {
					minipoolDetails.CanDistribute = false
					return nil
				}

				// Can't distribute a minipool that's already finalized
				if minipoolDetails.IsFinalized {
					minipoolDetails.CanDistribute = false
					return nil
				}

				// Get the node share of the balance
				minipoolDetails.NodeShareOfBalance, err = mp.CalculateNodeShare(effectiveBalance, nil)
				if err != nil {
					return fmt.Errorf("error calculating node share for minipool %s: %w", address.Hex(), err)
				}

				minipoolDetails.CanDistribute = true
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
