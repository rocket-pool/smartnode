package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/units"
)

type NodeFeeResponse struct {
	APIResponse
	NodeFee       units.Eth `json:"nodeFee"`
	MinNodeFee    units.Eth `json:"minNodeFee"`
	TargetNodeFee units.Eth `json:"targetNodeFee"`
	MaxNodeFee    units.Eth `json:"maxNodeFee"`
}

type RplPriceResponse struct {
	APIResponse
	RplPrice      units.Wei `json:"rplPrice"`
	RplPriceBlock uint64    `json:"rplPriceBlock"`
}

type NetworkStatsResponse struct {
	APIResponse
	TotalValueLocked                 units.Eth      `json:"totalValueLocked"`
	DepositPoolBalance               units.Eth      `json:"depositPoolBalance"`
	MinipoolCapacity                 units.Eth      `json:"minipoolCapacity"`
	StakerUtilization                units.Eth      `json:"stakerUtilization"`
	NodeFee                          units.Eth      `json:"nodeFee"`
	NodeCount                        uint64         `json:"nodeCount"`
	InitializedMinipoolCount         uint64         `json:"initializedMinipoolCount"`
	PrelaunchMinipoolCount           uint64         `json:"prelaunchMinipoolCount"`
	StakingMinipoolCount             uint64         `json:"stakingMinipoolCount"`
	WithdrawableMinipoolCount        uint64         `json:"withdrawableMinipoolCount"`
	DissolvedMinipoolCount           uint64         `json:"dissolvedMinipoolCount"`
	FinalizedMinipoolCount           uint64         `json:"finalizedMinipoolCount"`
	RplPrice                         units.Eth      `json:"rplPrice"`
	TotalRplStaked                   units.Eth      `json:"totalRplStaked"`
	TotalMegapoolRplStaked           units.Eth      `json:"totalMegapoolRplStaked"`
	TotalLegacyRplStaked             units.Eth      `json:"totalLegacyRplStaked"`
	EffectiveRplStaked               float64        `json:"effectiveRplStaked"`
	RethPrice                        units.Eth      `json:"rethPrice"`
	SmoothingPoolNodes               uint64         `json:"smoothingPoolNodes"`
	SmoothingPoolAddress             common.Address `json:"SmoothingPoolAddress"`
	SmoothingPoolBalance             units.Eth      `json:"smoothingPoolBalance"`
	MegapoolContractCount            uint64         `json:"megapoolContractCount"`
	MegapoolValidatorCount           uint64         `json:"megapoolValidatorCount"`
	MegapoolValidatorStakingCount    uint64         `json:"megapoolValidatorStakingCount"`
	MegapoolValidatorInPrestakeCount uint64         `json:"megapoolValidatorInPrestakeCount"`
	MegapoolValidatorInQueueCount    uint64         `json:"megapoolValidatorInQueueCount"`
	MegapoolValidatorExitedCount     uint64         `json:"megapoolValidatorExitedCount"`
	MegapoolValidatorLockedCount     uint64         `json:"megapoolValidatorLockedCount"`
	MegapoolValidatorExitingCount    uint64         `json:"megapoolValidatorExitingCount"`
	MegapoolValidatorDissolvedCount  uint64         `json:"megapoolValidatorDissolvedCount"`
}

type NetworkTimezonesResponse struct {
	APIResponse
	TimezoneCounts map[string]uint64 `json:"timezoneCounts"`
	TimezoneTotal  uint64            `json:"timezoneTotal"`
	NodeTotal      uint64            `json:"nodeTotal"`
}

type CanNetworkGenerateRewardsTreeResponse struct {
	APIResponse
	CurrentIndex   uint64 `json:"currentIndex"`
	TreeFileExists bool   `json:"treeFileExists"`
}

type NetworkGenerateRewardsTreeResponse struct {
	APIResponse
}

type SnapshotResponseStruct struct {
	Error                   string                 `json:"error"`
	ProposalVotes           []SnapshotProposalVote `json:"proposalVotes"`
	ActiveSnapshotProposals []SnapshotProposal     `json:"activeSnapshotProposals"`
}

type NetworkDAOProposalsResponse struct {
	APIResponse
	AccountAddress                 common.Address         `json:"accountAddress"`
	AccountAddressFormatted        string                 `json:"accountAddressFormatted"`
	TotalDelegatedVp               units.Wei              `json:"totalDelegateVp"`
	SumVotingPower                 units.Wei              `json:"sumVotingPower"`
	VotingDelegate                 common.Address         `json:"votingDelegate"`
	VotingPower                    units.Wei              `json:"votingPower"`
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
	APIResponse
}

type GetLatestDelegateResponse struct {
	APIResponse
	Address common.Address `json:"address"`
}
