package watchtower

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	nmc_utils "github.com/rocket-pool/node-manager-core/utils"
	"github.com/rocket-pool/rocketpool-go/v2/rewards"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/tokens"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/eth1"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/gas"
	rprewards "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/rewards"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/watchtower/utils"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
	sharedtypes "github.com/rocket-pool/smartnode/v2/shared/types"
)

// Process balances and rewards task
type SubmitRewardsTree_Rolling struct {
	ctx         context.Context
	sp          *services.ServiceProvider
	logger      *slog.Logger
	cfg         *config.SmartNodeConfig
	w           *wallet.Wallet
	ec          eth.IExecutionClient
	rp          *rocketpool.RocketPool
	bc          beacon.IBeaconClient
	rewardsPool *rewards.RewardsPool
	genesisTime time.Time
	recordMgr   *rprewards.RollingRecordManager
	stateMgr    *state.NetworkStateManager

	lock      *sync.Mutex
	isRunning bool
}

// Create submit rewards tree with rolling record support
func NewSubmitRewardsTree_Rolling(ctx context.Context, sp *services.ServiceProvider, logger *log.Logger, stateMgr *state.NetworkStateManager) (*SubmitRewardsTree_Rolling, error) {
	// Get services
	cfg := sp.GetConfig()
	rp := sp.GetRocketPool()
	bc := sp.GetBeaconClient()
	rewardsPool, err := rewards.NewRewardsPool(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating Rewards Pool binding: %w", err)
	}

	// Get the beacon config
	beaconCfg, err := bc.GetEth2Config(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting beacon config: %w", err)
	}

	// Get the Beacon genesis time
	genesisTime := time.Unix(int64(beaconCfg.GenesisTime), 0)

	// Get the current interval index
	err = rp.Query(nil, nil, rewardsPool.RewardIndex)
	if err != nil {
		return nil, fmt.Errorf("error getting rewards index: %w", err)
	}
	currentIndex := rewardsPool.RewardIndex.Formatted()
	if currentIndex == 0 {
		return nil, fmt.Errorf("rolling records cannot be used for the first rewards interval")
	}

	// Get the previous RocketRewardsPool addresses
	rs := cfg.GetRocketPoolResources()
	prevAddresses := rs.PreviousRewardsPoolAddresses

	// Get the last rewards event and starting epoch
	found, event, err := rewardsPool.GetRewardsEvent(rp, currentIndex-1, prevAddresses, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting event for rewards interval %d: %w", currentIndex-1, err)
	}
	if !found {
		return nil, fmt.Errorf("event for rewards interval %d not found", currentIndex-1)
	}

	// Get the start slot of the current interval
	startSlot, err := rprewards.GetStartSlotForInterval(ctx, event, bc, beaconCfg)
	if err != nil {
		return nil, fmt.Errorf("error getting start slot for interval %d: %w", currentIndex, err)
	}

	// Create the task
	lock := &sync.Mutex{}
	task := &SubmitRewardsTree_Rolling{
		ctx:         ctx,
		sp:          sp,
		logger:      logger.With(),
		cfg:         cfg,
		w:           sp.GetWallet(),
		ec:          sp.GetEthClient(),
		rp:          sp.GetRocketPool(),
		bc:          bc,
		stateMgr:    stateMgr,
		genesisTime: genesisTime,
		lock:        lock,
		isRunning:   false,
	}

	// Make a new rolling manager
	recordMgr, err := rprewards.NewRollingRecordManager(task.logger, cfg, rp, bc, stateMgr, startSlot, beaconCfg, currentIndex)
	if err != nil {
		return nil, fmt.Errorf("error creating rolling record manager: %w", err)
	}

	// Load the latest checkpoint
	beaconHead, err := bc.GetBeaconHead(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting beacon head: %w", err)
	}
	latestFinalizedSlot := (beaconHead.FinalizedEpoch+1)*beaconCfg.SlotsPerEpoch - 1
	_, err = recordMgr.LoadBestRecordFromDisk(startSlot, latestFinalizedSlot, currentIndex)
	if err != nil {
		return nil, fmt.Errorf("error loading rolling record checkpoint from disk: %w", err)
	}

	// Return
	task.recordMgr = recordMgr
	return task, nil
}

// Update the rolling record and run the submission process if applicable
func (t *SubmitRewardsTree_Rolling) Run(headState *state.NetworkState) error {
	t.lock.Lock()
	if t.isRunning {
		t.logger.Info("Record update is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()
		t.logger.Info("Running record update in a separate thread.")

		// Update contract bindings
		nodeAddress, _ := t.w.GetAddress()
		var err error
		t.rewardsPool, err = rewards.NewRewardsPool(t.rp)
		if err != nil {
			t.handleError(fmt.Errorf("error creating Rewards Pool binding: %w", err))
			return
		}

		// Capture the latest head state if one isn't passed in
		if headState == nil {
			// Get the latest Beacon block
			latestBlock, err := t.stateMgr.GetLatestBeaconBlock(t.ctx)
			if err != nil {
				t.handleError(fmt.Errorf("error getting latest Beacon block: %w", err))
				return
			}

			// Get the state of the network
			headState, err = t.stateMgr.GetStateForSlot(t.ctx, latestBlock.Header.Slot)
			if err != nil {
				t.handleError(fmt.Errorf("error getting network state: %w", err))
				return
			}
		}

		// Check whether or not the node is in the Oracle DAO
		isInOdao := false
		for _, details := range headState.OracleDaoMemberDetails {
			if details.Address == nodeAddress {
				isInOdao = true
				break
			}
		}

		// Get the latest finalized slot and epoch
		latestFinalizedBlock, err := t.stateMgr.GetLatestFinalizedBeaconBlock(t.ctx)
		if err != nil {
			t.handleError(fmt.Errorf("error getting latest finalized block: %w", err))
			return
		}
		latestFinalizedEpoch := latestFinalizedBlock.Header.Slot / headState.BeaconConfig.SlotsPerEpoch

		// Check if a rewards interval is due
		isRewardsSubmissionDue, rewardsSlot, intervalsPassed, startTime, endTime, err := t.isRewardsIntervalSubmissionRequired(headState)
		if err != nil {
			t.handleError(fmt.Errorf("error checking if rewards submission is required: %w", err))
			return
		}

		// If no special upcoming state is required, update normally
		if !isRewardsSubmissionDue {
			err = t.recordMgr.UpdateRecordToState(t.ctx, headState, latestFinalizedBlock.Header.Slot)
			if err != nil {
				t.handleError(fmt.Errorf("error updating record: %w", err))
				return
			}

			t.lock.Lock()
			t.isRunning = false
			t.lock.Unlock()
			return
		}

		// Check if rewards reporting is ready
		rewardsEpoch := rewardsSlot / headState.BeaconConfig.SlotsPerEpoch
		requiredRewardsEpoch := rewardsEpoch + 1
		isRewardsReadyForReport := isRewardsSubmissionDue && (latestFinalizedEpoch >= requiredRewardsEpoch)

		// Run updates and submissions as required
		if isRewardsReadyForReport {
			// Check if there's an existing file for this interval, and try submitting that
			existingRewardsFile, valid, mustRegenerate := t.isExistingRewardsFileValid(headState.NetworkDetails.RewardIndex, intervalsPassed, nodeAddress, isInOdao)
			if existingRewardsFile != nil {
				if valid && !mustRegenerate {
					// We already have a valid file and submission
					t.logger.Info("Rewards tree has already been submitted and is still valid but consensus hasn't been reached yet; nothing to do.", slog.Uint64(keys.IntervalKey, headState.NetworkDetails.RewardIndex))
					t.lock.Lock()
					t.isRunning = false
					t.lock.Unlock()
					return
				} else if !valid && !mustRegenerate {
					// We already have a valid file but need to submit again
					t.logger.Info("Rewards tree has already been created but hasn't been submitted yet, attempting resubmission.", slog.Uint64(keys.IntervalKey, headState.NetworkDetails.RewardIndex))
				} else if !valid && mustRegenerate {
					// We have a file but it's not valid (probably because too many intervals have passed)
					t.logger.Info("Rewards submission is due and current file is no longer valid (likely too many intervals have passed since its creation), regenerating it.", slog.Uint64(keys.IntervalKey, headState.NetworkDetails.RewardIndex))
				}
			}

			// Get the actual slot to report on
			var elBlockNumber uint64
			rewardsSlot, elBlockNumber, err = t.getTrueRewardsIntervalSubmissionSlot(rewardsSlot)
			if err != nil {
				t.handleError(fmt.Errorf("error getting the true rewards interval slot: %w", err))
				return
			}

			// Get an appropriate client that has access to the target state - this is required if the state gets pruned by the local EC and the
			// archive EC is required
			client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.logger, big.NewInt(0).SetUint64(elBlockNumber))
			if err != nil {
				t.handleError(fmt.Errorf("error getting best API client during rewards submission: %w", err))
				return
			}

			// Generate the rewards state
			stateMgr, err := state.NewNetworkStateManager(t.ctx, client, t.cfg, client.Client, t.bc, t.logger)
			if err != nil {
				t.handleError(fmt.Errorf("error creating state manager for rewards slot: %w", err))
				return
			}
			state, err := stateMgr.GetStateForSlot(t.ctx, rewardsSlot)
			if err != nil {
				t.handleError(fmt.Errorf("error getting state for rewards slot: %w", err))
				return
			}

			// Process the rewards interval
			t.logger.Info("Running rewards interval submission.")
			err = t.runRewardsIntervalReport(client, state, isInOdao, intervalsPassed, startTime, endTime, mustRegenerate, existingRewardsFile)
			if err != nil {
				t.handleError(fmt.Errorf("error running rewards interval report: %w", err))
				return
			}
		} else {
			t.logger.Info("Rewards submission is due... waiting for target epoch to be finalized.", slog.Uint64(keys.IntervalKey, headState.NetworkDetails.RewardIndex), slog.Uint64(keys.TargetEpochKey, requiredRewardsEpoch), slog.Uint64(keys.FinalizedEpochKey, latestFinalizedEpoch))
		}

		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

	return nil
}

// Check if a rewards interval submission is required and if so, the slot number for the update
func (t *SubmitRewardsTree_Rolling) isRewardsIntervalSubmissionRequired(state *state.NetworkState) (bool, uint64, uint64, time.Time, time.Time, error) {
	// Check if a rewards interval has passed and needs to be calculated
	startTime := state.NetworkDetails.IntervalStart
	intervalTime := state.NetworkDetails.IntervalDuration

	// Adjust for the first interval by making the start time the RPL inflation interval start time
	if startTime == time.Unix(0, 0) {
		opts := &bind.CallOpts{
			BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
		}
		rpl, err := tokens.NewTokenRpl(t.rp)
		if err != nil {
			return false, 0, 0, time.Time{}, time.Time{}, fmt.Errorf("error creating RPL token binding: %w", err)
		}
		err = t.rp.Query(nil, opts, rpl.InflationIntervalStartTime)
		if err != nil {
			return false, 0, 0, time.Time{}, time.Time{}, fmt.Errorf("start time is zero, but error getting RPL token inflation interval start time: %w", err)
		}
		startTime = rpl.InflationIntervalStartTime.Formatted()
		t.logger.Info("NOTE: rewards pool interval start time is 0, using the inflation interval start time according to the RPL token.", slog.Time(keys.StartKey, startTime))
	}

	// Calculate the end time, which is the number of intervals that have gone by since the current one's start
	secondsSinceGenesis := time.Duration(state.BeaconConfig.SecondsPerSlot*state.BeaconSlotNumber) * time.Second
	stateTime := t.genesisTime.Add(secondsSinceGenesis)
	timeSinceStart := stateTime.Sub(startTime)
	intervalsPassed := timeSinceStart / intervalTime
	endTime := startTime.Add(intervalTime * intervalsPassed)
	if intervalsPassed == 0 {
		return false, 0, 0, time.Time{}, time.Time{}, nil
	}

	// Get the target slot number
	eth2Config := state.BeaconConfig
	totalTimespan := endTime.Sub(t.genesisTime)
	targetSlot := uint64(math.Ceil(totalTimespan.Seconds() / float64(eth2Config.SecondsPerSlot)))
	targetSlotEpoch := targetSlot / eth2Config.SlotsPerEpoch
	targetSlot = (targetSlotEpoch+1)*eth2Config.SlotsPerEpoch - 1 // The target slot becomes the last one in the Epoch

	return true, targetSlot, uint64(intervalsPassed), startTime, endTime, nil
}

// Get the actual slot to be used for a rewards interval submission instead of the naively-determined one
// NOTE: only call this once the required epoch (targetSlotEpoch + 1) has been finalized
func (t *SubmitRewardsTree_Rolling) getTrueRewardsIntervalSubmissionSlot(targetSlot uint64) (uint64, uint64, error) {
	// Get the first successful block
	for {
		// Try to get the current block
		block, exists, err := t.bc.GetBeaconBlock(t.ctx, fmt.Sprint(targetSlot))
		if err != nil {
			return 0, 0, fmt.Errorf("error getting Beacon block %d: %w", targetSlot, err)
		}

		// If the block was missing, try the previous one
		if !exists {
			t.logger.Info("Slot was missing, trying the previous one...", slog.Uint64(keys.SlotKey, targetSlot))
			targetSlot--
		} else {
			// Ok, we have the first proposed finalized block - this is the one to use for the snapshot!
			return targetSlot, block.ExecutionBlockNumber, nil
		}
	}
}

// Checks to see if an existing rewards file is still valid and whether or not it should be regenerated or just resubmitted
func (t *SubmitRewardsTree_Rolling) isExistingRewardsFileValid(rewardIndex uint64, intervalsPassed uint64, nodeAddress common.Address, isInOdao bool) (*rprewards.LocalRewardsFile, bool, bool) {
	rewardsTreePath := t.cfg.GetRewardsTreePath(rewardIndex)

	// Check if the rewards file exists
	_, err := os.Stat(rewardsTreePath)
	if os.IsNotExist(err) {
		return nil, false, true
	}
	if err != nil {
		t.logger.Warn("Failed to check if rewards file exists; regenerating file...\n", slog.String(keys.FileKey, rewardsTreePath), log.Err(err))
		return nil, false, true
	}

	// The file already exists, attempt to read it
	localRewardsFile, err := rprewards.ReadLocalRewardsFile(rewardsTreePath)
	if err != nil {
		t.logger.Warn("Failed to read rewards file; regenerating file...\n", slog.String(keys.FileKey, rewardsTreePath), log.Err(err))
		return nil, false, true
	}

	proofWrapper := localRewardsFile.Impl()
	header := proofWrapper.GetHeader()

	if isInOdao {
		// Save the compressed file and get the CID for it
		cid, err := localRewardsFile.CreateCompressedFileAndCid()
		if err != nil {
			t.logger.Warn("Failed to get CID for rewards file; regenerating file...\n", slog.String(keys.FileKey, rewardsTreePath), log.Err(err))
			return nil, false, true
		}

		// Check if this file has already been submitted
		submission := rewards.RewardSubmission{
			RewardIndex:     big.NewInt(0).SetUint64(header.Index),
			ExecutionBlock:  big.NewInt(0).SetUint64(header.ExecutionEndBlock),
			ConsensusBlock:  big.NewInt(0).SetUint64(header.ConsensusEndBlock),
			MerkleRoot:      common.HexToHash(header.MerkleRoot),
			MerkleTreeCID:   cid.String(),
			IntervalsPassed: big.NewInt(0).SetUint64(header.IntervalsPassed),
			TreasuryRPL:     &header.TotalRewards.ProtocolDaoRpl.Int,
			TrustedNodeRPL:  []*big.Int{&header.TotalRewards.TotalOracleDaoRpl.Int},
			NodeRPL:         []*big.Int{&header.TotalRewards.TotalCollateralRpl.Int},
			NodeETH:         []*big.Int{&header.TotalRewards.NodeOperatorSmoothingPoolEth.Int},
			UserETH:         &header.TotalRewards.PoolStakerSmoothingPoolEth.Int,
		}

		var hasSubmitted bool
		err = t.rp.Query(func(mc *batch.MultiCaller) error {
			t.rewardsPool.GetTrustedNodeSubmittedSpecificRewards(mc, &hasSubmitted, nodeAddress, submission)
			return nil
		}, nil)
		if err != nil {
			t.logger.Warn("Could not check if node has previously submitted rewards file; regenerating file...\n", slog.String(keys.FileKey, rewardsTreePath), log.Err(err))
			return nil, false, true
		}
		if !hasSubmitted {
			if header.IntervalsPassed != intervalsPassed {
				t.logger.Info("Existing file has too few rounds, regenerating file...", slog.Uint64(keys.IntervalKey, header.Index), slog.Uint64(keys.FileRoundsKey, header.IntervalsPassed), slog.Uint64(keys.ActualRoundsKey, intervalsPassed))
				return localRewardsFile, false, true
			}
			t.logger.Info("Existing file has not been submitted yet.", slog.Uint64(keys.IntervalKey, header.Index))
			return localRewardsFile, false, false
		}
	}

	// Check if the file's valid (same number of intervals passed as the current time)
	if header.IntervalsPassed != intervalsPassed {
		t.logger.Info("Existing file has too few rounds, regenerating file...", slog.Uint64(keys.IntervalKey, header.Index), slog.Uint64(keys.FileRoundsKey, header.IntervalsPassed), slog.Uint64(keys.ActualRoundsKey, intervalsPassed))
		return localRewardsFile, false, true
	}

	// File's good and it has the same number of intervals passed, so use it
	return localRewardsFile, true, false
}

// Run a rewards interval report submission
func (t *SubmitRewardsTree_Rolling) runRewardsIntervalReport(client *rocketpool.RocketPool, state *state.NetworkState, isInOdao bool, intervalsPassed uint64, startTime time.Time, endTime time.Time, mustRegenerate bool, existingRewardsFile *rprewards.LocalRewardsFile) error {
	// Prep the record for reporting
	err := t.recordMgr.PrepareRecordForReport(t.ctx, state)
	if err != nil {
		return fmt.Errorf("error preparing record for report: %w", err)
	}

	// Initialize some variables
	snapshotBeaconBlock := state.BeaconSlotNumber
	elBlockNumber := state.ElBlockNumber

	// Get the number of the EL block matching the CL snapshot block
	snapshotElBlockHeader, err := t.rp.Client.HeaderByNumber(context.Background(), big.NewInt(int64(elBlockNumber)))
	if err != nil {
		return err
	}
	elBlockIndex := snapshotElBlockHeader.Number.Uint64()

	// Get the current interval
	currentIndex := state.NetworkDetails.RewardIndex
	currentIndexBig := big.NewInt(0).SetUint64(currentIndex)

	// Get the expected file paths
	rewardsTreePath := t.cfg.GetRewardsTreePath(currentIndex)
	compressedRewardsTreePath := rewardsTreePath + config.RewardsTreeIpfsExtension
	minipoolPerformancePath := t.cfg.GetMinipoolPerformancePath(currentIndex)
	compressedMinipoolPerformancePath := minipoolPerformancePath + config.RewardsTreeIpfsExtension

	// Check if we can reuse an existing file for this interval
	if !mustRegenerate {
		if !isInOdao {
			t.logger.Info("Node is not in the Oracle DAO, skipping submission..", slog.Uint64(keys.IntervalKey, currentIndex))
			return nil
		}

		t.logger.Info("Merkle rewards tree already exists, attempting to resubmit...", slog.Uint64(keys.IntervalKey, currentIndex), slog.String(keys.FileKey, rewardsTreePath))

		// Save the compressed file and get the CID for it
		cid, err := existingRewardsFile.CreateCompressedFileAndCid()
		if err != nil {
			return fmt.Errorf("error getting CID for file %s: %w", compressedRewardsTreePath, err)
		}
		t.logger.Info("Calculated rewards tree CID", slog.String(keys.CidKey, cid.String()))

		// Submit to the contracts
		err = t.submitRewardsSnapshot(currentIndexBig, snapshotBeaconBlock, elBlockIndex, existingRewardsFile.Impl().GetHeader(), cid.String(), big.NewInt(int64(intervalsPassed)))
		if err != nil {
			return fmt.Errorf("error submitting rewards snapshot: %w", err)
		}

		t.logger.Info("Successfully submitted rewards snapshot.", slog.Uint64(keys.IntervalKey, currentIndex))
		return nil
	}

	// Generate the tree
	err = t.generateTree(client, state, intervalsPassed, isInOdao, currentIndex, snapshotBeaconBlock, elBlockIndex, startTime, endTime, snapshotElBlockHeader, rewardsTreePath, compressedRewardsTreePath, minipoolPerformancePath, compressedMinipoolPerformancePath)
	if err != nil {
		return fmt.Errorf("error generating rewards tree: %w", err)
	}

	return nil
}

// Implementation for rewards tree generation using a viable EC
func (t *SubmitRewardsTree_Rolling) generateTree(rp *rocketpool.RocketPool, state *state.NetworkState, intervalsPassed uint64, nodeTrusted bool, currentIndex uint64, snapshotBeaconBlock uint64, elBlockIndex uint64, startTime time.Time, endTime time.Time, snapshotElBlockHeader *types.Header, rewardsTreePath string, compressedRewardsTreePath string, minipoolPerformancePath string, compressedMinipoolPerformancePath string) error {
	// Log
	if intervalsPassed > 1 {
		t.logger.Warn("Multiple intervals have passed since the last rewards checkpoint was submitted! Rolling them into one...", slog.Uint64(keys.RoundsKey, intervalsPassed))
	}
	t.logger.Info("Rewards checkpoint has passed, starting Merkle tree generation in the background.", slog.Uint64(keys.IntervalKey, currentIndex), slog.Uint64(keys.SlotKey, snapshotBeaconBlock), slog.Uint64(keys.BlockKey, elBlockIndex), slog.Time(keys.StartKey, startTime), slog.Time(keys.EndKey, endTime))

	// Generate the rewards file
	treegen, err := rprewards.NewTreeGenerator(t.logger, rp, t.cfg, t.bc, currentIndex, startTime, endTime, snapshotBeaconBlock, snapshotElBlockHeader, uint64(intervalsPassed), state, t.recordMgr.Record)
	if err != nil {
		return fmt.Errorf("error creating Merkle tree generator: %w", err)
	}
	rewardsFile, err := treegen.GenerateTree(t.ctx)
	if err != nil {
		return fmt.Errorf("error generating Merkle tree: %w", err)
	}
	for address, network := range rewardsFile.GetHeader().InvalidNetworkNodes {
		t.logger.Warn("Node has invalid network assigned! Using 0 (mainnet) instead.", slog.String(keys.NodeKey, address.Hex()), slog.Uint64(keys.NetworkKey, network))
	}

	// Serialize the minipool performance file
	localMinipoolPerformanceFile := rprewards.NewLocalFile[sharedtypes.IMinipoolPerformanceFile](
		rewardsFile.GetMinipoolPerformanceFile(),
		minipoolPerformancePath,
	)
	err = localMinipoolPerformanceFile.Write()
	if err != nil {
		return fmt.Errorf("error serializing minipool performance file into JSON: %w", err)
	}

	if nodeTrusted {
		minipoolPerformanceCid, err := localMinipoolPerformanceFile.CreateCompressedFileAndCid()
		if err != nil {
			return fmt.Errorf("error getting the CID for file %s: %w", compressedMinipoolPerformancePath, err)
		}
		t.logger.Info("Calculated minipool performance CID", slog.String(keys.CidKey, minipoolPerformanceCid.String()))
		rewardsFile.SetMinipoolPerformanceFileCID(minipoolPerformanceCid.String())
	} else {
		t.logger.Info("Saved minipool performance file.")
		rewardsFile.SetMinipoolPerformanceFileCID("---")
	}

	// Serialize the rewards tree to JSON
	localRewardsFile := rprewards.NewLocalFile[sharedtypes.IRewardsFile](
		rewardsFile,
		rewardsTreePath,
	)
	t.logger.Info("Generation complete! Saving tree...")

	// Write the rewards tree to disk
	err = localRewardsFile.Write()
	if err != nil {
		return fmt.Errorf("error saving rewards tree file to %s: %w", rewardsTreePath, err)
	}

	if nodeTrusted {
		cid, err := localRewardsFile.CreateCompressedFileAndCid()
		if err != nil {
			return fmt.Errorf("error getting CID for file %s: %w", compressedRewardsTreePath, err)
		}
		t.logger.Info("Calculated rewards tree CID", slog.String(keys.CidKey, cid.String()))
		// Submit to the contracts
		err = t.submitRewardsSnapshot(big.NewInt(int64(currentIndex)), snapshotBeaconBlock, elBlockIndex, rewardsFile.GetHeader(), cid.String(), big.NewInt(int64(intervalsPassed)))
		if err != nil {
			return fmt.Errorf("error submitting rewards snapshot: %w", err)
		}

		t.logger.Info("Successfully submitted rewards snapshot.", slog.Uint64(keys.IntervalKey, currentIndex))
	} else {
		t.logger.Info("Successfully generated rewards snapshot.", slog.Uint64(keys.IntervalKey, currentIndex))
	}

	return nil
}

// Submit rewards info to the contracts
func (t *SubmitRewardsTree_Rolling) submitRewardsSnapshot(index *big.Int, consensusBlock uint64, executionBlock uint64, rewardsFileHeader *sharedtypes.RewardsFileHeader, cid string, intervalsPassed *big.Int) error {
	treeRootBytes, err := hex.DecodeString(nmc_utils.RemovePrefix(rewardsFileHeader.MerkleRoot))
	if err != nil {
		return fmt.Errorf("error decoding merkle root: %w", err)
	}
	treeRoot := common.BytesToHash(treeRootBytes)

	// Create the arrays of rewards per network
	collateralRplRewards := []*big.Int{}
	oDaoRplRewards := []*big.Int{}
	smoothingPoolEthRewards := []*big.Int{}

	// Create the total rewards for each network
	network := uint64(0)
	for {
		networkRewards, exists := rewardsFileHeader.NetworkRewards[network]
		if !exists {
			break
		}

		collateralRplRewards = append(collateralRplRewards, &networkRewards.CollateralRpl.Int)
		oDaoRplRewards = append(oDaoRplRewards, &networkRewards.OracleDaoRpl.Int)
		smoothingPoolEthRewards = append(smoothingPoolEthRewards, &networkRewards.SmoothingPoolEth.Int)

		network++
	}

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return err
	}

	// Create the submission
	submission := rewards.RewardSubmission{
		RewardIndex:     index,
		ExecutionBlock:  big.NewInt(0).SetUint64(executionBlock),
		ConsensusBlock:  big.NewInt(0).SetUint64(consensusBlock),
		MerkleRoot:      treeRoot,
		MerkleTreeCID:   cid,
		IntervalsPassed: intervalsPassed,
		TreasuryRPL:     &rewardsFileHeader.TotalRewards.ProtocolDaoRpl.Int,
		NodeRPL:         collateralRplRewards,
		TrustedNodeRPL:  oDaoRplRewards,
		NodeETH:         smoothingPoolEthRewards,
		UserETH:         &rewardsFileHeader.TotalRewards.PoolStakerSmoothingPoolEth.Int,
	}

	// Get the gas limit
	txInfo, err := t.rewardsPool.SubmitRewardSnapshot(submission, opts)
	if err != nil {
		if enableSubmissionAfterConsensus_RewardsTree && strings.Contains(err.Error(), "Can only submit snapshot for next period") {
			// Set a gas limit which will intentionally be too low and revert
			txInfo.SimulationResult = eth.SimulationResult{
				EstimatedGasLimit: utils.RewardsSubmissionForcedGas,
				SafeGasLimit:      utils.RewardsSubmissionForcedGas,
			}
			t.logger.Info("Rewards period consensus has already been reached but submitting anyway for the health check.")
		} else {
			return fmt.Errorf("error getting TX for submitting the rewards tree: %w", err)
		}
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating TX for submitting the rewards tree failed: %s", txInfo.SimulationResult.SimulationError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, t.logger, maxFee, 0) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, t.logger, txInfo, opts)
	if err != nil {
		return err
	}

	// Return
	return nil
}

// Print an error and unlock the mutex
func (t *SubmitRewardsTree_Rolling) handleError(err error) {
	t.logger.Error("*** Rolling Record processing failed. ***", log.Err(err))
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}
