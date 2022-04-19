package rewards

import (
	"encoding/hex"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
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
	MerkleRoot     string `json:"merkleRoot,omitempty"`
	NetworkRewards struct {
		CollateralRplPerNetwork    map[uint64]*big.Int `json:"collateralRplPerNetwork,omitempty"`
		OracleDaoRplPerNetwork     map[uint64]*big.Int `json:"oracleDaoRplPerNetwork,omitempty"`
		SmoothingPoolEthPerNetwork map[uint64]*big.Int `json:"smoothingPoolEthPerNetwork,omitempty"`
	} `json:"networkRewards,omitempty"`
	TotalRewards struct {
		TotalCollateralRpl    *big.Int `json:"totalCollateralRpl,omitempty"`
		TotalOracleDaoRpl     *big.Int `json:"totalOracleDaoRpl,omitempty"`
		TotalSmoothingPoolEth *big.Int `json:"totalSmoothingPoolEth,omitempty"`
	} `json:"totalRewards,omitempty"`
	NodeRewards map[common.Address]NodeRewards `json:"nodeRewards,omitempty"`
}

// Information about an interval
type IntervalInfo struct {
	Index                  uint64    `json:"index"`
	TreeFilePath           string    `json:"treeFilePath"`
	TreeFileExists         bool      `json:"treeFileExists"`
	MerkleRootValid        bool      `json:"merkleRootValid"`
	StartTime              time.Time `json:"startTime"`
	EndTime                time.Time `json:"endTime"`
	NodeExists             bool      `json:"nodeExists"`
	CollateralRplAmount    *big.Int  `json:"collateralRplAmount"`
	ODaoRplAmount          *big.Int  `json:"oDaoRplAmount"`
	SmoothingPoolEthAmount *big.Int  `json:"smoothingPoolEthAmount"`
	MerkleProof            [][]byte  `json:"merkleProof"`
}

// Get the deserialized Merkle Proof bytes
func (n *NodeRewards) GetMerkleProof() ([][]byte, error) {
	proofBytes := [][]byte{}
	for _, proofLevel := range n.MerkleProof {
		proofLevelBytes, err := hex.DecodeString(hexutil.RemovePrefix(proofLevel))
		if err != nil {
			return nil, err
		}
		proofBytes = append(proofBytes, proofLevelBytes)
	}
	return proofBytes, nil
}
