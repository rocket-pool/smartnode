package watchtower

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/client"
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
	ec        *client.EthClientProxy
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
	ec, err := services.GetEthClientProxy(c)
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
	endTime := startTime.Add(intervalTime)
	if time.Until(endTime) > 0 {
		return nil
	}

	// Get the number of the snapshot block which ended the rewards interval
	latestBlockHeader, err := t.ec.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error getting latest block header: %w", err)
	}
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
		t.log.Println("Rewards checkpoint has passed, starting Merkle tree generation in the background...")
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
		err = ioutil.WriteFile(path, wrapperBytes, 0755)
		if err != nil {
			t.handleError(fmt.Errorf("%s Error saving file to %s: %w", generationPrefix, path, err))
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

// Get the number of the first block after the given time
func (t *submitRewardsTree) getBlockHeaderForTime(targetTime time.Time, candidateNumber *big.Int) (*types.Header, error) {

	blockNumber := candidateNumber
	one := big.NewInt(1)

	for {
		// Get the preceding block
		previousNumber := big.NewInt(0).Sub(blockNumber, one)
		previousBlock, err := t.ec.HeaderByNumber(context.Background(), previousNumber)
		if err != nil {
			return nil, fmt.Errorf("error getting header for block %s : %w", previousNumber.String(), err)
		}

		previousBlockTime := time.Unix(int64(previousBlock.Time), 0)
		if targetTime.Sub(previousBlockTime) > 0 {
			// This block happened before the end, so return the prior candidate
			return previousBlock, nil
		}

		blockNumber = previousNumber
	}

}

// Check whether the rewards tree for the current interval been submitted by the node
func (t *submitRewardsTree) hasSubmittedTree(nodeAddress common.Address, index *big.Int) (bool, error) {
	indexBuffer := make([]byte, 32)
	index.FillBytes(indexBuffer)
	return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("rewards.snapshot.submitted.node"), nodeAddress.Bytes(), indexBuffer))
}
