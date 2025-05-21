package rp

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"golang.org/x/sync/errgroup"
)

type FeeRecipientInfo struct {
	SmoothingPoolAddress  common.Address `json:"smoothingPoolAddress"`
	FeeDistributorAddress common.Address `json:"feeDistributorAddress"`
	IsInSmoothingPool     bool           `json:"isInSmoothingPool"`
	IsInOptOutCooldown    bool           `json:"isInOptOutCooldown"`
	OptOutEpoch           uint64         `json:"optOutEpoch"`
}

func GetFeeRecipientInfo(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address, state *state.NetworkState) (*FeeRecipientInfo, error) {

	info := &FeeRecipientInfo{
		IsInOptOutCooldown: false,
		OptOutEpoch:        0,
	}

	mpd := state.NodeDetailsByAddress[nodeAddress]

	// Get info
	info.SmoothingPoolAddress = state.NetworkDetails.SmoothingPoolAddress
	info.FeeDistributorAddress = mpd.FeeDistributorAddress
	info.IsInSmoothingPool = mpd.SmoothingPoolRegistrationState

	// Calculate the safe opt-out epoch if applicable
	if !info.IsInSmoothingPool {
		// Get the opt out time
		optOutTime := time.Unix(mpd.SmoothingPoolRegistrationChanged.Int64(), 0)

		// Get the Beacon info
		beaconConfig := state.BeaconConfig
		beaconHead, err := bc.GetBeaconHead()
		if err != nil {
			return nil, fmt.Errorf("Error getting Beacon head: %w", err)
		}

		// Check if the user just opted out
		if optOutTime != time.Unix(0, 0) {
			// Get the epoch for that time
			genesisTime := time.Unix(int64(beaconConfig.GenesisTime), 0)
			secondsSinceGenesis := optOutTime.Sub(genesisTime)
			epoch := uint64(secondsSinceGenesis.Seconds()) / beaconConfig.SecondsPerEpoch

			// Make sure epoch + 1 is finalized - if not, they're still on cooldown
			targetEpoch := epoch + 1
			if beaconHead.FinalizedEpoch < targetEpoch {
				info.IsInOptOutCooldown = true
				info.OptOutEpoch = targetEpoch
			}
		}
	}

	return info, nil

}

func GetFeeRecipientInfoWithoutState(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address, opts *bind.CallOpts) (*FeeRecipientInfo, error) {
	info := &FeeRecipientInfo{
		IsInOptOutCooldown: false,
		OptOutEpoch:        0,
	}

	// Sync
	var wg errgroup.Group

	// Get the smoothing pool address
	wg.Go(func() error {
		smoothingPoolContract, err := rp.GetContract("rocketSmoothingPool", opts)
		if err != nil {
			return fmt.Errorf("Error getting smoothing pool contract: %w", err)
		}
		info.SmoothingPoolAddress = *smoothingPoolContract.Address
		return nil
	})

	// Get the node's fee distributor
	wg.Go(func() error {
		distributorAddress, err := node.GetDistributorAddress(rp, nodeAddress, opts)
		if err != nil {
			return fmt.Errorf("Error getting the fee distributor for %s: %w", nodeAddress.Hex(), err)
		}
		info.FeeDistributorAddress = distributorAddress
		return nil
	})

	// Check if the user's opted into the smoothing pool
	wg.Go(func() error {
		isOptedIn, err := node.GetSmoothingPoolRegistrationState(rp, nodeAddress, opts)
		if err != nil {
			return err
		}
		info.IsInSmoothingPool = isOptedIn
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Calculate the safe opt-out epoch if applicable
	if !info.IsInSmoothingPool {
		// Get the opt out time
		optOutTime, err := node.GetSmoothingPoolRegistrationChanged(rp, nodeAddress, opts)
		if err != nil {
			return nil, fmt.Errorf("Error getting smoothing pool opt-out time: %w", err)
		}

		// Get the Beacon info
		beaconConfig, err := bc.GetEth2Config()
		if err != nil {
			return nil, fmt.Errorf("Error getting Beacon config: %w", err)
		}
		beaconHead, err := bc.GetBeaconHead()
		if err != nil {
			return nil, fmt.Errorf("Error getting Beacon head: %w", err)
		}

		// Check if the user just opted out
		if optOutTime != time.Unix(0, 0) {
			// Get the epoch for that time
			genesisTime := time.Unix(int64(beaconConfig.GenesisTime), 0)
			secondsSinceGenesis := optOutTime.Sub(genesisTime)
			epoch := uint64(secondsSinceGenesis.Seconds()) / beaconConfig.SecondsPerEpoch

			// Make sure epoch + 1 is finalized - if not, they're still on cooldown
			targetEpoch := epoch + 1
			if beaconHead.FinalizedEpoch < targetEpoch {
				info.IsInOptOutCooldown = true
				info.OptOutEpoch = targetEpoch
			}
		}
	}

	return info, nil

}
