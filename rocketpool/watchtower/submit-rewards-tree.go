package watchtower

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
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
	c   *cli.Context
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   *wallet.Wallet
	rp  *rocketpool.RocketPool
	ec  *client.EthClientProxy
}

// Create submit rewards Merkle Tree task
func newSubmitRewardsTree(c *cli.Context, logger log.ColorLogger) (*submitRewardsTree, error) {

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
	generator := &submitRewardsTree{
		c:   c,
		log: logger,
		cfg: cfg,
		ec:  ec,
		w:   w,
		rp:  rp,
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

	// Get the total pending rewards and respective distribution percentages
	nodeRewardsMap, networkRewardsMap, invalidNodeNetworks, err := rprewards.CalculateRplRewards(t.rp, snapshotBlockHeader, intervalTime)
	if err != nil {
		return fmt.Errorf("error calculating node operator rewards: %w", err)
	}
	for address, network := range invalidNodeNetworks {
		t.log.Printlnf("WARNING: Node %s has invalid network %d assigned!\n", address.Hex(), network)
	}

	// Generate the Merkle tree
	tree, err := rprewards.GenerateMerkleTree(nodeRewardsMap)
	if err != nil {
		return fmt.Errorf("error generating Merkle tree: %w", err)
	}

	// Create the JSON proof wrapper and encode it
	proofWrapper := rprewards.GenerateTreeJson(tree.Root(), nodeRewardsMap, networkRewardsMap)
	wrapperBytes, err := json.Marshal(proofWrapper)
	if err != nil {
		return fmt.Errorf("error serializing proof wrapper into JSON: %w", err)
	}

	// Write the file (TEMP)
	ioutil.WriteFile("rocket-pool-rewards-0.json", wrapperBytes, 0755)

	// Done
	return nil

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
