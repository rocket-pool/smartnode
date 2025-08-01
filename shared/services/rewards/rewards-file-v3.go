package rewards

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/shared/services/rewards/ssz_types"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/keccak256"
)

// JSON struct for a complete rewards file
type RewardsFile_v3 struct {
	*RewardsFileHeader
	NodeRewards             map[common.Address]*NodeRewardsInfo_v2 `json:"nodeRewards"`
	MinipoolPerformanceFile MinipoolPerformanceFile_v2             `json:"-"`
}

// Type assertion to implement IRewardsFile
var _ IRewardsFile = (*RewardsFile_v3)(nil)

// Serialize a rewards file into bytes
func (f *RewardsFile_v3) Serialize() ([]byte, error) {
	return json.Marshal(f)
}

// Serialize as SSZ
func (f *RewardsFile_v3) SerializeSSZ() ([]byte, error) {
	// In order to avoid multiple code paths, we won't bother making a RewardsFile_v3 <-> SSZFile_v1 function
	// Instead, we can serialize json, parse to SSZFile_v1, and then serialize that as SSZ
	data, err := f.Serialize()
	if err != nil {
		return nil, fmt.Errorf("error converting RewardsFile v3 to json so it could be parsed as SSZFile_v1: %w", err)
	}

	s := &ssz_types.SSZFile_v1{}
	err = s.UnmarshalSSZ(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing RewardsFile v3 json as SSZFile_v1: %w", err)
	}

	return s.SerializeSSZ()
}

// Deserialize a rewards file from bytes
func (f *RewardsFile_v3) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, &f)
}

// Get the rewards file version
func (f *RewardsFile_v3) GetRewardsFileVersion() uint64 {
	return rewardsFileVersionThree
}

// Get the rewards file index
func (f *RewardsFile_v3) GetIndex() uint64 {
	return f.RewardsFileHeader.Index
}

// Get the TotalNodeWeight (only added in v3)
func (f *RewardsFile_v3) GetTotalNodeWeight() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.TotalNodeWeight.Int
}

// Get the merkle root
func (f *RewardsFile_v3) GetMerkleRoot() string {
	return f.RewardsFileHeader.MerkleRoot
}

// Get network rewards for a specific network
func (f *RewardsFile_v3) GetNetworkRewards(network uint64) *NetworkRewardsInfo {
	return f.RewardsFileHeader.NetworkRewards[network]
}

// Get the number of intervals that have passed
func (f *RewardsFile_v3) GetIntervalsPassed() uint64 {
	return f.RewardsFileHeader.IntervalsPassed
}

// Get the total RPL sent to the pDAO
func (f *RewardsFile_v3) GetTotalProtocolDaoRpl() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.ProtocolDaoRpl.Int
}

// Get the total RPL sent to the pDAO
func (f *RewardsFile_v3) GetTotalOracleDaoRpl() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.TotalOracleDaoRpl.Int
}

// Get the total Eth sent to pool stakers from the SP
func (f *RewardsFile_v3) GetTotalPoolStakerSmoothingPoolEth() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.PoolStakerSmoothingPoolEth.Int
}

// Get the total rpl sent to stakers
func (f *RewardsFile_v3) GetTotalCollateralRpl() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.TotalCollateralRpl.Int
}

// Get the total smoothing pool eth sent to node operators
func (f *RewardsFile_v3) GetTotalNodeOperatorSmoothingPoolEth() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.NodeOperatorSmoothingPoolEth.Int
}

// Get the execution end block
func (f *RewardsFile_v3) GetExecutionEndBlock() uint64 {
	return f.RewardsFileHeader.ExecutionEndBlock
}

// Get the consensus end block
func (f *RewardsFile_v3) GetConsensusEndBlock() uint64 {
	return f.RewardsFileHeader.ConsensusEndBlock
}

// Get the execution start block
func (f *RewardsFile_v3) GetExecutionStartBlock() uint64 {
	return f.RewardsFileHeader.ExecutionStartBlock
}

// Get the consensus start block
func (f *RewardsFile_v3) GetConsensusStartBlock() uint64 {
	return f.RewardsFileHeader.ConsensusStartBlock
}

// Get the start time
func (f *RewardsFile_v3) GetStartTime() time.Time {
	return f.RewardsFileHeader.StartTime
}

// Get the end time
func (f *RewardsFile_v3) GetEndTime() time.Time {
	return f.RewardsFileHeader.EndTime
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

func (f *RewardsFile_v3) getNodeRewardsInfo(address common.Address) (*NodeRewardsInfo_v2, bool) {
	rewards, exists := f.NodeRewards[address]
	return rewards, exists
}

func (f *RewardsFile_v3) HasRewardsFor(addr common.Address) bool {
	_, ok := f.NodeRewards[addr]
	return ok
}

func (f *RewardsFile_v3) GetNodeCollateralRpl(addr common.Address) *big.Int {
	nr, ok := f.NodeRewards[addr]
	if !ok {
		return big.NewInt(0)
	}
	return &nr.CollateralRpl.Int
}

func (f *RewardsFile_v3) GetNodeOracleDaoRpl(addr common.Address) *big.Int {
	nr, ok := f.NodeRewards[addr]
	if !ok {
		return big.NewInt(0)
	}
	return &nr.OracleDaoRpl.Int
}

func (f *RewardsFile_v3) GetNodeSmoothingPoolEth(addr common.Address) *big.Int {
	nr, ok := f.NodeRewards[addr]
	if !ok {
		return big.NewInt(0)
	}
	return &nr.SmoothingPoolEth.Int
}

func (f *RewardsFile_v3) GetNodeVoterShareEth(addr common.Address) *big.Int {
	return big.NewInt(0)
}

func (f *RewardsFile_v3) GetNodeEth(addr common.Address) *big.Int {
	return f.GetNodeSmoothingPoolEth(addr)
}

func (f *RewardsFile_v3) GetMerkleProof(addr common.Address) ([]common.Hash, error) {
	nr, ok := f.getNodeRewardsInfo(addr)
	if !ok {
		return nil, nil
	}
	proof := make([]common.Hash, 0, len(nr.MerkleProof))
	for _, proofLevel := range nr.MerkleProof {
		proof = append(proof, common.HexToHash(proofLevel))
	}
	return proof, nil
}

// Getters for network info
func (f *RewardsFile_v3) HasRewardsForNetwork(network uint64) bool {
	_, ok := f.NetworkRewards[network]
	return ok
}

func (f *RewardsFile_v3) GetNetworkCollateralRpl(network uint64) *big.Int {
	nr, ok := f.NetworkRewards[network]
	if !ok {
		return big.NewInt(0)
	}

	return &nr.CollateralRpl.Int
}

func (f *RewardsFile_v3) GetNetworkOracleDaoRpl(network uint64) *big.Int {
	nr, ok := f.NetworkRewards[network]
	if !ok {
		return big.NewInt(0)
	}

	return &nr.OracleDaoRpl.Int
}

func (f *RewardsFile_v3) GetNetworkSmoothingPoolEth(network uint64) *big.Int {
	nr, ok := f.NetworkRewards[network]
	if !ok {
		return big.NewInt(0)
	}

	return &nr.SmoothingPoolEth.Int
}

// Sets the CID of the minipool performance file corresponding to this rewards file
func (f *RewardsFile_v3) SetMinipoolPerformanceFileCID(cid string) {
	f.MinipoolPerformanceFileCID = cid
}

// Generates a merkle tree from the provided rewards map
func (f *RewardsFile_v3) GenerateMerkleTree() error {
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
