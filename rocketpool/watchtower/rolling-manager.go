package watchtower

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/klauspost/compress/zstd"
	rprewards "github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

const (
	recordsFilenameFormat    string = "%d-%d.json.zst"
	recordsFilenamePattern   string = "(?P<slot>\\d+)\\-(?P<epoch>\\d+)\\.json\\.zst"
	checksumTableFilename    string = "checksums.sha384"
	recordCheckpointInterval uint64 = 45
	checkpointRetentionLimit uint64 = 200
)

// Manager for RollingRecords
type RollingRecordManager struct {
	Record                       *rewards.RollingRecord
	LatestFinalizedEpoch         uint64
	ExpectedBalancesBlock        uint64
	ExpectedRewardsIntervalBlock uint64

	log         *log.ColorLogger
	errLog      *log.ColorLogger
	logPrefix   string
	cfg         *config.RocketPoolConfig
	w           *wallet.Wallet
	nodeAddress *common.Address
	rp          *rocketpool.RocketPool
	bc          beacon.Client
	mgr         *state.NetworkStateManager
	startSlot   uint64

	submitNetworkBalances *submitNetworkBalances
	submitRewardsTree     *submitRewardsTree
	beaconCfg             beacon.Eth2Config
	genesisTime           time.Time
	compressor            *zstd.Encoder
	decompressor          *zstd.Decoder
	recordsFilenameRegex  *regexp.Regexp

	lock      *sync.Mutex
	isRunning bool
}

// Creates a new manager for rolling records.
func NewRollingRecordManager(log *log.ColorLogger, errLog *log.ColorLogger, cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, bc beacon.Client, mgr *state.NetworkStateManager, w *wallet.Wallet, startSlot uint64, beaconCfg beacon.Eth2Config) (*RollingRecordManager, error) {
	// Get the Beacon genesis time
	genesisTime := time.Unix(int64(beaconCfg.GenesisTime), 0)

	// Create the zstd compressor and decompressor
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return nil, fmt.Errorf("error creating zstd compressor for rolling record manager: %w", err)
	}
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating zstd decompressor for rolling record manager: %w", err)
	}

	// Create the records filename regex
	recordsFilenameRegex := regexp.MustCompile(recordsFilenamePattern)

	// Make the records folder if it doesn't exist
	recordsPath := cfg.Smartnode.GetRecordsPath()
	fileInfo, err := os.Stat(recordsPath)
	if os.IsNotExist(err) {
		err2 := os.MkdirAll(recordsPath, 0755)
		if err2 != nil {
			return nil, fmt.Errorf("error creating rolling records folder: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error checking rolling records folder: %w", err)
	} else if !fileInfo.IsDir() {
		return nil, fmt.Errorf("rolling records folder location exists (%s), but is not a folder", recordsPath)
	}

	var nodeAddress *common.Address
	var submitNetworkBalances *submitNetworkBalances
	var submitRewardsTree *submitRewardsTree
	if w != nil {
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return nil, fmt.Errorf("error getting node account: %w", err)
		}
		nodeAddress = &nodeAccount.Address

		submitNetworkBalances = newSubmitNetworkBalances(log, errLog, cfg, w, rp, bc)
		submitRewardsTree = newSubmitRewardsTree(log, errLog, cfg, w, rp, bc)
	}

	lock := &sync.Mutex{}
	logPrefix := "[Rolling Record]"
	return &RollingRecordManager{
		Record: rewards.NewRollingRecord(log, logPrefix, bc, startSlot, &beaconCfg),

		log:                   log,
		errLog:                errLog,
		logPrefix:             logPrefix,
		cfg:                   cfg,
		rp:                    rp,
		bc:                    bc,
		mgr:                   mgr,
		w:                     w,
		nodeAddress:           nodeAddress,
		startSlot:             startSlot,
		submitNetworkBalances: submitNetworkBalances,
		submitRewardsTree:     submitRewardsTree,
		beaconCfg:             beaconCfg,
		genesisTime:           genesisTime,
		compressor:            encoder,
		decompressor:          decoder,
		recordsFilenameRegex:  recordsFilenameRegex,
		lock:                  lock,
		isRunning:             false,
	}, nil
}

// Process the details of the latest head state
func (r *RollingRecordManager) ProcessNewHeadState(state *state.NetworkState) error {

	r.log.Printlnf("Updating record to head state (slot %d)...", state.BeaconSlotNumber)

	r.lock.Lock()
	if r.isRunning {
		r.log.Println("Record update is already running in the background.")
		r.lock.Unlock()
		return nil
	}
	r.lock.Unlock()

	// Get the latest finalized slot and epoch
	latestFinalizedBlock, err := r.mgr.GetLatestFinalizedBeaconBlock()
	if err != nil {
		return fmt.Errorf("error getting latest finalized block: %w", err)
	}
	latestFinalizedEpoch := latestFinalizedBlock.Slot / state.BeaconConfig.SlotsPerEpoch

	// Check if a network balance update is due
	isNetworkBalanceUpdateDue, networkBalanceSlot, err := r.isNetworkBalanceUpdateRequired(state)
	if err != nil {
		return fmt.Errorf("error checking if network balance update is required: %w", err)
	}

	// Check if this node has already submitted network balances
	hasSubmittedNetworkBalances, err := r.hasSubmittedNetworkBalances(state.NetworkDetails.LatestReportableBalancesBlock.Uint64())
	if err != nil {
		return fmt.Errorf("error checking if node has submitted network balances: %w", err)
	}

	// Check if a rewards interval is due
	isRewardsSubmissionDue, rewardsSlot, err := r.isRewardsIntervalSubmissionRequired(state)
	if err != nil {
		return fmt.Errorf("error checking if rewards submission is required: %w", err)
	}

	// Check if this node has already submitted rewards info
	hasSubmittedRewards, err := r.hasSubmittedRewards(state.NetworkDetails.RewardIndex)
	if err != nil {
		return fmt.Errorf("error checking if node has submitted rewards: %w", err)
	}

	// If no special upcoming state is required, update normally
	if !isNetworkBalanceUpdateDue && !isRewardsSubmissionDue {
		go func() {
			r.lock.Lock()
			r.isRunning = true
			r.lock.Unlock()

			// Update the record
			r.log.Printlnf("%s Starting record update in a separate thread.", r.logPrefix)
			err = r.updateToSlot(state, latestFinalizedBlock.Slot)
			if err != nil {
				r.handleError(fmt.Errorf("%s error during rolling record update: %w", r.logPrefix, err))
				return
			}

			// Log and return
			r.log.Printlnf("%s Record update complete.", r.logPrefix)
			r.lock.Lock()
			r.isRunning = false
			r.lock.Unlock()
		}()
		return nil
	}

	// Handle cases where at only one report is due but we've already submitted it, or when we've already submitted both
	if hasSubmittedNetworkBalances && !isRewardsSubmissionDue {
		r.log.Printlnf("%s Network balances have already been submitted for block %s but consensus hasn't been reached yet, skipping record update.", r.logPrefix, state.NetworkDetails.LatestReportableBalancesBlock.String())
		return nil
	}
	if hasSubmittedRewards && !isNetworkBalanceUpdateDue {
		r.log.Printlnf("%s Rewards tree has already been submitted for interval %d but consensus hasn't been reached yet, skipping record update.", r.logPrefix, state.NetworkDetails.RewardIndex)
		return nil
	}
	if hasSubmittedNetworkBalances && hasSubmittedRewards {
		r.log.Printlnf("%s Network balances have already been submitted for block %s and rewards tree has already been submitted for interval %d but consensus hasn't been reached yet for either one, skipping record update.", r.logPrefix, state.NetworkDetails.LatestReportableBalancesBlock.String(), state.NetworkDetails.RewardIndex)
		return nil
	}

	// Check if network balance reporting is ready
	networkBalanceEpoch := networkBalanceSlot / r.beaconCfg.SlotsPerEpoch
	requiredNetworkBalanceEpoch := networkBalanceEpoch + 1
	isNetworkBalanceReadyForReport := isNetworkBalanceUpdateDue && (latestFinalizedEpoch >= requiredNetworkBalanceEpoch) && !hasSubmittedNetworkBalances

	// Check if rewards reporting is ready
	var rewardsElBlock uint64
	rewardsEpoch := rewardsSlot / r.beaconCfg.SlotsPerEpoch
	requiredRewardsEpoch := rewardsEpoch + 1
	isRewardsReadyForReport := isRewardsSubmissionDue && (latestFinalizedEpoch >= requiredRewardsEpoch) && !hasSubmittedRewards
	if isRewardsReadyForReport {
		// Get the actual slot to report on
		rewardsSlot, rewardsElBlock, err = r.getTrueRewardsIntervalSubmissionSlot(rewardsSlot)
		if err != nil {
			return fmt.Errorf("error getting the true rewards interval slot: %w", err)
		}
	}

	// Run updates and submissions as required
	if isNetworkBalanceReadyForReport || isRewardsReadyForReport {
		go func() {
			r.lock.Lock()
			r.isRunning = true
			r.lock.Unlock()

			if isNetworkBalanceReadyForReport && isRewardsReadyForReport { // Report network balance and rewards
				// If balances are due after rewards (but before rewards have been submitted), report the balances according to the rewards slot
				if rewardsSlot < networkBalanceSlot {
					r.log.Printlnf("%s NOTE: network balance report is due for block %s (slot %d) but this is after the rewards interval due for block %d (slot %d); setting the network balance report to the rewards interval block.", r.logPrefix, state.NetworkDetails.LatestReportableBalancesBlock.String(), networkBalanceSlot, rewardsElBlock, rewardsSlot)
					networkBalanceSlot = rewardsSlot
				}

				// Generate the network balances state
				state, err := r.mgr.GetStateForSlot(networkBalanceSlot)
				if err != nil {
					r.handleError(fmt.Errorf("error getting state for network balances slot: %w", err))
					return
				}

				// Process network balances
				r.log.Printlnf("%s Running network balance report in a separate thread.", r.logPrefix)
				err = r.runNetworkBalancesReport(state)
				if err != nil {
					r.handleError(fmt.Errorf("error running network balances report: %w", err))
					return
				}
				r.log.Printlnf("%s Network Balance report complete.", r.logPrefix)

				// Generate the rewards state
				if networkBalanceSlot != rewardsSlot {
					state, err = r.mgr.GetStateForSlot(networkBalanceSlot)
					if err != nil {
						r.handleError(fmt.Errorf("error getting state for network balances slot: %w", err))
						return
					}
				}

				// Process the rewards interval
				r.log.Printlnf("%s Running rewards interval submission in a separate thread.", r.logPrefix)
				err = r.runRewardsIntervalReport(state)
				if err != nil {
					r.handleError(fmt.Errorf("error running rewards interval report: %w", err))
					return
				}
				r.log.Printlnf("%s Rewards Interval submission complete.", r.logPrefix)

			} else if isNetworkBalanceReadyForReport { // Report network balance only
				if hasSubmittedRewards {
					// Special situation where network balances are required but consensus is still pending on the last rewards interval
					// In this case, use the rewards slot instead
					r.log.Printlnf("%s NOTE: network balance report is due for block %s (slot %d) but this is after the rewards interval due for block %d (slot %d) which hasn't reached consensus yet; setting the network balance report to the rewards interval block.", r.logPrefix, state.NetworkDetails.LatestReportableBalancesBlock.String(), networkBalanceSlot, rewardsElBlock, rewardsSlot)
					networkBalanceSlot = rewardsSlot
				}

				// Generate the network balances state
				state, err := r.mgr.GetStateForSlot(networkBalanceSlot)
				if err != nil {
					r.handleError(fmt.Errorf("error getting state for network balances slot: %w", err))
					return
				}

				// Process network balances
				r.log.Printlnf("%s Running network balance report in a separate thread.", r.logPrefix)
				err = r.runNetworkBalancesReport(state)
				if err != nil {
					r.handleError(fmt.Errorf("error running network balances report: %w", err))
					return
				}

			} else { // Report rewards only
				// Generate the rewards state
				state, err := r.mgr.GetStateForSlot(rewardsSlot)
				if err != nil {
					r.handleError(fmt.Errorf("error getting state for network balances slot: %w", err))
					return
				}

				// Process the rewards interval
				r.log.Printlnf("%s Running rewards interval submission in a separate thread.", r.logPrefix)
				err = r.runRewardsIntervalReport(state)
				if err != nil {
					r.handleError(fmt.Errorf("error running rewards interval report: %w", err))
					return
				}

			}

			r.lock.Lock()
			r.isRunning = false
			r.lock.Unlock()
		}()
	}

	return nil

}

// Print an error and unlock the mutex
func (r *RollingRecordManager) handleError(err error) {
	r.errLog.Printlnf("%s %s", r.logPrefix, err.Error())
	r.errLog.Println("*** Rolling Record processing failed. ***")
	r.lock.Lock()
	r.isRunning = false
	r.lock.Unlock()
}

// Run a network balances submission
func (r *RollingRecordManager) runNetworkBalancesReport(state *state.NetworkState) error {
	networkBalanceSlot := state.BeaconSlotNumber

	// Check if the current record has gone past the requested slot or if it can be updated / used
	if networkBalanceSlot < r.Record.LastDutiesSlot {
		r.log.Printlnf("%s Current record has extended too far (need slot %d, but record has processed slot %d)... reverting to a previous checkpoint.", r.logPrefix, networkBalanceSlot, r.Record.LastDutiesSlot)

		newRecord, err := r.GenerateRecordForState(state)
		if err != nil {
			return fmt.Errorf("error creating record for network balance slot: %w", err)
		}

		r.Record = newRecord
	} else {
		r.log.Printlnf("%s Current record can be used (need slot %d, record has only processed slot %d).", r.logPrefix, networkBalanceSlot, r.Record.LastDutiesSlot)
		err := r.Record.UpdateToSlot(networkBalanceSlot, state)
		if err != nil {
			return fmt.Errorf("error updating record to network balance slot: %w", err)
		}
	}

	// Run the network balance submission with the given state and record
	if r.submitNetworkBalances != nil {
		r.submitNetworkBalances.run(state)
	}

	return nil
}

// Run a rewards interval report submission
func (r *RollingRecordManager) runRewardsIntervalReport(state *state.NetworkState) error {
	rewardsSlot := state.BeaconSlotNumber

	// Check if the current record has gone past the requested slot or if it can be updated / used
	if rewardsSlot < r.Record.LastDutiesSlot {
		r.log.Printlnf("%s Current record has extended too far (need slot %d, but record has processed slot %d)... reverting to a previous checkpoint.", r.logPrefix, rewardsSlot, r.Record.LastDutiesSlot)

		newRecord, err := r.GenerateRecordForState(state)
		if err != nil {
			return fmt.Errorf("error creating record for rewards slot: %w", err)
		}

		r.Record = newRecord
	} else {
		r.log.Printlnf("%s Current record can be used (need slot %d, record has only processed slot %d).", r.logPrefix, rewardsSlot, r.Record.LastDutiesSlot)
		err := r.Record.UpdateToSlot(rewardsSlot, state)
		if err != nil {
			return fmt.Errorf("error updating record to rewards slot: %w", err)
		}
	}

	// Run the rewards interval submission with the given state and record
	if r.submitRewardsTree != nil {
		r.submitRewardsTree.run(true, state)
	}

	return nil
}

// Start an update to a given slot
func (r *RollingRecordManager) updateToSlot(state *state.NetworkState, slot uint64) error {
	// Skip it if the latest record is already up to date or is in the process
	if r.Record.PendingSlot >= slot {
		return nil
	}

	// Get the latest finalized state
	var err error
	if state.BeaconSlotNumber != slot {
		state, err = r.mgr.GetStateForSlot(slot)
		if err != nil {
			return fmt.Errorf("error getting state at slot %d: %w", slot, err)
		}
	}

	// Run an update on it
	err = r.Record.UpdateToSlot(slot, state)
	if err != nil {
		return fmt.Errorf("error updating rolling record to slot %d, block %d: %w", state.BeaconSlotNumber, state.ElBlockNumber, err)
	}

	return nil
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

// Check if the node wallet has already submitted network balances for the given block number
func (r *RollingRecordManager) hasSubmittedNetworkBalances(blockNumber uint64) (bool, error) {
	// Ignore if there isn't a node address set
	if r.nodeAddress == nil {
		return false, nil
	}

	// Check if the address has submitted for the given block number
	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)
	hasSubmitted, err := r.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.balances.submitted.node"), r.nodeAddress.Bytes(), blockNumberBuf))
	if err != nil {
		return false, fmt.Errorf("error checking if node has already submitted network balance for block %d: %w", blockNumber, err)
	}

	return hasSubmitted, nil
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

	// Get the target slot number
	eth2Config := state.BeaconConfig
	totalTimespan := endTime.Sub(r.genesisTime)
	targetSlot := uint64(math.Ceil(totalTimespan.Seconds() / float64(eth2Config.SecondsPerSlot)))
	targetSlotEpoch := targetSlot / eth2Config.SlotsPerEpoch
	targetSlot = (targetSlotEpoch+1)*eth2Config.SlotsPerEpoch - 1 // The target slot becomes the last one in the Epoch

	return true, targetSlot, nil
}

// Check if the node wallet has already submitted rewards for the given interval
func (r *RollingRecordManager) hasSubmittedRewards(index uint64) (bool, error) {
	// Ignore if there isn't a node address set
	if r.nodeAddress == nil {
		return false, nil
	}

	// Check if the address has submitted for the given index
	return rprewards.GetTrustedNodeSubmitted(r.rp, *r.nodeAddress, index, nil)
}

// Get the actual slot to be used for a rewards interval submission instead of the naively-determined one
// NOTE: only call this once the required epoch (targetSlotEpoch + 1) has been finalized
func (r *RollingRecordManager) getTrueRewardsIntervalSubmissionSlot(targetSlot uint64) (uint64, uint64, error) {
	// Get the first successful block
	for {
		// Try to get the current block
		block, exists, err := r.bc.GetBeaconBlock(fmt.Sprint(targetSlot))
		if err != nil {
			return 0, 0, fmt.Errorf("error getting Beacon block %d: %w", targetSlot, err)
		}

		// If the block was missing, try the previous one
		if !exists {
			r.log.Printlnf("Slot %d was missing, trying the previous one...", targetSlot)
			targetSlot--
		} else {
			// Ok, we have the first proposed finalized block - this is the one to use for the snapshot!
			return targetSlot, block.ExecutionBlockNumber, nil
		}
	}
}

// Generate a new record for the provided slot using the latest viable saved record
func (r *RollingRecordManager) GenerateRecordForState(state *state.NetworkState) (*rewards.RollingRecord, error) {

	// Load the latest viable record
	slot := state.BeaconSlotNumber
	record, err := r.LoadBestRecordFromDisk(r.startSlot, slot)
	if err != nil {
		return nil, fmt.Errorf("error loading best record for slot %d: %w", slot, err)
	}

	if record.LastDutiesSlot == slot {
		// Already have a full snapshot so we don't have to do anything
		r.log.Printf("%s Loaded record was already up-to-date for slot %d.", r.logPrefix, slot)
		return record, nil
	} else if record.LastDutiesSlot > slot {
		// This should never happen but sanity check it anyway
		return nil, fmt.Errorf("loaded record has duties completed for slot %d, which is too far forward (targeting slot %d)", record.LastDutiesSlot, slot)
	}

	// Get the slot to start processing from and the target for the first round
	nextStartSlot := record.LastDutiesSlot + 1
	if record.LastDutiesSlot == 0 {
		nextStartSlot = r.startSlot
	}

	nextStartEpoch := nextStartSlot / r.beaconCfg.SlotsPerEpoch
	nextTargetEpoch := nextStartEpoch + recordCheckpointInterval - 1
	nextTargetSlot := (nextTargetEpoch+1)*r.beaconCfg.SlotsPerEpoch - 1 // Target is the last slot of the epoch
	if nextTargetSlot > slot {
		nextTargetSlot = slot
		nextTargetEpoch = nextTargetSlot / r.beaconCfg.SlotsPerEpoch
	}
	finalEpoch := slot / r.beaconCfg.SlotsPerEpoch
	totalSlots := float64(slot - nextStartSlot + 1)
	initialSlot := nextStartSlot

	r.log.Printlnf("%s Collecting records from slot %d (epoch %d) to slot %d (epoch %d).", r.logPrefix, nextStartSlot, nextStartEpoch, slot, finalEpoch)
	r.log.Printlnf("%s Progress will be reported at each saved checkpoint (%d epochs).", r.logPrefix, recordCheckpointInterval)
	startTime := time.Now()
	for {
		if nextStartSlot > slot {
			break
		}

		// Update the record to the target state
		err = record.UpdateToSlot(nextTargetSlot, state)
		if err != nil {
			return nil, fmt.Errorf("error updating record to slot %d: %w", slot, err)
		}

		err = r.SaveRecordToFile(record)
		if err != nil {
			return nil, fmt.Errorf("error saving record: %w", err)
		}

		slotsProcessed := nextTargetSlot - initialSlot + 1
		r.log.Printf("%s (%.2f%%) Updated from slot %d (epoch %d) to slot %d (epoch %d)... (%s so far) ", r.logPrefix, float64(slotsProcessed)/totalSlots*100.0, nextStartSlot, nextStartEpoch, nextTargetSlot, nextTargetEpoch, time.Since(startTime))

		nextStartSlot = nextTargetSlot + 1
		nextStartEpoch = nextStartSlot / r.beaconCfg.SlotsPerEpoch
		nextTargetEpoch = nextStartEpoch + recordCheckpointInterval - 1
		nextTargetSlot = (nextTargetEpoch+1)*r.beaconCfg.SlotsPerEpoch - 1 // Target is the last slot of the epoch
		if nextTargetSlot > slot {
			nextTargetSlot = slot
			nextTargetEpoch = nextTargetSlot / r.beaconCfg.SlotsPerEpoch
		}
	}

	r.log.Printlnf("%s Finished in %s.", r.logPrefix, time.Since(startTime))

	return record, nil

}

// Save the rolling record to a file and update the record info catalog
func (r *RollingRecordManager) SaveRecordToFile(record *rewards.RollingRecord) error {

	// Serialize the record
	bytes, err := record.Serialize()
	if err != nil {
		return fmt.Errorf("error saving rolling record: %w", err)
	}

	// Compress the record
	compressedBytes := r.compressor.EncodeAll(bytes, make([]byte, 0, len(bytes)))

	// Get the record filename
	slot := record.LastDutiesSlot
	epoch := record.LastDutiesSlot / r.beaconCfg.SlotsPerEpoch
	recordsPath := r.cfg.Smartnode.GetRecordsPath()
	filename := filepath.Join(recordsPath, fmt.Sprintf(recordsFilenameFormat, slot, epoch))

	// Write it to a file
	err = os.WriteFile(filename, compressedBytes, 0664)
	if err != nil {
		return fmt.Errorf("error writing file [%s]: %w", filename, err)
	}

	// Compute the SHA384 hash to act as a checksum
	checksum := sha512.Sum384(compressedBytes)

	// Load the existing checksum table
	_, lines, err := r.parseChecksumFile()
	if err != nil {
		return fmt.Errorf("error parsing checkpoint file: %w", err)
	}
	if lines == nil {
		lines = []string{}
	}

	// Add the new record checksum
	checksumLine := fmt.Sprintf("%s  %s", hex.EncodeToString(checksum[:]), filepath.Base(filename))
	lines = append(lines, checksumLine)

	// Sort the lines by their slot
	err = r.sortChecksumEntries(lines)
	if err != nil {
		return fmt.Errorf("error sorting checkpoint file entries: %w", err)
	}

	// Get the number of lines to write
	var newLines []string
	if len(lines) > int(checkpointRetentionLimit) {
		numberOfNewLines := int(checkpointRetentionLimit)
		cullCount := len(lines) - numberOfNewLines

		// Remove old lines and delete the corresponding files that shouldn't be retained
		for i := 0; i < cullCount; i++ {
			line := lines[i]

			// Extract the filename
			elems := strings.Split(line, "  ")
			if len(elems) != 2 {
				return fmt.Errorf("error parsing checkpoint line (%s): expected 2 elements, but got %d", line, len(elems))
			}
			filename := elems[1]
			fullFilename := filepath.Join(recordsPath, filename)

			// Delete the file if it exists
			_, err := os.Stat(fullFilename)
			if os.IsNotExist(err) {
				r.log.Printlnf("%s NOTE: tried removing checkpoint file [%s] based on the retention limit, but it didn't exist.", r.logPrefix, filename)
				continue
			}
			err = os.Remove(fullFilename)
			if err != nil {
				return fmt.Errorf("error deleting file [%s]: %w", fullFilename, err)
			}

			r.log.Printlnf("%s Removed checkpoint file [%s] based on the retention limit.", r.logPrefix, filename)
		}

		// Store the rest
		newLines = make([]string, numberOfNewLines)
		for i := cullCount; i <= numberOfNewLines; i++ {
			newLines[i-cullCount] = lines[i]
		}
	} else {
		newLines = lines
	}

	fileContents := strings.Join(newLines, "\n")
	checksumBytes := []byte(fileContents)

	// Save the new file
	checksumFilename := filepath.Join(recordsPath, checksumTableFilename)
	err = os.WriteFile(checksumFilename, checksumBytes, 0644)
	if err != nil {
		return fmt.Errorf("error writing checksum file after culling: %w", err)
	}

	return nil
}

// Load the most recent appropriate rolling record from disk, using the checksum table as an index
func (r *RollingRecordManager) LoadBestRecordFromDisk(startSlot uint64, targetSlot uint64) (*rewards.RollingRecord, error) {

	// Parse the checksum file
	exists, lines, err := r.parseChecksumFile()
	if err != nil {
		return nil, fmt.Errorf("error parsing checkpoint file: %w", err)
	}
	if !exists {
		// There isn't a checksum file so start over
		r.log.Printlnf("%s Checksum file not found, creating a new record from the start.", r.logPrefix)
		return rewards.NewRollingRecord(r.log, r.logPrefix, r.bc, startSlot, &r.beaconCfg), nil
	}

	// Iterate over each file, counting backwards from the bottom
	recordsPath := r.cfg.Smartnode.GetRecordsPath()
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		// Extract the checksum, filename, and slot number
		checksumString, filename, slot, err := r.parseChecksumEntry(line)
		if err != nil {
			return nil, err
		}

		// Check if the slot was too far into the future
		if slot > targetSlot {
			r.log.Printlnf("%s File [%s] was too far into the future, trying an older one...", r.logPrefix, filename)
			continue
		}

		// Check if it was too far into the past
		if slot < startSlot {
			r.log.Printlnf("%s File [%s] was too old (generated before the target start slot), none of the remaining records can be used.", r.logPrefix, filename)
			break
		}

		// Make sure the checksum parses properly
		checksum, err := hex.DecodeString(checksumString)
		if err != nil {
			return nil, fmt.Errorf("error scanning checkpoint line (%s): checksum (%s) could not be parsed", line, checksumString)
		}

		// Try to load it
		fullFilename := filepath.Join(recordsPath, filename)
		record, err := r.loadRecordFromFile(fullFilename, checksum)
		if err != nil {
			r.log.Printlnf("%s WARNING: error loading record from file [%s]: %s... attempting previous file", r.logPrefix, fullFilename, err.Error())
			continue
		}
		epoch := slot / r.beaconCfg.SlotsPerEpoch
		r.log.Printlnf("%s Loaded file [%s] which ended on slot %d (epoch %d).", r.logPrefix, filename, slot, epoch)
		return record, nil

	}

	// If we got here then none of the saved files worked so we have to make a new record
	r.log.Printlnf("%s None of the saved record checkpoint files were eligible for use, creating a new record from the start.", r.logPrefix)
	return rewards.NewRollingRecord(r.log, r.logPrefix, r.bc, startSlot, &r.beaconCfg), nil

}

// Get the slot number from a record filename
func (r *RollingRecordManager) getSlotFromFilename(filename string) (uint64, error) {
	matches := r.recordsFilenameRegex.FindStringSubmatch(filename)
	if matches == nil {
		return 0, fmt.Errorf("filename (%s) did not match the expected format", filename)
	}
	slotIndex := r.recordsFilenameRegex.SubexpIndex("slot")
	if slotIndex == -1 {
		return 0, fmt.Errorf("slot number not found in filename (%s)", filename)
	}
	slotString := matches[slotIndex]
	slot, err := strconv.ParseUint(slotString, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("slot (%s) could not be parsed to a number")
	}

	return slot, nil
}

// Load a record from a file, making sure its contents match the provided checksum
func (r *RollingRecordManager) loadRecordFromFile(filename string, expectedChecksum []byte) (*rewards.RollingRecord, error) {
	// Read the file
	compressedBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Calculate the hash and validate it
	checksum := sha512.Sum384(compressedBytes)
	if !bytes.Equal(expectedChecksum, checksum[:]) {
		expectedString := hex.EncodeToString(expectedChecksum)
		actualString := hex.EncodeToString(checksum[:])
		return nil, fmt.Errorf("checksum mismatch (expected %s, but it was %s)", expectedString, actualString)
	}

	// Decompress it
	bytes, err := r.decompressor.DecodeAll(compressedBytes, []byte{})
	if err != nil {
		return nil, fmt.Errorf("error decompressing data: %w", err)
	}

	// Create a new record from the data
	return rewards.DeserializeRollingRecord(r.log, r.logPrefix, r.bc, &r.beaconCfg, bytes)
}

// Get the lines from the checksum file
func (r *RollingRecordManager) parseChecksumFile() (bool, []string, error) {
	// Get the checksum filename
	recordsPath := r.cfg.Smartnode.GetRecordsPath()
	checksumFilename := filepath.Join(recordsPath, checksumTableFilename)

	// Check if the file exists
	_, err := os.Stat(checksumFilename)
	if os.IsNotExist(err) {
		return false, nil, nil
	}

	// Open the checksum file
	checksumTable, err := os.ReadFile(checksumFilename)
	if err != nil {
		return false, nil, fmt.Errorf("error loading checksum table (%s): %w", checksumFilename, err)
	}

	// Parse out each line
	originalLines := strings.Split(string(checksumTable), "\n")

	// Remove empty lines
	lines := make([]string, 0, len(originalLines))
	for _, line := range originalLines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			lines = append(lines, line)
		}
	}

	return true, lines, nil
}

// Sort the checksum file entries by their slot
func (r *RollingRecordManager) sortChecksumEntries(lines []string) error {
	var sortErr error
	sort.Slice(lines, func(i int, j int) bool {
		_, _, firstSlot, err := r.parseChecksumEntry(lines[i])
		if err != nil && sortErr == nil {
			sortErr = err
			return false
		}

		_, _, secondSlot, err := r.parseChecksumEntry(lines[j])
		if err != nil && sortErr == nil {
			sortErr = err
			return false
		}

		return firstSlot < secondSlot
	})
	return sortErr
}

// Get the checksum, the filename, and the slot number from a checksum entry.
func (r *RollingRecordManager) parseChecksumEntry(line string) (string, string, uint64, error) {
	// Extract the checksum and filename
	elems := strings.Split(line, "  ")
	if len(elems) != 2 {
		return "", "", 0, fmt.Errorf("error parsing checkpoint line (%s): expected 2 elements, but got %d", line, len(elems))
	}
	checksumString := elems[0]
	filename := elems[1]

	// Extract the slot number for this file
	slot, err := r.getSlotFromFilename(filename)
	if err != nil {
		return "", "", 0, fmt.Errorf("error scanning checkpoint line (%s): %w", line, err)
	}

	return checksumString, filename, slot, nil
}
