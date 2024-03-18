package types

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/wealdtech/go-merkletree"
)

type RewardsFileVersion uint64

const (
	RewardsFileVersionUnknown = iota
	RewardsFileVersionOne
	RewardsFileVersionTwo
	RewardsFileVersionThree
	RewardsFileVersionMax = iota - 1
)

// Interface for version-agnostic minipool performance
type IMinipoolPerformanceFile interface {
	// Serialize a minipool performance file into bytes
	Serialize() ([]byte, error)

	// Serialize a minipool performance file into bytes designed for human readability
	SerializeHuman() ([]byte, error)

	// Deserialize a rewards file from bytes
	Deserialize([]byte) error

	// Get all of the minipool addresses with rewards in this file
	// NOTE: the order of minipool addresses is not guaranteed to be stable, so don't rely on it
	GetMinipoolAddresses() []common.Address

	// Get a minipool's smoothing pool performance if it was present
	GetSmoothingPoolPerformance(minipoolAddress common.Address) (ISmoothingPoolMinipoolPerformance, bool)
}

// Interface for version-agnostic rewards files
type IRewardsFile interface {
	// Serialize a rewards file into bytes
	Serialize() ([]byte, error)

	// Deserialize a rewards file from bytes
	Deserialize([]byte) error

	// Get the rewards file's header
	GetHeader() *RewardsFileHeader

	// Get all of the node addresses with rewards in this file
	// NOTE: the order of node addresses is not guaranteed to be stable, so don't rely on it
	GetNodeAddresses() []common.Address

	// Get info about a node's rewards
	GetNodeRewardsInfo(address common.Address) (INodeRewardsInfo, bool)

	// Gets the minipool performance file corresponding to this rewards file
	GetMinipoolPerformanceFile() IMinipoolPerformanceFile

	// Sets the CID of the minipool performance file corresponding to this rewards file
	SetMinipoolPerformanceFileCID(cid string)

	// Generate the Merkle Tree and its root from the rewards file's proofs
	GenerateMerkleTree() error
}

// Rewards per network
type NetworkRewardsInfo struct {
	CollateralRpl    *QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl     *QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth *QuotedBigInt `json:"smoothingPoolEth"`
}

// Total cumulative rewards for an interval
type TotalRewards struct {
	ProtocolDaoRpl               *QuotedBigInt `json:"protocolDaoRpl"`
	TotalCollateralRpl           *QuotedBigInt `json:"totalCollateralRpl"`
	TotalOracleDaoRpl            *QuotedBigInt `json:"totalOracleDaoRpl"`
	TotalSmoothingPoolEth        *QuotedBigInt `json:"totalSmoothingPoolEth"`
	PoolStakerSmoothingPoolEth   *QuotedBigInt `json:"poolStakerSmoothingPoolEth"`
	NodeOperatorSmoothingPoolEth *QuotedBigInt `json:"nodeOperatorSmoothingPoolEth"`
	TotalNodeWeight              *QuotedBigInt `json:"totalNodeWeight,omitempty"`
}

// Minipool stats
type ISmoothingPoolMinipoolPerformance interface {
	GetPubkey() (beacon.ValidatorPubkey, error)
	GetSuccessfulAttestationCount() uint64
	GetMissedAttestationCount() uint64
	GetMissingAttestationSlots() []uint64
	GetEthEarned() *big.Int
}

// Interface for version-agnostic node operator rewards
type INodeRewardsInfo interface {
	GetRewardNetwork() uint64
	GetCollateralRpl() *QuotedBigInt
	GetOracleDaoRpl() *QuotedBigInt
	GetSmoothingPoolEth() *QuotedBigInt
	GetMerkleProof() ([]common.Hash, error)
}

// Small struct to test version information for rewards files during deserialization
type VersionHeader struct {
	RewardsFileVersion RewardsFileVersion `json:"rewardsFileVersion,omitempty"`
}

// General version-agnostic information about a rewards file
type RewardsFileHeader struct {
	// Serialized fields
	RewardsFileVersion         RewardsFileVersion             `json:"rewardsFileVersion"`
	RulesetVersion             uint64                         `json:"rulesetVersion,omitempty"`
	Index                      uint64                         `json:"index"`
	Network                    string                         `json:"network"`
	StartTime                  time.Time                      `json:"startTime,omitempty"`
	EndTime                    time.Time                      `json:"endTime"`
	ConsensusStartBlock        uint64                         `json:"consensusStartBlock,omitempty"`
	ConsensusEndBlock          uint64                         `json:"consensusEndBlock"`
	ExecutionStartBlock        uint64                         `json:"executionStartBlock,omitempty"`
	ExecutionEndBlock          uint64                         `json:"executionEndBlock"`
	IntervalsPassed            uint64                         `json:"intervalsPassed"`
	MerkleRoot                 string                         `json:"merkleRoot,omitempty"`
	MinipoolPerformanceFileCID string                         `json:"minipoolPerformanceFileCid,omitempty"`
	TotalRewards               *TotalRewards                  `json:"totalRewards"`
	NetworkRewards             map[uint64]*NetworkRewardsInfo `json:"networkRewards"`

	// Non-serialized fields
	MerkleTree          *merkletree.MerkleTree    `json:"-"`
	InvalidNetworkNodes map[common.Address]uint64 `json:"-"`
}

// Information about an interval
type IntervalInfo struct {
	Index                  uint64        `json:"index"`
	TreeFilePath           string        `json:"treeFilePath"`
	TreeFileExists         bool          `json:"treeFileExists"`
	MerkleRootValid        bool          `json:"merkleRootValid"`
	MerkleRoot             common.Hash   `json:"merkleRoot"`
	CID                    string        `json:"cid"`
	StartTime              time.Time     `json:"startTime"`
	EndTime                time.Time     `json:"endTime"`
	NodeExists             bool          `json:"nodeExists"`
	CollateralRplAmount    *QuotedBigInt `json:"collateralRplAmount"`
	ODaoRplAmount          *QuotedBigInt `json:"oDaoRplAmount"`
	SmoothingPoolEthAmount *QuotedBigInt `json:"smoothingPoolEthAmount"`
	MerkleProof            []common.Hash `json:"merkleProof"`

	TotalNodeWeight        *QuotedBigInt `json:"-"`
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
