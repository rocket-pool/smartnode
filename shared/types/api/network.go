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
	Status                         string   `json:"status"`
	Error                          string   `json:"error"`
	RplPrice                       *big.Int `json:"rplPrice"`
	RplPriceBlock                  uint64   `json:"rplPriceBlock"`
	FivePercentBorrowedRplStake    *big.Int `json:"fivePercentBorrowedRplStake"`
	TenPercentBorrowedRplStake     *big.Int `json:"tenPercentRplStake"`
	FifteenPercentBorrowedRplStake *big.Int `json:"fifteenPercentRplStake"`
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

type SnapshotResponseStruct struct {
	Error                   string                 `json:"error"`
	ProposalVotes           []SnapshotProposalVote `json:"proposalVotes"`
	ActiveSnapshotProposals []SnapshotProposal     `json:"activeSnapshotProposals"`
}

type NetworkDAOProposalsResponse struct {
	Status                         string                 `json:"status"`
	Error                          string                 `json:"error"`
	AccountAddress                 common.Address         `json:"accountAddress"`
	AccountAddressFormatted        string                 `json:"accountAddressFormatted"`
	TotalDelegatedVp               *big.Int               `json:"totalDelegateVp"`
	SumVotingPower                 *big.Int               `json:"sumVotingPower"`
	VotingDelegate                 common.Address         `json:"votingDelegate"`
	IsVotingInitialized            bool                   `json:"isVotingInitialized"`
	VotingPower                    *big.Int               `json:"votingPower"`
	BlockNumber                    uint32                 `json:"blockNumber"`
	IsNodeRegistered               bool                   `json:"isNodeRegistered"`
	OnchainVotingDelegate          common.Address         `json:"onchainVotingDelegate"`
	OnchainVotingDelegateFormatted string                 `json:"onchainVotingDelegateFormatted"`
	SnapshotResponse               SnapshotResponseStruct `json:"snapshotResponse"`
	SignallingAddress              common.Address         `json:"signallingAddress"`
	SignallingAddressFormatted     string                 `json:"SignallingAddressFormatted"`
}

func (s *SnapshotResponseStruct) VoteCount() uint {
	voteCount := uint(0)
	for _, activeProposal := range s.ActiveSnapshotProposals {
		for _, votedProposal := range s.ProposalVotes {
			if votedProposal.Proposal.Id == activeProposal.Id {
				voteCount++
				break
			}
		}
	}
	return voteCount
}

type DownloadRewardsFileResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type GetLatestDelegateResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}

type IsHoustonHotfixDeployedResponse struct {
	Status                  string `json:"status"`
	Error                   string `json:"error"`
	IsHoustonHotfixDeployed bool   `json:"isHoustonHotfixDeployed"`
}
