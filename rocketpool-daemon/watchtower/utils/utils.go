package utils

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
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

func FindLastBlockWithExecutionPayload(ctx context.Context, bc beacon.IBeaconClient, slotNumber uint64) (beacon.BeaconBlock, error) {
	beaconBlock := beacon.BeaconBlock{}
	var err error
	for blockExists, searchSlot := false, slotNumber; !blockExists; searchSlot -= 1 {
		beaconBlock, blockExists, err = bc.GetBeaconBlock(ctx, strconv.FormatUint(searchSlot, 10))
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

func FindNextSubmissionTarget(ctx context.Context, rp *rocketpool.RocketPool, eth2Config beacon.Eth2Config, bc beacon.IBeaconClient, ec eth.IExecutionClient, lastSubmissionBlock uint64, referenceTimestamp int64, submissionIntervalInSeconds int64) (uint64, time.Time, *types.Header, error) {
	// Get the time of the last submission
	lastSubmissionBlockHeader, err := rp.Client.HeaderByNumber(context.Background(), big.NewInt(int64(lastSubmissionBlock)))
	if err != nil {
		return 0, time.Time{}, nil, fmt.Errorf("can't get the latest submission block header: %w", err)
	}
	
	// Get the time of the latest block
	latestEth1Block, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, time.Time{}, nil, fmt.Errorf("can't get the latest block time: %w", err)
	}
	latestBlockTimestamp := int64(latestEth1Block.Time)

	if int64(lastSubmissionBlockHeader.Time)+submissionIntervalInSeconds > latestBlockTimestamp {
		return 0, time.Time{}, nil, fmt.Errorf("not enough time has passed for the next price/balances submission")
	}

	// Calculate the next submission timestamp
	submissionTimestamp, err := FindNextSubmissionTimestamp(latestBlockTimestamp, referenceTimestamp, submissionIntervalInSeconds)
	if err != nil {
		return 0, time.Time{}, nil, err
	}

	// Convert the submission timestamp to time.Time
	nextSubmissionTime := time.Unix(submissionTimestamp, 0)

	// Get the Beacon block corresponding to this time
	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)
	timeSinceGenesis := nextSubmissionTime.Sub(genesisTime)
	slotNumber := uint64(timeSinceGenesis.Seconds()) / eth2Config.SecondsPerSlot

	// Search for the last existing EL block, going back up to 32 slots if the block is not found.
	targetBlock, err := FindLastBlockWithExecutionPayload(ctx, bc, slotNumber)
	if err != nil {
		return 0, time.Time{}, nil, err
	}

	targetBlockNumber := targetBlock.ExecutionBlockNumber

	targetBlockHeader, err := ec.HeaderByNumber(context.Background(), big.NewInt(int64(targetBlockNumber)))
	if err != nil {
		return 0, time.Time{}, nil, err
	}
	requiredEpoch := slotNumber / eth2Config.SlotsPerEpoch

	// Check if the required epoch is finalized yet
	beaconHead, err := bc.GetBeaconHead(ctx)
	if err != nil {
		return 0, time.Time{}, nil, err
	}
	finalizedEpoch := beaconHead.FinalizedEpoch
	if requiredEpoch > finalizedEpoch {
		return 0, time.Time{}, nil, fmt.Errorf("balances must be reported for EL block %d, waiting until Epoch %d is finalized (currently %d)", targetBlockNumber, requiredEpoch, finalizedEpoch)
	}

	return slotNumber, nextSubmissionTime, targetBlockHeader, nil
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
