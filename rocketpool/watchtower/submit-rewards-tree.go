package watchtower

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/goccy/go-json"
	"github.com/klauspost/compress/zstd"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/web3-storage/go-w3s-client"
)

// Submit rewards Merkle Tree task
type submitRewardsTree struct {
	log       *log.ColorLogger
	errLog    *log.ColorLogger
	cfg       *config.RocketPoolConfig
	w         *wallet.Wallet
	rp        *rocketpool.RocketPool
	bc        beacon.Client
	logPrefix string
}

// Create submit rewards Merkle Tree task
func newSubmitRewardsTree(logger *log.ColorLogger, errorLogger *log.ColorLogger, cfg *config.RocketPoolConfig, w *wallet.Wallet, rp *rocketpool.RocketPool, bc beacon.Client) *submitRewardsTree {
	return &submitRewardsTree{
		log:       logger,
		errLog:    errorLogger,
		cfg:       cfg,
		bc:        bc,
		w:         w,
		rp:        rp,
		logPrefix: "[Rewards Tree]",
	}
}

// Submit rewards Merkle Tree
func (t *submitRewardsTree) run(nodeTrusted bool, state *state.NetworkState) error {

	snapshotBeaconBlock := state.BeaconSlotNumber
	elBlockNumber := state.ElBlockNumber

	// Check if a rewards interval has passed and needs to be calculated
	startTime := state.NetworkDetails.IntervalStart
	intervalTime := state.NetworkDetails.IntervalDuration

	// Calculate the end time, which is the number of intervals that have gone by since the current one's start
	genesisTime := time.Unix(int64(state.BeaconConfig.GenesisTime), 0)
	secondsSinceGenesis := time.Duration(state.BeaconConfig.SecondsPerSlot*snapshotBeaconBlock) * time.Second
	stateTime := genesisTime.Add(secondsSinceGenesis)
	timeSinceStart := stateTime.Sub(startTime)
	intervalsPassed := timeSinceStart / intervalTime
	endTime := startTime.Add(intervalTime * intervalsPassed)
	if intervalsPassed == 0 {
		return nil
	}

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
	rewardsTreePath := t.cfg.Smartnode.GetRewardsTreePath(currentIndex, true)
	compressedRewardsTreePath := rewardsTreePath + config.RewardsTreeIpfsExtension
	minipoolPerformancePath := t.cfg.Smartnode.GetMinipoolPerformancePath(currentIndex, true)
	compressedMinipoolPerformancePath := minipoolPerformancePath + config.RewardsTreeIpfsExtension

	// Check if we can reuse an existing file for this interval
	if t.isExistingFileValid(rewardsTreePath, uint64(intervalsPassed)) {
		if !nodeTrusted {
			t.log.Printlnf("Merkle rewards tree for interval %d already exists at %s.", currentIndex, rewardsTreePath)
			return nil
		}

		t.log.Printlnf("Merkle rewards tree for interval %d already exists at %s, attempting to resubmit...", currentIndex, rewardsTreePath)

		// Deserialize the file
		wrapperBytes, err := os.ReadFile(rewardsTreePath)
		if err != nil {
			return fmt.Errorf("Error reading rewards tree file: %w", err)
		}

		proofWrapper := new(rprewards.RewardsFile)
		err = json.Unmarshal(wrapperBytes, proofWrapper)
		if err != nil {
			return fmt.Errorf("Error deserializing rewards tree file: %w", err)
		}

		// Upload the file
		cid, err := t.uploadFileToWeb3Storage(wrapperBytes, compressedRewardsTreePath, "compressed rewards tree")
		if err != nil {
			return fmt.Errorf("Error uploading Merkle tree to Web3.Storage: %w", err)
		}
		t.log.Printlnf("Uploaded Merkle tree with CID %s", cid)

		// Submit to the contracts
		err = t.submitRewardsSnapshot(currentIndexBig, snapshotBeaconBlock, elBlockIndex, proofWrapper, cid, big.NewInt(int64(intervalsPassed)))
		if err != nil {
			return fmt.Errorf("Error submitting rewards snapshot: %w", err)
		}

		t.log.Printlnf("Successfully submitted rewards snapshot for interval %d.", currentIndex)
		return nil
	}

	// Generate the tree
	t.generateTree(intervalsPassed, nodeTrusted, currentIndex, snapshotBeaconBlock, elBlockIndex, startTime, endTime, snapshotElBlockHeader, rewardsTreePath, compressedRewardsTreePath, minipoolPerformancePath, compressedMinipoolPerformancePath)

	// Done
	return nil

}

// Print a message from the tree generation goroutine
func (t *submitRewardsTree) printMessage(message string) {
	t.log.Printlnf("%s %s", t.logPrefix, message)
}

// Checks to see if an existing rewards file is still valid
func (t *submitRewardsTree) isExistingFileValid(rewardsTreePath string, intervalsPassed uint64) bool {

	_, err := os.Stat(rewardsTreePath)
	if !os.IsNotExist(err) {
		// The file already exists, attempt to read it
		var proofWrapper rprewards.RewardsFile
		fileBytes, err := os.ReadFile(rewardsTreePath)
		if err != nil {
			t.log.Printlnf("WARNING: failed to read %s: %s\nRegenerating file...\n", rewardsTreePath, err.Error())
			return false
		}

		err = json.Unmarshal(fileBytes, &proofWrapper)
		if err != nil {
			t.log.Printlnf("WARNING: failed to deserialize %s: %s\nRegenerating file...\n", rewardsTreePath, err.Error())
			return false
		}

		// Compare the number of intervals in it with the current number of intervals
		if proofWrapper.IntervalsPassed != intervalsPassed {
			t.log.Printlnf("Existing file for interval %d had %d intervals passed but %d have passed now, regenerating file...\n", proofWrapper.Index, proofWrapper.IntervalsPassed, intervalsPassed)
			return false
		}

		// File's good and it has the same number of intervals passed, so use it
		return true
	}

	return false

}

// Kick off the tree generation goroutine
func (t *submitRewardsTree) generateTree(intervalsPassed time.Duration, nodeTrusted bool, currentIndex uint64, snapshotBeaconBlock uint64, elBlockIndex uint64, startTime time.Time, endTime time.Time, snapshotElBlockHeader *types.Header, rewardsTreePath string, compressedRewardsTreePath string, minipoolPerformancePath string, compressedMinipoolPerformancePath string) error {
	// Get an appropriate client
	client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.printMessage, snapshotElBlockHeader.Number)
	if err != nil {
		return fmt.Errorf("error getting best API client during rewards submission: %w", err)
	}

	// Generate the tree
	err = t.generateTreeImpl(client, intervalsPassed, nodeTrusted, currentIndex, snapshotBeaconBlock, elBlockIndex, startTime, endTime, snapshotElBlockHeader, rewardsTreePath, compressedRewardsTreePath, minipoolPerformancePath, compressedMinipoolPerformancePath)
	if err != nil {
		return fmt.Errorf("error generating rewards tree: %w", err)
	}

	return nil
}

// Implementation for rewards tree generation using a viable EC
func (t *submitRewardsTree) generateTreeImpl(rp *rocketpool.RocketPool, intervalsPassed time.Duration, nodeTrusted bool, currentIndex uint64, snapshotBeaconBlock uint64, elBlockIndex uint64, startTime time.Time, endTime time.Time, snapshotElBlockHeader *types.Header, rewardsTreePath string, compressedRewardsTreePath string, minipoolPerformancePath string, compressedMinipoolPerformancePath string) error {

	// Log
	if uint64(intervalsPassed) > 1 {
		t.log.Printlnf("WARNING: %d intervals have passed since the last rewards checkpoint was submitted! Rolling them into one...", uint64(intervalsPassed))
	}
	t.log.Printlnf("Rewards checkpoint has passed, starting Merkle tree generation for interval %d in the background.\n%s Snapshot Beacon block = %d, EL block = %d, running from %s to %s", currentIndex, t.logPrefix, snapshotBeaconBlock, elBlockIndex, startTime, endTime)

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
	treegen, err := rprewards.NewTreeGenerator(t.log, t.logPrefix, rp, t.cfg, t.bc, currentIndex, startTime, endTime, snapshotBeaconBlock, snapshotElBlockHeader, uint64(intervalsPassed), state, nil)
	if err != nil {
		return fmt.Errorf("Error creating Merkle tree generator: %w", err)
	}
	rewardsFile, err := treegen.GenerateTree()
	if err != nil {
		return fmt.Errorf("Error generating Merkle tree: %w", err)
	}
	for address, network := range rewardsFile.InvalidNetworkNodes {
		t.printMessage(fmt.Sprintf("WARNING: Node %s has invalid network %d assigned! Using 0 (mainnet) instead.", address.Hex(), network))
	}

	// Serialize the minipool performance file
	minipoolPerformanceBytes, err := json.Marshal(rewardsFile.MinipoolPerformanceFile)
	if err != nil {
		return fmt.Errorf("Error serializing minipool performance file into JSON: %w", err)
	}

	// Write it to disk
	err = os.WriteFile(minipoolPerformancePath, minipoolPerformanceBytes, 0644)
	if err != nil {
		return fmt.Errorf("Error saving minipool performance file to %s: %w", minipoolPerformancePath, err)
	}

	// Upload it if this is an Oracle DAO node
	if nodeTrusted {
		t.printMessage("Uploading minipool performance file to Web3.Storage...")
		minipoolPerformanceCid, err := t.uploadFileToWeb3Storage(minipoolPerformanceBytes, compressedMinipoolPerformancePath, "compressed minipool performance")
		if err != nil {
			return fmt.Errorf("Error uploading minipool performance file to Web3.Storage: %w", err)
		}
		t.printMessage(fmt.Sprintf("Uploaded minipool performance file with CID %s", minipoolPerformanceCid))
		rewardsFile.MinipoolPerformanceFileCID = minipoolPerformanceCid
	} else {
		t.printMessage("Saved minipool performance file.")
		rewardsFile.MinipoolPerformanceFileCID = "---"
	}

	// Serialize the rewards tree to JSON
	wrapperBytes, err := json.Marshal(rewardsFile)
	if err != nil {
		return fmt.Errorf("Error serializing proof wrapper into JSON: %w", err)
	}
	t.printMessage("Generation complete! Saving tree...")

	// Write the rewards tree to disk
	err = os.WriteFile(rewardsTreePath, wrapperBytes, 0644)
	if err != nil {
		return fmt.Errorf("Error saving rewards tree file to %s: %w", rewardsTreePath, err)
	}

	// Only do the upload and submission process if this is an Oracle DAO node
	if nodeTrusted {
		// Upload the rewards tree file
		t.printMessage("Uploading to Web3.Storage and submitting results to the contracts...")
		cid, err := t.uploadFileToWeb3Storage(wrapperBytes, compressedRewardsTreePath, "compressed rewards tree")
		if err != nil {
			return fmt.Errorf("Error uploading Merkle tree to Web3.Storage: %w", err)
		}
		t.printMessage(fmt.Sprintf("Uploaded Merkle tree with CID %s", cid))

		// Submit to the contracts
		err = t.submitRewardsSnapshot(big.NewInt(int64(currentIndex)), snapshotBeaconBlock, elBlockIndex, rewardsFile, cid, big.NewInt(int64(intervalsPassed)))
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
func (t *submitRewardsTree) submitRewardsSnapshot(index *big.Int, consensusBlock uint64, executionBlock uint64, rewardsFile *rprewards.RewardsFile, cid string, intervalsPassed *big.Int) error {

	treeRootBytes, err := hex.DecodeString(hexutil.RemovePrefix(rewardsFile.MerkleRoot))
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
		networkRewards, exists := rewardsFile.NetworkRewards[network]
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
		TreasuryRPL:     &rewardsFile.TotalRewards.ProtocolDaoRpl.Int,
		NodeRPL:         collateralRplRewards,
		TrustedNodeRPL:  oDaoRplRewards,
		NodeETH:         smoothingPoolEthRewards,
		UserETH:         &rewardsFile.TotalRewards.PoolStakerSmoothingPoolEth.Int,
	}

	// Get the gas limit
	gasInfo, err := rewards.EstimateSubmitRewardSnapshotGas(t.rp, submission, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to submit the rewards tree: %w", err)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(GetWatchtowerMaxFee(t.cfg))
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log, maxFee, 0) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(GetWatchtowerPrioFee(t.cfg))
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

// Compress and upload a file to Web3.Storage and get the CID for it
func (t *submitRewardsTree) uploadFileToWeb3Storage(wrapperBytes []byte, compressedPath string, description string) (string, error) {

	// Get the API token
	apiToken := t.cfg.Smartnode.Web3StorageApiToken.Value.(string)
	if apiToken == "" {
		return "", fmt.Errorf("***ERROR***\nYou have not configured your Web3.Storage API token yet, so you cannot submit Merkle rewards trees.\nPlease get an API token from https://web3.storage and enter it in the Smartnode section of the `service config` TUI (or use `--smartnode-web3StorageApiToken` if you configure your system headlessly).")
	}

	// Create the client
	w3sClient, err := w3s.NewClient(w3s.WithToken(apiToken))
	if err != nil {
		return "", fmt.Errorf("Error creating new Web3.Storage client: %w", err)
	}

	// Compress the file
	encoder, _ := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	compressedBytes := encoder.EncodeAll(wrapperBytes, make([]byte, 0, len(wrapperBytes)))

	// Create the compressed tree file
	compressedFile, err := os.Create(compressedPath)
	if err != nil {
		return "", fmt.Errorf("Error creating %s file [%s]: %w", description, compressedPath, err)
	}
	defer compressedFile.Close()

	// Write the compressed data to the file
	_, err = compressedFile.Write(compressedBytes)
	if err != nil {
		return "", fmt.Errorf("Error writing %s to %s: %w", description, compressedPath, err)
	}

	// Rewind it to the start
	compressedFile.Seek(0, 0)

	// Upload it
	cid, err := w3sClient.Put(context.Background(), compressedFile)
	if err != nil {
		return "", fmt.Errorf("Error uploading %s: %w", description, err)
	}

	return cid.String(), nil

}
