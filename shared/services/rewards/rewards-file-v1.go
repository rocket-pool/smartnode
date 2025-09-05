package rewards

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/keccak256"
)

type MinipoolPerformanceFile_v1 struct {
	Index               uint64                                     `json:"index"`
	Network             string                                     `json:"network"`
	StartTime           time.Time                                  `json:"startTime,omitempty"`
	EndTime             time.Time                                  `json:"endTime,omitempty"`
	ConsensusStartBlock uint64                                     `json:"consensusStartBlock,omitempty"`
	ConsensusEndBlock   uint64                                     `json:"consensusEndBlock,omitempty"`
	ExecutionStartBlock uint64                                     `json:"executionStartBlock,omitempty"`
	ExecutionEndBlock   uint64                                     `json:"executionEndBlock,omitempty"`
	MinipoolPerformance map[common.Address]*MinipoolPerformance_v1 `json:"minipoolPerformance"`
}

// Type assertion to implement IPerformanceFile
var _ IPerformanceFile = (*MinipoolPerformanceFile_v1)(nil)

// Type assertion to implement IRewardsFile
var _ IRewardsFile = (*RewardsFile_v1)(nil)

// Serialize a minipool performance file into bytes
func (f *MinipoolPerformanceFile_v1) Serialize() ([]byte, error) {
	return json.Marshal(f)
}

func (f *MinipoolPerformanceFile_v1) SerializeSSZ() ([]byte, error) {
	return nil, fmt.Errorf("ssz format not implemented for minipool performance files")
}

// Serialize a minipool performance file into bytes designed for human readability
func (f *MinipoolPerformanceFile_v1) SerializeHuman() ([]byte, error) {
	return json.MarshalIndent(f, "", "\t")
}

// Deserialize a minipool performance file from bytes
func (f *MinipoolPerformanceFile_v1) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, &f)
}

// Get all of the minipool addresses with rewards in this file
// NOTE: the order of minipool addresses is not guaranteed to be stable, so don't rely on it
func (f *MinipoolPerformanceFile_v1) GetMinipoolAddresses() []common.Address {
	addresses := make([]common.Address, len(f.MinipoolPerformance))
	i := 0
	for address := range f.MinipoolPerformance {
		addresses[i] = address
		i++
	}
	return addresses
}

func (f *MinipoolPerformanceFile_v1) GetMegapoolAddresses() []common.Address {
	return nil
}

func (f *MinipoolPerformanceFile_v1) GetMegapoolPerformance(megapoolAddress common.Address, pubkey types.ValidatorPubkey) (ISmoothingPoolPerformance, bool) {
	return nil, false
}

func (f *MinipoolPerformanceFile_v1) GetMegapoolValidatorPubkeys(megapoolAddress common.Address) ([]types.ValidatorPubkey, error) {
	return nil, nil
}

// Get a minipool's smoothing pool performance if it was present
func (f *MinipoolPerformanceFile_v1) GetMinipoolPerformance(minipoolAddress common.Address) (ISmoothingPoolPerformance, bool) {
	perf, exists := f.MinipoolPerformance[minipoolAddress]
	return perf, exists
}

// Minipool stats
type MinipoolPerformance_v1 struct {
	Pubkey                  string   `json:"pubkey"`
	StartSlot               uint64   `json:"startSlot,omitempty"`
	EndSlot                 uint64   `json:"endSlot,omitempty"`
	ActiveFraction          float64  `json:"activeFraction,omitempty"`
	SuccessfulAttestations  uint64   `json:"successfulAttestations"`
	MissedAttestations      uint64   `json:"missedAttestations"`
	ParticipationRate       float64  `json:"participationRate"`
	MissingAttestationSlots []uint64 `json:"missingAttestationSlots"`
	EthEarned               float64  `json:"ethEarned"`
}

// Type assertion to implement ISmoothingPoolPerformance
var _ ISmoothingPoolPerformance = (*MinipoolPerformance_v1)(nil)

func (p *MinipoolPerformance_v1) GetPubkey() (types.ValidatorPubkey, error) {
	return types.HexToValidatorPubkey(p.Pubkey)
}
func (p *MinipoolPerformance_v1) GetSuccessfulAttestationCount() uint64 {
	return p.SuccessfulAttestations
}
func (p *MinipoolPerformance_v1) GetMissedAttestationCount() uint64 {
	return p.MissedAttestations
}
func (p *MinipoolPerformance_v1) GetMissingAttestationSlots() []uint64 {
	return p.MissingAttestationSlots
}
func (p *MinipoolPerformance_v1) GetEthEarned() *big.Int {
	return eth.EthToWei(p.EthEarned)
}
func (p *MinipoolPerformance_v1) GetBonusEthEarned() *big.Int {
	return big.NewInt(0)
}
func (p *MinipoolPerformance_v1) GetEffectiveCommission() *big.Int {
	return big.NewInt(0)
}
func (p *MinipoolPerformance_v1) GetConsensusIncome() *big.Int {
	return big.NewInt(0)
}
func (p *MinipoolPerformance_v1) GetAttestationScore() *big.Int {
	return big.NewInt(0)
}

// Node operator rewards
type NodeRewardsInfo_v1 struct {
	RewardNetwork                uint64        `json:"rewardNetwork"`
	CollateralRpl                *QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl                 *QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth             *QuotedBigInt `json:"smoothingPoolEth"`
	SmoothingPoolEligibilityRate float64       `json:"smoothingPoolEligibilityRate"`
	MerkleData                   []byte        `json:"-"`
	MerkleProof                  []string      `json:"merkleProof"`
}

func (f *RewardsFile_v1) GetMerkleProof(addr common.Address) ([]common.Hash, error) {
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
type RewardsFile_v1 struct {
	*RewardsFileHeader
	NodeRewards             map[common.Address]*NodeRewardsInfo_v1 `json:"nodeRewards"`
	MinipoolPerformanceFile MinipoolPerformanceFile_v1             `json:"-"`
}

// Serialize a rewards file into bytes
func (f *RewardsFile_v1) Serialize() ([]byte, error) {
	return json.Marshal(f)
}

func (f *RewardsFile_v1) SerializeSSZ() ([]byte, error) {
	return nil, fmt.Errorf("ssz format not implemented for rewards file v1")
}

// Deserialize a rewards file from bytes
func (f *RewardsFile_v1) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, &f)
}

// Get the rewards file version
func (f *RewardsFile_v1) GetRewardsFileVersion() uint64 {
	return rewardsFileVersionOne
}

// Get the rewards file index
func (f *RewardsFile_v1) GetIndex() uint64 {
	return f.RewardsFileHeader.Index
}

// Get the TotalNodeWeight (only added in v3)
func (f *RewardsFile_v1) GetTotalNodeWeight() *big.Int {
	return nil
}

// Get the merkle root
func (f *RewardsFile_v1) GetMerkleRoot() string {
	return f.RewardsFileHeader.MerkleRoot
}

// Get network rewards for a specific network
func (f *RewardsFile_v1) GetNetworkRewards(network uint64) *NetworkRewardsInfo {
	return f.RewardsFileHeader.NetworkRewards[network]
}

// Get the number of intervals that have passed
func (f *RewardsFile_v1) GetIntervalsPassed() uint64 {
	return f.RewardsFileHeader.IntervalsPassed
}

// Get the total RPL sent to the pDAO
func (f *RewardsFile_v1) GetTotalProtocolDaoRpl() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.ProtocolDaoRpl.Int
}

// Get the total RPL sent to the pDAO
func (f *RewardsFile_v1) GetTotalOracleDaoRpl() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.TotalOracleDaoRpl.Int
}

// Get the total Eth sent to pool stakers from the SP
func (f *RewardsFile_v1) GetTotalPoolStakerSmoothingPoolEth() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.PoolStakerSmoothingPoolEth.Int
}

// Get the total rpl sent to stakers
func (f *RewardsFile_v1) GetTotalCollateralRpl() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.TotalCollateralRpl.Int
}

// Get the total smoothing pool eth sent to node operators
func (f *RewardsFile_v1) GetTotalNodeOperatorSmoothingPoolEth() *big.Int {
	return &f.RewardsFileHeader.TotalRewards.NodeOperatorSmoothingPoolEth.Int
}

// Get The execution start block
func (f *RewardsFile_v1) GetExecutionStartBlock() uint64 {
	return f.RewardsFileHeader.ExecutionStartBlock
}

// Get The consensus start block
func (f *RewardsFile_v1) GetConsensusStartBlock() uint64 {
	return f.RewardsFileHeader.ConsensusStartBlock
}

// Get The execution end block
func (f *RewardsFile_v1) GetExecutionEndBlock() uint64 {
	return f.RewardsFileHeader.ExecutionEndBlock
}

// Get The consensus end block
func (f *RewardsFile_v1) GetConsensusEndBlock() uint64 {
	return f.RewardsFileHeader.ConsensusEndBlock
}

// Get the start time
func (f *RewardsFile_v1) GetStartTime() time.Time {
	return f.RewardsFileHeader.StartTime
}

// Get the end time
func (f *RewardsFile_v1) GetEndTime() time.Time {
	return f.RewardsFileHeader.EndTime
}

// Get all of the node addresses with rewards in this file
// NOTE: the order of node addresses is not guaranteed to be stable, so don't rely on it
func (f *RewardsFile_v1) GetNodeAddresses() []common.Address {
	addresses := make([]common.Address, len(f.NodeRewards))
	i := 0
	for address := range f.NodeRewards {
		addresses[i] = address
		i++
	}
	return addresses
}

func (f *RewardsFile_v1) getNodeRewardsInfo(address common.Address) (*NodeRewardsInfo_v1, bool) {
	rewards, exists := f.NodeRewards[address]
	return rewards, exists
}

func (f *RewardsFile_v1) HasRewardsFor(addr common.Address) bool {
	_, ok := f.NodeRewards[addr]
	return ok
}

func (f *RewardsFile_v1) GetNodeCollateralRpl(addr common.Address) *big.Int {
	nr, ok := f.NodeRewards[addr]
	if !ok {
		return big.NewInt(0)
	}
	return &nr.CollateralRpl.Int
}

func (f *RewardsFile_v1) GetNodeOracleDaoRpl(addr common.Address) *big.Int {
	nr, ok := f.NodeRewards[addr]
	if !ok {
		return big.NewInt(0)
	}
	return &nr.OracleDaoRpl.Int
}

func (f *RewardsFile_v1) GetNodeSmoothingPoolEth(addr common.Address) *big.Int {
	nr, ok := f.NodeRewards[addr]
	if !ok {
		return big.NewInt(0)
	}
	return &nr.SmoothingPoolEth.Int
}

func (f *RewardsFile_v1) GetNodeVoterShareEth(addr common.Address) *big.Int {
	return big.NewInt(0)
}

func (f *RewardsFile_v1) GetNodeEth(addr common.Address) *big.Int {
	return f.GetNodeSmoothingPoolEth(addr)
}

// Getters for network info
func (f *RewardsFile_v1) HasRewardsForNetwork(network uint64) bool {
	_, ok := f.NetworkRewards[network]
	return ok
}

func (f *RewardsFile_v1) GetNetworkCollateralRpl(network uint64) *big.Int {
	nr, ok := f.NetworkRewards[network]
	if !ok {
		return big.NewInt(0)
	}

	return &nr.CollateralRpl.Int
}

func (f *RewardsFile_v1) GetNetworkOracleDaoRpl(network uint64) *big.Int {
	nr, ok := f.NetworkRewards[network]
	if !ok {
		return big.NewInt(0)
	}

	return &nr.OracleDaoRpl.Int
}

func (f *RewardsFile_v1) GetNetworkSmoothingPoolEth(network uint64) *big.Int {
	nr, ok := f.NetworkRewards[network]
	if !ok {
		return big.NewInt(0)
	}

	return &nr.SmoothingPoolEth.Int
}

// Sets the CID of the minipool performance file corresponding to this rewards file
func (f *RewardsFile_v1) SetMinipoolPerformanceFileCID(cid string) {
	f.MinipoolPerformanceFileCID = cid
}

// Generates a merkle tree from the provided rewards map
func (f *RewardsFile_v1) GenerateMerkleTree() error {
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
