package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type NodeFeeResponse struct {
	Status        string  `json:"status"`
	Error         string  `json:"error"`
	NodeFee       float64 `json:"nodeFee"`
	MinNodeFee    float64 `json:"minNodeFee"`
	TargetNodeFee float64 `json:"targetNodeFee"`
	MaxNodeFee    float64 `json:"maxNodeFee"`
}

type RplPriceResponse struct {
	Status                      string   `json:"status"`
	Error                       string   `json:"error"`
	RplPrice                    *big.Int `json:"rplPrice"`
	RplPriceBlock               uint64   `json:"rplPriceBlock"`
	MinPer8EthMinipoolRplStake  *big.Int `json:"minPer8EthMinipoolRplStake"`
	MaxPer8EthMinipoolRplStake  *big.Int `json:"maxPer8EthMinipoolRplStake"`
	MinPer16EthMinipoolRplStake *big.Int `json:"minPer16EthMinipoolRplStake"`
	MaxPer16EthMinipoolRplStake *big.Int `json:"maxPer16EthMinipoolRplStake"`
}

type NetworkStatsResponse struct {
	Status                    string         `json:"status"`
	Error                     string         `json:"error"`
	TotalValueLocked          *big.Int       `json:"totalValueLocked"`
	DepositPoolBalance        *big.Int       `json:"depositPoolBalance"`
	MinipoolCapacity          *big.Int       `json:"minipoolCapacity"`
	StakerUtilization         *big.Int       `json:"stakerUtilization"`
	NodeFee                   *big.Int       `json:"nodeFee"`
	NodeCount                 uint64         `json:"nodeCount"`
	InitializedMinipoolCount  uint64         `json:"initializedMinipoolCount"`
	PrelaunchMinipoolCount    uint64         `json:"prelaunchMinipoolCount"`
	StakingMinipoolCount      uint64         `json:"stakingMinipoolCount"`
	WithdrawableMinipoolCount uint64         `json:"withdrawableMinipoolCount"`
	DissolvedMinipoolCount    uint64         `json:"dissolvedMinipoolCount"`
	FinalizedMinipoolCount    uint64         `json:"finalizedMinipoolCount"`
	RplPrice                  *big.Int       `json:"rplPrice"`
	TotalRplStaked            *big.Int       `json:"totalRplStaked"`
	EffectiveRplStaked        *big.Int       `json:"effectiveRplStaked"`
	RethPrice                 *big.Int       `json:"rethPrice"`
	SmoothingPoolNodes        uint64         `json:"smoothingPoolNodes"`
	SmoothingPoolAddress      common.Address `json:"SmoothingPoolAddress"`
	SmoothingPoolBalance      *big.Int       `json:"smoothingPoolBalance"`
}

type NetworkTimezonesResponse struct {
	Status         string            `json:"status"`
	Error          string            `json:"error"`
	TimezoneCounts map[string]uint64 `json:"timezoneCounts"`
	TimezoneTotal  uint64            `json:"timezoneTotal"`
	NodeTotal      uint64            `json:"nodeTotal"`
}

type CanNetworkGenerateRewardsTreeResponse struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
	CurrentIndex   uint64 `json:"currentIndex"`
	TreeFileExists bool   `json:"treeFileExists"`
}

type NetworkGenerateRewardsTreeResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type NetworkDAOProposalsResponse struct {
	Status                  string                 `json:"status"`
	Error                   string                 `json:"error"`
	AccountAddress          common.Address         `json:"accountAddress"`
	VotingDelegate          common.Address         `json:"votingDelegate"`
	ActiveSnapshotProposals []SnapshotProposal     `json:"activeSnapshotProposals"`
	ProposalVotes           []SnapshotProposalVote `json:"proposalVotes"`
}

type DownloadRewardsFileResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type IsAtlasDeployedResponse struct {
	Status          string `json:"status"`
	Error           string `json:"error"`
	IsAtlasDeployed bool   `json:"isAtlasDeployed"`
}

type GetLatestDelegateResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}
