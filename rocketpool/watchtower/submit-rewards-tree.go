package watchtower

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)

// Settings
const SubmitFollowDistanceRewardsTree = 2

// Submit rewards Merkle Tree task
type submitRewardsTree struct {
	c         *cli.Context
	log       log.ColorLogger
	errLog    log.ColorLogger
	cfg       *config.RocketPoolConfig
	w         *wallet.Wallet
	rp        *rocketpool.RocketPool
	ec        rocketpool.ExecutionClient
	bc        beacon.Client
	lock      *sync.Mutex
	isRunning bool
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
	lock := &sync.Mutex{}
	generator := &submitRewardsTree{
		c:         c,
		log:       logger,
		errLog:    errorLogger,
		cfg:       cfg,
		ec:        ec,
		bc:        bc,
		w:         w,
		rp:        rp,
		lock:      lock,
		isRunning: false,
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
	if time.Until(endTime) > 0 {
		return nil
	} else if int64(intervalsPassed) > 1 {
		t.log.Printlnf("WARNING: %d intervals have passed since the last rewards checkpoint was submitted! Rolling them into one...", int64(intervalsPassed))
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

	// Return if this node has already submitted the tree for the current interval
	hasSubmitted, err := t.hasSubmittedTree(nodeAccount.Address, currentIndexBig)
	if err != nil {
		return fmt.Errorf("error checking if Merkle tree submission has already been processed: %w", err)
	}
	if hasSubmitted {
		return nil
	}

	// Check if rewards generation is already running
	t.lock.Lock()
	if t.isRunning {
		t.log.Println("Tree generation is already running in the background.")
		return nil
	}
	t.lock.Unlock()

	// Run the tree generation
	go func() {
		// Log
		t.lock.Lock()
		t.log.Printlnf("Rewards checkpoint has passed, starting Merkle tree generation in the background... snapshot Beacon block = %d, EL block = %d, running from %s to %s", snapshotBeaconBlock, snapshotElBlockHeader.Number.Uint64(), startTime, endTime)
		t.lock.Unlock()

		generationPrefix := "[Merkle Tree]"

		// Get the total pending rewards and respective distribution percentages
		nodeRewardsMap, networkRewardsMap, invalidNodeNetworks, err := rprewards.CalculateRplRewards(t.rp, snapshotElBlockHeader, intervalTime)
		if err != nil {
			t.handleError(fmt.Errorf("%s Error calculating node operator rewards: %w", generationPrefix, err))
			return
		}
		for address, network := range invalidNodeNetworks {
			t.log.Printlnf("%s WARNING: Node %s has invalid network %d assigned!\n", generationPrefix, address.Hex(), network)
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

		// Write the file
		path := t.cfg.Smartnode.GetRewardsTreePath(currentIndex)
		err = ioutil.WriteFile(path, wrapperBytes, 0644)
		if err != nil {
			t.handleError(fmt.Errorf("%s Error saving file to %s: %w", generationPrefix, path, err))
			return
		}

		// Upload the file
		// TODO

		// Submit to the contracts
		// TODO

		t.log.Println(fmt.Sprintf("%s Generation complete! CID = [Placeholder]", generationPrefix))
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
