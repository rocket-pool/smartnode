package rewards

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	sharedtypes "github.com/rocket-pool/smartnode/shared/types"
)

// Holds information
type MinipoolPerformanceFile_v1 struct {
	Index               uint64                                                  `json:"index"`
	Network             string                                                  `json:"network"`
	StartTime           time.Time                                               `json:"startTime,omitempty"`
	EndTime             time.Time                                               `json:"endTime,omitempty"`
	ConsensusStartBlock uint64                                                  `json:"consensusStartBlock,omitempty"`
	ConsensusEndBlock   uint64                                                  `json:"consensusEndBlock,omitempty"`
	ExecutionStartBlock uint64                                                  `json:"executionStartBlock,omitempty"`
	ExecutionEndBlock   uint64                                                  `json:"executionEndBlock,omitempty"`
	MinipoolPerformance map[common.Address]*SmoothingPoolMinipoolPerformance_v1 `json:"minipoolPerformance"`
}

// Serialize a minipool performance file into bytes
func (f *MinipoolPerformanceFile_v1) Serialize() ([]byte, error) {
	return json.Marshal(f)
}

// Serialize a minipool performance file into bytes designed for human readability
func (f *MinipoolPerformanceFile_v1) SerializeHuman() ([]byte, error) {
	return json.MarshalIndent(f, "", "\t")
}

// Minipool stats
type SmoothingPoolMinipoolPerformance_v1 struct {
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

// Node operator rewards
type NodeRewardsInfo_v1 struct {
	*sharedtypes.NodeRewardsInfo
	SmoothingPoolEligibilityRate float64 `json:"smoothingPoolEligibilityRate"`
}

// JSON struct for a complete rewards file
type RewardsFile_v1 struct {
	*sharedtypes.RewardsFileHeader
	NodeRewards             map[common.Address]*NodeRewardsInfo_v1 `json:"nodeRewards"`
	MinipoolPerformanceFile MinipoolPerformanceFile_v1             `json:"-"`
}

// Serialize a rewards file into bytes
func (f *RewardsFile_v1) Serialize() ([]byte, error) {
	return json.Marshal(f)
}

// Deserialize a rewards file from bytes
func (f *RewardsFile_v1) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, &f)
}

// Get the rewards file's header
func (f *RewardsFile_v1) GetHeader() *sharedtypes.RewardsFileHeader {
	return f.RewardsFileHeader
}

// Get info about a node's rewards
func (f *RewardsFile_v1) GetNodeRewardsInfo(address common.Address) (sharedtypes.INodeRewardsInfo, bool) {
	rewards, exists := f.NodeRewards[address]
	return rewards, exists
}

// Gets the minipool performance file corresponding to this rewards file
func (f *RewardsFile_v1) GetMinipoolPerformanceFile() sharedtypes.IMinipoolPerformanceFile {
	return &f.MinipoolPerformanceFile
}

// Sets the CID of the minipool performance file corresponding to this rewards file
func (f *RewardsFile_v1) SetMinipoolPerformanceFileCID(cid string) {
	f.MinipoolPerformanceFileCID = cid
}
