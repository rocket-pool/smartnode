package rewards

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Node operator rewards
type NodeRewards struct {
	RewardNetwork    uint64   `json:"rewardNetwork,omitempty"`
	CollateralRpl    *big.Int `json:"collateralRpl,omitempty"`
	OracleDaoRpl     *big.Int `json:"oracleDaoRpl,omitempty"`
	SmoothingPoolEth *big.Int `json:"smoothingPoolEth,omitempty"`
	MerkleData       []byte   `json:"-"`
	MerkleProof      []string `json:"merkleProof,omitempty"`
}

// JSON struct for a complete Merkle Tree proof list
type ProofWrapper struct {
	RewardsFileVersion uint64 `json:"rewardsFileVersion,omitempty"`
	Index              uint64 `json:"index,omitempty"`
	ConsensusBlock     uint64 `json:"consensusBlock,omitempty"`
	ExecutionBlock     uint64 `json:"executionBlock,omitempty"`
	IntervalsPassed    uint64 `json:"intervalsPassed,omitempty"`
	MerkleRoot         string `json:"merkleRoot,omitempty"`
	NetworkRewards     struct {
		CollateralRplPerNetwork    map[uint64]*big.Int `json:"collateralRplPerNetwork,omitempty"`
		OracleDaoRplPerNetwork     map[uint64]*big.Int `json:"oracleDaoRplPerNetwork,omitempty"`
		SmoothingPoolEthPerNetwork map[uint64]*big.Int `json:"smoothingPoolEthPerNetwork,omitempty"`
	} `json:"networkRewards,omitempty"`
	TotalRewards struct {
		ProtocolDaoRpl        *big.Int `json:"protocolDaoRpl,omitempty"`
		TotalCollateralRpl    *big.Int `json:"totalCollateralRpl,omitempty"`
		TotalOracleDaoRpl     *big.Int `json:"totalOracleDaoRpl,omitempty"`
		TotalSmoothingPoolEth *big.Int `json:"totalSmoothingPoolEth,omitempty"`
	} `json:"totalRewards,omitempty"`
	NodeRewards map[common.Address]NodeRewards `json:"nodeRewards,omitempty"`
}

// Information about an interval
type IntervalInfo struct {
	Index                  uint64        `json:"index"`
	TreeFilePath           string        `json:"treeFilePath"`
	TreeFileExists         bool          `json:"treeFileExists"`
	MerkleRootValid        bool          `json:"merkleRootValid"`
	CID                    string        `json:"cid"`
	StartTime              time.Time     `json:"startTime"`
	EndTime                time.Time     `json:"endTime"`
	NodeExists             bool          `json:"nodeExists"`
	CollateralRplAmount    *big.Int      `json:"collateralRplAmount"`
	ODaoRplAmount          *big.Int      `json:"oDaoRplAmount"`
	SmoothingPoolEthAmount *big.Int      `json:"smoothingPoolEthAmount"`
	MerkleProof            []common.Hash `json:"merkleProof"`
}

// Get the deserialized Merkle Proof bytes
func (n *NodeRewards) GetMerkleProof() ([]common.Hash, error) {
	proof := []common.Hash{}
	for _, proofLevel := range n.MerkleProof {
		proof = append(proof, common.HexToHash(proofLevel))
	}
	return proof, nil
}
