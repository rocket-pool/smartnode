package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/types"
)

var VoteDirectionMap = map[string]types.VoteDirection{
	"abstain": types.VoteDirection_Abstain,
	"for":     types.VoteDirection_For,
	"against": types.VoteDirection_Against,
	"veto":    types.VoteDirection_AgainstWithVeto,
}

var VoteDirectionNameMap = map[types.VoteDirection]string{
	types.VoteDirection_Abstain:         "abstain",
	types.VoteDirection_For:             "for",
	types.VoteDirection_Against:         "against",
	types.VoteDirection_AgainstWithVeto: "veto",
}

type ProtocolDaoProposalDetails struct {
	ID                   uint64                         `json:"id"`
	ProposerAddress      common.Address                 `json:"proposerAddress"`
	TargetBlock          uint32                         `json:"targetBlock"`
	Message              string                         `json:"message"`
	ChallengeWindow      time.Duration                  `json:"challengeWindow"`
	CreatedTime          time.Time                      `json:"createdTime"`
	VotingStartTime      time.Time                      `json:"votingStartTime"`
	Phase1EndTime        time.Time                      `json:"phase1EndTime"`
	Phase2EndTime        time.Time                      `json:"phase2EndTime"`
	ExpiryTime           time.Time                      `json:"expiryTime"`
	VotingPowerRequired  *big.Int                       `json:"votingPowerRequired"`
	VotingPowerFor       *big.Int                       `json:"votingPowerFor"`
	VotingPowerAgainst   *big.Int                       `json:"votingPowerAgainst"`
	VotingPowerAbstained *big.Int                       `json:"votingPowerAbstained"`
	VotingPowerToVeto    *big.Int                       `json:"votingPowerToVeto"`
	IsDestroyed          bool                           `json:"isDestroyed"`
	IsFinalized          bool                           `json:"isFinalized"`
	IsExecuted           bool                           `json:"isExecuted"`
	IsVetoed             bool                           `json:"isVetoed"`
	Payload              []byte                         `json:"payload"`
	PayloadStr           string                         `json:"payloadStr"`
	State                types.ProtocolDaoProposalState `json:"state"`
	ProposalBond         *big.Int                       `json:"proposalBond"`
	ChallengeBond        *big.Int                       `json:"challengeBond"`
	DefeatIndex          uint64                         `json:"defeatIndex"`
	NodeVoteDirection    types.VoteDirection            `json:"nodeVoteDirection"`
}
type ProtocolDaoProposalsData struct {
	Proposals []ProtocolDaoProposalDetails `json:"proposals"`
}

type ProtocolDaoVoteOnProposalData struct {
	CanVote           bool                 `json:"canVote"`
	DoesNotExist      bool                 `json:"doesNotExist"`
	InvalidState      bool                 `json:"invalidState"`
	InsufficientPower bool                 `json:"insufficientPower"`
	AlreadyVoted      bool                 `json:"alreadyVoted"`
	VotingPower       *big.Int             `json:"votingPower"`
	TxInfo            *eth.TransactionInfo `json:"txInfo"`
}

type ProtocolDaoExecuteProposalData struct {
	CanExecute   bool                 `json:"canExecute"`
	DoesNotExist bool                 `json:"doesNotExist"`
	InvalidState bool                 `json:"invalidState"`
	TxInfo       *eth.TransactionInfo `json:"txInfo"`
}

type ProtocolDaoSettingsData struct {
	Auction struct {
		IsCreateLotEnabled    bool          `json:"isCreateLotEnabled"`
		IsBidOnLotEnabled     bool          `json:"isBidOnLotEnabled"`
		LotMinimumEthValue    *big.Int      `json:"lotMinimumEthValue"`
		LotMaximumEthValue    *big.Int      `json:"lotMaximumEthValue"`
		LotDuration           time.Duration `json:"lotDuration"`
		LotStartingPriceRatio *big.Int      `json:"lotStartingPriceRatio"`
		LotReservePriceRatio  *big.Int      `json:"lotReservePriceRatio"`
	} `json:"auction"`

	Deposit struct {
		IsDepositingEnabled                    bool     `json:"isDepositingEnabled"`
		AreDepositAssignmentsEnabled           bool     `json:"areDepositAssignmentsEnabled"`
		MinimumDeposit                         *big.Int `json:"minimumDeposit"`
		MaximumDepositPoolSize                 *big.Int `json:"maximumDepositPoolSize"`
		MaximumAssignmentsPerDeposit           uint64   `json:"maximumAssignmentsPerDeposit"`
		MaximumSocialisedAssignmentsPerDeposit uint64   `json:"maximumSocialisedAssignmentsPerDeposit"`
		DepositFee                             *big.Int `json:"depositFee"`
	} `json:"deposit"`

	Inflation struct {
		IntervalRate *big.Int  `json:"intervalRate"`
		StartTime    time.Time `json:"startTime"`
	} `json:"inflation"`

	Minipool struct {
		IsSubmitWithdrawableEnabled bool          `json:"isSubmitWithdrawableEnabled"`
		LaunchTimeout               time.Duration `json:"launchTimeout"`
		IsBondReductionEnabled      bool          `json:"isBondReductionEnabled"`
		MaximumCount                uint64        `json:"maximumCount"`
		UserDistributeWindowStart   time.Duration `json:"userDistributeWindowStart"`
		UserDistributeWindowLength  time.Duration `json:"userDistributeWindowLength"`
	} `json:"minipool"`

	Network struct {
		OracleDaoConsensusThreshold *big.Int      `json:"oracleDaoConsensusThreshold"`
		NodePenaltyThreshold        *big.Int      `json:"nodePenaltyThreshold"`
		PerPenaltyRate              *big.Int      `json:"perPenaltyRate"`
		IsSubmitBalancesEnabled     bool          `json:"isSubmitBalancesEnabled"`
		SubmitBalancesFrequency     time.Duration `json:"submitBalancesFrequency"`
		IsSubmitPricesEnabled       bool          `json:"isSubmitPricesEnabled"`
		SubmitPricesFrequency       time.Duration `json:"submitPricesFrequency"`
		MinimumNodeFee              *big.Int      `json:"minimumNodeFee"`
		TargetNodeFee               *big.Int      `json:"targetNodeFee"`
		MaximumNodeFee              *big.Int      `json:"maximumNodeFee"`
		NodeFeeDemandRange          *big.Int      `json:"nodeFeeDemandRange"`
		TargetRethCollateralRate    *big.Int      `json:"targetRethCollateralRate"`
		IsSubmitRewardsEnabled      bool          `json:"isSubmitRewardsEnabled"`
	} `json:"network"`

	Node struct {
		IsRegistrationEnabled              bool     `json:"isRegistrationEnabled"`
		IsSmoothingPoolRegistrationEnabled bool     `json:"isSmoothingPoolRegistrationEnabled"`
		IsDepositingEnabled                bool     `json:"isDepositingEnabled"`
		AreVacantMinipoolsEnabled          bool     `json:"areVacantMinipoolsEnabled"`
		MinimumPerMinipoolStake            *big.Int `json:"minimumPerMinipoolStake"`
		MaximumPerMinipoolStake            *big.Int `json:"maximumPerMinipoolStake"`
	} `json:"node"`

	Proposals struct {
		VotePhase1Time  time.Duration `json:"votePhase1Time"`
		VotePhase2Time  time.Duration `json:"votePhase2Time"`
		VoteDelayTime   time.Duration `json:"voteDelayTime"`
		ExecuteTime     time.Duration `json:"executeTime"`
		ProposalBond    *big.Int      `json:"proposalBond"`
		ChallengeBond   *big.Int      `json:"challengeBond"`
		ChallengePeriod time.Duration `json:"challengePeriod"`
		Quorum          *big.Int      `json:"quorum"`
		VetoQuorum      *big.Int      `json:"vetoQuorum"`
		MaxBlockAge     uint64        `json:"maxBlockAge"`
	} `json:"proposals"`

	Rewards struct {
		IntervalTime time.Duration `json:"intervalTime"`
	} `json:"rewards"`

	Security struct {
		MembersQuorum       *big.Int      `json:"membersQuorum"`
		MembersLeaveTime    time.Duration `json:"membersLeaveTime"`
		ProposalVoteTime    time.Duration `json:"proposalVoteTime"`
		ProposalExecuteTime time.Duration `json:"proposalExecuteTime"`
		ProposalActionTime  time.Duration `json:"proposalActionTime"`
	} `json:"security"`
}

type ProtocolDaoRewardsPercentagesData struct {
	Node        *big.Int `json:"node"`
	OracleDao   *big.Int `json:"odao"`
	ProtocolDao *big.Int `json:"pdao"`
}

type ProtocolDaoProposeSettingData struct {
	CanPropose      bool                 `json:"canPropose"`
	UnknownSetting  bool                 `json:"unknownSetting"`
	InsufficientRpl bool                 `json:"insufficientRpl"`
	StakedRpl       *big.Int             `json:"stakedRpl"`
	LockedRpl       *big.Int             `json:"lockedRpl"`
	ProposalBond    *big.Int             `json:"proposalBond"`
	TxInfo          *eth.TransactionInfo `json:"txInfo"`
}

type ProtocolDaoGeneralProposeData struct {
	CanPropose      bool                 `json:"canPropose"`
	InsufficientRpl bool                 `json:"insufficientRpl"`
	StakedRpl       *big.Int             `json:"stakedRpl"`
	LockedRpl       *big.Int             `json:"lockedRpl"`
	ProposalBond    *big.Int             `json:"proposalBond"`
	TxInfo          *eth.TransactionInfo `json:"txInfo"`
}

type ProtocolDaoProposeInviteToSecurityCouncilData struct {
	CanPropose          bool                 `json:"canPropose"`
	MemberAlreadyExists bool                 `json:"memberAlreadyExists"`
	InsufficientRpl     bool                 `json:"insufficientRpl"`
	StakedRpl           *big.Int             `json:"stakedRpl"`
	LockedRpl           *big.Int             `json:"lockedRpl"`
	ProposalBond        *big.Int             `json:"proposalBond"`
	TxInfo              *eth.TransactionInfo `json:"txInfo"`
}

type ProtocolDaoProposeKickFromSecurityCouncilData struct {
	CanPropose         bool                 `json:"canPropose"`
	MemberDoesNotExist bool                 `json:"memberDoesNotExist"`
	InsufficientRpl    bool                 `json:"insufficientRpl"`
	StakedRpl          *big.Int             `json:"stakedRpl"`
	LockedRpl          *big.Int             `json:"lockedRpl"`
	ProposalBond       *big.Int             `json:"proposalBond"`
	TxInfo             *eth.TransactionInfo `json:"txInfo"`
}

type ProtocolDaoProposeKickMultiFromSecurityCouncilData struct {
	CanPropose         bool                 `json:"canPropose"`
	NonexistingMembers []common.Address     `json:"nonexistingMembers"`
	InsufficientRpl    bool                 `json:"insufficientRpl"`
	StakedRpl          *big.Int             `json:"stakedRpl"`
	LockedRpl          *big.Int             `json:"lockedRpl"`
	ProposalBond       *big.Int             `json:"proposalBond"`
	TxInfo             *eth.TransactionInfo `json:"txInfo"`
}

type ProtocolDaoProposeReplaceMemberOfSecurityCouncilData struct {
	CanPropose             bool                 `json:"canPropose"`
	OldMemberDoesNotExist  bool                 `json:"oldMemberDoesNotExist"`
	NewMemberAlreadyExists bool                 `json:"newMemberAlreadyExists"`
	InsufficientRpl        bool                 `json:"insufficientRpl"`
	StakedRpl              *big.Int             `json:"stakedRpl"`
	LockedRpl              *big.Int             `json:"lockedRpl"`
	ProposalBond           *big.Int             `json:"proposalBond"`
	TxInfo                 *eth.TransactionInfo `json:"txInfo"`
}

type BondClaimResult struct {
	ProposalID        uint64   `json:"proposalId"`
	IsProposer        bool     `json:"isProposer"`
	UnlockableIndices []uint64 `json:"unlockableIndices"`
	RewardableIndices []uint64 `json:"rewardableIndices"`
	UnlockAmount      *big.Int `json:"unlockAmount"`
	RewardAmount      *big.Int `json:"rewardAmount"`
}
type ProtocolDaoGetClaimableBondsData struct {
	ClaimableBonds []BondClaimResult `json:"claimableBonds"`
}

type ProtocolDaoClaimBonds struct {
	ProposalID uint64   `json:"proposalId"`
	Indices    []uint64 `json:"indices"`
}

type ProtocolDaoClaimBondsBody struct {
	Claims []ProtocolDaoClaimBonds `json:"claims"`
}

type ProtocolDaoClaimBondsData struct {
	CanClaim     bool                 `json:"canClaim"`
	IsProposer   bool                 `json:"isProposer"`
	DoesNotExist bool                 `json:"doesNotExist"`
	InvalidState bool                 `json:"invalidState"`
	TxInfo       *eth.TransactionInfo `json:"txInfo"`
}

type ProtocolDaoDefeatProposalData struct {
	CanDefeat              bool                 `json:"canDefeat"`
	DoesNotExist           bool                 `json:"doesNotExist"`
	AlreadyDefeated        bool                 `json:"alreadyDefeated"`
	StillInChallengeWindow bool                 `json:"stillInChallengeWindow"`
	InvalidChallengeState  bool                 `json:"invalidChallengeState"`
	TxInfo                 *eth.TransactionInfo `json:"txInfo"`
}

type ProtocolDaoFinalizeProposalData struct {
	CanFinalize      bool                 `json:"canFinalize"`
	DoesNotExist     bool                 `json:"doesNotExist"`
	InvalidState     bool                 `json:"invalidState"`
	AlreadyFinalized bool                 `json:"alreadyFinalized"`
	TxInfo           *eth.TransactionInfo `json:"txInfo"`
}

type ProtocolDaoProposeRecurringSpendUpdateData struct {
	CanPropose      bool                 `json:"canPropose"`
	DoesNotExist    bool                 `json:"doesNotExist"`
	InsufficientRpl bool                 `json:"insufficientRpl"`
	StakedRpl       *big.Int             `json:"stakedRpl"`
	LockedRpl       *big.Int             `json:"lockedRpl"`
	ProposalBond    *big.Int             `json:"proposalBond"`
	TxInfo          *eth.TransactionInfo `json:"txInfo"`
}

type ProtocolDaoInitializeVotingData struct {
	CanInitialize     bool                 `json:"canInitialize"`
	VotingInitialized bool                 `json:"votingInitialized"`
	TxInfo            *eth.TransactionInfo `json:"txInfo"`
}

type ProtocolDaoCurrentVotingDelegateData struct {
	AccountAddress common.Address `json:"accountAddress"`
	VotingDelegate common.Address `json:"votingDelegate"`
}
