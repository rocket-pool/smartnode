package voting

import (
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	depthPerRound              uint64 = 5
	threadLimit                int    = 6
)

// Struct for generating proposal voting trees and pollards
type VotingTreeGenerator struct {
	rp        *rocketpool.RocketPool
	mcAddress common.Address
}

// Creates a new VotingTreeGenerator instance
func NewVotingTreeGenerator(rp *rocketpool.RocketPool, multicallerAddress common.Address) (*VotingTreeGenerator, error) {
	g := &VotingTreeGenerator{
		rp:        rp,
		mcAddress: multicallerAddress,
	}
	return g, nil
}

// Gets the voting power and delegation info for every node at the specified block
func (g *VotingTreeGenerator) GetNodeVotingInfo(blockNumber uint32, opts *bind.CallOpts) ([]types.NodeVotingInfo, error) {
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

// Gets a complete Pollard row for a new proposal based on the target block number.
func (g *VotingTreeGenerator) CreatePollardRowForProposal(votingInfo []types.NodeVotingInfo) []types.VotingTreeNode {
	// Get the 2D voting power subtree for the main tree
	votingPowers := g.getDelegatedVotingPower(votingInfo)

	// Get the leaf nodes of the tree
	leafNodes := g.constructLeafNodes(votingPowers)

	// Create the Pollard row from the leaf nodes - don't need a proof just for the proposal
	_, nodes := g.generatePollard(leafNodes, 1)
	return nodes
}

// Gets a complete proof and corresponding Pollard row to challenge an existing proposal.
func (g *VotingTreeGenerator) CreatePollardForChallenge(targetIndex uint64, votingInfo []types.NodeVotingInfo) ([]types.VotingTreeNode, []types.VotingTreeNode) {
	// Get the 2D voting power subtree for the main tree
	votingPowers := g.getDelegatedVotingPower(votingInfo)

	// Get the leaf nodes of the tree
	leafNodes := g.constructLeafNodes(votingPowers)

	// Create the proof and Pollard row from the leaf nodes
	proof, nodes := g.generatePollard(leafNodes, targetIndex)
	return proof, nodes
}

// Get the 2D array of voting delegation and total power for each node
func (g *VotingTreeGenerator) getDelegatedVotingPower(votingInfo []types.NodeVotingInfo) [][]*big.Int {
	nodeCount := uint64(len(votingInfo))

	// For each node, create an array of nodes that have delegated to it
	votingPowers := make([][]*big.Int, nodeCount)
	for i := uint64(0); i < nodeCount; i++ {
		nodeAddress := votingInfo[i].NodeAddress

		votingPower := make([]*big.Int, nodeCount)
		for j := uint64(0); j < nodeCount; j++ {
			info := votingInfo[j]
			if info.Delegate == nodeAddress {
				votingPower[j] = info.VotingPower
			} else {
				votingPower[j] = big.NewInt(0)
			}
		}
		votingPowers[i] = votingPower
	}

	// Return
	return votingPowers
}

// Create the complete set of subtree leaf nodes
func (g *VotingTreeGenerator) constructLeafNodes(votingPowers [][]*big.Int) []types.VotingTreeNode {
	nodeCount := uint64(len(votingPowers))
	if nodeCount == 0 {
		return []types.VotingTreeNode{}
	}

	// Create the slice of leaf nodes for the subtree
	subTreeDepth := uint64(math.Ceil(math.Log2(float64(nodeCount))))              // First power of 2 greater than nodeCount
	subTreeLeafCountPerMainTreeNode := uint64(math.Pow(2, float64(subTreeDepth))) // Number of leaf nodes in the sub-tree that correspond to a single leaf of the main tree
	totalSubTreeLeafNodes := subTreeLeafCountPerMainTreeNode * subTreeLeafCountPerMainTreeNode

	// Create the leaf nodes
	leafNodes := make([]types.VotingTreeNode, totalSubTreeLeafNodes)
	for i := uint64(0); i < subTreeLeafCountPerMainTreeNode; i++ {
		for j := uint64(0); j < subTreeLeafCountPerMainTreeNode; j++ {
			index := i*subTreeLeafCountPerMainTreeNode + j
			var balance *big.Int

			// Get the balance if i and j are both in-bounds
			if i < nodeCount && j < nodeCount {
				balance = votingPowers[i][j]
			} else {
				balance = big.NewInt(0)
			}

			leafNode := types.VotingTreeNode{
				Sum:  balance,
				Hash: getHashForBalance(balance),
			}
			leafNodes[index] = leafNode
		}
	}
	return leafNodes
}

// Generates a complete Pollard, either for a new proposal or for a challenge.
// For new proposals use index = 1.
// For challenges, the index is the index of the node being challenged.
// Returns the aggregated proof, and the list of nodes in the pollard row.
func (g *VotingTreeGenerator) generatePollard(leafNodes []types.VotingTreeNode, index uint64) ([]types.VotingTreeNode, []types.VotingTreeNode) {
	order := depthPerRound
	offset := uint64(math.Floor(math.Log2(float64(index)))) // Depth of the node being challenged, if not building a proposal
	depth := uint64(math.Log2(float64(len(leafNodes))))     // Total depth of the tree

	// If the target is out of bounds, bring the order up enough levels to make the target a leaf node
	if order+offset > depth {
		order = depth - offset
	}

	// Get the pollard parameters
	pollardSize := uint64(math.Pow(2, float64(order)))
	pollardDepth := offset + order
	pollardOffset := index*uint64(math.Pow(2, float64(order))) - uint64(math.Pow(2, float64(order+offset)))

	// Get the list of nodes corresponding to the pollard row
	var nodes []types.VotingTreeNode
	if depth == pollardDepth {
		// The pollard row is the last one so just grab the final row from the leaf nodes
		nodes = make([]types.VotingTreeNode, pollardSize)
		copy(nodes, leafNodes[pollardOffset:pollardOffset+pollardSize])
		//nodes = leafNodes[pollardOffset : pollardOffset+pollardSize]
	}
	// The pollard row is above the last one, so crawl up the tree calculating the values of each node until getting to it
	for level := depth; level > offset; level-- {
		n := uint64(math.Pow(2, float64(level)))

		for i := uint64(0); i < n/2; i++ {
			a := i * 2 // Index of the first node
			b := a + 1 // Index of the second node, directly to the right of it
			node := getParentNodeFromChildren(leafNodes[a], leafNodes[b])
			leafNodes[i] = node
		}

		// Slice out the nodes for the pollard once we've reached the right level
		if level-1 == offset+order {
			nodes = make([]types.VotingTreeNode, pollardSize)
			copy(nodes, leafNodes[pollardOffset:pollardOffset+pollardSize])
			//nodes = leafNodes[pollardOffset : pollardOffset+pollardSize]
		}
	}

	// Build a proof from the offset up to the root node
	proof := []types.VotingTreeNode{}
	for level := offset; level > 0; level-- {
		indexOffset := uint64(math.Pow(2, float64(level)))

		for i := uint64(0); i < indexOffset/2; i++ {
			a := i * 2 // Index of the first node
			b := a + 1 // Index of the second node, directly to the right of it

			if indexOffset+a == index {
				proof = append(proof, leafNodes[b])
			} else if indexOffset+b == index {
				proof = append(proof, leafNodes[a])
			}

			leafNodes[i] = getParentNodeFromChildren(leafNodes[a], leafNodes[b])
		}

		index = index / 2
	}

	return proof, nodes
}

// Get all node addresses using a multicaller
func (g *VotingTreeGenerator) getNodeAddressesFast(opts *bind.CallOpts) ([]common.Address, error) {
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

// Get the keccak hash of a parent node with two children
func getParentNodeFromChildren(leftChild types.VotingTreeNode, rightChild types.VotingTreeNode) types.VotingTreeNode {
	leftBuffer := [32]byte{}
	rightBuffer := [32]byte{}
	leftChild.Sum.FillBytes(leftBuffer[:])
	rightChild.Sum.FillBytes(rightBuffer[:])
	hash := crypto.Keccak256Hash(leftChild.Hash[:], leftBuffer[:], rightChild.Hash[:], rightBuffer[:])

	sum := big.NewInt(0).Add(leftChild.Sum, rightChild.Sum)
	return types.VotingTreeNode{
		Hash: hash,
		Sum:  sum,
	}
}

// Get the keccak hash of a balance as a uint256
func getHashForBalance(balance *big.Int) common.Hash {
	buffer := [32]byte{}
	balance.FillBytes(buffer[:])
	hash := crypto.Keccak256Hash(buffer[:])
	return hash
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
