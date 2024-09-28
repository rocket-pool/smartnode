package rewards

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/wealdtech/go-merkletree"
)

const (
	FarEpoch uint64 = 18446744073709551615
)

const (
	rewardsFileVersionUnknown uint64 = iota
	rewardsFileVersionOne
	rewardsFileVersionTwo
	rewardsFileVersionThree
	rewardsFileVersionMax = iota - 1

	minRewardsFileVersionSSZ = rewardsFileVersionThree
)

// RewardsExecutionClient defines and interface
// that contains only the functions from rocketpool.RocketPool
// required for rewards generation.
// This facade makes it easier to perform dependency injection in tests.
type RewardsExecutionClient interface {
	GetNetworkEnabled(networkId *big.Int, opts *bind.CallOpts) (bool, error)
	HeaderByNumber(context.Context, *big.Int) (*ethtypes.Header, error)
	GetRewardsEvent(index uint64, rocketRewardsPoolAddresses []common.Address, opts *bind.CallOpts) (bool, rewards.RewardsEvent, error)
	GetRewardSnapshotEvent(previousRewardsPoolAddresses []common.Address, interval uint64, opts *bind.CallOpts) (rewards.RewardsEvent, error)
}

// RewardsBeaconClient defines and interface
// that contains only the functions from beacon.Client
// required for rewards generation.
// This facade makes it easier to perform dependency injection in tests.
type RewardsBeaconClient interface {
	GetBeaconBlock(slot string) (beacon.BeaconBlock, bool, error)
	GetCommitteesForEpoch(epoch *uint64) (beacon.Committees, error)
	GetAttestations(slot string) ([]beacon.AttestationInfo, bool, error)
}

// Interface for version-agnostic minipool performance
type IMinipoolPerformanceFile interface {
	// Serialize a minipool performance file into bytes
	Serialize() ([]byte, error)
	SerializeSSZ() ([]byte, error)

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
	SerializeSSZ() ([]byte, error)

	// Deserialize a rewards file from bytes
	Deserialize([]byte) error

	// Getters for general interval info
	GetRewardsFileVersion() uint64
	GetIndex() uint64
	GetTotalNodeWeight() *big.Int
	GetMerkleRoot() string
	GetIntervalsPassed() uint64
	GetTotalProtocolDaoRpl() *big.Int
	GetTotalOracleDaoRpl() *big.Int
	GetTotalCollateralRpl() *big.Int
	GetTotalNodeOperatorSmoothingPoolEth() *big.Int
	GetTotalPoolStakerSmoothingPoolEth() *big.Int
	GetExecutionStartBlock() uint64
	GetConsensusStartBlock() uint64
	GetExecutionEndBlock() uint64
	GetConsensusEndBlock() uint64
	GetStartTime() time.Time
	GetEndTime() time.Time

	// Get all of the node addresses with rewards in this file
	// NOTE: the order of node addresses is not guaranteed to be stable, so don't rely on it
	GetNodeAddresses() []common.Address

	// Getters for into about specific node's rewards
	HasRewardsFor(common.Address) bool
	GetNodeCollateralRpl(common.Address) *big.Int
	GetNodeOracleDaoRpl(common.Address) *big.Int
	GetNodeSmoothingPoolEth(common.Address) *big.Int
	GetMerkleProof(common.Address) ([]common.Hash, error)

	// Getters for network info
	HasRewardsForNetwork(network uint64) bool
	GetNetworkCollateralRpl(network uint64) *big.Int
	GetNetworkOracleDaoRpl(network uint64) *big.Int
	GetNetworkSmoothingPoolEth(network uint64) *big.Int

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
	GetPubkey() (types.ValidatorPubkey, error)
	GetSuccessfulAttestationCount() uint64
	GetMissedAttestationCount() uint64
	GetMissingAttestationSlots() []uint64
	GetEthEarned() *big.Int
}

// Small struct to test version information for rewards files during deserialization
type VersionHeader struct {
	RewardsFileVersion uint64 `json:"rewardsFileVersion,omitempty"`
}

// General version-agnostic information about a rewards file
type RewardsFileHeader struct {
	// Serialized fields
	RewardsFileVersion         uint64                         `json:"rewardsFileVersion"`
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
	MerkleTree *merkletree.MerkleTree `json:"-"`
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

	TotalNodeWeight *big.Int `json:"-"`
}

type MinipoolInfo struct {
	Address                 common.Address        `json:"address"`
	ValidatorPubkey         types.ValidatorPubkey `json:"pubkey"`
	ValidatorIndex          string                `json:"index"`
	NodeAddress             common.Address        `json:"nodeAddress"`
	NodeIndex               uint64                `json:"-"`
	Fee                     *big.Int              `json:"-"`
	MissedAttestations      uint64                `json:"-"`
	GoodAttestations        uint64                `json:"-"`
	MinipoolShare           *big.Int              `json:"-"`
	MissingAttestationSlots map[uint64]bool       `json:"missingAttestationSlots"`
	WasActive               bool                  `json:"-"`
	StartSlot               uint64                `json:"-"`
	EndSlot                 uint64                `json:"-"`
	AttestationScore        *QuotedBigInt         `json:"attestationScore"`
	CompletedAttestations   map[uint64]bool       `json:"-"`
	AttestationCount        int                   `json:"attestationCount"`
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
	Address          common.Address
	IsEligible       bool
	IsOptedIn        bool
	StatusChangeTime time.Time
	Minipools        []*MinipoolInfo
	EligibleSeconds  *big.Int
	StartSlot        uint64
	EndSlot          uint64
	SmoothingPoolEth *big.Int
	RewardsNetwork   uint64

	// v2 Fields
	OptInTime  time.Time
	OptOutTime time.Time
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

func (versionHeader *VersionHeader) checkVersion() error {
	if versionHeader.RewardsFileVersion == rewardsFileVersionUnknown {
		return fmt.Errorf("unexpected rewards file version [%d]", versionHeader.RewardsFileVersion)
	}

	if versionHeader.RewardsFileVersion > rewardsFileVersionMax {
		return fmt.Errorf("unexpected rewards file version [%d]... highest supported version is [%d], you may need to update Smartnode", versionHeader.RewardsFileVersion, rewardsFileVersionMax)
	}

	return nil
}

func (versionHeader *VersionHeader) deserializeRewardsFile(bytes []byte) (IRewardsFile, error) {
	if err := versionHeader.checkVersion(); err != nil {
		return nil, err
	}

	switch versionHeader.RewardsFileVersion {
	case rewardsFileVersionOne:
		file := &RewardsFile_v1{}
		return file, file.Deserialize(bytes)
	case rewardsFileVersionTwo:
		file := &RewardsFile_v2{}
		return file, file.Deserialize(bytes)
	case rewardsFileVersionThree:
		file := &RewardsFile_v3{}
		return file, file.Deserialize(bytes)
	}

	panic("unreachable section of code reached, please report this error to the maintainers")
}

func (versionHeader *VersionHeader) deserializeMinipoolPerformanceFile(bytes []byte) (IMinipoolPerformanceFile, error) {
	if err := versionHeader.checkVersion(); err != nil {
		return nil, err
	}

	switch versionHeader.RewardsFileVersion {
	case rewardsFileVersionOne:
		file := &MinipoolPerformanceFile_v1{}
		return file, file.Deserialize(bytes)
	case rewardsFileVersionTwo:
		file := &MinipoolPerformanceFile_v2{}
		return file, file.Deserialize(bytes)
	case rewardsFileVersionThree:
		file := &MinipoolPerformanceFile_v2{}
		return file, file.Deserialize(bytes)
	}

	panic("unreachable section of code reached, please report this error to the maintainers")
}
