package rp

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/state"
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
