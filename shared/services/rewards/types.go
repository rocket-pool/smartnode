package rewards

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/types"
)

// Node operator rewards
type NodeRewards struct {
	RewardNetwork    uint64        `json:"rewardNetwork,omitempty"`
	CollateralRpl    *QuotedBigInt `json:"collateralRpl,omitempty"`
	OracleDaoRpl     *QuotedBigInt `json:"oracleDaoRpl,omitempty"`
	SmoothingPoolEth *QuotedBigInt `json:"smoothingPoolEth,omitempty"`
	MerkleData       []byte        `json:"-"`
	MerkleProof      []string      `json:"merkleProof,omitempty"`
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
		CollateralRplPerNetwork    map[uint64]*QuotedBigInt `json:"collateralRplPerNetwork,omitempty"`
		OracleDaoRplPerNetwork     map[uint64]*QuotedBigInt `json:"oracleDaoRplPerNetwork,omitempty"`
		SmoothingPoolEthPerNetwork map[uint64]*QuotedBigInt `json:"smoothingPoolEthPerNetwork,omitempty"`
	} `json:"networkRewards,omitempty"`
	TotalRewards struct {
		ProtocolDaoRpl        *QuotedBigInt `json:"protocolDaoRpl,omitempty"`
		TotalCollateralRpl    *QuotedBigInt `json:"totalCollateralRpl,omitempty"`
		TotalOracleDaoRpl     *QuotedBigInt `json:"totalOracleDaoRpl,omitempty"`
		TotalSmoothingPoolEth *QuotedBigInt `json:"totalSmoothingPoolEth,omitempty"`
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
	CollateralRplAmount    *QuotedBigInt `json:"collateralRplAmount"`
	ODaoRplAmount          *QuotedBigInt `json:"oDaoRplAmount"`
	SmoothingPoolEthAmount *QuotedBigInt `json:"smoothingPoolEthAmount"`
	MerkleProof            []common.Hash `json:"merkleProof"`
}

type MinipoolInfo struct {
	Address            common.Address
	ValidatorPubkey    types.ValidatorPubkey
	ValidatorIndex     uint64
	NodeAddress        common.Address
	NodeIndex          uint64
	Fee                *big.Int
	MissedAttestations uint64
	GoodAttestations   uint64
	MinipoolShare      *big.Int
}

type IntervalDutiesInfo struct {
	Index uint64
	Slots map[uint64]*SlotInfo
}

type SlotInfo struct {
	Index      uint64
	Committees map[uint64]*CommitteeInfo
}

type CommitteeInfo struct {
	Index     uint64
	Positions map[int]*MinipoolInfo
}

// Details about a node for the Smoothing Pool
type NodeSmoothingDetails struct {
	Address           common.Address
	IsEligible        bool
	IsOptedIn         bool
	StatusChangeTime  time.Time
	Minipools         []*MinipoolInfo
	EligibilityFactor float64
	CheaterInfo       CheaterInfo
	SmoothingPoolEth  *big.Int
}

type CheaterInfo struct {
	CheatingDetected bool
	OffendingSlot    uint64
	Minipool         common.Address
	FeeRecipient     common.Address
}

type QuotedBigInt struct {
	big.Int
}

func NewQuotedBigInt(x int64) *QuotedBigInt {
	q := QuotedBigInt{}
	native := big.NewInt(x)
	q.Int = *native
	return &q
}

func (b *QuotedBigInt) MarshalJSON() ([]byte, error) {
	return []byte("\"" + b.String() + "\""), nil
}

func (b *QuotedBigInt) UnmarshalJSON(p []byte) error {
	strippedString := strings.Trim(string(p), "\"")
	nativeInt, success := big.NewInt(0).SetString(strippedString, 0)
	if !success {
		return fmt.Errorf("%s is not a valid big integer", strippedString)
	}

	b.Int = *nativeInt
	return nil
}

// Get the deserialized Merkle Proof bytes
func (n *NodeRewards) GetMerkleProof() ([]common.Hash, error) {
	proof := []common.Hash{}
	for _, proofLevel := range n.MerkleProof {
		proof = append(proof, common.HexToHash(proofLevel))
	}
	return proof, nil
}
