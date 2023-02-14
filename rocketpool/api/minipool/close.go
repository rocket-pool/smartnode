package minipool

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

func getMinipoolCloseDetailsForNode(c *cli.Context) (*api.GetMinipoolCloseDetailsForNodeResponse, error) {

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
	response := api.GetMinipoolCloseDetailsForNodeResponse{}

	// Check if Atlas has been deployed
	isAtlasDeployed, err := rputils.IsAtlasDeployed(rp)
	if err != nil {
		return nil, fmt.Errorf("error checking if Atlas has been deployed: %w", err)
	}
	response.IsAtlasDeployed = isAtlasDeployed
	if !isAtlasDeployed {
		return &response, nil
	}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get the minipool addresses for this node
	addresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Get the transaction opts
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Iterate over each minipool to get its close details
	details := make([]api.MinipoolCloseDetails, len(addresses))
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
				mpDetails, err := getMinipoolCloseDetails(rp, address, nodeAccount.Address, opts)
				if err == nil {
					details[mi] = mpDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return nil, err
		}

	}

	// Aggregate the closeable ones
	closeableMinipools := []api.MinipoolCloseDetails{}
	for _, mp := range details {
		if mp.CanClose {
			closeableMinipools = append(closeableMinipools, mp)
		}
	}
	response.Details = closeableMinipools

	return &response, nil

}

func getMinipoolCloseDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address, nodeAddress common.Address, opts *bind.TransactOpts) (api.MinipoolCloseDetails, error) {

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return api.MinipoolCloseDetails{}, err
	}

	// Validate minipool owner
	if err := validateMinipoolOwner(mp, nodeAddress); err != nil {
		return api.MinipoolCloseDetails{}, err
	}

	var details api.MinipoolCloseDetails
	details.Address = mp.GetAddress()

	// Get the balance / share info and status details
	var wg errgroup.Group
	balance, err := rp.Client.BalanceAt(context.Background(), mp.GetAddress(), nil)
	if err != nil {
		return api.MinipoolCloseDetails{}, fmt.Errorf("error getting balance of minipool %s: %w", mp.GetAddress().Hex(), err)
	}
	wg.Go(func() error {
		var err error
		details.NodeShare, err = mp.CalculateNodeShare(balance, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.UserShare, err = mp.CalculateUserShare(balance, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.IsFinalized, err = mp.GetFinalised(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.MinipoolStatus, err = mp.GetStatus(nil)
		return err
	})

	if err := wg.Wait(); err != nil {
		return api.MinipoolCloseDetails{}, err
	}

	// Can't close a minipool that's already finalized
	if details.IsFinalized {
		details.CanClose = false
		return details, nil
	}

	// Make sure it's in a closeable state
	switch details.MinipoolStatus {
	case types.Staking, types.Withdrawable, types.Dissolved:
		details.CanClose = true
	case types.Initialized, types.Prelaunch:
		details.CanClose = false
		return details, nil
	}

	// If it's dissolved, just close it
	if details.MinipoolStatus == types.Dissolved {
		// Get gas estimate
		gasInfo, err := mp.EstimateCloseGas(opts)
		if err != nil {
			return api.MinipoolCloseDetails{}, err
		}
		details.GasInfo = gasInfo
	} else {
		// Check if it's an upgraded Atlas-era minipool
		mpv3, success := minipool.GetMinipoolAsV3(mp)
		if success {
			// It is, so check if it's already been distributed
			distributed, err := mpv3.GetUserDistributed(nil)
			if err != nil {
				return api.MinipoolCloseDetails{}, err
			}
			if distributed {
				// It's already been distributed so just finalize it
				gasInfo, err := mpv3.EstimateFinaliseGas(opts)
				if err != nil {
					return api.MinipoolCloseDetails{}, err
				}
				details.GasInfo = gasInfo
			} else {
				// Do a distribution, which will finalize it
				gasInfo, err := mpv3.EstimateDistributeBalanceGas(opts)
				if err != nil {
					return api.MinipoolCloseDetails{}, err
				}
				details.GasInfo = gasInfo
			}

		} else {
			// Check if it's a vanilla / Redstone-era minipool
			mpv2, success := minipool.GetMinipoolAsV2(mp)
			if !success {
				return api.MinipoolCloseDetails{}, fmt.Errorf("minipool version %d doesn't have a proper close binding", mp.GetVersion())
			}
			// Distribute and finalize
			gasInfo, err := mpv2.EstimateDistributeBalanceAndFinaliseGas(opts)
			if err != nil {
				return api.MinipoolCloseDetails{}, err
			}
			details.GasInfo = gasInfo
		}
	}

	return details, nil

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

	// Check if Atlas has been deployed
	isAtlasDeployed, err := rputils.IsAtlasDeployed(rp)
	if err != nil {
		return nil, fmt.Errorf("error checking if Atlas has been deployed: %w", err)
	}
	if !isAtlasDeployed {
		return nil, fmt.Errorf("Atlas has not been deployed yet.")
	}

	status, err := mp.GetStatus(nil)
	if err != nil {
		return nil, err
	}

	// If it's dissolved, just close it
	if status == types.Dissolved {
		hash, err := mp.Close(opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = hash
	} else {
		// Check if it's an upgraded Atlas-era minipool
		mpv3, success := minipool.GetMinipoolAsV3(mp)
		if success {
			// It is, so check if it's already been distributed
			distributed, err := mpv3.GetUserDistributed(nil)
			if err != nil {
				return nil, err
			}
			if distributed {
				// It's already been distributed so just finalize it
				hash, err := mpv3.Finalise(opts)
				if err != nil {
					return nil, err
				}
				response.TxHash = hash
			} else {
				// Do a distribution, which will finalize it
				hash, err := mpv3.DistributeBalance(opts)
				if err != nil {
					return nil, err
				}
				response.TxHash = hash
			}

		} else {
			// Check if it's a vanilla / Redstone-era minipool
			mpv2, success := minipool.GetMinipoolAsV2(mp)
			if !success {
				return nil, fmt.Errorf("minipool version %d doesn't have a proper close binding", mp.GetVersion())
			}
			// Distribute and finalize
			hash, err := mpv2.DistributeBalanceAndFinalise(opts)
			if err != nil {
				return nil, err
			}
			response.TxHash = hash
		}
	}

	// Return response
	return &response, nil

}
