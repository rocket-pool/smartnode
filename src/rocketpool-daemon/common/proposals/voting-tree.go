package proposals

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/shared"
)

type VotingTree struct {
	SmartnodeVersion string                  `json:"smartnodeVersion"`
	Network          config.Network          `json:"network"`
	BlockNumber      uint32                  `json:"blockNumber"`
	Depth            uint64                  `json:"depth"`
	VirtualRootIndex uint64                  `json:"virtualRootIndex"`
	DepthPerRound    uint64                  `json:"depthPerRound"`
	Nodes            []*types.VotingTreeNode `json:"nodes"`
}

// Creates a new NetworkVotingPowerTree instance from leaf nodes.
func CreateTreeFromLeaves(blockNumber uint32, network config.Network, leaves []*types.VotingTreeNode, virtualRootIndex uint64, depthPerRound uint64) *VotingTree {
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
		SmartnodeVersion: shared.RocketPoolVersion,
		BlockNumber:      blockNumber,
		Network:          network,
		Nodes:            nodes,
		Depth:            uint64(ceilingPower),
		DepthPerRound:    depthPerRound,
		VirtualRootIndex: virtualRootIndex,
	}
}

// Create a pollard from the tree's root node, to be used for new proposals
func (t *VotingTree) GetPollardForProposal() (*types.VotingTreeNode, []*types.VotingTreeNode) {
	return t.generatePollard(t.VirtualRootIndex)
}

// Create a pollard for a challenged tree node, to be used as a challenge response
func (t *VotingTree) GetArtifactsForChallengeResponse(challengedIndex uint64) (*types.VotingTreeNode, []*types.VotingTreeNode) {
	return t.generatePollard(challengedIndex)
}

// Compare a pollard used in a proposal / root submission with the corresponding pollard in this tree, getting the challenge artifacts for the first mismatch
func (t *VotingTree) CheckForChallengeableArtifacts(virtualRootIndex uint64, proposedPollard []types.VotingTreeNode) (uint64, *types.VotingTreeNode, []*types.VotingTreeNode, error) {
	_, localPollard := t.generatePollard(virtualRootIndex)
	if len(localPollard) != len(proposedPollard) {
		return 0, nil, nil, fmt.Errorf("pollard size mismatch: local pollard = %d nodes, proposed pollard size = %d nodes", len(t.Nodes), len(proposedPollard))
	}

	for i, localNode := range localPollard {
		proposedNode := proposedPollard[i]
		if localNode.Hash != proposedNode.Hash || localNode.Sum.Cmp(proposedNode.Sum) != 0 {
			// Get the local index from the pollard offset being used
			firstPollardIndex := len(localPollard) // First index is just the length of the pollard row because it's 1-indexed
			localIndex := uint64(firstPollardIndex + i)
			virtualIndex := t.getVirtualIndexFromLocalIndex(localIndex, virtualRootIndex)
			//fmt.Printf("[Challenge] i = %d, firstPollardIndex = %d, localIndex = %d, virtualIndex = %d\n", i, firstPollardIndex, localIndex, virtualIndex)

			// Create the pollard as a slice of pointers
			pollardPtrs := make([]*types.VotingTreeNode, len(proposedPollard))
			for i := range proposedPollard {
				pollardPtrs[i] = &proposedPollard[i]
			}

			// Create a new tree from the proposed pollard
			proposedSubtree := CreateTreeFromLeaves(t.BlockNumber, t.Network, pollardPtrs, virtualRootIndex, t.DepthPerRound)
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
	challengedNode := t.Nodes[localTargetIndex-1] // 0-indexed

	return challengedNode, t.generateMerkleProof(localTargetIndex)
}

// Create a Merkle summation proof for the provided tree node
// Note the index is a *local* one, not a virtual one
func (t *VotingTree) generateMerkleProof(localIndex uint64) []*types.VotingTreeNode {
	proof := []*types.VotingTreeNode{}
	index := localIndex
	for index > 1 { // Recurse until we get to the root node, stop before that
		var partnerIndex uint64
		if index%2 == 0 {
			// The target is even so grab the node to the right
			partnerIndex = index + 1
		} else {
			// The target is odd so grab the node to the left
			partnerIndex = index - 1
		}
		node := t.Nodes[partnerIndex-1] // Indices in the real array are 0-indexed
		proof = append(proof, node)

		// Go up a level
		index = index / 2
	}
	return proof
}

// Construct a pollard using the provided node index as the root node
func (t *VotingTree) generatePollard(virtualRootIndex uint64) (*types.VotingTreeNode, []*types.VotingTreeNode) {
	index := t.getLocalIndexFromVirtualIndex(virtualRootIndex)
	//fmt.Printf("[Pollard Gen] Virtual index = %d, local index = %d\n", virtualRootIndex, index)
	rootNode := t.Nodes[index-1] // 0-indexed

	rootLevel := uint64(math.Floor(math.Log2(float64(index)))) // The level of the root node
	absoluteDepth := rootLevel + t.DepthPerRound               // The actual level in the tree that this pollard must come from
	if absoluteDepth > t.Depth {
		absoluteDepth = t.Depth // Clamp it to the level of the leaf nodes
	}
	relativeDepth := absoluteDepth - rootLevel // How far the pollard level is below the root node level
	//fmt.Printf("[Pollard Gen] Root level = %d, absolute depth = %d, relative depth = %d\n", rootLevel, absoluteDepth, relativeDepth)

	// Get the indices of the pollard
	pollardSize := uint64(math.Pow(2, float64(relativeDepth)))
	firstIndex := index*pollardSize - 1 // Subtract 1 to make it 0-indexed
	lastIndex := firstIndex + pollardSize
	//fmt.Printf("[Pollard Gen] Pollard size = %d, first index = %d, last index = %d\n", pollardSize, firstIndex, lastIndex)
	return rootNode, t.Nodes[firstIndex:lastIndex]
}

// Get the local index of the tree node (where the root node has index 1) corresponding to a virtual one
func (t *VotingTree) getLocalIndexFromVirtualIndex(virtualIndex uint64) uint64 {
	if t.VirtualRootIndex == 1 {
		return virtualIndex
	}

	levelStartIndex := virtualIndex / t.VirtualRootIndex
	offset := virtualIndex % t.VirtualRootIndex
	return levelStartIndex + offset
}

// Get the virtual index of the tree (where the root node has index 1) corresponding to a local one, using the provided virtual index of the root node
func (t *VotingTree) getVirtualIndexFromLocalIndex(localIndex uint64, virtualRootIndex uint64) uint64 {
	if virtualRootIndex == 1 {
		return localIndex
	}

	level := uint64(math.Floor(math.Log2(float64(localIndex))))
	firstLevelIndex := uint64(math.Pow(2, float64(level)))
	offset := localIndex - firstLevelIndex

	virtualFirstLevelIndex := firstLevelIndex * virtualRootIndex
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
