package voting

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/types"
)

const (
	depthPerRound uint64 = 5
)

type NetworkVotingPowerTree struct {
	nodes            []*types.VotingTreeNode
	depth            uint64
	virtualRootIndex uint64
}

// Creates a new NetworkVotingPowerTree instance from leaf nodes.
func CreateTreeFromLeaves(leaves []*types.VotingTreeNode, virtualRootIndex uint64) *NetworkVotingPowerTree {
	// Determine the total number of nodes from the leaf count
	originalPower := math.Log2(float64(len(leaves)))
	ceilingPower := int(math.Ceil(originalPower))
	totalLeafNodes := int(math.Pow(2.0, float64(ceilingPower)))

	// Create the new tree, which is internally managed as an array since it never changes and that provides the fastest indexing
	nodes := make([]*types.VotingTreeNode, totalLeafNodes*2-1)

	// Copy the leaves to the end of the tree
	leafStart := totalLeafNodes - 1
	copy(nodes[leafStart:], leaves)

	// Add padding to the end if the size of leaves isn't a power of 2
	if totalLeafNodes != len(leaves) {
		zeroHash := getHashForBalance(common.Big0)
		for i := leafStart + len(leaves); i < len(nodes); i++ {
			nodes[i] = &types.VotingTreeNode{
				Sum:  big.NewInt(0),
				Hash: zeroHash,
			}
		}
	}

	// Make the tree from the leaves
	currentLevel := ceilingPower - 1
	for i := currentLevel; i >= 0; i-- {
		// Go through each level (row) of the tree linearly
		levelLength := int(math.Pow(2.0, float64(i)))
		startIndex := levelLength - 1
		endIndex := startIndex + levelLength
		for j := startIndex; j < endIndex; j++ {
			// Create the node from its children below
			leftChildIndex := j*2 + 1
			rightChildIndex := leftChildIndex + 1
			nodes[j] = getParentNodeFromChildren(nodes[leftChildIndex], nodes[rightChildIndex])
		}
	}

	return &NetworkVotingPowerTree{
		nodes:            nodes,
		depth:            uint64(math.Floor(originalPower)),
		virtualRootIndex: virtualRootIndex,
	}
}

func (t *NetworkVotingPowerTree) GetPollardForProposal() []*types.VotingTreeNode {
	return t.generatePollard(1)
}

func (t *NetworkVotingPowerTree) GetArtifactsForChallenge(targetIndex uint64) ([]types.VotingTreeNode, []types.VotingTreeNode) {

}

func (t *NetworkVotingPowerTree) generatePollard(index uint64) []*types.VotingTreeNode {
	rootLevel := uint64(math.Floor(math.Log2(float64(index)))) // The level of the root node
	absoluteDepth := rootLevel + depthPerRound                 // The actual level in the tree that this pollard must come from
	if absoluteDepth > t.depth {
		absoluteDepth = t.depth // Clamp it to the level of the leaf nodes
	}
	relativeDepth := absoluteDepth - rootLevel // How far the pollard level is below the root node level

	// Get the indices of the pollard
	pollardSize := uint64(math.Pow(2, float64(relativeDepth)))
	firstIndex := index * pollardSize
	lastIndex := firstIndex + pollardSize
	return t.nodes[firstIndex:lastIndex]
}

func (t *NetworkVotingPowerTree) createProofForIndex(index uint64) []*types.VotingTreeNode {
	// Create the proof for the index, starting from the bottom up
	proof := []*types.VotingTreeNode{}
	for index > 1 {
		var partnerIndex uint64
		if index%2 == 0 {
			// The target is even so grab the node to the right
			partnerIndex = index + 1
		} else {
			// The target is odd so grab the node to the left
			partnerIndex = index - 1
		}
		node := t.nodes[partnerIndex-1] // Indices in the real array are 0-indexed
		proof = append(proof, node)

		// Go up a level
		index = index / 2
	}
	return proof
}

func (t *NetworkVotingPowerTree) getPhysicalIndexFromVirtualIndex(virtualIndex uint64) uint64 {

}

// Get the keccak hash of a parent node with two children
func getParentNodeFromChildren(leftChild *types.VotingTreeNode, rightChild *types.VotingTreeNode) *types.VotingTreeNode {
	leftBuffer := [32]byte{}
	rightBuffer := [32]byte{}
	leftChild.Sum.FillBytes(leftBuffer[:])
	rightChild.Sum.FillBytes(rightBuffer[:])
	hash := crypto.Keccak256Hash(leftChild.Hash[:], leftBuffer[:], rightChild.Hash[:], rightBuffer[:])

	sum := big.NewInt(0).Add(leftChild.Sum, rightChild.Sum)
	return &types.VotingTreeNode{
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
