package voting

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"golang.org/x/sync/errgroup"
)

const (
	NodeVotingDetailsBatchSize uint64 = 20
	DepthPerRound              uint64 = 5
)

type NodeVotingInfo struct {
	NodeAddress common.Address
	VotingPower *big.Int
	Delegate    common.Address
}

type VotingTreeGenerator struct {
	rp *rocketpool.RocketPool
}

func NewVotingTreeGenerator(rp *rocketpool.RocketPool) (*VotingTreeGenerator, error) {
	g := &VotingTreeGenerator{
		rp: rp,
	}
	return g, nil
}

// Gets a complete Pollard row for a new proposal based on the target block number.
func (g *VotingTreeGenerator) CreatePollardRowForProposal(blockNumber uint32, opts *bind.CallOpts) ([]types.VotingTreeNode, error) {
	// Get an nxn array of each node's voting power and delegating status
	votingPowers, err := g.getDelegatedVotingPower(blockNumber, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting voting power details: %w", err)
	}

	// Get the leaf nodes of the tree
	leaves := g.constructLeaves(votingPowers)

	// Create the Pollard row from the leaf nodes - don't need a proof just for the proposal
	_, nodes := g.generatePollard(leaves, 1)
	return nodes, nil
}

// Gets a complete proof and corresponding Pollard row to challenge an existing proposal.
func (g *VotingTreeGenerator) CreatePollardForChallenge(blockNumber uint32, targetIndex uint64, opts *bind.CallOpts) ([]types.VotingTreeNode, []types.VotingTreeNode, error) {
	// Get an nxn array of each node's voting power and delegating status
	votingPowers, err := g.getDelegatedVotingPower(blockNumber, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting voting power details: %w", err)
	}

	// Get the leaf nodes of the tree
	leaves := g.constructLeaves(votingPowers)

	// Create the proof and Pollard row from the leaf nodes
	proof, nodes := g.generatePollard(leaves, targetIndex)
	return proof, nodes, nil
}

func (g *VotingTreeGenerator) getDelegatedVotingPower(blockNumber uint32, opts *bind.CallOpts) ([][]*big.Int, error) {
	// Get the number of voting nodes
	nodeCountBig, err := network.GetVotingNodeCount(g.rp, blockNumber, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting voting node count: %w", err)
	}
	nodeCount := nodeCountBig.Uint64()

	// Get the node addresses
	nodeAddresses, err := node.GetNodeAddresses(g.rp, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}

	// Load node voting details in batches
	votingInfos := make([]NodeVotingInfo, nodeCount)
	//delegateIndices := map[common.Address]uint64{}
	for bsi := uint64(0); bsi < nodeCount; bsi += NodeVotingDetailsBatchSize {

		// Get batch start & end index
		nsi := bsi
		nei := bsi + NodeVotingDetailsBatchSize
		if nei > nodeCount {
			nei = nodeCount
		}

		// Load details
		var wg errgroup.Group
		for ni := nsi; ni < nei; ni++ {
			ni := ni
			wg.Go(func() error {
				nodeAddress := nodeAddresses[ni]
				votingPower, err := network.GetVotingPower(g.rp, nodeAddress, blockNumber, opts)
				if err != nil {
					return err
				}
				delegate, err := network.GetVotingDelegate(g.rp, nodeAddress, blockNumber, opts)
				if err != nil {
					return err
				}
				votingInfos[ni] = NodeVotingInfo{
					NodeAddress: nodeAddress,
					VotingPower: votingPower,
					Delegate:    delegate,
				}
				//delegateIndices[nodeAddress] = ni
				return nil
			})
		}
		if err := wg.Wait(); err != nil {
			return nil, err
		}

	}

	// For each node, create an array of nodes that have delegated to it
	votingPowers := make([][]*big.Int, nodeCount)
	for i := uint64(0); i < nodeCount; i++ {
		nodeAddress := nodeAddresses[i]

		votingPower := make([]*big.Int, nodeCount)
		for j := uint64(0); j < nodeCount; j++ {
			votingInfo := votingInfos[j]
			if votingInfo.Delegate == nodeAddress {
				votingPower[j] = votingInfo.VotingPower
			} else {
				votingPower[j] = big.NewInt(0)
			}
		}
		votingPowers[i] = votingPower
	}

	// Return
	return votingPowers, nil
}

// Create the complete set of subtree leaf nodes
func (g *VotingTreeGenerator) constructLeaves(votingPowers [][]*big.Int) []types.VotingTreeNode {
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
	order := DepthPerRound
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
