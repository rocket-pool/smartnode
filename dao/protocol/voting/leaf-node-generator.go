package voting

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	nodeVotingDetailsBatchSize uint64 = 250
	nodeAddressBatchSize       int    = 1000
	threadLimit                int    = 6
)

// Struct for generating the leaf nodes required to create voting trees
type LeafNodeGenerator struct {
	rp        *rocketpool.RocketPool
	mcAddress common.Address
}

// Creates a new LeafNodeGenerator instance
func NewLeafNodeGenerator(rp *rocketpool.RocketPool, multicallerAddress common.Address) (*LeafNodeGenerator, error) {
	g := &LeafNodeGenerator{
		rp:        rp,
		mcAddress: multicallerAddress,
	}
	return g, nil
}

// Gets the voting power and delegation info for every node at the specified block
func (g *LeafNodeGenerator) GetNodeVotingInfo(blockNumber uint32, opts *bind.CallOpts) ([]types.NodeVotingInfo, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(g.rp, nil)
	if err != nil {
		return nil, err
	}

	// Get the number of voting nodes
	nodeCountBig, err := network.GetVotingNodeCount(g.rp, blockNumber, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting voting node count: %w", err)
	}
	nodeCount := nodeCountBig.Uint64()

	// Get the node addresses
	nodeAddresses, err := g.getNodeAddressesFast(opts)
	if err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)

	// Run the getters in batches
	votingInfos := make([]types.NodeVotingInfo, nodeCount)
	for i := uint64(0); i < nodeCount; i += nodeVotingDetailsBatchSize {
		i := i
		max := i + nodeVotingDetailsBatchSize
		if max > nodeCount {
			max = nodeCount
		}

		// Load details
		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(g.rp.Client, g.mcAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				nodeAddress := nodeAddresses[j]
				votingInfos[j].NodeAddress = nodeAddress
				mc.AddCall(rocketNetworkVoting, &votingInfos[j].VotingPower, "getVotingPower", nodeAddress, blockNumber)
				mc.AddCall(rocketNetworkVoting, &votingInfos[j].Delegate, "getDelegate", nodeAddress, blockNumber)
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	return votingInfos, nil
}

// Get the leaves of a Network Voting Power tree based on node voting info
func (g *LeafNodeGenerator) CreateLeavesForNetwork(infos []types.NodeVotingInfo) []*types.VotingTreeNode {
	// Create a map of the voting power of each node, accounting for delegation
	votingPower := map[common.Address]*big.Int{}
	for _, info := range infos {
		delegateVp, exists := votingPower[info.Delegate]
		if !exists {
			delegateVp = big.NewInt(0)
			votingPower[info.Delegate] = delegateVp
		}
		delegateVp.Add(delegateVp, info.VotingPower)
	}

	// Make the tree leaves
	leaves := make([]*types.VotingTreeNode, len(infos))
	zeroHash := getHashForBalance(common.Big0)
	for i, info := range infos {
		vp, exists := votingPower[info.NodeAddress]
		if !exists || vp.Cmp(common.Big0) == 0 {
			leaves[i] = &types.VotingTreeNode{
				Sum:  big.NewInt(0),
				Hash: zeroHash,
			}
		} else {
			leaves[i] = &types.VotingTreeNode{
				Sum:  big.NewInt(0).Set(vp),
				Hash: getHashForBalance(vp),
			}
		}
	}
	return leaves
}

// Get the leaves of a Node Voting Power tree based on node voting info
func (g *LeafNodeGenerator) CreateLeavesForNode(infos []types.NodeVotingInfo, address common.Address) []*types.VotingTreeNode {
	leaves := make([]*types.VotingTreeNode, len(infos))
	zeroHash := getHashForBalance(common.Big0)
	for i, info := range infos {
		if info.Delegate == address {
			leaves[i] = &types.VotingTreeNode{
				Sum:  info.VotingPower,
				Hash: getHashForBalance(info.VotingPower),
			}
		} else {
			leaves[i] = &types.VotingTreeNode{
				Sum:  big.NewInt(0),
				Hash: zeroHash,
			}
		}
	}
	return leaves
}

// Get all node addresses using a multicaller
func (g *LeafNodeGenerator) getNodeAddressesFast(opts *bind.CallOpts) ([]common.Address, error) {
	rocketNodeManager, err := getRocketNodeManager(g.rp, opts)
	if err != nil {
		return nil, err
	}

	// Get minipool count
	nodeCount, err := node.GetNodeCount(g.rp, opts)
	if err != nil {
		return []common.Address{}, err
	}

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	addresses := make([]common.Address, nodeCount)

	// Run the getters in batches
	count := int(nodeCount)
	for i := 0; i < count; i += nodeAddressBatchSize {
		i := i
		max := i + nodeAddressBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(g.rp.Client, g.mcAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				mc.AddCall(rocketNodeManager, &addresses[j], "getNodeAt", big.NewInt(int64(j)))
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}

	return addresses, nil
}

// Get contracts
var rocketNodeManagerLock sync.Mutex
var rocketNetworkVotingLock sync.Mutex

func getRocketNodeManager(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNodeManagerLock.Lock()
	defer rocketNodeManagerLock.Unlock()
	return rp.GetContract("rocketNodeManager", opts)
}

func getRocketNetworkVoting(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkVotingLock.Lock()
	defer rocketNetworkVotingLock.Unlock()
	return rp.GetContract("rocketNetworkVoting", opts)
}
