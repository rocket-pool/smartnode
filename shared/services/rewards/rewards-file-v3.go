package rewards

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/keccak256"
)

// JSON struct for a complete rewards file
type RewardsFile_v3 struct {
	*RewardsFileHeader
	NodeRewards             map[common.Address]*NodeRewardsInfo_v2 `json:"nodeRewards"`
	MinipoolPerformanceFile MinipoolPerformanceFile_v2             `json:"-"`
}

// Serialize a rewards file into bytes
func (f *RewardsFile_v3) Serialize() ([]byte, error) {
	return json.Marshal(f)
}

// Deserialize a rewards file from bytes
func (f *RewardsFile_v3) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, &f)
}

// Get the rewards file's header
func (f *RewardsFile_v3) GetHeader() *RewardsFileHeader {
	return f.RewardsFileHeader
}

// Get all of the node addresses with rewards in this file
// NOTE: the order of node addresses is not guaranteed to be stable, so don't rely on it
func (f *RewardsFile_v3) GetNodeAddresses() []common.Address {
	addresses := make([]common.Address, len(f.NodeRewards))
	i := 0
	for address := range f.NodeRewards {
		addresses[i] = address
		i++
	}
	return addresses
}

// Get info about a node's rewards
func (f *RewardsFile_v3) GetNodeRewardsInfo(address common.Address) (INodeRewardsInfo, bool) {
	rewards, exists := f.NodeRewards[address]
	return rewards, exists
}

// Sets the CID of the minipool performance file corresponding to this rewards file
func (f *RewardsFile_v3) SetMinipoolPerformanceFileCID(cid string) {
	f.MinipoolPerformanceFileCID = cid
}

// Generates a merkle tree from the provided rewards map
func (f *RewardsFile_v3) generateMerkleTree() error {
	// Generate the leaf data for each node
	totalData := make([][]byte, 0, len(f.NodeRewards))
	for address, rewardsForNode := range f.NodeRewards {
		// Ignore nodes that didn't receive any rewards
		if rewardsForNode.CollateralRpl.Cmp(common.Big0) == 0 && rewardsForNode.OracleDaoRpl.Cmp(common.Big0) == 0 && rewardsForNode.SmoothingPoolEth.Cmp(common.Big0) == 0 {
			continue
		}

		// Node data is address[20] :: network[32] :: RPL[32] :: ETH[32]
		nodeData := make([]byte, 0, 20+32*3)

		// Node address
		addressBytes := address.Bytes()
		nodeData = append(nodeData, addressBytes...)

		// Node network
		network := big.NewInt(0).SetUint64(rewardsForNode.RewardNetwork)
		networkBytes := make([]byte, 32)
		network.FillBytes(networkBytes)
		nodeData = append(nodeData, networkBytes...)

		// RPL rewards
		rplRewards := big.NewInt(0)
		rplRewards.Add(&rewardsForNode.CollateralRpl.Int, &rewardsForNode.OracleDaoRpl.Int)
		rplRewardsBytes := make([]byte, 32)
		rplRewards.FillBytes(rplRewardsBytes)
		nodeData = append(nodeData, rplRewardsBytes...)

		// ETH rewards
		ethRewardsBytes := make([]byte, 32)
		rewardsForNode.SmoothingPoolEth.FillBytes(ethRewardsBytes)
		nodeData = append(nodeData, ethRewardsBytes...)

		// Assign it to the node rewards tracker and add it to the leaf data slice
		rewardsForNode.MerkleData = nodeData
		totalData = append(totalData, nodeData)
	}

	// Generate the tree
	tree, err := merkletree.NewUsing(totalData, keccak256.New(), false, true)
	if err != nil {
		return fmt.Errorf("error generating Merkle Tree: %w", err)
	}

	// Generate the proofs for each node
	for address, rewardsForNode := range f.NodeRewards {
		// Get the proof
		proof, err := tree.GenerateProof(rewardsForNode.MerkleData, 0)
		if err != nil {
			return fmt.Errorf("error generating proof for node %s: %w", address.Hex(), err)
		}

		// Convert the proof into hex strings
		proofStrings := make([]string, len(proof.Hashes))
		for i, hash := range proof.Hashes {
			proofStrings[i] = fmt.Sprintf("0x%s", hex.EncodeToString(hash))
		}

		// Assign the hex strings to the node rewards struct
		rewardsForNode.MerkleProof = proofStrings
	}

	f.MerkleTree = tree
	f.MerkleRoot = common.BytesToHash(tree.Root()).Hex()
	return nil
}
