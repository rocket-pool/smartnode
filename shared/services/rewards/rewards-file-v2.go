package rewards

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/rocketpool-go/types"
)

// Holds information
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

func (i *NodeRewardsInfo_v2) GetRewardNetwork() uint64 {
	return i.RewardNetwork
}
func (i *NodeRewardsInfo_v2) GetCollateralRpl() *QuotedBigInt {
	return i.CollateralRpl
}
func (i *NodeRewardsInfo_v2) GetOracleDaoRpl() *QuotedBigInt {
	return i.OracleDaoRpl
}
func (i *NodeRewardsInfo_v2) GetSmoothingPoolEth() *QuotedBigInt {
	return i.SmoothingPoolEth
}
func (i *NodeRewardsInfo_v2) GetMerkleProof() ([]common.Hash, error) {
	proof := []common.Hash{}
	for _, proofLevel := range i.MerkleProof {
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

// Deserialize a rewards file from bytes
func (f *RewardsFile_v2) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, &f)
}

// Get the rewards file's header
func (f *RewardsFile_v2) GetHeader() *RewardsFileHeader {
	return f.RewardsFileHeader
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

// Get info about a node's rewards
func (f *RewardsFile_v2) GetNodeRewardsInfo(address common.Address) (INodeRewardsInfo, bool) {
	rewards, exists := f.NodeRewards[address]
	return rewards, exists
}

// Gets the minipool performance file corresponding to this rewards file
func (f *RewardsFile_v2) GetMinipoolPerformanceFile() IMinipoolPerformanceFile {
	return &f.MinipoolPerformanceFile
}

// Sets the CID of the minipool performance file corresponding to this rewards file
func (f *RewardsFile_v2) SetMinipoolPerformanceFileCID(cid string) {
	f.MinipoolPerformanceFileCID = cid
}
