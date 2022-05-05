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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
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
		w:         w,
		rp:        rp,
		lock:      lock,
		isRunning: false,
	}

	return generator, nil
}

// Submit rewards Merkle Tree
func (t *submitRewardsTree) run() error {

	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
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

	// Get the number of the snapshot block which ended the rewards interval
	snapshotBlockHeader, err := t.getBlockHeaderForTime(endTime, latestBlockHeader.Number)
	if err != nil {
		return err
	}

	// Allow some blocks to pass in case of a short reorg
	blockWithBuffer := big.NewInt(SubmitFollowDistanceRewardsTree)
	blockWithBuffer.Add(snapshotBlockHeader.Number, blockWithBuffer)
	if blockWithBuffer.Cmp(latestBlockHeader.Number) == 1 {
		return nil
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
		t.log.Printlnf("Rewards checkpoint has passed, starting Merkle tree generation in the background... snapshot block = %d, running from %s to %s", snapshotBlockHeader.Number.Uint64(), startTime, endTime)
		t.lock.Unlock()

		generationPrefix := "[Merkle Tree]"

		// Get the total pending rewards and respective distribution percentages
		nodeRewardsMap, networkRewardsMap, invalidNodeNetworks, err := rprewards.CalculateRplRewards(t.rp, snapshotBlockHeader, intervalTime)
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

// Get the number of the first block that was created after the given timestamp
func (t *submitRewardsTree) getBlockHeaderForTime(targetTime time.Time, latestBlock *big.Int) (*types.Header, error) {
	// Start at the halfway point
	candidateBlockNumber := big.NewInt(0).Div(latestBlock, big.NewInt(2))
	candidateBlock, err := t.ec.HeaderByNumber(context.Background(), candidateBlockNumber)
	if err != nil {
		return nil, err
	}
	bestBlock := candidateBlock
	pivotSize := candidateBlock.Number.Uint64()
	minimumDistance := +math.Inf(1)
	targetTimeUnix := float64(targetTime.Unix())

	for {
		// Get the distance from the candidate block to the target time
		candidateTime := float64(candidateBlock.Time)
		delta := targetTimeUnix - candidateTime
		distance := math.Abs(delta)

		// If it's better, replace the best candidate with it
		if distance < minimumDistance {
			minimumDistance = distance
			bestBlock = candidateBlock
		} else if pivotSize == 1 {
			// If the pivot is down to size 1 and we didn't find anything better after another iteration, this is the best block!

			// If this block happened before the target timestamp, return the one after it.
			if bestBlock.Time < uint64(targetTime.Unix()) {
				return t.ec.HeaderByNumber(context.Background(), bestBlock.Number.Add(bestBlock.Number, big.NewInt(1)))
			}
			return bestBlock, nil
		}

		// Iterate over the correct half, setting the pivot to the halfway point of that half (rounded up)
		pivotSize = uint64(math.Ceil(float64(pivotSize) / 2))
		if delta < 0 {
			// Go left
			candidateBlockNumber = big.NewInt(0).Sub(candidateBlockNumber, big.NewInt(int64(pivotSize)))
		} else {
			// Go right
			candidateBlockNumber = big.NewInt(0).Add(candidateBlockNumber, big.NewInt(int64(pivotSize)))
		}

		// Clamp the new candidate to the latest block
		if candidateBlockNumber.Uint64() > (latestBlock.Uint64() - 1) {
			candidateBlockNumber.SetUint64(latestBlock.Uint64() - 1)
		}

		candidateBlock, err = t.ec.HeaderByNumber(context.Background(), candidateBlockNumber)
		if err != nil {
			return nil, err
		}
	}
}

// Check whether the rewards tree for the current interval been submitted by the node
func (t *submitRewardsTree) hasSubmittedTree(nodeAddress common.Address, index *big.Int) (bool, error) {
	indexBuffer := make([]byte, 32)
	index.FillBytes(indexBuffer)
	return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("rewards.snapshot.submitted.node"), nodeAddress.Bytes(), indexBuffer))
}
