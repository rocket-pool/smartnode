package rewards

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/keccak256"
)

type MinipoolPerformanceFile_v2 struct {
	RewardsFileVersion  uint64                                                  `json:"rewardsFileVersion"`
	RulesetVersion      uint64                                                  `json:"rulesetVersion"`
	Index               uint64                                                  `json:"index"`
	Network             string                                                  `json:"network"`
	StartTime           time.Time                                               `json:"startTime,omitempty"`
	EndTime             time.Time                                               `json:"endTime,omitempty"`
	ConsensusStartBlock uint64                                                  `json:"consensusStartBlock,omitempty"`
	ConsensusEndBlock   uint64                                                  `json:"consensusEndBlock,omitempty"`
	ExecutionStartBlock uint64                                                  `json:"executionStartBlock,omitempty"`
	ExecutionEndBlock   uint64                                                  `json:"executionEndBlock,omitempty"`
	MinipoolPerformance map[common.Address]*SmoothingPoolMinipoolPerformance_v2 `json:"minipoolPerformance"`
}

// Serialize a minipool performance file into bytes
func (f *MinipoolPerformanceFile_v2) Serialize() ([]byte, error) {
	return json.Marshal(f)
}

func (f *MinipoolPerformanceFile_v2) SerializeSSZ() ([]byte, error) {
	return nil, fmt.Errorf("ssz format not implemented for minipool performance files")
}

// Serialize a minipool performance file into bytes designed for human readability
func (f *MinipoolPerformanceFile_v2) SerializeHuman() ([]byte, error) {
	return json.MarshalIndent(f, "", "\t")
}

// Deserialize a minipool performance file from bytes
func (f *MinipoolPerformanceFile_v2) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, &f)
}

// Get all of the minipool addresses with rewards in this file
// NOTE: the order of minipool addresses is not guaranteed to be stable, so don't rely on it
func (f *MinipoolPerformanceFile_v2) GetMinipoolAddresses() []common.Address {
	addresses := make([]common.Address, len(f.MinipoolPerformance))
	i := 0
	for address := range f.MinipoolPerformance {
		addresses[i] = address
		i++
	}
	return addresses
}

// Get a minipool's smoothing pool performance if it was present
func (f *MinipoolPerformanceFile_v2) GetSmoothingPoolPerformance(minipoolAddress common.Address) (ISmoothingPoolMinipoolPerformance, bool) {
	perf, exists := f.MinipoolPerformance[minipoolAddress]
	return perf, exists
}

// Minipool stats
type SmoothingPoolMinipoolPerformance_v2 struct {
	Pubkey                  string        `json:"pubkey"`
	SuccessfulAttestations  uint64        `json:"successfulAttestations"`
	MissedAttestations      uint64        `json:"missedAttestations"`
	AttestationScore        *QuotedBigInt `json:"attestationScore"`
	MissingAttestationSlots []uint64      `json:"missingAttestationSlots"`
	EthEarned               *QuotedBigInt `json:"ethEarned"`
}

func (p *SmoothingPoolMinipoolPerformance_v2) GetPubkey() (types.ValidatorPubkey, error) {
	return types.HexToValidatorPubkey(p.Pubkey)
}
func (p *SmoothingPoolMinipoolPerformance_v2) GetSuccessfulAttestationCount() uint64 {
	return p.SuccessfulAttestations
}
func (p *SmoothingPoolMinipoolPerformance_v2) GetMissedAttestationCount() uint64 {
	return p.MissedAttestations
}
func (p *SmoothingPoolMinipoolPerformance_v2) GetMissingAttestationSlots() []uint64 {
	return p.MissingAttestationSlots
}
func (p *SmoothingPoolMinipoolPerformance_v2) GetEthEarned() *big.Int {
	return &p.EthEarned.Int
}

// Node operator rewards
type NodeRewardsInfo_v2 struct {
	RewardNetwork    uint64        `json:"rewardNetwork"`
	CollateralRpl    *QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl     *QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth *QuotedBigInt `json:"smoothingPoolEth"`
	MerkleData       []byte        `json:"-"`
	MerkleProof      []string      `json:"merkleProof"`
}

func (f *RewardsFile_v2) GetMerkleProof(addr common.Address) ([]common.Hash, error) {
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

// JSON struct for a complete rewards file
type RewardsFile_v2 struct {
	*RewardsFileHeader
	NodeRewards             map[common.Address]*NodeRewardsInfo_v2 `json:"nodeRewards"`
	MinipoolPerformanceFile MinipoolPerformanceFile_v2             `json:"-"`
}

// Serialize a rewards file into bytes
func (f *RewardsFile_v2) Serialize() ([]byte, error) {
	return json.Marshal(f)
}

func (f *RewardsFile_v2) SerializeSSZ() ([]byte, error) {
	return nil, fmt.Errorf("ssz format not implemented for rewards file v2")
}

// Deserialize a rewards file from bytes
func (f *RewardsFile_v2) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, &f)
}

// Get the rewards file version
func (f *RewardsFile_v2) GetRewardsFileVersion() uint64 {
	return rewardsFileVersionTwo
}

// Get the rewards file index
func (f *RewardsFile_v2) GetIndex() uint64 {
	return f.RewardsFileHeader.Index
}

// Get the TotalNodeWeight (only added in v3)
func (f *RewardsFile_v2) GetTotalNodeWeight() *big.Int {
	return nil
}

// Get the merkle root
func (f *RewardsFile_v2) GetMerkleRoot() string {
	return f.RewardsFileHeader.MerkleRoot
}

// Get network rewards for a specific network
func (f *RewardsFile_v2) GetNetworkRewards(network uint64) *NetworkRewardsInfo {
	return f.RewardsFileHeader.NetworkRewards[network]
}

// Get the number of intervals that have passed
func (f *RewardsFile_v2) GetIntervalsPassed() uint64 {
	return f.RewardsFileHeader.IntervalsPassed
}

// Get the total RPL sent to the pDAO
func (f *RewardsFile_v2) GetTotalProtocolDaoRpl() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.ProtocolDaoRpl.Int
}

// Get the total RPL sent to the pDAO
func (f *RewardsFile_v2) GetTotalOracleDaoRpl() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.TotalOracleDaoRpl.Int
}

// Get the total Eth sent to pool stakers from the SP
func (f *RewardsFile_v2) GetTotalPoolStakerSmoothingPoolEth() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.PoolStakerSmoothingPoolEth.Int
}

// Get the total rpl sent to stakers
func (f *RewardsFile_v2) GetTotalCollateralRpl() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.TotalCollateralRpl.Int
}

// Get the total smoothing pool eth sent to node operators
func (f *RewardsFile_v2) GetTotalNodeOperatorSmoothingPoolEth() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.NodeOperatorSmoothingPoolEth.Int
}

// Get the the execution start block
func (f *RewardsFile_v2) GetExecutionStartBlock() uint64 {
	return f.RewardsFileHeader.ExecutionStartBlock
}

// Get the the consensus start block
func (f *RewardsFile_v2) GetConsensusStartBlock() uint64 {
	return f.RewardsFileHeader.ConsensusStartBlock
}

// Get the the execution end block
func (f *RewardsFile_v2) GetExecutionEndBlock() uint64 {
	return f.RewardsFileHeader.ExecutionEndBlock
}

// Get the the consensus end block
func (f *RewardsFile_v2) GetConsensusEndBlock() uint64 {
	return f.RewardsFileHeader.ConsensusEndBlock
}

// Get the start time
func (f *RewardsFile_v2) GetStartTime() time.Time {
	return f.RewardsFileHeader.StartTime
}

// Get the end time
func (f *RewardsFile_v2) GetEndTime() time.Time {
	return f.RewardsFileHeader.EndTime
}

// Get all of the node addresses with rewards in this file
// NOTE: the order of node addresses is not guaranteed to be stable, so don't rely on it
func (f *RewardsFile_v2) GetNodeAddresses() []common.Address {
	addresses := make([]common.Address, len(f.NodeRewards))
	i := 0
	for address := range f.NodeRewards {
		addresses[i] = address
		i++
	}
	return addresses
}

func (f *RewardsFile_v2) getNodeRewardsInfo(address common.Address) (*NodeRewardsInfo_v2, bool) {
	rewards, exists := f.NodeRewards[address]
	return rewards, exists
}

func (f *RewardsFile_v2) HasRewardsFor(addr common.Address) bool {
	_, ok := f.NodeRewards[addr]
	return ok
}

func (f *RewardsFile_v2) GetNodeCollateralRpl(addr common.Address) *big.Int {
	nr, ok := f.NodeRewards[addr]
	if !ok {
		return big.NewInt(0)
	}
	return &nr.CollateralRpl.Int
}

func (f *RewardsFile_v2) GetNodeOracleDaoRpl(addr common.Address) *big.Int {
	nr, ok := f.NodeRewards[addr]
	if !ok {
		return big.NewInt(0)
	}
	return &nr.OracleDaoRpl.Int
}

func (f *RewardsFile_v2) GetNodeSmoothingPoolEth(addr common.Address) *big.Int {
	nr, ok := f.NodeRewards[addr]
	if !ok {
		return big.NewInt(0)
	}
	return &nr.SmoothingPoolEth.Int
}

// Getters for network info
func (f *RewardsFile_v2) HasRewardsForNetwork(network uint64) bool {
	_, ok := f.NetworkRewards[network]
	return ok
}

func (f *RewardsFile_v2) GetNetworkCollateralRpl(network uint64) *big.Int {
	nr, ok := f.NetworkRewards[network]
	if !ok {
		return big.NewInt(0)
	}

	return &nr.CollateralRpl.Int
}

func (f *RewardsFile_v2) GetNetworkOracleDaoRpl(network uint64) *big.Int {
	nr, ok := f.NetworkRewards[network]
	if !ok {
		return big.NewInt(0)
	}

	return &nr.OracleDaoRpl.Int
}

func (f *RewardsFile_v2) GetNetworkSmoothingPoolEth(network uint64) *big.Int {
	nr, ok := f.NetworkRewards[network]
	if !ok {
		return big.NewInt(0)
	}

	return &nr.SmoothingPoolEth.Int
}

// Sets the CID of the minipool performance file corresponding to this rewards file
func (f *RewardsFile_v2) SetMinipoolPerformanceFileCID(cid string) {
	f.MinipoolPerformanceFileCID = cid
}

// Generates a merkle tree from the provided rewards map
func (f *RewardsFile_v2) GenerateMerkleTree() error {
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
