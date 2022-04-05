package watchtower

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/client"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/keccak256"
)

// Settings
const SubmitFollowDistanceRewardsTree = 2

// Node operator rewards
type nodeRewards struct {
	rewardNetwork    uint64   `json:"rewardNetwork,omitempty"`
	collateralRpl    *big.Int `json:"collateralRpl,omitempty"`
	oracleDaoRpl     *big.Int `json:"oracleDaoRpl,omitempty"`
	smoothingPoolEth *big.Int `json:"smoothingPoolEth,omitempty"`
	merkleData       []byte   `json:"-"`
}

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
	nodeRewardsMap, networkRewardsMap, err := t.calculateNodeOperatorRewards(snapshotBlockHeader, intervalTime)

	// Generate the Merkle tree
	tree, err := t.generateMerkleTree(nodeRewardsMap)
	if err != nil {
		return err
	}

	// TODO: serialize the tree, generate proofs, submit it to the chain
	/*

		for address, rewardsForNode := range claims {
			proof, err := tree.GenerateProof(rewardsForNode.merkleData, 0)
			if err != nil {
				panic(fmt.Errorf("error generating proof for node %s: %w", address.Hex(), err))
			}

			//fmt.Printf("Data for Node %s:\n", address.Hex())
			//fmt.Printf("\t0x%s\n", hex.EncodeToString(rewardsForNode.merkleData))

			fmt.Printf("Proof for Node %s:\n", address.Hex())
			for _, hash := range proof.Hashes {
				fmt.Printf("\t0x%s\n", hex.EncodeToString(hash))
			}
			fmt.Println()
		}

		fmt.Printf("Merkle Root: 0x%s\n", hex.EncodeToString(tree.Root()))
	*/

	// Done
	return nil

}

// Generates a merkle tree from the provided rewards map
func (t *submitRewardsTree) generateMerkleTree(nodeRewardsMap map[common.Address]nodeRewards) (*merkletree.MerkleTree, error) {

	// Create a slice of every address
	keys := make([]common.Address, 0, len(nodeRewardsMap))
	for address := range nodeRewardsMap {
		keys = append(keys, address)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return bytes.Compare(keys[i].Bytes(), keys[j].Bytes()) == -1
	})

	totalData := make([][]byte, 0, len(nodeRewardsMap))

	for _, address := range keys {
		rewardsForNode := nodeRewardsMap[address]
		nodeData := make([]byte, 0, 20+32*3)

		addressBytes := address.Bytes()
		nodeData = append(nodeData, addressBytes...)

		network := big.NewInt(0).SetUint64(rewardsForNode.rewardNetwork)
		networkBytes := make([]byte, 32)
		network.FillBytes(networkBytes)
		nodeData = append(nodeData, networkBytes...)

		rplRewards := big.NewInt(0)
		rplRewards.Add(rewardsForNode.collateralRpl, rewardsForNode.oracleDaoRpl)
		rplRewardsBytes := make([]byte, 32)
		rplRewards.FillBytes(rplRewardsBytes)
		nodeData = append(nodeData, rplRewardsBytes...)

		ethRewardsBytes := make([]byte, 32)
		rewardsForNode.smoothingPoolEth.FillBytes(ethRewardsBytes)
		nodeData = append(nodeData, ethRewardsBytes...)

		totalData = append(totalData, nodeData)
	}

	tree, err := merkletree.NewUsing(totalData, keccak256.New(), false, true)
	if err != nil {
		return nil, fmt.Errorf("error generating Merkle Tree: %w", err)
	}

	return tree, nil

}

// Calculates the RPL rewards for regular node operators for this interval
func (t *submitRewardsTree) calculateNodeOperatorRewards(snapshotBlockHeader *types.Header, rewardsInterval time.Duration) (map[common.Address]nodeRewards, map[uint64]nodeRewards, error) {

	nodeRewardsMap := map[common.Address]nodeRewards{}
	networkRewardsMap := map[uint64]nodeRewards{}
	opts := &bind.CallOpts{
		BlockNumber: snapshotBlockHeader.Number,
	}
	snapshotBlockTime := time.Unix(int64(snapshotBlockHeader.Time), 0)

	// Handle node operator rewards
	nodeOpPercent, err := rewards.GetNodeOperatorRewardsPercentRaw(t.rp, opts)
	if err != nil {
		return nil, nil, err
	}
	pendingRewards, err := rewards.GetPendingRPLRewardsRaw(t.rp, opts)
	if err != nil {
		return nil, nil, err
	}
	totalNodeRewards := big.NewInt(0)
	totalNodeRewards.Mul(pendingRewards, nodeOpPercent)
	totalNodeRewards.Div(totalNodeRewards, eth.EthToWei(1))

	nodeAddresses, err := node.GetNodeAddresses(t.rp, opts)
	if err != nil {
		return nil, nil, err
	}
	totalRplStake, err := node.GetTotalEffectiveRPLStake(t.rp, opts)
	if err != nil {
		return nil, nil, err
	}

	for _, address := range nodeAddresses {
		// Make sure this node is eligible for rewards
		regTime, err := rewards.GetNodeRegistrationTime(t.rp, address, opts)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting registration time for node %s: %w", err)
		}
		if snapshotBlockTime.Sub(regTime) < rewardsInterval {
			continue
		}

		// Get how much RPL goes to this node: effective stake / total stake * total RPL rewards for nodes
		nodeStake, err := node.GetNodeEffectiveRPLStake(t.rp, address, opts)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting effective stake for node %s: %w", address.Hex(), err)
		}
		nodeRplRewards := big.NewInt(0)
		nodeRplRewards.Mul(nodeStake, totalNodeRewards)
		nodeRplRewards.Div(nodeRplRewards, totalRplStake)

		// If there are pending rewards, add it to the map
		if nodeRplRewards.Cmp(big.NewInt(0)) == 1 {
			rewardsForNode, exists := nodeRewardsMap[address]
			if !exists {
				// Get the network the rewards should go to
				network, err := node.GetRewardNetwork(t.rp, address, opts)
				if err != nil {
					return nil, nil, err
				}
				if !t.validateNetwork(network) {
					t.log.Printlnf("WARNING: Node %s has an invalid reward network assigned (%d)", address, network)
					continue
				}

				rewardsForNode = nodeRewards{
					rewardNetwork:    network,
					collateralRpl:    big.NewInt(0),
					oracleDaoRpl:     big.NewInt(0),
					smoothingPoolEth: big.NewInt(0),
				}
			}
			rewardsForNode.collateralRpl.Add(rewardsForNode.collateralRpl, nodeRplRewards)
			nodeRewardsMap[address] = rewardsForNode

			// Add the rewards to the running total for the specified network
			rewardsForNetwork, exists := networkRewardsMap[rewardsForNode.rewardNetwork]
			if !exists {
				rewardsForNetwork = nodeRewards{
					rewardNetwork:    rewardsForNode.rewardNetwork,
					collateralRpl:    big.NewInt(0),
					oracleDaoRpl:     big.NewInt(0),
					smoothingPoolEth: big.NewInt(0),
				}
			}
			rewardsForNetwork.collateralRpl.Add(rewardsForNetwork.collateralRpl, nodeRplRewards)
			networkRewardsMap[rewardsForNode.rewardNetwork] = rewardsForNetwork
		}
	}

	// Handle Oracle DAO rewards
	oDaoPercent, err := rewards.GetTrustedNodeOperatorRewardsPercentRaw(t.rp, opts)
	if err != nil {
		return nil, nil, err
	}
	totalODaoRewards := big.NewInt(0)
	totalODaoRewards.Mul(pendingRewards, oDaoPercent)
	totalODaoRewards.Div(totalODaoRewards, eth.EthToWei(1))

	oDaoAddresses, err := trustednode.GetMemberAddresses(t.rp, opts)
	if err != nil {
		return nil, nil, err
	}
	memberCount := big.NewInt(int64(len(oDaoAddresses)))
	individualOdaoRewards := big.NewInt(0)
	individualOdaoRewards.Div(totalODaoRewards, memberCount)

	for _, address := range oDaoAddresses {
		rewardsForNode, exists := nodeRewardsMap[address]
		if !exists {
			// Get the network the rewards should go to
			network, err := node.GetRewardNetwork(t.rp, address, opts)
			if err != nil {
				return nil, nil, err
			}
			if !t.validateNetwork(network) {
				t.log.Printlnf("WARNING: Oracle DAO member %s has an invalid reward network assigned (%d)", address, network)
				continue
			}

			rewardsForNode = nodeRewards{
				rewardNetwork:    network,
				collateralRpl:    big.NewInt(0),
				oracleDaoRpl:     big.NewInt(0),
				smoothingPoolEth: big.NewInt(0),
			}
		}
		rewardsForNode.collateralRpl.Add(rewardsForNode.collateralRpl, individualOdaoRewards)
		nodeRewardsMap[address] = rewardsForNode

		// Add the rewards to the running total for the specified network
		rewardsForNetwork, exists := networkRewardsMap[rewardsForNode.rewardNetwork]
		if !exists {
			rewardsForNetwork = nodeRewards{
				rewardNetwork:    rewardsForNode.rewardNetwork,
				collateralRpl:    big.NewInt(0),
				oracleDaoRpl:     big.NewInt(0),
				smoothingPoolEth: big.NewInt(0),
			}
		}
		rewardsForNetwork.oracleDaoRpl.Add(rewardsForNetwork.oracleDaoRpl, individualOdaoRewards)
		networkRewardsMap[rewardsForNode.rewardNetwork] = rewardsForNetwork
	}

	// Return the rewards maps
	return nodeRewardsMap, networkRewardsMap, nil
}

// Validates that the provided network is legal
func (t *submitRewardsTree) validateNetwork(network uint64) bool {

	// TODO: add more of these as we add L2 support
	switch network {
	case 0:
		return true
	default:
		return false
	}

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
