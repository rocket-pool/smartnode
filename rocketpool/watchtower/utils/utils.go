package utils

import (
	"fmt"
	"strconv"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

const (
	MinWatchtowerMaxFee        float64 = 200
	MinWatchtowerPriorityFee   float64 = 3
	BalanceSubmissionForcedGas uint64  = 64000
	RewardsSubmissionForcedGas uint64  = 64000
)

// Get the max fee for watchtower transactions
func GetWatchtowerMaxFee(cfg *config.RocketPoolConfig) float64 {
	setting := cfg.Smartnode.WatchtowerMaxFeeOverride.Value.(float64)
	if setting < MinWatchtowerMaxFee {
		return MinWatchtowerMaxFee
	}
	return setting
}

// Get the priority fee for watchtower transactions
func GetWatchtowerPrioFee(cfg *config.RocketPoolConfig) float64 {
	setting := cfg.Smartnode.WatchtowerPrioFeeOverride.Value.(float64)
	if setting < MinWatchtowerPriorityFee {
		return MinWatchtowerPriorityFee
	}
	return setting
}

func FindLastBlockWithExecutionPayload(bc beacon.Client, slotNumber uint64) (beacon.BeaconBlock, error) {
	beaconBlock := beacon.BeaconBlock{}
	var err error
	for blockExists, searchSlot := false, slotNumber; !blockExists; searchSlot -= 1 {
		beaconBlock, blockExists, err = bc.GetBeaconBlock(strconv.FormatUint(searchSlot, 10))
		if err != nil {
			return beacon.BeaconBlock{}, err
		}
		// If we go back more than 32 slots, error out
		if slotNumber-searchSlot > 32 {
			return beacon.BeaconBlock{}, fmt.Errorf("could not find EL block from slot %d", slotNumber)
		}
	}
	return beaconBlock, nil
}

func FindNextSubmissionTimestamp(latestBlockTimestamp int64, referenceTimestamp int64, submissionIntervalInSeconds int64) (int64, error) {
	if latestBlockTimestamp == 0 || referenceTimestamp == 0 || submissionIntervalInSeconds == 0 {
		return 0, fmt.Errorf("FindNextSubmissionTimestamp can't use zero values")
	}

	// Calculate the difference between latestBlockTime and the reference timestamp
	timeDifference := latestBlockTimestamp - referenceTimestamp
	if timeDifference < 0 {
		return 0, fmt.Errorf("FindNextSubmissionTimestamp referenceTimestamp in the future")
	}

	// Calculate the remainder to find out how far off from a multiple of the interval the current time is
	remainder := timeDifference % submissionIntervalInSeconds

	// Subtract the remainder from current time to find the first multiple of the interval in the past
	submissionTimeRef := latestBlockTimestamp - remainder
	return submissionTimeRef, nil
}
