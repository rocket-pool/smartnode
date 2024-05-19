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
	"github.com/ethereum/go-ethereum/crypto"
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
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)

// Submit rewards Merkle Tree task
type submitRewardsTree_Stateless struct {
	c                *cli.Context
	log              *log.ColorLogger
	errLog           *log.ColorLogger
	cfg              *config.RocketPoolConfig
	w                *wallet.Wallet
	rp               *rocketpool.RocketPool
	ec               rocketpool.ExecutionClient
	bc               beacon.Client
	lock             *sync.Mutex
	isRunning        bool
	generationPrefix string
	m                *state.NetworkStateManager
}

// Create submit rewards Merkle Tree task
func newSubmitRewardsTree_Stateless(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger, m *state.NetworkStateManager) (*submitRewardsTree_Stateless, error) {

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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	lock := &sync.Mutex{}
	generator := &submitRewardsTree_Stateless{
		c:                c,
		log:              &logger,
		errLog:           &errorLogger,
		cfg:              cfg,
		ec:               ec,
		bc:               bc,
		w:                w,
		rp:               rp,
		lock:             lock,
		isRunning:        false,
		generationPrefix: "[Merkle Tree]",
		m:                m,
	}

	return generator, nil
}

// Submit rewards Merkle Tree
func (t *submitRewardsTree_Stateless) Run(nodeTrusted bool, state *state.NetworkState, beaconSlot uint64) error {

	// Wait for clients to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	if err := services.WaitBeaconClientSynced(t.c, true); err != nil {
		return err
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Check node trusted status
	if !nodeTrusted {
		if t.cfg.Smartnode.RewardsTreeMode.Value.(cfgtypes.RewardsMode) != cfgtypes.RewardsMode_Generate {
			return nil
		} else {
			// Create the state, since it's not done except for manual generators
			state, err = t.m.GetStateForSlot(beaconSlot)
			if err != nil {
				return fmt.Errorf("error getting state for beacon slot %d: %w", beaconSlot, err)
			}
		}
	}

	// Log
	t.log.Println("Checking for rewards checkpoint...")

	// Check if a rewards interval has passed and needs to be calculated
	startTime := state.NetworkDetails.IntervalStart
	intervalTime := state.NetworkDetails.IntervalDuration

	// Adjust for the first interval by making the start time the RPL inflation interval start time
	if startTime == time.Unix(0, 0) {
		opts := &bind.CallOpts{
			BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
		}
		startTime, err = tokens.GetRPLInflationIntervalStartTime(t.rp, opts)
		if err != nil {
			return fmt.Errorf("start time is zero, but error getting Rocket Pool deployment block: %w", err)
		}
		t.log.Printlnf("NOTE: rewards pool interval start time is 0, using the inflation interval start time according to the RPL token (%s)", startTime.String())
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
		t.log.Println("Tree generation is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	// Get the expected file paths
	rewardsTreePathJSON := t.cfg.Smartnode.GetRewardsTreePath(currentIndex, true, config.RewardsExtensionJSON)
	compressedRewardsTreePathJSON := rewardsTreePathJSON + config.RewardsTreeIpfsExtension

	// Check if we can reuse an existing file for this interval
	if t.isExistingRewardsFileValid(rewardsTreePathJSON, uint64(intervalsPassed)) {
		if !nodeTrusted {
			t.log.Printlnf("Merkle rewards tree for interval %d already exists at %s.", currentIndex, rewardsTreePathJSON)
			return nil
		}

		// Return if this node has already submitted the tree for the current interval and there's a file present
		hasSubmitted, err := t.hasSubmittedTree(nodeAccount.Address, currentIndexBig)
		if err != nil {
			return fmt.Errorf("error checking if Merkle tree submission has already been processed: %w", err)
		}
		if hasSubmitted {
			return nil
		}

		t.log.Printlnf("Merkle rewards tree for interval %d already exists at %s, attempting to resubmit...", currentIndex, rewardsTreePathJSON)

		// Deserialize the file
		localRewardsFile, err := rprewards.ReadLocalRewardsFile(rewardsTreePathJSON)
		if err != nil {
			return fmt.Errorf("Error reading rewards tree file: %w", err)
		}

		proofWrapper := localRewardsFile.Impl()

		// Save the compressed file and get the CID for it
		_, cid, err := localRewardsFile.CreateCompressedFileAndCid()
		if err != nil {
			return fmt.Errorf("Error getting CID for file %s: %w", compressedRewardsTreePathJSON, err)
		}

		t.printMessage(fmt.Sprintf("Calculated rewards tree CID: %s", cid))

		// Submit to the contracts
		err = t.submitRewardsSnapshot(currentIndexBig, snapshotBeaconBlock, elBlockIndex, proofWrapper.GetHeader(), cid.String(), big.NewInt(int64(intervalsPassed)))
		if err != nil {
			return fmt.Errorf("Error submitting rewards snapshot: %w", err)
		}

		t.log.Printlnf("Successfully submitted rewards snapshot for interval %d.", currentIndex)
		return nil
	}

	// Generate the tree
	t.generateTree(intervalsPassed, nodeTrusted, currentIndex, snapshotBeaconBlock, elBlockIndex, startTime, endTime, snapshotElBlockHeader)

	// Done
	return nil

}

func (t *submitRewardsTree_Stateless) handleError(err error) {
	t.errLog.Println(fmt.Errorf("%s %w", t.generationPrefix, err))
	t.errLog.Println("*** Rewards tree generation failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Print a message from the tree generation goroutine
func (t *submitRewardsTree_Stateless) printMessage(message string) {
	t.log.Printlnf("%s %s", t.generationPrefix, message)
}

// Checks to see if an existing rewards file is still valid
func (t *submitRewardsTree_Stateless) isExistingRewardsFileValid(rewardsTreePath string, intervalsPassed uint64) bool {

	_, err := os.Stat(rewardsTreePath)
	if os.IsNotExist(err) {
		return false
	}

	// The file already exists, attempt to read it
	localRewardsFile, err := rprewards.ReadLocalRewardsFile(rewardsTreePath)
	if err != nil {
		t.log.Printlnf("WARNING: failed to read %s: %s\nRegenerating file...\n", rewardsTreePath, err.Error())
		return false
	}

	// Compare the number of intervals in it with the current number of intervals
	proofWrapper := localRewardsFile.Impl()
	header := proofWrapper.GetHeader()
	if header.IntervalsPassed != intervalsPassed {
		t.log.Printlnf("Existing file for interval %d had %d intervals passed but %d have passed now, regenerating file...\n", header.Index, header.IntervalsPassed, intervalsPassed)
		return false
	}

	// File's good and it has the same number of intervals passed, so use it
	return true

}

// Kick off the tree generation goroutine
func (t *submitRewardsTree_Stateless) generateTree(intervalsPassed time.Duration, nodeTrusted bool, currentIndex uint64, snapshotBeaconBlock uint64, elBlockIndex uint64, startTime time.Time, endTime time.Time, snapshotElBlockHeader *types.Header) {

	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()

		// Get an appropriate client
		client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.printMessage, snapshotElBlockHeader.Number)
		if err != nil {
			t.handleError(err)
			return
		}

		// Generate the tree
		err = t.generateTreeImpl(client, intervalsPassed, nodeTrusted, currentIndex, snapshotBeaconBlock, elBlockIndex, startTime, endTime, snapshotElBlockHeader)
		if err != nil {
			t.handleError(err)
		}

		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

}

// Implementation for rewards tree generation using a viable EC
func (t *submitRewardsTree_Stateless) generateTreeImpl(rp *rocketpool.RocketPool, intervalsPassed time.Duration, nodeTrusted bool, currentIndex uint64, snapshotBeaconBlock uint64, elBlockIndex uint64, startTime time.Time, endTime time.Time, snapshotElBlockHeader *types.Header) error {

	// Log
	if uint64(intervalsPassed) > 1 {
		t.log.Printlnf("WARNING: %d intervals have passed since the last rewards checkpoint was submitted! Rolling them into one...", uint64(intervalsPassed))
	}
	t.log.Printlnf("Rewards checkpoint has passed, starting Merkle tree generation for interval %d in the background.\n%s Snapshot Beacon block = %d, EL block = %d, running from %s to %s", currentIndex, t.generationPrefix, snapshotBeaconBlock, elBlockIndex, startTime, endTime)

	// Create a new state gen manager
	mgr, err := state.NewNetworkStateManager(rp, t.cfg, rp.Client, t.bc, t.log)
	if err != nil {
		return fmt.Errorf("error creating network state manager for EL block %d, Beacon slot %d: %w", elBlockIndex, snapshotBeaconBlock, err)
	}

	// Create a new state for the target block
	state, err := mgr.GetStateForSlot(snapshotBeaconBlock)
	if err != nil {
		return fmt.Errorf("couldn't get network state for EL block %d, Beacon slot %d: %w", elBlockIndex, snapshotBeaconBlock, err)
	}

	// Generate the rewards file
	treegen, err := rprewards.NewTreeGenerator(t.log, t.generationPrefix, rp, t.cfg, t.bc, currentIndex, startTime, endTime, snapshotBeaconBlock, snapshotElBlockHeader, uint64(intervalsPassed), state, nil)
	if err != nil {
		return fmt.Errorf("Error creating Merkle tree generator: %w", err)
	}
	rewardsFile, err := treegen.GenerateTree()
	if err != nil {
		return fmt.Errorf("Error generating Merkle tree: %w", err)
	}
	for address, network := range rewardsFile.GetHeader().InvalidNetworkNodes {
		t.printMessage(fmt.Sprintf("WARNING: Node %s has invalid network %d assigned! Using 0 (mainnet) instead.", address.Hex(), network))
	}

	// Save the files
	t.printMessage("Generation complete! Saving files...")
	cid, cids, err := treegen.SaveFiles(rewardsFile, nodeTrusted)
	if err != nil {
		return fmt.Errorf("Error writing rewards artifacts to disk: %w", err)
	}
	for filename, cid := range cids {
		t.printMessage(fmt.Sprintf("\t%s - CID %s", filename, cid.String()))
	}

	if nodeTrusted {
		t.printMessage(fmt.Sprintf("Calculated rewards tree CID: %s", cid))

		// Submit to the contracts
		err = t.submitRewardsSnapshot(big.NewInt(int64(currentIndex)), snapshotBeaconBlock, elBlockIndex, rewardsFile.GetHeader(), cid.String(), big.NewInt(int64(intervalsPassed)))
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
func (t *submitRewardsTree_Stateless) submitRewardsSnapshot(index *big.Int, consensusBlock uint64, executionBlock uint64, rewardsFileHeader *rprewards.RewardsFileHeader, cid string, intervalsPassed *big.Int) error {

	treeRootBytes, err := hex.DecodeString(hexutil.RemovePrefix(rewardsFileHeader.MerkleRoot))
	if err != nil {
		return fmt.Errorf("Error decoding merkle root: %w", err)
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
		TreasuryRPL:     &rewardsFileHeader.TotalRewards.ProtocolDaoRpl.Int,
		NodeRPL:         collateralRplRewards,
		TrustedNodeRPL:  oDaoRplRewards,
		NodeETH:         smoothingPoolEthRewards,
		UserETH:         &rewardsFileHeader.TotalRewards.PoolStakerSmoothingPoolEth.Int,
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
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log, maxFee, 0) {
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
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Return
	return nil
}

// Get the first finalized, successful consensus block that occurred after the given target time
func (t *submitRewardsTree_Stateless) getSnapshotConsensusBlock(endTime time.Time, state *state.NetworkState) (uint64, uint64, error) {

	// Get the beacon head
	beaconHead, err := t.bc.GetBeaconHead()
	if err != nil {
		return 0, 0, fmt.Errorf("Error getting Beacon head: %w", err)
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
		return 0, 0, fmt.Errorf("Snapshot end time = %s, slot (epoch) = %d (%d)... waiting until epoch %d is finalized (currently %d).", endTime, targetSlot, targetSlotEpoch, requiredEpoch, beaconHead.FinalizedEpoch)
	}

	// Get the first successful block
	for {
		// Try to get the current block
		block, exists, err := t.bc.GetBeaconBlock(fmt.Sprint(targetSlot))
		if err != nil {
			return 0, 0, fmt.Errorf("Error getting Beacon block %d: %w", targetSlot, err)
		}

		// If the block was missing, try the previous one
		if !exists {
			t.log.Printlnf("Slot %d was missing, trying the previous one...", targetSlot)
			targetSlot--
		} else {
			// Ok, we have the first proposed finalized block - this is the one to use for the snapshot!
			return targetSlot, block.ExecutionBlockNumber, nil
		}
	}

}

// Check whether the rewards tree for the current interval been submitted by the node
func (t *submitRewardsTree_Stateless) hasSubmittedTree(nodeAddress common.Address, index *big.Int) (bool, error) {
	indexBuffer := make([]byte, 32)
	index.FillBytes(indexBuffer)
	return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("rewards.snapshot.submitted.node"), nodeAddress.Bytes(), indexBuffer))
}
