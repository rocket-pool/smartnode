package rewards

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	svctypes "github.com/rocket-pool/smartnode/shared/types"
	"github.com/wealdtech/go-merkletree"
)

// Holds information
type MinipoolPerformanceFile struct {
	Index               uint64                                               `json:"index"`
	Network             string                                               `json:"network"`
	StartTime           time.Time                                            `json:"startTime,omitempty"`
	EndTime             time.Time                                            `json:"endTime,omitempty"`
	ConsensusStartBlock uint64                                               `json:"consensusStartBlock,omitempty"`
	ConsensusEndBlock   uint64                                               `json:"consensusEndBlock,omitempty"`
	ExecutionStartBlock uint64                                               `json:"executionStartBlock,omitempty"`
	ExecutionEndBlock   uint64                                               `json:"executionEndBlock,omitempty"`
	MinipoolPerformance map[common.Address]*SmoothingPoolMinipoolPerformance `json:"minipoolPerformance"`
}

// Minipool stats
type SmoothingPoolMinipoolPerformance struct {
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
type NodeRewardsInfo struct {
	RewardNetwork                uint64                 `json:"rewardNetwork"`
	CollateralRpl                *svctypes.QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl                 *svctypes.QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth             *svctypes.QuotedBigInt `json:"smoothingPoolEth"`
	SmoothingPoolEligibilityRate float64                `json:"smoothingPoolEligibilityRate"`
	MerkleData                   []byte                 `json:"-"`
	MerkleProof                  []string               `json:"merkleProof"`
}

// Rewards per network
type NetworkRewardsInfo struct {
	CollateralRpl    *svctypes.QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl     *svctypes.QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth *svctypes.QuotedBigInt `json:"smoothingPoolEth"`
}

// Total cumulative rewards for an interval
type TotalRewards struct {
	ProtocolDaoRpl               *svctypes.QuotedBigInt `json:"protocolDaoRpl"`
	TotalCollateralRpl           *svctypes.QuotedBigInt `json:"totalCollateralRpl"`
	TotalOracleDaoRpl            *svctypes.QuotedBigInt `json:"totalOracleDaoRpl"`
	TotalSmoothingPoolEth        *svctypes.QuotedBigInt `json:"totalSmoothingPoolEth"`
	PoolStakerSmoothingPoolEth   *svctypes.QuotedBigInt `json:"poolStakerSmoothingPoolEth"`
	NodeOperatorSmoothingPoolEth *svctypes.QuotedBigInt `json:"nodeOperatorSmoothingPoolEth"`
}

// JSON struct for a complete rewards file
type RewardsFile struct {
	// Serialized fields
	RewardsFileVersion         uint64                              `json:"rewardsFileVersion"`
	RulesetVersion             uint64                              `json:"rulesetVersion,omitempty"`
	Index                      uint64                              `json:"index"`
	Network                    string                              `json:"network"`
	StartTime                  time.Time                           `json:"startTime,omitempty"`
	EndTime                    time.Time                           `json:"endTime"`
	ConsensusStartBlock        uint64                              `json:"consensusStartBlock,omitempty"`
	ConsensusEndBlock          uint64                              `json:"consensusEndBlock"`
	ExecutionStartBlock        uint64                              `json:"executionStartBlock,omitempty"`
	ExecutionEndBlock          uint64                              `json:"executionEndBlock"`
	IntervalsPassed            uint64                              `json:"intervalsPassed"`
	MerkleRoot                 string                              `json:"merkleRoot,omitempty"`
	MinipoolPerformanceFileCID string                              `json:"minipoolPerformanceFileCid,omitempty"`
	TotalRewards               *TotalRewards                       `json:"totalRewards"`
	NetworkRewards             map[uint64]*NetworkRewardsInfo      `json:"networkRewards"`
	NodeRewards                map[common.Address]*NodeRewardsInfo `json:"nodeRewards"`
	MinipoolPerformanceFile    MinipoolPerformanceFile             `json:"-"`

	// Non-serialized fields
	MerkleTree          *merkletree.MerkleTree    `json:"-"`
	InvalidNetworkNodes map[common.Address]uint64 `json:"-"`
}

// Get the deserialized Merkle Proof bytes
func (n *NodeRewardsInfo) GetMerkleProof() ([]common.Hash, error) {
	proof := []common.Hash{}
	for _, proofLevel := range n.MerkleProof {
		proof = append(proof, common.HexToHash(proofLevel))
	}
	return proof, nil
}
