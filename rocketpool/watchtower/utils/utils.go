package utils

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
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

// Walks back from slotNumber until it finds a block that actually has an Execution Layer block
// behind it, and returns it with ExecutionBlockNumber populated. Gloas (EIP-7732) blocks carry
// only the payload bid's block hash, so the number is resolved through the EL.
func FindLastBlockWithExecutionPayload(bc beacon.Client, ec rocketpool.ExecutionClient, slotNumber uint64) (beacon.BeaconBlock, error) {
	for searchSlot := slotNumber; ; searchSlot-- {
		// If we go back more than 32 slots, error out
		if slotNumber-searchSlot > 32 {
			return beacon.BeaconBlock{}, fmt.Errorf("could not find EL block from slot %d", slotNumber)
		}

		beaconBlock, blockExists, err := bc.GetBeaconBlock(strconv.FormatUint(searchSlot, 10))
		if err != nil {
			return beacon.BeaconBlock{}, err
		}
		if !blockExists {
			continue
		}

		// Resolve the EL block number
		elBlockNumber, found, err := beacon.ResolveExecutionBlockNumber(context.Background(), ec, beaconBlock)
		if err != nil {
			return beacon.BeaconBlock{}, err
		}
		if !found {
			continue
		}

		beaconBlock.ExecutionBlockNumber = elBlockNumber
		return beaconBlock, nil
	}
}

func FindNextSubmissionTarget(rp *rocketpool.RocketPool, eth2Config beacon.Eth2Config, bc beacon.Client, ec rocketpool.ExecutionClient, lastSubmissionBlock uint64, referenceTimestamp int64, submissionIntervalInSeconds int64) (uint64, time.Time, *types.Header, bool, error) {
	lastSubmissionSlotTimestamp := referenceTimestamp

	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)

	if lastSubmissionBlock > 0 {
		lastSubmissionBlockHeader, err := rp.Client.HeaderByNumber(context.Background(), big.NewInt(int64(lastSubmissionBlock)))
		if err != nil {
			return 0, time.Time{}, nil, false, fmt.Errorf("can't get the latest submission block header: %w", err)
		}

		lastSubmissionParent, _, err := bc.GetBeaconBlock(lastSubmissionBlockHeader.ParentBeaconRoot.Hex())
		if err != nil {
			return 0, time.Time{}, nil, false, fmt.Errorf("can't get the parent block: %w", err)
		}

		lastSubmissionSlot := lastSubmissionParent.Slot + 1
		lastSubmissionSlotTimestamp = genesisTime.Add(time.Duration((lastSubmissionSlot)*eth2Config.SecondsPerSlot) * time.Second).Unix()
	}

	beaconHead, err := bc.GetBeaconHead()
	if err != nil {
		return 0, time.Time{}, nil, false, err
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
	targetBlock, err := FindLastBlockWithExecutionPayload(bc, ec, slotNumber)
	if err != nil {
		return 0, time.Time{}, nil, false, err
	}

	targetBlockNumber := targetBlock.ExecutionBlockNumber

	targetBlockHeader, err := ec.HeaderByNumber(context.Background(), big.NewInt(int64(targetBlockNumber)))
	if err != nil {
		return 0, time.Time{}, nil, false, err
	}

	targetSlot := targetBlock.Slot

	targetSlotTime := eth2Config.GetSlotTime(targetSlot)

	// Don't submit if targetBlockNumber <= lastSubmissionBlock, but pass the
	// targetSlot, targetSlotTime, targetBlockHeader so caller can handle logging
	if targetBlockNumber <= lastSubmissionBlock {
		return targetSlot, targetSlotTime, targetBlockHeader, false, nil
	}

	return targetSlot, targetSlotTime, targetBlockHeader, true, nil
}

// Determines if the primary EC can be used for historical queries, or if the Archive EC is required
func GetBestApiClient(primary *rocketpool.RocketPool, cfg *config.RocketPoolConfig, printMessage func(string), blockNumber *big.Int) (*rocketpool.RocketPool, error) {

	client := primary

	// Try getting the rETH address as a canary to see if the block is available
	opts := &bind.CallOpts{
		BlockNumber: blockNumber,
	}
	address, err := client.RocketStorage.GetAddress(opts, crypto.Keccak256Hash([]byte("contract.addressrocketTokenRETH")))
	if err != nil {
		errMessage := err.Error()
		printMessage(fmt.Sprintf("Error getting state for block %d: %s", blockNumber.Uint64(), errMessage))
		// The state was missing so fall back to the archive node
		archiveEcUrl := cfg.Smartnode.ArchiveECUrl.Value.(string)
		if archiveEcUrl != "" {
			printMessage(fmt.Sprintf("Primary EC cannot retrieve state for historical block %d, using archive EC [%s]", blockNumber.Uint64(), archiveEcUrl))
			ec, err := services.NewEthClient(archiveEcUrl)
			if err != nil {
				return nil, fmt.Errorf("Error connecting to archive EC: %w", err)
			}
			client, err = rocketpool.NewRocketPool(ec, common.HexToAddress(cfg.Smartnode.GetStorageAddress()))
			if err != nil {
				return nil, fmt.Errorf("Error creating Rocket Pool client connected to archive EC: %w", err)
			}

			// Get the rETH address from the archive EC
			address, err = client.RocketStorage.GetAddress(opts, crypto.Keccak256Hash([]byte("contract.addressrocketTokenRETH")))
			if err != nil {
				return nil, fmt.Errorf("Error verifying rETH address with Archive EC: %w", err)
			}
		} else {
			// No archive node specified
			return nil, fmt.Errorf("***ERROR*** Primary EC cannot retrieve state for historical block %d and the Archive EC is not specified.", blockNumber.Uint64())
		}
	}

	// Sanity check the rETH address to make sure the client is working right
	if address != cfg.Smartnode.GetRethAddress() {
		return nil, fmt.Errorf("***ERROR*** Your Primary EC provided %s as the rETH address, but it should have been %s!", address.Hex(), cfg.Smartnode.GetRethAddress().Hex())
	}

	return client, nil

}
