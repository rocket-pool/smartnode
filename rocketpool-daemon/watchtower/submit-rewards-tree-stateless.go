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
	"github.com/ethereum/go-ethereum/crypto"
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

// Submit rewards Merkle Tree task
type SubmitRewardsTree_Stateless struct {
	ctx         context.Context
	sp          *services.ServiceProvider
	logger      *slog.Logger
	cfg         *config.SmartNodeConfig
	w           *wallet.Wallet
	rp          *rocketpool.RocketPool
	ec          eth.IExecutionClient
	bc          beacon.IBeaconClient
	rewardsPool *rewards.RewardsPool
	lock        *sync.Mutex
	isRunning   bool
	m           *state.NetworkStateManager
}

// Create submit rewards Merkle Tree task
func NewSubmitRewardsTree_Stateless(ctx context.Context, sp *services.ServiceProvider, logger *log.Logger, m *state.NetworkStateManager) *SubmitRewardsTree_Stateless {
	lock := &sync.Mutex{}
	return &SubmitRewardsTree_Stateless{
		ctx:       ctx,
		sp:        sp,
		logger:    logger.With(slog.String(keys.RoutineKey, "Merkle Tree")),
		cfg:       sp.GetConfig(),
		w:         sp.GetWallet(),
		rp:        sp.GetRocketPool(),
		ec:        sp.GetEthClient(),
		bc:        sp.GetBeaconClient(),
		lock:      lock,
		isRunning: false,
		m:         m,
	}
}

// Submit rewards Merkle Tree
func (t *SubmitRewardsTree_Stateless) Run(nodeTrusted bool, state *state.NetworkState, beaconSlot uint64) error {
	// Check node trusted status
	if !nodeTrusted {
		if t.cfg.RewardsTreeMode.Value != config.RewardsMode_Generate {
			return nil
		} else {
			// Create the state, since it's not done except for manual generators
			var err error
			state, err = t.m.GetStateForSlot(t.ctx, beaconSlot)
			if err != nil {
				return fmt.Errorf("error getting state for beacon slot %d: %w", beaconSlot, err)
			}
		}
	}

	// Log
	t.logger.Info("Checking for rewards checkpoint...")

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
			return fmt.Errorf("error creating RPL token binding: %w", err)
		}
		err = t.rp.Query(nil, opts, rpl.InflationIntervalStartTime)
		if err != nil {
			return fmt.Errorf("start time is zero, but error getting RPL token inflation interval start time: %w", err)
		}
		startTime = rpl.InflationIntervalStartTime.Formatted()
		t.logger.Info("NOTE: rewards pool interval start time is 0, using the inflation interval start time according to the RPL token", slog.Time(keys.StartKey, startTime))
	}

	// Calculate the end time, which is the number of intervals that have gone by since the current one's start
	genesisTime := time.Unix(int64(state.BeaconConfig.GenesisTime), 0)
	secondsSinceGenesis := time.Duration(state.BeaconConfig.SecondsPerSlot*state.BeaconSlotNumber) * time.Second
	stateTime := genesisTime.Add(secondsSinceGenesis)
	timeSinceStart := stateTime.Sub(startTime)
	intervalsPassed := timeSinceStart / intervalTime
	endTime := startTime.Add(intervalTime * intervalsPassed)
	if intervalsPassed == 0 {
		return nil
	}

	// Get the block and timestamp of the consensus block that best matches the end time
	snapshotBeaconBlock, elBlockNumber, err := t.getSnapshotConsensusBlock(endTime, state)
	if err != nil {
		return err
	}

	// Get the number of the EL block matching the CL snapshot block
	snapshotElBlockHeader, err := t.ec.HeaderByNumber(context.Background(), big.NewInt(int64(elBlockNumber)))
	if err != nil {
		return err
	}
	elBlockIndex := snapshotElBlockHeader.Number.Uint64()

	// Get the current interval
	currentIndex := state.NetworkDetails.RewardIndex
	currentIndexBig := big.NewInt(0).SetUint64(currentIndex)

	// Check if rewards generation is already running
	t.lock.Lock()
	if t.isRunning {
		t.logger.Info("Tree generation is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	// Refresh contract bindings
	nodeAddress, _ := t.w.GetAddress()
	t.rewardsPool, err = rewards.NewRewardsPool(t.rp)
	if err != nil {
		t.handleError(fmt.Errorf("error creating Rewards Pool binding: %w", err))
		return nil
	}

	// Get the expected file paths
	rewardsTreePath := t.cfg.GetRewardsTreePath(currentIndex)
	compressedRewardsTreePath := rewardsTreePath + config.RewardsTreeIpfsExtension
	minipoolPerformancePath := t.cfg.GetMinipoolPerformancePath(currentIndex)
	compressedMinipoolPerformancePath := minipoolPerformancePath + config.RewardsTreeIpfsExtension

	// Check if we can reuse an existing file for this interval
	if t.isExistingRewardsFileValid(rewardsTreePath, uint64(intervalsPassed)) {
		if !nodeTrusted {
			t.logger.Info("Merkle rewards tree file already exists.", slog.Uint64(keys.IntervalKey, currentIndex), slog.String(keys.FileKey, rewardsTreePath))
			return nil
		}

		// Return if this node has already submitted the tree for the current interval and there's a file present
		hasSubmitted, err := t.hasSubmittedTree(nodeAddress, currentIndexBig)
		if err != nil {
			return fmt.Errorf("error checking if Merkle tree submission has already been processed: %w", err)
		}
		if hasSubmitted {
			return nil
		}

		t.logger.Info("Merkle rewards tree file already exists, attempting to resubmit...", slog.Uint64(keys.IntervalKey, currentIndex), slog.String(keys.FileKey, rewardsTreePath))

		// Deserialize the file
		localRewardsFile, err := rprewards.ReadLocalRewardsFile(rewardsTreePath)
		if err != nil {
			return fmt.Errorf("error reading rewards tree file: %w", err)
		}

		proofWrapper := localRewardsFile.Impl()

		// Save the compressed file and get the CID for it
		cid, err := localRewardsFile.CreateCompressedFileAndCid()
		if err != nil {
			return fmt.Errorf("error getting CID for file %s: %w", compressedRewardsTreePath, err)
		}

		t.logger.Info("Calculated rewards tree CID", slog.String(keys.CidKey, cid.String()))

		// Submit to the contracts
		err = t.submitRewardsSnapshot(currentIndexBig, snapshotBeaconBlock, elBlockIndex, proofWrapper.GetHeader(), cid.String(), big.NewInt(int64(intervalsPassed)))
		if err != nil {
			return fmt.Errorf("error submitting rewards snapshot: %w", err)
		}

		t.logger.Info("Successfully submitted rewards snapshot.", slog.Uint64(keys.IntervalKey, currentIndex))
		return nil
	}

	// Generate the tree
	t.generateTree(intervalsPassed, nodeTrusted, currentIndex, snapshotBeaconBlock, elBlockIndex, startTime, endTime, snapshotElBlockHeader, rewardsTreePath, compressedRewardsTreePath, minipoolPerformancePath, compressedMinipoolPerformancePath)

	// Done
	return nil
}

// Checks to see if an existing rewards file is still valid
func (t *SubmitRewardsTree_Stateless) isExistingRewardsFileValid(rewardsTreePath string, intervalsPassed uint64) bool {
	_, err := os.Stat(rewardsTreePath)
	if os.IsNotExist(err) {
		return false
	}

	// The file already exists, attempt to read it
	localRewardsFile, err := rprewards.ReadLocalRewardsFile(rewardsTreePath)
	if err != nil {
		t.logger.Warn("Failed to read rewards file, regenerating file...\n", slog.String(keys.FileKey, rewardsTreePath), log.Err(err))
		return false
	}

	// Compare the number of intervals in it with the current number of intervals
	proofWrapper := localRewardsFile.Impl()
	header := proofWrapper.GetHeader()
	if header.IntervalsPassed != intervalsPassed {
		t.logger.Info("Existing file has too few rounds, regenerating file...", slog.Uint64(keys.IntervalKey, header.Index), slog.Uint64(keys.FileRoundsKey, header.IntervalsPassed), slog.Uint64(keys.ActualRoundsKey, intervalsPassed))
		return false
	}

	// File's good and it has the same number of intervals passed, so use it
	return true
}

// Kick off the tree generation goroutine
func (t *SubmitRewardsTree_Stateless) generateTree(intervalsPassed time.Duration, nodeTrusted bool, currentIndex uint64, snapshotBeaconBlock uint64, elBlockIndex uint64, startTime time.Time, endTime time.Time, snapshotElBlockHeader *types.Header, rewardsTreePath string, compressedRewardsTreePath string, minipoolPerformancePath string, compressedMinipoolPerformancePath string) {
	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()

		// Get an appropriate client
		client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.logger, snapshotElBlockHeader.Number)
		if err != nil {
			t.handleError(err)
			return
		}

		// Generate the tree
		err = t.generateTreeImpl(client, intervalsPassed, nodeTrusted, currentIndex, snapshotBeaconBlock, elBlockIndex, startTime, endTime, snapshotElBlockHeader, rewardsTreePath, compressedRewardsTreePath, minipoolPerformancePath, compressedMinipoolPerformancePath)
		if err != nil {
			t.handleError(err)
		}

		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()
}

// Implementation for rewards tree generation using a viable EC
func (t *SubmitRewardsTree_Stateless) generateTreeImpl(rp *rocketpool.RocketPool, intervalsPassed time.Duration, nodeTrusted bool, currentIndex uint64, snapshotBeaconBlock uint64, elBlockIndex uint64, startTime time.Time, endTime time.Time, snapshotElBlockHeader *types.Header, rewardsTreePath string, compressedRewardsTreePath string, minipoolPerformancePath string, compressedMinipoolPerformancePath string) error {
	// Log
	if uint64(intervalsPassed) > 1 {
		t.logger.Warn("Multiple intervals have passed since the last rewards checkpoint was submitted! Rolling them into one...", slog.Uint64(keys.RoundsKey, uint64(intervalsPassed)))
	}
	t.logger.Info("Rewards checkpoint has passed, starting Merkle tree generation in the background.", slog.Uint64(keys.IntervalKey, currentIndex), slog.Uint64(keys.SlotKey, snapshotBeaconBlock), slog.Uint64(keys.BlockKey, elBlockIndex), slog.Time(keys.StartKey, startTime), slog.Time(keys.EndKey, endTime))

	// Create a new state gen manager
	mgr, err := state.NewNetworkStateManager(t.ctx, rp, t.cfg, rp.Client, t.bc, t.logger)
	if err != nil {
		return fmt.Errorf("error creating network state manager for EL block %d, Beacon slot %d: %w", elBlockIndex, snapshotBeaconBlock, err)
	}

	// Create a new state for the target block
	state, err := mgr.GetStateForSlot(t.ctx, snapshotBeaconBlock)
	if err != nil {
		return fmt.Errorf("couldn't get network state for EL block %d, Beacon slot %d: %w", elBlockIndex, snapshotBeaconBlock, err)
	}

	// Generate the rewards file
	treegen, err := rprewards.NewTreeGenerator(t.logger, rp, t.cfg, t.bc, currentIndex, startTime, endTime, snapshotBeaconBlock, snapshotElBlockHeader, uint64(intervalsPassed), state, nil)
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

	// Write it to disk
	err = localMinipoolPerformanceFile.Write()
	if err != nil {
		return fmt.Errorf("error saving minipool performance file to %s: %w", minipoolPerformancePath, err)
	}

	if nodeTrusted {
		minipoolPerformanceCid, err := localMinipoolPerformanceFile.CreateCompressedFileAndCid()
		if err != nil {
			return fmt.Errorf("error getting CID for file %s: %w", compressedMinipoolPerformancePath, err)
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
		// Save the compressed file and get the CID for it
		cid, err := localRewardsFile.CreateCompressedFileAndCid()
		if err != nil {
			return fmt.Errorf("error getting CID for file %s : %w", rewardsTreePath, err)
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
func (t *SubmitRewardsTree_Stateless) submitRewardsSnapshot(index *big.Int, consensusBlock uint64, executionBlock uint64, rewardsFileHeader *sharedtypes.RewardsFileHeader, cid string, intervalsPassed *big.Int) error {
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

	// Get the tx info
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

// Get the first finalized, successful consensus block that occurred after the given target time
func (t *SubmitRewardsTree_Stateless) getSnapshotConsensusBlock(endTime time.Time, state *state.NetworkState) (uint64, uint64, error) {
	// Get the beacon head
	beaconHead, err := t.bc.GetBeaconHead(t.ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("error getting Beacon head: %w", err)
	}

	// Get the target block number
	eth2Config := state.BeaconConfig
	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)
	totalTimespan := endTime.Sub(genesisTime)
	targetSlot := uint64(math.Ceil(totalTimespan.Seconds() / float64(eth2Config.SecondsPerSlot)))
	targetSlotEpoch := targetSlot / eth2Config.SlotsPerEpoch
	targetSlot = targetSlotEpoch*eth2Config.SlotsPerEpoch + (eth2Config.SlotsPerEpoch - 1) // The target slot becomes the last one in the Epoch
	requiredEpoch := targetSlotEpoch + 1                                                   // The smoothing pool requires 1 epoch beyond the target to be finalized, to check for late attestations

	// Check if the required epoch is finalized yet
	if beaconHead.FinalizedEpoch < requiredEpoch {
		return 0, 0, fmt.Errorf("snapshot end time = %s, slot (epoch) = %d (%d)... waiting until epoch %d is finalized (currently %d)", endTime, targetSlot, targetSlotEpoch, requiredEpoch, beaconHead.FinalizedEpoch)
	}

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

// Check whether the rewards tree for the current interval been submitted by the node
func (t *SubmitRewardsTree_Stateless) hasSubmittedTree(nodeAddress common.Address, index *big.Int) (bool, error) {
	indexBuffer := make([]byte, 32)
	index.FillBytes(indexBuffer)
	var result bool
	err := t.rp.Query(func(mc *batch.MultiCaller) error {
		t.rp.Storage.GetBool(mc, &result, crypto.Keccak256Hash([]byte("rewards.snapshot.submitted.node"), nodeAddress.Bytes(), indexBuffer))
		return nil
	}, nil)
	return result, err
}

func (t *SubmitRewardsTree_Stateless) handleError(err error) {
	t.logger.Error("*** Rewards tree generation failed. ***", log.Err(err))
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}
