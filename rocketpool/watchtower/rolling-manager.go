package watchtower

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Manager for RollingRecords
type RollingRecordManager struct {
	Record                       *rewards.RollingRecord
	LatestFinalizedEpoch         uint64
	ExpectedBalancesBlock        uint64
	ExpectedRewardsIntervalBlock uint64

	log         log.ColorLogger
	logPrefix   string
	rp          *rocketpool.RocketPool
	bc          beacon.Client
	genesisTime time.Time
	nodeAddress common.Address
	mgr         *state.NetworkStateManager
}

// Creates a new manager for RollingRecords.
func NewRollingRecordManager(log log.ColorLogger, logPrefix string, rp *rocketpool.RocketPool, bc beacon.Client, mgr *state.NetworkStateManager, nodeAddress common.Address, startSlot uint64, beaconConfig *beacon.Eth2Config) (*RollingRecordManager, error) {
	cfg, err := bc.GetEth2Config()
	if err != nil {
		return nil, fmt.Errorf("error getting beacon config: %w", err)
	}
	genesisTime := time.Unix(int64(cfg.GenesisTime), 0)

	return &RollingRecordManager{
		Record: rewards.NewRollingRecord(log, logPrefix, bc, startSlot, beaconConfig),

		log:         log,
		logPrefix:   logPrefix,
		rp:          rp,
		bc:          bc,
		genesisTime: genesisTime,
		nodeAddress: nodeAddress,
		mgr:         mgr,
	}, nil
}

// Process the details of the latest head state
func (r *RollingRecordManager) ProcessNewHeadState(state *state.NetworkState) error {

	// Get the latest finalized slot
	latestFinalizedBlock, err := r.mgr.GetLatestFinalizedBeaconBlock()
	if err != nil {
		return fmt.Errorf("error getting latest finalized block: %w", err)
	}

	// Check if a network balance update is due
	isNetworkBalanceUpdateDue, networkBalanceSlot, err := r.isNetworkBalanceUpdateRequired(state)
	if err != nil {
		return fmt.Errorf("error checking if network balance update is required: %w", err)
	}

	// Check if a rewards interval is due
	isRewardsSubmissionDue, rewardsSlot, err := r.isRewardsIntervalSubmissionRequired(state)
	if err != nil {
		return fmt.Errorf("error checking if rewards submission is required: %w", err)
	}

	if !isNetworkBalanceUpdateDue && !isRewardsSubmissionDue {
		// No special upcoming state required, so update normally
		_, err = r.updateToSlot(latestFinalizedBlock.Slot)
		if err != nil {
			return fmt.Errorf("error during previous rolling record update: %w", err)
		}
		return nil
	}

	var earliestRequiredSlot uint64
	if isNetworkBalanceUpdateDue && isRewardsSubmissionDue {
		if networkBalanceSlot < rewardsSlot {
			earliestRequiredSlot = networkBalanceSlot
		} else {
			earliestRequiredSlot = rewardsSlot
		}
	} else if isNetworkBalanceUpdateDue {
		earliestRequiredSlot = networkBalanceSlot
	} else if isRewardsSubmissionDue {
		earliestRequiredSlot = rewardsSlot
	}

	// Do NOT do an update until the required slot has been finalized
	if latestFinalizedBlock.Slot < earliestRequiredSlot {
		r.log.Printlnf("%s TODO: Message goes here")
	}
}

// Start an update to a given slot
func (r *RollingRecordManager) updateToSlot(slot uint64) (bool, error) {
	// Skip it if the latest record is already up to date or is in the process
	if r.Record.PendingSlot >= slot {
		return false, nil
	}

	// Get the latest finalized state
	state, err := r.mgr.GetStateForSlot(slot)
	if err != nil {
		return false, fmt.Errorf("error getting state at slot %d: %w", slot, err)
	}

	// Run an update on it
	updateStarted, err := r.Record.UpdateToState(state, false)
	if err != nil {
		return false, fmt.Errorf("error updating rolling record to slot %d, block %d: %w", state.BeaconSlotNumber, state.ElBlockNumber, err)
	}

	return updateStarted, nil
}

// Check if a network balance submission is required and if so, the slot number for the update
func (r *RollingRecordManager) isNetworkBalanceUpdateRequired(state *state.NetworkState) (bool, uint64, error) {
	// Get block to submit balances for
	blockNumberBig := state.NetworkDetails.LatestReportableBalancesBlock
	blockNumber := blockNumberBig.Uint64()

	// Check if a submission needs to be made
	if blockNumber <= state.NetworkDetails.BalancesBlock.Uint64() {
		return false, 0, nil
	}

	// Check if this node has already submitted a balance
	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)
	hasSubmitted, err := r.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.balances.submitted.node"), r.nodeAddress.Bytes(), blockNumberBuf))
	if err != nil {
		return false, 0, fmt.Errorf("error checking if node has already submitted network balance for block %d: %w", blockNumber, err)
	}
	if hasSubmitted {
		return false, 0, nil
	}

	// Get the time of the block
	header, err := r.rp.Client.HeaderByNumber(context.Background(), big.NewInt(0).SetUint64(blockNumber))
	if err != nil {
		return false, 0, fmt.Errorf("error getting header for block %d: %w", blockNumber, err)
	}
	blockTime := time.Unix(int64(header.Time), 0)

	// Get the Beacon block corresponding to this time
	eth2Config := state.BeaconConfig
	timeSinceGenesis := blockTime.Sub(r.genesisTime)
	slotNumber := uint64(timeSinceGenesis.Seconds()) / eth2Config.SecondsPerSlot
	return true, slotNumber, nil
}

// Check if a rewards interval submission is required and if so, the slot number for the update
func (r *RollingRecordManager) isRewardsIntervalSubmissionRequired(state *state.NetworkState) (bool, uint64, error) {
	// Check if a rewards interval has passed and needs to be calculated
	startTime := state.NetworkDetails.IntervalStart
	intervalTime := state.NetworkDetails.IntervalDuration

	// Calculate the end time, which is the number of intervals that have gone by since the current one's start
	secondsSinceGenesis := time.Duration(state.BeaconConfig.SecondsPerSlot*state.BeaconSlotNumber) * time.Second
	stateTime := r.genesisTime.Add(secondsSinceGenesis)
	timeSinceStart := stateTime.Sub(startTime)
	intervalsPassed := timeSinceStart / intervalTime
	endTime := startTime.Add(intervalTime * intervalsPassed)
	if intervalsPassed == 0 {
		return false, 0, nil
	}

	// Check if this node already submitted a tree
	currentIndex := state.NetworkDetails.RewardIndex
	currentIndexBig := big.NewInt(0).SetUint64(currentIndex)
	indexBuffer := make([]byte, 32)
	currentIndexBig.FillBytes(indexBuffer)
	hasSubmitted, err := r.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("rewards.snapshot.submitted.node"), r.nodeAddress.Bytes(), indexBuffer))
	if err != nil {
		return false, 0, fmt.Errorf("error checking if node has already submitted for rewards interval %d: %w", currentIndex, err)
	}
	if hasSubmitted {
		return false, 0, nil
	}

	// Get the target slot number
	eth2Config := state.BeaconConfig
	totalTimespan := endTime.Sub(r.genesisTime)
	targetSlot := uint64(math.Ceil(totalTimespan.Seconds() / float64(eth2Config.SecondsPerSlot)))
	targetSlotEpoch := targetSlot / eth2Config.SlotsPerEpoch
	targetSlot = targetSlotEpoch*eth2Config.SlotsPerEpoch + (eth2Config.SlotsPerEpoch - 1) // The target slot becomes the last one in the Epoch

	// Get the first successful block
	for {
		// Try to get the current block
		_, exists, err := r.bc.GetBeaconBlock(fmt.Sprint(targetSlot))
		if err != nil {
			return false, 0, fmt.Errorf("error getting Beacon block %d: %w", targetSlot, err)
		}

		// If the block was missing, try the previous one
		if !exists {
			r.log.Printlnf("Slot %d was missing, trying the previous one...", targetSlot)
			targetSlot--
		} else {
			// Ok, we have the first proposed block - this is the one to use for the snapshot!
			return true, targetSlot, nil
		}
	}
}
