package utils

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/smartnode/v2/shared/config"
)

const (
	MinWatchtowerMaxFee        float64 = 200
	MinWatchtowerPriorityFee   float64 = 3
	BalanceSubmissionForcedGas uint64  = 64000
	RewardsSubmissionForcedGas uint64  = 64000
)

// Get the max fee for watchtower transactions
func GetWatchtowerMaxFee(cfg *config.SmartNodeConfig) float64 {
	setting := cfg.WatchtowerMaxFeeOverride.Value
	if setting < MinWatchtowerMaxFee {
		return MinWatchtowerMaxFee
	}
	return setting
}

// Get the priority fee for watchtower transactions
func GetWatchtowerPrioFee(cfg *config.SmartNodeConfig) float64 {
	setting := cfg.WatchtowerPriorityFeeOverride.Value
	if setting < MinWatchtowerPriorityFee {
		return MinWatchtowerPriorityFee
	}
	return setting
}

func FindLastExistingELBlockFromSlot(ctx context.Context, bc beacon.IBeaconClient, slotNumber uint64) (beacon.Eth1Data, error) {
	ecBlock := beacon.Eth1Data{}
	var err error
	for blockExists, searchSlot := false, slotNumber; !blockExists; searchSlot -= 1 {
		ecBlock, blockExists, err = bc.GetEth1DataForEth2Block(ctx, strconv.FormatUint(searchSlot, 10))
		if err != nil {
			return ecBlock, err
		}
		// If we go back more than 32 slots, error out
		if slotNumber-searchSlot > 32 {
			return ecBlock, fmt.Errorf("could not find EL block from slot %d", slotNumber)
		}
	}
	return ecBlock, nil
}
