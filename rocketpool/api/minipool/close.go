package minipool

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetMinipoolCloseDetailsForNodeResponse{}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check if Saturn is already deployed
	response.IsSaturnDeployed, err = state.IsSaturnDeployed(rp, nil)
	if err != nil {
		return nil, err
	}

	// Check if express tickets have been provisioned for the node
	if response.IsSaturnDeployed {
		response.ExpressTicketsProvisioned, err = node.GetExpressTicketsProvisioned(rp, nodeAccount.Address, nil)
		if err != nil {
			return nil, fmt.Errorf("error checking if the node's express tickets are provisioned: %w", err)
		}
	}

	// Check the fee distributor
	response.IsFeeDistributorInitialized, err = node.GetFeeDistributorInitialized(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking if the node's fee distributor is initialized: %w", err)
	}
	if !response.IsFeeDistributorInitialized {
		return &response, nil
	}

	// Get the minipool addresses for this node
	addresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
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

	// Get the beacon statuses for each closeable minipool
	pubkeys := []types.ValidatorPubkey{}
	pubkeyMap := map[common.Address]types.ValidatorPubkey{}
	for _, mp := range details {
		if mp.MinipoolStatus == types.Dissolved {
			// Ignore dissolved minipools
			continue
		}
		pubkey, err := minipool.GetMinipoolPubkey(rp, mp.Address, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting pubkey for minipool %s: %w", mp.Address.Hex(), err)
		}
		pubkeyMap[mp.Address] = pubkey
		pubkeys = append(pubkeys, pubkey)
	}
	statusMap, err := bc.GetValidatorStatuses(pubkeys, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting beacon status of minipools: %w", err)
	}

	// Review closeability based on validator status
	for i, mp := range details {
		pubkey := pubkeyMap[mp.Address]
		validator := statusMap[pubkey]
		if mp.MinipoolStatus != types.Dissolved {
			details[i].BeaconState = validator.Status
			if validator.Status != beacon.ValidatorState_WithdrawalDone {
				details[i].CanClose = false
			}
		}
	}

	response.Details = details
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
	details.MinipoolVersion = mp.GetVersion()
	details.Balance = big.NewInt(0)
	details.Refund = big.NewInt(0)
	details.NodeShare = big.NewInt(0)

	// Ignore minipools that are too old
	if details.MinipoolVersion < 3 {
		details.CanClose = false
		return details, nil
	}

	// Get the balance / share info and status details
	var wg1 errgroup.Group
	wg1.Go(func() error {
		var err error
		details.Balance, err = rp.Client.BalanceAt(context.Background(), minipoolAddress, nil)
		if err != nil {
			return fmt.Errorf("error getting finalized status of minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})
	wg1.Go(func() error {
		var err error
		details.Refund, err = mp.GetNodeRefundBalance(nil)
		if err != nil {
			return fmt.Errorf("error getting refund balance of minipool %s: %w", mp.GetAddress().Hex(), err)
		}
		return nil
	})
	wg1.Go(func() error {
		var err error
		details.IsFinalized, err = mp.GetFinalised(nil)
		if err != nil {
			return fmt.Errorf("error getting finalized status of minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})
	wg1.Go(func() error {
		var err error
		details.MinipoolStatus, err = mp.GetStatus(nil)
		if err != nil {
			return fmt.Errorf("error getting status of minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})
	wg1.Go(func() error {
		var err error
		details.UserDepositBalance, err = mp.GetUserDepositBalance(nil)
		if err != nil {
			return fmt.Errorf("error getting user deposit balance of minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})

	if err := wg1.Wait(); err != nil {
		return api.MinipoolCloseDetails{}, err
	}

	// Can't close a minipool that's already finalized
	if details.IsFinalized {
		details.CanClose = false
		return details, nil
	}

	// Make sure it's in a closeable state
	effectiveBalance := big.NewInt(0).Sub(details.Balance, details.Refund)
	switch details.MinipoolStatus {
	case types.Dissolved:
		details.CanClose = true

	case types.Staking, types.Withdrawable:
		// Ignore minipools with a balance lower than the refund
		if details.Balance.Cmp(details.Refund) == -1 {
			details.CanClose = false
			return details, nil
		}

		// Ignore minipools with an effective balance lower than v3 rewards-vs-exit cap
		eight := eth.EthToWei(8)
		if effectiveBalance.Cmp(eight) == -1 {
			details.CanClose = false
			return details, nil
		}

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
			return api.MinipoolCloseDetails{}, fmt.Errorf("error estimating close gas for MP %s: %w", minipoolAddress.Hex(), err)
		}
		details.GasInfo = gasInfo
	} else {
		// Check if it's an upgraded Atlas-era minipool
		mpv3, success := minipool.GetMinipoolAsV3(mp)
		if success {
			// Create another wait group
			var wg2 errgroup.Group
			wg2.Go(func() error {
				var err error
				details.NodeShare, err = mp.CalculateNodeShare(effectiveBalance, nil)
				if err != nil {
					return fmt.Errorf("error getting node share of minipool %s: %w", mp.GetAddress().Hex(), err)
				}
				return nil
			})
			wg2.Go(func() error {
				var err error
				details.Distributed, err = mpv3.GetUserDistributed(nil)
				if err != nil {
					return fmt.Errorf("error checking if user distributed minipool %s: %w", mp.GetAddress().Hex(), err)
				}
				return nil
			})

			if err := wg2.Wait(); err != nil {
				return api.MinipoolCloseDetails{}, err
			}

			if details.Distributed {
				// It's already been distributed so just finalize it
				gasInfo, err := mpv3.EstimateFinaliseGas(opts)
				if err != nil {
					return api.MinipoolCloseDetails{}, fmt.Errorf("error estimating finalise gas for MP %s: %w", minipoolAddress.Hex(), err)
				}
				details.GasInfo = gasInfo
			} else {
				// Do a distribution, which will finalize it
				gasInfo, err := mpv3.EstimateDistributeBalanceGas(false, opts)
				if err != nil {
					return api.MinipoolCloseDetails{}, fmt.Errorf("error estimating distribute balance gas for MP %s: %w", minipoolAddress.Hex(), err)
				}
				details.GasInfo = gasInfo
			}
		} else {
			return api.MinipoolCloseDetails{}, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", minipoolAddress.Hex(), mp.GetVersion())
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

	// Check if it's an upgraded Atlas-era minipool
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return nil, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", minipoolAddress.Hex(), mp.GetVersion())
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

	// Get some details
	var status types.MinipoolStatus
	var distributed bool
	var wg errgroup.Group
	wg.Go(func() error {
		var err error
		status, err = mp.GetStatus(nil)
		if err != nil {
			return fmt.Errorf("error getting status of minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})
	wg.Go(func() error {
		var err error
		distributed, err = mpv3.GetUserDistributed(nil)
		if err != nil {
			return fmt.Errorf("error checking distributed flag of minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	if status == types.Dissolved {
		// If it's dissolved, just close it
		hash, err := mp.Close(opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = hash
	} else if distributed {
		// It's already been distributed so just finalize it
		hash, err := mpv3.Finalise(opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = hash
	} else {
		// Do a distribution, which will finalize it
		hash, err := mpv3.DistributeBalance(false, opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = hash
	}

	// Return response
	return &response, nil

}
