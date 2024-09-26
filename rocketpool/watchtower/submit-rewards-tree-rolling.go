package watchtower

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)

// Process balances and rewards task
type submitRewardsTree_Rolling struct {
	c           *cli.Context
	log         log.ColorLogger
	errLog      log.ColorLogger
	cfg         *config.RocketPoolConfig
	w           *wallet.Wallet
	ec          rocketpool.ExecutionClient
	rp          *rocketpool.RocketPool
	bc          beacon.Client
	genesisTime time.Time
	recordMgr   *rprewards.RollingRecordManager
	stateMgr    *state.NetworkStateManager
	logPrefix   string

	lock      *sync.Mutex
	isRunning bool
}

// Create submit rewards tree with rolling record support
func newSubmitRewardsTree_Rolling(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger, stateMgr *state.NetworkStateManager) (*submitRewardsTree_Rolling, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Get the beacon config
	beaconCfg, err := bc.GetEth2Config()
	if err != nil {
		return nil, fmt.Errorf("error getting beacon config: %w", err)
	}

	// Get the Beacon genesis time
	genesisTime := time.Unix(int64(beaconCfg.GenesisTime), 0)

	// Get the current interval index
	currentIndexBig, err := rewards.GetRewardIndex(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting rewards index: %w", err)
	}
	currentIndex := currentIndexBig.Uint64()
	if currentIndex == 0 {
		return nil, fmt.Errorf("rolling records cannot be used for the first rewards interval")
	}

	// Get the previous RocketRewardsPool addresses
	prevAddresses := cfg.Smartnode.GetPreviousRewardsPoolAddresses()

	// Get the last rewards event and starting epoch
	found, event, err := rewards.GetRewardsEvent(rp, currentIndex-1, prevAddresses, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting event for rewards interval %d: %w", currentIndex-1, err)
	}
	if !found {
		return nil, fmt.Errorf("event for rewards interval %d not found", currentIndex-1)
	}

	// Get the start slot of the current interval
	startSlot, err := rprewards.GetStartSlotForInterval(event, bc, beaconCfg)
	if err != nil {
		return nil, fmt.Errorf("error getting start slot for interval %d: %w", currentIndex, err)
	}

	// Create the task
	lock := &sync.Mutex{}
	logPrefix := "[Rolling Record]"
	task := &submitRewardsTree_Rolling{
		c:           c,
		log:         logger,
		errLog:      errorLogger,
		cfg:         cfg,
		ec:          ec,
		w:           w,
		rp:          rp,
		bc:          bc,
		stateMgr:    stateMgr,
		genesisTime: genesisTime,
		logPrefix:   logPrefix,
		lock:        lock,
		isRunning:   false,
	}

	// Make a new rolling manager
	recordMgr, err := rprewards.NewRollingRecordManager(&task.log, &task.errLog, cfg, rp, bc, stateMgr, startSlot, beaconCfg, currentIndex)
	if err != nil {
		return nil, fmt.Errorf("error creating rolling record manager: %w", err)
	}

	// Load the latest checkpoint
	beaconHead, err := bc.GetBeaconHead()
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
func (t *submitRewardsTree_Rolling) run(headState *state.NetworkState) error {
	// Wait for clients to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	if err := services.WaitBeaconClientSynced(t.c, true); err != nil {
		return err
	}

	t.lock.Lock()
	if t.isRunning {
		t.log.Println("Record update is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return fmt.Errorf("error loading node account: %w", err)
	}
	nodeAddress := nodeAccount.Address

	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()
		t.log.Printlnf("%s Running record update in a separate thread.", t.logPrefix)

		// Capture the latest head state if one isn't passed in
		if headState == nil {
			// Get the latest Beacon block
			latestBlock, err := t.stateMgr.GetLatestBeaconBlock()
			if err != nil {
				t.handleError(fmt.Errorf("error getting latest Beacon block: %w", err))
				return
			}

			// Get the state of the network
			headState, err = t.stateMgr.GetStateForSlot(latestBlock.Slot)
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
		latestFinalizedBlock, err := t.stateMgr.GetLatestFinalizedBeaconBlock()
		if err != nil {
			t.handleError(fmt.Errorf("error getting latest finalized block: %w", err))
			return
		}
		latestFinalizedEpoch := latestFinalizedBlock.Slot / headState.BeaconConfig.SlotsPerEpoch

		// Check if a rewards interval is due
		isRewardsSubmissionDue, snapshotEnd, intervalsPassed, startTime, endTime, err := t.isRewardsIntervalSubmissionRequired(headState)
		if err != nil {
			t.handleError(fmt.Errorf("error checking if rewards submission is required: %w", err))
			return
		}

		// If no special upcoming state is required, update normally
		if !isRewardsSubmissionDue {
			err = t.recordMgr.UpdateRecordToState(headState, latestFinalizedBlock.Slot)
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
		rewardsEpoch := snapshotEnd.Slot / headState.BeaconConfig.SlotsPerEpoch
		requiredRewardsEpoch := rewardsEpoch + 1
		isRewardsReadyForReport := isRewardsSubmissionDue && (latestFinalizedEpoch >= requiredRewardsEpoch)

		// Run updates and submissions as required
		if isRewardsReadyForReport {
			// Check if there's an existing file for this interval, and try submitting that
			existingRewardsFile, valid, mustRegenerate := t.isExistingRewardsFileValid(headState.NetworkDetails.RewardIndex, intervalsPassed, nodeAddress, isInOdao)
			if existingRewardsFile != nil {
				if valid && !mustRegenerate {
					// We already have a valid file and submission
					t.log.Printlnf("%s Rewards tree has already been submitted for interval %d and is still valid but consensus hasn't been reached yet; nothing to do.", t.logPrefix, headState.NetworkDetails.RewardIndex)
					t.lock.Lock()
					t.isRunning = false
					t.lock.Unlock()
					return
				} else if !valid && !mustRegenerate {
					// We already have a valid file but need to submit again
					t.log.Printlnf("%s Rewards tree has already been created for interval %d but hasn't been submitted yet, attempting resubmission.", t.logPrefix, headState.NetworkDetails.RewardIndex)
				} else if !valid && mustRegenerate {
					// We have a file but it's not valid (probably because too many intervals have passed)
					t.log.Printlnf("%s Rewards submission for interval %d is due and current file is no longer valid (likely too many intervals have passed since its creation), regenerating it.", t.logPrefix, headState.NetworkDetails.RewardIndex)
				}
			}

			// Get the actual slot to report on
			err = t.getTrueRewardsIntervalSubmissionSlot(snapshotEnd)
			if err != nil {
				t.handleError(fmt.Errorf("error getting the true rewards interval slot: %w", err))
				return
			}

			// Get an appropriate client that has access to the target state - this is required if the state gets pruned by the local EC and the
			// archive EC is required
			client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.printMessage, big.NewInt(0).SetUint64(snapshotEnd.ExecutionBlock))
			if err != nil {
				t.handleError(fmt.Errorf("error getting best API client during rewards submission: %w", err))
				return
			}

			// Generate the rewards state
			stateMgr := state.NewNetworkStateManager(client, t.cfg.Smartnode.GetStateManagerContracts(), t.bc, &t.log)
			state, err := stateMgr.GetStateForSlot(snapshotEnd.ConsensusBlock)
			if err != nil {
				t.handleError(fmt.Errorf("error getting state for rewards slot: %w", err))
				return
			}

			// Process the rewards interval
			t.log.Printlnf("%s Running rewards interval submission.", t.logPrefix)
			err = t.runRewardsIntervalReport(
				client,
				state,
				isInOdao,
				intervalsPassed,
				startTime, endTime, snapshotEnd,
				mustRegenerate,
				existingRewardsFile,
			)
			if err != nil {
				t.handleError(fmt.Errorf("error running rewards interval report: %w", err))
				return
			}
		} else {
			t.log.Printlnf("%s Rewards submission for interval %d is due... waiting for epoch %d to be finalized (currently on epoch %d)", t.logPrefix, headState.NetworkDetails.RewardIndex, requiredRewardsEpoch, latestFinalizedEpoch)
		}

		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

	return nil
}

// Print a message from the tree generation goroutine
func (t *submitRewardsTree_Rolling) printMessage(message string) {
	t.log.Printlnf("%s %s", t.logPrefix, message)
}

// Print an error and unlock the mutex
func (t *submitRewardsTree_Rolling) handleError(err error) {
	t.errLog.Printlnf("%s %s", t.logPrefix, err.Error())
	t.errLog.Println("*** Rolling Record processing failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Check if a rewards interval submission is required and if so, the slot number for the update
func (t *submitRewardsTree_Rolling) isRewardsIntervalSubmissionRequired(state *state.NetworkState) (bool, *rprewards.SnapshotEnd, uint64, time.Time, time.Time, error) {
	// Check if a rewards interval has passed and needs to be calculated
	startTime := state.NetworkDetails.IntervalStart
	intervalTime := state.NetworkDetails.IntervalDuration

	// Adjust for the first interval by making the start time the RPL inflation interval start time
	if startTime == time.Unix(0, 0) {
		var err error
		opts := &bind.CallOpts{
			BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
		}
		startTime, err = tokens.GetRPLInflationIntervalStartTime(t.rp, opts)
		if err != nil {
			return false, nil, 0, time.Time{}, time.Time{}, fmt.Errorf("start time is zero, but error getting Rocket Pool deployment block: %w", err)
		}
		t.log.Printlnf("NOTE: rewards pool interval start time is 0, using the inflation interval start time according to the RPL token (%s)", startTime.String())
	}

	// Calculate the end time, which is the number of intervals that have gone by since the current one's start
	secondsSinceGenesis := time.Duration(state.BeaconConfig.SecondsPerSlot*state.BeaconSlotNumber) * time.Second
	stateTime := t.genesisTime.Add(secondsSinceGenesis)
	timeSinceStart := stateTime.Sub(startTime)
	intervalsPassed := timeSinceStart / intervalTime
	endTime := startTime.Add(intervalTime * intervalsPassed)
	if intervalsPassed == 0 {
		return false, nil, 0, time.Time{}, time.Time{}, nil
	}

	// Get the target slot number
	eth2Config := state.BeaconConfig
	totalTimespan := endTime.Sub(t.genesisTime)
	targetSlot := uint64(math.Ceil(totalTimespan.Seconds() / float64(eth2Config.SecondsPerSlot)))
	targetSlotEpoch := targetSlot / eth2Config.SlotsPerEpoch
	targetSlot = (targetSlotEpoch+1)*eth2Config.SlotsPerEpoch - 1 // The target slot becomes the last one in the Epoch

	snapshotEnd := &rprewards.SnapshotEnd{
		Slot: targetSlot,
	}

	return true, snapshotEnd, uint64(intervalsPassed), startTime, endTime, nil
}

// Get the actual slot to be used for a rewards interval submission instead of the naively-determined one
// NOTE: only call this once the required epoch (targetSlotEpoch + 1) has been finalized
func (t *submitRewardsTree_Rolling) getTrueRewardsIntervalSubmissionSlot(snapshotEnd *rprewards.SnapshotEnd) error {
	targetSlot := snapshotEnd.Slot
	// Get the first successful block
	for {
		// Try to get the current block
		block, exists, err := t.bc.GetBeaconBlock(fmt.Sprint(targetSlot))
		if err != nil {
			return fmt.Errorf("error getting Beacon block %d: %w", targetSlot, err)
		}

		// If the block was missing, try the previous one
		if !exists {
			t.log.Printlnf("%s Slot %d was missing, trying the previous one...", t.logPrefix, targetSlot)
			targetSlot--
			continue
		}
		// Ok, we have the first proposed finalized block - this is the one to use for the snapshot!
		snapshotEnd.ConsensusBlock = targetSlot
		snapshotEnd.ExecutionBlock = block.ExecutionBlockNumber
		return nil
	}
}

// Checks to see if an existing rewards file is still valid and whether or not it should be regenerated or just resubmitted
func (t *submitRewardsTree_Rolling) isExistingRewardsFileValid(rewardIndex uint64, intervalsPassed uint64, nodeAddress common.Address, isInOdao bool) (*rprewards.LocalRewardsFile, bool, bool) {
	rewardsTreePath := t.cfg.Smartnode.GetRewardsTreePath(rewardIndex, true, config.RewardsExtensionJSON)

	// Check if the rewards file exists
	_, err := os.Stat(rewardsTreePath)
	if os.IsNotExist(err) {
		return nil, false, true
	}
	if err != nil {
		t.log.Printlnf("%s WARNING: failed to check if [%s] exists: %s; regenerating file...\n", t.logPrefix, rewardsTreePath, err.Error())
		return nil, false, true
	}

	// The file already exists, attempt to read it
	localRewardsFile, err := rprewards.ReadLocalRewardsFile(rewardsTreePath)
	if err != nil {
		t.log.Printlnf("%s WARNING: failed to read %s: %s; regenerating file...\n", t.logPrefix, rewardsTreePath, err.Error())
		return nil, false, true
	}

	proofWrapper := localRewardsFile.Impl()

	if isInOdao {
		// Save the compressed file and get the CID for it
		_, cid, err := localRewardsFile.CreateCompressedFileAndCid()
		if err != nil {
			t.log.Printlnf("%s WARNING: failed to get CID for %s: %s; regenerating file...\n", t.logPrefix, rewardsTreePath, err.Error())
			return nil, false, true
		}

		// Check if this file has already been submitted
		submission := rewards.RewardSubmission{
			RewardIndex:     big.NewInt(0).SetUint64(proofWrapper.GetIndex()),
			ExecutionBlock:  big.NewInt(0).SetUint64(proofWrapper.GetExecutionEndBlock()),
			ConsensusBlock:  big.NewInt(0).SetUint64(proofWrapper.GetConsensusEndBlock()),
			MerkleRoot:      common.HexToHash(proofWrapper.GetMerkleRoot()),
			MerkleTreeCID:   cid.String(),
			IntervalsPassed: big.NewInt(0).SetUint64(proofWrapper.GetIntervalsPassed()),
			TreasuryRPL:     proofWrapper.GetTotalProtocolDaoRpl(),
			TrustedNodeRPL:  []*big.Int{proofWrapper.GetTotalOracleDaoRpl()},
			NodeRPL:         []*big.Int{proofWrapper.GetTotalCollateralRpl()},
			NodeETH:         []*big.Int{proofWrapper.GetTotalNodeOperatorSmoothingPoolEth()},
			UserETH:         proofWrapper.GetTotalPoolStakerSmoothingPoolEth(),
		}

		hasSubmitted, err := rewards.GetTrustedNodeSubmittedSpecificRewards(t.rp, nodeAddress, submission, nil)
		if err != nil {
			t.log.Printlnf("%s WARNING: could not check if node has previously submitted file %s: %s; regenerating file...\n", t.logPrefix, rewardsTreePath, err.Error())
			return nil, false, true
		}
		if !hasSubmitted {
			if proofWrapper.GetIntervalsPassed() != intervalsPassed {
				t.log.Printlnf("%s Existing file for interval %d had %d intervals passed but %d have passed now, regenerating file...",
					t.logPrefix,
					proofWrapper.GetIndex(),
					proofWrapper.GetIntervalsPassed(),
					intervalsPassed,
				)
				return localRewardsFile, false, true
			}
			t.log.Printlnf("%s Existing file for interval %d has not been submitted yet.", t.logPrefix, proofWrapper.GetIndex())
			return localRewardsFile, false, false
		}
	}

	// Check if the file's valid (same number of intervals passed as the current time)
	if proofWrapper.GetIntervalsPassed() != intervalsPassed {
		t.log.Printlnf("%s Existing file for interval %d had %d intervals passed but %d have passed now, regenerating file...",
			t.logPrefix,
			proofWrapper.GetIndex(),
			proofWrapper.GetIntervalsPassed(),
			intervalsPassed,
		)
		return localRewardsFile, false, true
	}

	// File's good and it has the same number of intervals passed, so use it
	return localRewardsFile, true, false
}

// Run a rewards interval report submission
func (t *submitRewardsTree_Rolling) runRewardsIntervalReport(
	client *rocketpool.RocketPool,
	state *state.NetworkState,
	isInOdao bool,
	intervalsPassed uint64,
	startTime time.Time, endTime time.Time, snapshotEnd *rprewards.SnapshotEnd,
	mustRegenerate bool,
	existingRewardsFile *rprewards.LocalRewardsFile,
) error {
	// Prep the record for reporting
	err := t.recordMgr.PrepareRecordForReport(state)
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
	rewardsTreePathJSON := t.cfg.Smartnode.GetRewardsTreePath(currentIndex, true, config.RewardsExtensionJSON)
	compressedRewardsTreePathJSON := rewardsTreePathJSON + config.RewardsTreeIpfsExtension

	// Check if we can reuse an existing file for this interval
	if !mustRegenerate {
		if !isInOdao {
			t.log.Printlnf("%s Node is not in the Oracle DAO, skipping submission for interval %d.", t.logPrefix, currentIndex)
			return nil
		}

		t.log.Printlnf("%s Merkle rewards tree for interval %d already exists at %s, attempting to resubmit...", t.logPrefix, currentIndex, rewardsTreePathJSON)

		// Save the compressed file and get the CID for it
		_, cid, err := existingRewardsFile.CreateCompressedFileAndCid()
		if err != nil {
			return fmt.Errorf("error getting CID for file %s: %w", compressedRewardsTreePathJSON, err)
		}
		t.printMessage(fmt.Sprintf("Calculated rewards tree CID: %s", cid))

		// Submit to the contracts
		err = t.submitRewardsSnapshot(currentIndexBig, snapshotBeaconBlock, elBlockIndex, existingRewardsFile.Impl(), cid.String(), big.NewInt(int64(intervalsPassed)))
		if err != nil {
			return fmt.Errorf("error submitting rewards snapshot: %w", err)
		}

		t.log.Printlnf("%s Successfully submitted rewards snapshot for interval %d.", t.logPrefix, currentIndex)
		return nil
	}

	// Generate the tree
	err = t.generateTree(client, state, intervalsPassed, isInOdao, currentIndex, snapshotEnd, elBlockIndex, startTime, endTime, snapshotElBlockHeader)
	if err != nil {
		return fmt.Errorf("error generating rewards tree: %w", err)
	}

	return nil
}

// Implementation for rewards tree generation using a viable EC
func (t *submitRewardsTree_Rolling) generateTree(
	rp *rocketpool.RocketPool,
	state *state.NetworkState,
	intervalsPassed uint64,
	nodeTrusted bool,
	currentIndex uint64,
	snapshotEnd *rprewards.SnapshotEnd,
	elBlockIndex uint64,
	startTime time.Time, endTime time.Time,
	snapshotElBlockHeader *types.Header,
) error {
	snapshotBeaconBlock := snapshotEnd.ConsensusBlock

	// Log
	if intervalsPassed > 1 {
		t.log.Printlnf("WARNING: %d intervals have passed since the last rewards checkpoint was submitted! Rolling them into one...", intervalsPassed)
	}
	t.log.Printlnf("Rewards checkpoint has passed, starting Merkle tree generation for interval %d in the background.\n%s Snapshot Beacon block = %d, EL block = %d, running from %s to %s", currentIndex, t.logPrefix, snapshotBeaconBlock, elBlockIndex, startTime, endTime)

	// Generate the rewards file
	treegen, err := rprewards.NewTreeGenerator(&t.log, t.logPrefix, rp, t.cfg, t.bc, currentIndex, startTime, endTime, snapshotEnd, snapshotElBlockHeader, uint64(intervalsPassed), state, t.recordMgr.Record)
	if err != nil {
		return fmt.Errorf("Error creating Merkle tree generator: %w", err)
	}
	treeResult, err := treegen.GenerateTree()
	if err != nil {
		return fmt.Errorf("Error generating Merkle tree: %w", err)
	}
	rewardsFile := treeResult.RewardsFile
	for address, network := range treeResult.InvalidNetworkNodes {
		t.printMessage(fmt.Sprintf("WARNING: Node %s has invalid network %d assigned! Using 0 (mainnet) instead.", address.Hex(), network))
	}

	// Save the files
	t.printMessage("Generation complete! Saving files...")
	cid, cids, err := treegen.SaveFiles(treeResult, nodeTrusted)
	if err != nil {
		return fmt.Errorf("Error writing rewards artifacts to disk: %w", err)
	}
	for filename, cid := range cids {
		t.printMessage(fmt.Sprintf("\t%s - CID %s", filename, cid.String()))
	}

	if nodeTrusted {
		t.printMessage(fmt.Sprintf("Calculated rewards tree CID: %s", cid))
		// Submit to the contracts
		err = t.submitRewardsSnapshot(big.NewInt(int64(currentIndex)), snapshotBeaconBlock, elBlockIndex, rewardsFile, cid.String(), big.NewInt(int64(intervalsPassed)))
		if err != nil {
			return fmt.Errorf("Error submitting rewards snapshot: %w", err)
		}

		t.printMessage(fmt.Sprintf("Successfully submitted rewards snapshot for interval %d.", currentIndex))
	} else {
		t.printMessage(fmt.Sprintf("Successfully generated rewards snapshot for interval %d.", currentIndex))
	}

	return nil

}

// Submit rewards info to the contracts
func (t *submitRewardsTree_Rolling) submitRewardsSnapshot(index *big.Int, consensusBlock uint64, executionBlock uint64, rewardsFile rprewards.IRewardsFile, cid string, intervalsPassed *big.Int) error {

	treeRootBytes, err := hex.DecodeString(hexutil.RemovePrefix(rewardsFile.GetMerkleRoot()))
	if err != nil {
		return fmt.Errorf("Error decoding merkle root: %w", err)
	}
	treeRoot := common.BytesToHash(treeRootBytes)

	// Create the arrays of rewards per network
	collateralRplRewards := []*big.Int{}
	oDaoRplRewards := []*big.Int{}
	smoothingPoolEthRewards := []*big.Int{}

	// Create the total rewards for each network
	// Create the total rewards for each network
	for network := uint64(0); rewardsFile.HasRewardsForNetwork(network); network++ {

		collateralRplRewards = append(collateralRplRewards, rewardsFile.GetNetworkCollateralRpl(network))
		oDaoRplRewards = append(oDaoRplRewards, rewardsFile.GetNetworkOracleDaoRpl(network))
		smoothingPoolEthRewards = append(smoothingPoolEthRewards, rewardsFile.GetNetworkSmoothingPoolEth(network))
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
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
		TreasuryRPL:     rewardsFile.GetTotalProtocolDaoRpl(),
		NodeRPL:         collateralRplRewards,
		TrustedNodeRPL:  oDaoRplRewards,
		NodeETH:         smoothingPoolEthRewards,
		UserETH:         rewardsFile.GetTotalPoolStakerSmoothingPoolEth(),
	}

	// Get the gas limit
	gasInfo, err := rewards.EstimateSubmitRewardSnapshotGas(t.rp, submission, opts)
	if err != nil {
		if enableSubmissionAfterConsensus_RewardsTree && strings.Contains(err.Error(), "Can only submit snapshot for next period") {
			// Set a gas limit which will intentionally be too low and revert
			gasInfo = rocketpool.GasInfo{
				EstGasLimit:  utils.RewardsSubmissionForcedGas,
				SafeGasLimit: utils.RewardsSubmissionForcedGas,
			}
			t.log.Println("Rewards period consensus has already been reached but submitting anyway for the health check.")
		} else {
			return fmt.Errorf("Could not estimate the gas required to submit the rewards tree: %w", err)
		}
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, &t.log, maxFee, 0) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = gasInfo.SafeGasLimit

	// Submit RPL price
	hash, err := rewards.SubmitRewardSnapshot(t.rp, submission, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Return
	return nil
}
