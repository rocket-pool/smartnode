package proposals

import (
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum/common"
)

const (
	latestCompatibleVersionString string = "1.12.0-dev"
)

// Gets the address of the Rocket Pool Node corresponding to the tree index provided.
// If nil, this is an index into the network tree instead.
func getRPNodeIndexFromTreeNodeIndex(snapshot *VotingInfoSnapshot, virtualIndex uint64) *uint64 {
	// Determine the number of leaf nodes in the tree by the number of RP nodes
	originalPower := math.Log2(float64(len(snapshot.Info)))
	ceilingPower := math.Ceil(originalPower)
	totalLeafNodes := uint64(math.Pow(2.0, ceilingPower)) // This is also the index of the first leaf node

	if virtualIndex < totalLeafNodes {
		return nil
	}

	// Repeatedly divide the index by 2 until arriving at one of the network tree's leaf nodes
	maxLeafNodeIndex := totalLeafNodes*2 - 1
	index := virtualIndex
	for index > maxLeafNodeIndex {
		index /= 2
	}
	rootIndex := index - totalLeafNodes
	return &rootIndex
}

// Gets the index of the leaf node corresponding to the Rocket Pool Node index provided.
func getTreeNodeIndexFromRPNodeIndex(snapshot *VotingInfoSnapshot, nodeIndex uint64) uint64 {
	// Determine the number of leaf nodes in the tree by the number of RP nodes
	originalPower := math.Log2(float64(len(snapshot.Info)))
	ceilingPower := math.Ceil(originalPower)
	totalLeafNodes := uint64(math.Pow(2.0, ceilingPower)) // This is also the index of the first leaf node
	return totalLeafNodes + nodeIndex
}

// Gets the index of the provided Rocket Pool node by its address
func getRPNodeIndexFromSnapshot(snapshot *VotingInfoSnapshot, address common.Address) (uint64, error) {
	for i, info := range snapshot.Info {
		if info.NodeAddress == address {
			return uint64(i), nil
		}
	}

	return 0, fmt.Errorf("address %s is not in the RP node set", address.Hex())
}
