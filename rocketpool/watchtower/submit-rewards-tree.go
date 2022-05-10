package watchtower

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/klauspost/compress/zstd"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
	"github.com/web3-storage/go-w3s-client"
)

// Submit rewards Merkle Tree task
type submitRewardsTree struct {
	c              *cli.Context
	log            log.ColorLogger
	errLog         log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              *wallet.Wallet
	rp             *rocketpool.RocketPool
	ec             rocketpool.ExecutionClient
	bc             beacon.Client
	lock           *sync.Mutex
	isRunning      bool
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
}

// Create submit rewards Merkle Tree task
func newSubmitRewardsTree(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger) (*submitRewardsTree, error) {

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

	// Get the user-requested max fee
	maxFeeGwei := cfg.Smartnode.ManualMaxFee.Value.(float64)
	var maxFee *big.Int
	if maxFeeGwei == 0 {
		maxFee = nil
	} else {
		maxFee = eth.GweiToWei(maxFeeGwei)
	}

	// Get the user-requested max fee
	priorityFeeGwei := cfg.Smartnode.PriorityFee.Value.(float64)
	var priorityFee *big.Int
	if priorityFeeGwei == 0 {
		logger.Println("WARNING: priority fee was missing or 0, setting a default of 2.")
		priorityFee = eth.GweiToWei(2)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	lock := &sync.Mutex{}
	generator := &submitRewardsTree{
		c:              c,
		log:            logger,
		errLog:         errorLogger,
		cfg:            cfg,
		ec:             ec,
		bc:             bc,
		w:              w,
		rp:             rp,
		lock:           lock,
		isRunning:      false,
		maxFee:         maxFee,
		maxPriorityFee: priorityFee,
		gasLimit:       0,
	}

	return generator, nil
}

// Submit rewards Merkle Tree
func (t *submitRewardsTree) run() error {

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
	nodeTrusted, err := trustednode.GetMemberExists(t.rp, nodeAccount.Address, nil)
	if err != nil {
		return err
	}
	if !nodeTrusted {
		return nil
	}

	// Log
	t.log.Println("Checking for rewards checkpoint...")

	// Check if a rewards interval has passed and needs to be calculated
	startTime, err := rewards.GetClaimIntervalTimeStart(t.rp, nil)
	if err != nil {
		return fmt.Errorf("error getting claim interval start time: %w", err)
	}
	intervalTime, err := rewards.GetClaimIntervalTime(t.rp, nil)
	if err != nil {
		return fmt.Errorf("error getting claim interval time: %w", err)
	}

	// Calculate the end time, which is the number of intervals that have gone by since the current one's start
	latestBlockHeader, err := t.ec.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error getting latest block header: %w", err)
	}
	latestBlockTime := time.Unix(int64(latestBlockHeader.Time), 0)
	timeSinceStart := latestBlockTime.Sub(startTime)
	intervalsPassed := timeSinceStart / intervalTime
	endTime := startTime.Add(intervalTime * intervalsPassed)
	if intervalsPassed == 0 {
		return nil
	}

	// Get the block and timestamp of the consensus block that best matches the end time
	snapshotBeaconBlock, snapshotBeaconBlockTime, err := t.getSnapshotConsensusBlock(endTime)
	if err != nil {
		return err
	}

	// Get the number of the EL block matching the CL snapshot block
	snapshotElBlockHeader, err := rprewards.GetELBlockHeaderForTime(snapshotBeaconBlockTime, t.ec)
	if err != nil {
		return err
	}

	// Get the current interval
	currentIndexBig, err := rewards.GetRewardIndex(t.rp, nil)
	if err != nil {
		return err
	}
	currentIndex := currentIndexBig.Uint64()

	// Check if rewards generation is already running
	t.lock.Lock()
	if t.isRunning {
		t.log.Println("Tree generation is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	// Check if the file is already generated and reupload it without rebuilding it
	path := t.cfg.Smartnode.GetRewardsTreePath(currentIndex, true)
	compressedPath := t.cfg.Smartnode.GetCompressedRewardsTreePath(currentIndex, true)
	_, err = os.Stat(path)
	if !os.IsNotExist(err) {
		// Return if this node has already submitted the tree for the current interval and there's a file present
		hasSubmitted, err := t.hasSubmittedTree(nodeAccount.Address, currentIndexBig)
		if err != nil {
			return fmt.Errorf("error checking if Merkle tree submission has already been processed: %w", err)
		}
		if hasSubmitted {
			return nil
		}

		t.log.Printlnf("Merkle rewards tree for interval %d already exists at %s, attempting to resubmit...", currentIndex, path)

		// Deserialize the file
		wrapperBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("Error reading rewards tree file: %w", err)
		}

		proofWrapper := new(rprewards.ProofWrapper)
		err = json.Unmarshal(wrapperBytes, proofWrapper)
		if err != nil {
			return fmt.Errorf("Error deserializing rewards tree file: %w", err)
		}

		// Upload the file
		cid, err := t.uploadRewardsTreeToWeb3Storage(wrapperBytes, compressedPath)
		if err != nil {
			return fmt.Errorf("Error uploading Merkle tree to Web3.Storage: %w", err)
		}
		t.log.Printlnf("Uploaded Merkle tree with CID %s", cid)

		// Submit to the contracts
		err = t.submitRewardsSnapshot(currentIndexBig, snapshotBeaconBlock, proofWrapper, cid, big.NewInt(int64(intervalsPassed)))
		if err != nil {
			return fmt.Errorf("Error submitting rewards snapshot: %w", err)
		}

		t.log.Printlnf("Successfully submitted rewards snapshot for interval %d.", currentIndex)
		return nil
	}

	// Run the tree generation
	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()

		// Log
		generationPrefix := "[Merkle Tree]"
		if int64(intervalsPassed) > 1 {
			t.log.Printlnf("WARNING: %d intervals have passed since the last rewards checkpoint was submitted! Rolling them into one...", int64(intervalsPassed))
		}
		t.log.Printlnf("Rewards checkpoint has passed, starting Merkle tree generation for interval %d in the background.\n%s Snapshot Beacon block = %d, EL block = %d, running from %s to %s", currentIndex, generationPrefix, snapshotBeaconBlock, snapshotElBlockHeader.Number.Uint64(), startTime, endTime)

		// Get the total pending rewards and respective distribution percentages
		nodeRewardsMap, networkRewardsMap, invalidNodeNetworks, err := rprewards.CalculateRplRewards(t.rp, snapshotElBlockHeader, intervalTime)
		if err != nil {
			t.handleError(fmt.Errorf("%s Error calculating node operator rewards: %w", generationPrefix, err))
			return
		}
		for address, network := range invalidNodeNetworks {
			t.log.Printlnf("%s WARNING: Node %s has invalid network %d assigned!", generationPrefix, address.Hex(), network)
		}

		// Generate the Merkle tree
		tree, err := rprewards.GenerateMerkleTree(nodeRewardsMap)
		if err != nil {
			t.handleError(fmt.Errorf("%s Error generating Merkle tree: %w", generationPrefix, err))
			return
		}

		// Create the JSON proof wrapper and encode it
		proofWrapper := rprewards.GenerateTreeJson(tree.Root(), nodeRewardsMap, networkRewardsMap)
		wrapperBytes, err := json.Marshal(proofWrapper)
		if err != nil {
			t.handleError(fmt.Errorf("%s Error serializing proof wrapper into JSON: %w", generationPrefix, err))
			return
		}
		t.log.Println(fmt.Sprintf("%s Generation complete! Saving and uploading...", generationPrefix))

		// Write the file
		path := t.cfg.Smartnode.GetRewardsTreePath(currentIndex, true)
		err = ioutil.WriteFile(path, wrapperBytes, 0644)
		if err != nil {
			t.handleError(fmt.Errorf("%s Error saving file to %s: %w", generationPrefix, path, err))
			return
		}

		// Upload the file
		cid, err := t.uploadRewardsTreeToWeb3Storage(wrapperBytes, compressedPath)
		if err != nil {
			t.handleError(fmt.Errorf("%s Error uploading Merkle tree to Web3.Storage: %w", generationPrefix, err))
			return
		}
		t.log.Printlnf("%s Uploaded Merkle tree with CID %s", generationPrefix, cid)

		// Submit to the contracts
		err = t.submitRewardsSnapshot(currentIndexBig, snapshotBeaconBlock, proofWrapper, cid, big.NewInt(int64(intervalsPassed)))
		if err != nil {
			t.handleError(fmt.Errorf("%s Error submitting rewards snapshot: %w", generationPrefix, err))
			return
		}

		t.log.Printlnf("%s Successfully submitted rewards snapshot for interval %d.", generationPrefix, currentIndex)
		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

	// Done
	return nil

}

func (t *submitRewardsTree) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Rewards tree generation failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Submit rewards info to the contracts
func (t *submitRewardsTree) submitRewardsSnapshot(index *big.Int, consensusBlock uint64, proofWrapper *rprewards.ProofWrapper, cid string, intervalsPassed *big.Int) error {

	consensusBlockBig := big.NewInt(0).SetUint64(consensusBlock)
	treeRootBytes, err := hex.DecodeString(hexutil.RemovePrefix(proofWrapper.MerkleRoot))
	if err != nil {
		return fmt.Errorf("Error decoding merkle root: %w", err)
	}
	treeRoot := common.BytesToHash(treeRootBytes)

	// Create the array of RPL rewards per network
	rplRewards := []*big.Int{}

	// TODO: OTHER NETWORK SUPPORT
	mainnetRplRewards := big.NewInt(0)
	mainnetRplRewards.Add(mainnetRplRewards, proofWrapper.NetworkRewards.CollateralRplPerNetwork[0])
	mainnetRplRewards.Add(mainnetRplRewards, proofWrapper.NetworkRewards.OracleDaoRplPerNetwork[0])
	rplRewards = append(rplRewards, mainnetRplRewards)

	// Create the array of ETH rewards per network
	ethRewards := []*big.Int{}

	// TODO: OTHER NETWORK SUPPORT
	mainnetEthRewards := big.NewInt(0)
	mainnetEthRewards.Add(mainnetEthRewards, proofWrapper.NetworkRewards.SmoothingPoolEthPerNetwork[0])
	ethRewards = append(ethRewards, mainnetEthRewards)

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := rewards.EstimateSubmitRewardSnapshotGas(t.rp, index, consensusBlockBig, rplRewards, ethRewards, treeRoot, cid, intervalsPassed, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to submit the rewards tree: %w", err)
	}
	var gas *big.Int
	if t.gasLimit != 0 {
		gas = new(big.Int).SetUint64(t.gasLimit)
	} else {
		gas = new(big.Int).SetUint64(gasInfo.SafeGasLimit)
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWei()
		if err != nil {
			return err
		}
	}

	// Print the gas info
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee
	opts.GasLimit = gas.Uint64()

	// Submit RPL price
	hash, err := rewards.SubmitRewardSnapshot(t.rp, index, consensusBlockBig, rplRewards, ethRewards, treeRoot, cid, intervalsPassed, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be mined
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Return
	return nil
}

// Upload the Merkle rewards tree to Web3.Storage and get the CID for it
func (t *submitRewardsTree) uploadRewardsTreeToWeb3Storage(wrapperBytes []byte, compressedPath string) (string, error) {

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
		return "", fmt.Errorf("Error creating compressed rewards tree file [%s]: %w", compressedPath, err)
	}
	defer compressedFile.Close()

	// Write the compressed data to the tree file
	_, err = compressedFile.Write(compressedBytes)
	if err != nil {
		return "", fmt.Errorf("Error writing compressed rewards tree to %s: %w", compressedPath, err)
	}

	// Rewind it to the start
	compressedFile.Seek(0, 0)

	// Upload it
	cid, err := w3sClient.Put(context.Background(), compressedFile)
	if err != nil {
		return "", fmt.Errorf("Error uploading compressed rewards tree: %w", err)
	}

	return cid.String(), nil

}

// Get the first finalized, successful consensus block that occurred after the given target time
func (t *submitRewardsTree) getSnapshotConsensusBlock(endTime time.Time) (uint64, time.Time, error) {

	// Get the config
	eth2Config, err := t.bc.GetEth2Config()
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("Error getting Beacon config: %w", err)
	}

	// Get the beacon head
	beaconHead, err := t.bc.GetBeaconHead()
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("Error getting Beacon head: %w", err)
	}

	// Get the target block number
	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)
	totalTimespan := endTime.Sub(genesisTime)
	targetSlot := uint64(math.Ceil(totalTimespan.Seconds() / float64(eth2Config.SecondsPerSlot)))
	targetEpoch := targetSlot / eth2Config.SlotsPerEpoch

	// Check if the target epoch is finalized yet
	if beaconHead.FinalizedEpoch < targetEpoch {
		return 0, time.Time{}, fmt.Errorf("Snapshot end time = %s, slot (epoch) = %d (%d) but the latest finalized epoch is %d... waiting until the snapshot slot is finalized.", endTime, targetSlot, targetEpoch, beaconHead.FinalizedEpoch)
	}

	// Get the first successful block
	for {
		// Try to get the current block
		_, exists, err := t.bc.GetEth1DataForEth2Block(fmt.Sprint(targetSlot))
		if err != nil {
			return 0, time.Time{}, fmt.Errorf("Error getting Beacon block %d: %w", targetSlot, err)
		}

		// If the block was missing, try the next one (and make sure its epoch is finalized)
		if !exists {
			t.log.Printlnf("Slot %d was missing, trying the next one...", targetSlot)
			targetSlot++
			newEpoch := targetSlot / eth2Config.SlotsPerEpoch
			if newEpoch != targetEpoch {
				if beaconHead.FinalizedEpoch < targetEpoch {
					return 0, time.Time{}, fmt.Errorf("Snapshot end time = %s, slot (epoch) = %d (%d) but the latest finalized epoch is %d... waiting until the snapshot slot is finalized.", endTime, targetSlot, newEpoch, beaconHead.FinalizedEpoch)
				}
			}
			targetEpoch = newEpoch
			continue
		}

		// Ok, we have the first proposed finalized block - this is the one to use for the snapshot!
		blockTime := genesisTime.Add(time.Duration(targetSlot*eth2Config.SecondsPerSlot) * time.Second)
		return targetSlot, blockTime, nil
	}

}

// Check whether the rewards tree for the current interval been submitted by the node
func (t *submitRewardsTree) hasSubmittedTree(nodeAddress common.Address, index *big.Int) (bool, error) {
	indexBuffer := make([]byte, 32)
	index.FillBytes(indexBuffer)
	return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("rewards.snapshot.submitted.node"), nodeAddress.Bytes(), indexBuffer))
}
