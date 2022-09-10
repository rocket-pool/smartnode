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
	Status                 string   `json:"status"`
	Error                  string   `json:"error"`
	RplPrice               *big.Int `json:"rplPrice"`
	RplPriceBlock          uint64   `json:"rplPriceBlock"`
	MinPerMinipoolRplStake *big.Int `json:"minPerMinipoolRplStake"`
	MaxPerMinipoolRplStake *big.Int `json:"maxPerMinipoolRplStake"`
}

type NetworkStatsResponse struct {
	Status                    string         `json:"status"`
	Error                     string         `json:"error"`
	TotalValueLocked          float64        `json:"totalValueLocked"`
	DepositPoolBalance        float64        `json:"depositPoolBalance"`
	MinipoolCapacity          float64        `json:"minipoolCapacity"`
	StakerUtilization         float64        `json:"stakerUtilization"`
	NodeFee                   float64        `json:"nodeFee"`
	NodeCount                 uint64         `json:"nodeCount"`
	InitializedMinipoolCount  uint64         `json:"initializedMinipoolCount"`
	PrelaunchMinipoolCount    uint64         `json:"prelaunchMinipoolCount"`
	StakingMinipoolCount      uint64         `json:"stakingMinipoolCount"`
	WithdrawableMinipoolCount uint64         `json:"withdrawableMinipoolCount"`
	DissolvedMinipoolCount    uint64         `json:"dissolvedMinipoolCount"`
	FinalizedMinipoolCount    uint64         `json:"finalizedMinipoolCount"`
	RplPrice                  float64        `json:"rplPrice"`
	TotalRplStaked            float64        `json:"totalRplStaked"`
	EffectiveRplStaked        float64        `json:"effectiveRplStaked"`
	RethPrice                 float64        `json:"rethPrice"`
	SmoothingPoolNodes        uint64         `json:"smoothingPoolNodes"`
	SmoothingPoolAddress      common.Address `json:"SmoothingPoolAddress"`
	SmoothingPoolBalance      float64        `json:"smoothingPoolBalance"`
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
