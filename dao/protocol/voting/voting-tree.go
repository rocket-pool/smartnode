package voting

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/types"
)

const (
	depthPerRound uint64 = 5
)

type VotingTree struct {
	nodes            []*types.VotingTreeNode
	depth            uint64
	virtualRootIndex uint64
}

// Creates a new NetworkVotingPowerTree instance from leaf nodes.
func CreateTreeFromLeaves(leaves []*types.VotingTreeNode, virtualRootIndex uint64) *VotingTree {
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

	return &VotingTree{
		nodes:            nodes,
		depth:            uint64(math.Floor(originalPower)),
		virtualRootIndex: virtualRootIndex,
	}
}

// Create a pollard from the tree's root node, to be used for new proposals
func (t *VotingTree) GetPollardForProposal() []*types.VotingTreeNode {
	return t.generatePollard(t.virtualRootIndex)
}

// Create a pollard for a challenged tree node, to be used as a challenge response
func (t *VotingTree) GetPollardForChallengeResponse(challengedIndex uint64) []*types.VotingTreeNode {
	return t.generatePollard(challengedIndex)
}

// Compare a pollard used in a proposal / root submission with the corresponding pollard in this tree, getting the challenge artifacts for the first mismatch
func (t *VotingTree) CheckForChallengeableArtifacts(virtualRootIndex uint64, proposedPollard []*types.VotingTreeNode) (uint64, *types.VotingTreeNode, []*types.VotingTreeNode, error) {
	localPollard := t.generatePollard(virtualRootIndex)
	if len(localPollard) != len(proposedPollard) {
		return 0, nil, nil, fmt.Errorf("pollard size mismatch: local pollard = %d nodes, proposed pollard size = %d nodes", len(t.nodes), len(proposedPollard))
	}

	for i, localNode := range localPollard {
		proposedNode := proposedPollard[i]
		if localNode.Hash != proposedNode.Hash || localNode.Sum.Cmp(proposedNode.Sum) != 0 {
			// Get the local index from the pollard offset being used
			firstPollardIndex := len(localPollard)/2 + 1 // Add 1 because it's 1-indexed
			localIndex := uint64(firstPollardIndex + i)
			virtualIndex := t.getVirtualIndexFromLocalIndex(localIndex)

			// Create a new tree from the proposed pollard
			proposedSubtree := CreateTreeFromLeaves(proposedPollard, virtualRootIndex)
			challengedNode, proof := proposedSubtree.getArtifactsForChallenge(virtualIndex)
			return virtualIndex, challengedNode, proof, nil
		}
	}

	return 0, nil, nil, nil
}

// Get the challenged node and a Merkle proof for it
func (t *VotingTree) getArtifactsForChallenge(targetIndex uint64) (*types.VotingTreeNode, []*types.VotingTreeNode) {
	// Get the target node
	localTargetIndex := t.getLocalIndexFromVirtualIndex(targetIndex)
	challengedNode := t.nodes[localTargetIndex-1] // 0-indexed

	// Create a proof for the node using this tree, starting from the bottom up
	proof := []*types.VotingTreeNode{}
	index := localTargetIndex
	for index > 1 { // Recurse until we get to the root node, stop before that
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
	return challengedNode, proof
}

// Construct a pollard using the provided node index as the root node
func (t *VotingTree) generatePollard(virtualRootIndex uint64) []*types.VotingTreeNode {
	index := t.getLocalIndexFromVirtualIndex(virtualRootIndex)

	rootLevel := uint64(math.Floor(math.Log2(float64(index)))) // The level of the root node
	absoluteDepth := rootLevel + depthPerRound                 // The actual level in the tree that this pollard must come from
	if absoluteDepth > t.depth {
		absoluteDepth = t.depth // Clamp it to the level of the leaf nodes
	}
	relativeDepth := absoluteDepth - rootLevel // How far the pollard level is below the root node level

	// Get the indices of the pollard
	pollardSize := uint64(math.Pow(2, float64(relativeDepth)))
	firstIndex := (index - 1) * pollardSize // Subtract 1 to make it 0-indexed
	lastIndex := firstIndex + pollardSize
	return t.nodes[firstIndex:lastIndex]
}

// Get the local index of the tree node (where the root node has index 1) corresponding to a virtual one
func (t *VotingTree) getLocalIndexFromVirtualIndex(virtualIndex uint64) uint64 {
	if t.virtualRootIndex == 1 {
		return virtualIndex
	}

	levelStartIndex := virtualIndex / t.virtualRootIndex
	offset := virtualIndex % t.virtualRootIndex
	return levelStartIndex + offset
}

// Get the virtual index of the tree (where the root node has index 1) corresponding to a local one
func (t *VotingTree) getVirtualIndexFromLocalIndex(localIndex uint64) uint64 {
	if t.virtualRootIndex == 1 {
		return localIndex
	}

	level := uint64(math.Floor(math.Log2(float64(localIndex))))
	firstLevelIndex := uint64(math.Pow(2, float64(level)))
	offset := localIndex - firstLevelIndex

	virtualFirstLevelIndex := firstLevelIndex * t.virtualRootIndex
	return virtualFirstLevelIndex + offset
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
