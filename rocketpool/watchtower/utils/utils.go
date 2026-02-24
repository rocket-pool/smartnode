package utils

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

const (
	MinWatchtowerMaxFee        float64 = 20
	MinWatchtowerPriorityFee   float64 = 0.01
	BalanceSubmissionForcedGas uint64  = 300000
	RewardsSubmissionForcedGas uint64  = 300000
)

// Get the max fee for watchtower transactions
func GetWatchtowerMaxFee(cfg *config.RocketPoolConfig) float64 {
	setting := cfg.Smartnode.WatchtowerMaxFeeOverride.Value.(float64)
	return max(MinWatchtowerMaxFee, setting)
}

// Get the priority fee for watchtower transactions
func GetWatchtowerPrioFee(cfg *config.RocketPoolConfig) float64 {
	setting := cfg.Smartnode.WatchtowerPrioFeeOverride.Value.(float64)
	return max(MinWatchtowerPriorityFee, setting)
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

func FindNextSubmissionTarget(rp *rocketpool.RocketPool, eth2Config beacon.Eth2Config, bc beacon.Client, ec rocketpool.ExecutionClient, lastSubmissionBlock uint64, referenceTimestamp int64, submissionIntervalInSeconds int64) (uint64, time.Time, *types.Header, error) {
	lastSubmissionSlotTimestamp := referenceTimestamp

	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)

	if lastSubmissionBlock > 0 {
		lastSubmissionBlockHeader, err := rp.Client.HeaderByNumber(context.Background(), big.NewInt(int64(lastSubmissionBlock)))
		if err != nil {
			return 0, time.Time{}, nil, fmt.Errorf("can't get the latest submission block header: %w", err)
		}

		lastSubmissionParent, _, err := bc.GetBeaconBlock(lastSubmissionBlockHeader.ParentBeaconRoot.Hex())
		if err != nil {
			return 0, time.Time{}, nil, fmt.Errorf("can't get the parent block: %w", err)
		}

		lastSubmissionSlot := lastSubmissionParent.Slot + 1
		lastSubmissionSlotTimestamp = genesisTime.Add(time.Duration((lastSubmissionSlot)*eth2Config.SecondsPerSlot) * time.Second).Unix()
	}

	beaconHead, err := bc.GetBeaconHead()
	if err != nil {
		return 0, time.Time{}, nil, err
	}
	finalizedEpoch := beaconHead.FinalizedEpoch

	// Calculate the timestamp at the start of the head epoch
	finalizedEpochStartSlot := finalizedEpoch * eth2Config.SlotsPerEpoch
	finalizedEpochTimestamp := genesisTime.Add(time.Duration(finalizedEpochStartSlot*eth2Config.SecondsPerSlot) * time.Second)

	// Find the highest valid submissionTimestamp that is <= headTime
	maxSubmissionTimestamp := int64(0)
	for n := int64(0); ; n++ {
		ts := lastSubmissionSlotTimestamp + n*submissionIntervalInSeconds
		if ts > finalizedEpochTimestamp.Unix() {
			break
		}
		maxSubmissionTimestamp = ts
	}

	// Now, use this maxSubmissionTimestamp for slot calculations
	nextSubmissionTime := time.Unix(maxSubmissionTimestamp, 0)
	timeSinceGenesis := nextSubmissionTime.Sub(genesisTime)
	slotNumber := uint64(timeSinceGenesis.Seconds()) / eth2Config.SecondsPerSlot

	// Search for the last existing EL block, going back up to 32 slots if the block is not found.
	targetBlock, err := FindLastBlockWithExecutionPayload(bc, slotNumber)
	if err != nil {
		return 0, time.Time{}, nil, err
	}

	targetBlockNumber := targetBlock.ExecutionBlockNumber
	if targetBlockNumber <= lastSubmissionBlock {
		return 0, time.Time{}, nil, fmt.Errorf("target block number is the same as the last submission block")
	}

	targetBlockHeader, err := ec.HeaderByNumber(context.Background(), big.NewInt(int64(targetBlockNumber)))
	if err != nil {
		return 0, time.Time{}, nil, err
	}

	targetSlot := targetBlock.Slot

	targetSlotTime := eth2Config.GetSlotTime(targetSlot)

	return targetSlot, targetSlotTime, targetBlockHeader, nil
}
