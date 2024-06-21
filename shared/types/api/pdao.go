package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
)

type PDAOProposalWithNodeVoteDirection struct {
	protocol.ProtocolDaoProposalDetails
	NodeVoteDirection types.VoteDirection `json:"nodeVoteDirection"`
}

type PDAOProposalsResponse struct {
	Status    string                              `json:"status"`
	Error     string                              `json:"error"`
	Proposals []PDAOProposalWithNodeVoteDirection `json:"proposals"`
}

type PDAOProposalResponse struct {
	Status   string                            `json:"status"`
	Error    string                            `json:"error"`
	Proposal PDAOProposalWithNodeVoteDirection `json:"proposal"`
}

type CanCancelPDAOProposalResponse struct {
	Status          string             `json:"status"`
	Error           string             `json:"error"`
	CanCancel       bool               `json:"canCancel"`
	DoesNotExist    bool               `json:"doesNotExist"`
	InvalidState    bool               `json:"invalidState"`
	InvalidProposer bool               `json:"invalidProposer"`
	GasInfo         rocketpool.GasInfo `json:"gasInfo"`
}
type CancelPDAOProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanVoteOnPDAOProposalResponse struct {
	Status            string             `json:"status"`
	Error             string             `json:"error"`
	CanVote           bool               `json:"canVote"`
	DoesNotExist      bool               `json:"doesNotExist"`
	InvalidState      bool               `json:"invalidState"`
	InsufficientPower bool               `json:"insufficientPower"`
	AlreadyVoted      bool               `json:"alreadyVoted"`
	VotingPower       *big.Int           `json:"votingPower"`
	GasInfo           rocketpool.GasInfo `json:"gasInfo"`
}
type VoteOnPDAOProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanExecutePDAOProposalResponse struct {
	Status       string             `json:"status"`
	Error        string             `json:"error"`
	CanExecute   bool               `json:"canExecute"`
	DoesNotExist bool               `json:"doesNotExist"`
	InvalidState bool               `json:"invalidState"`
	GasInfo      rocketpool.GasInfo `json:"gasInfo"`
}
type ExecutePDAOProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type GetPDAOSettingsResponse struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
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

type CanProposePDAOSettingResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	CanPropose             bool               `json:"canPropose"`
	InsufficientRpl        bool               `json:"proposalCooldownActive"`
	StakedRpl              *big.Int           `json:"stakedRpl"`
	LockedRpl              *big.Int           `json:"lockedRpl"`
	ProposalBond           *big.Int           `json:"proposalBond"`
	BlockNumber            uint32             `json:"blockNumber"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
	IsRplLockingDisallowed bool               `json:"isRplLockingDisallowed"`
}
type ProposePDAOSettingResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOGetRewardsPercentagesResponse struct {
	Status      string   `json:"status"`
	Error       string   `json:"error"`
	Node        *big.Int `json:"node"`
	OracleDao   *big.Int `json:"odao"`
	ProtocolDao *big.Int `json:"pdao"`
}

type PDAOCanProposeRewardsPercentagesResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	BlockNumber            uint32             `json:"blockNumber"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
	CanPropose             bool               `json:"canPropose"`
	IsRplLockingDisallowed bool               `json:"isRplLockingDisallowed"`
}

type PDAOProposeRewardsPercentagesResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeOneTimeSpendResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	BlockNumber            uint32             `json:"blockNumber"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
	CanPropose             bool               `json:"canPropose"`
	IsRplLockingDisallowed bool               `json:"isRplLockingDisallowed"`
}
type PDAOProposeOneTimeSpendResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeRecurringSpendResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	BlockNumber            uint32             `json:"blockNumber"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
	CanPropose             bool               `json:"canPropose"`
	IsRplLockingDisallowed bool               `json:"isRplLockingDisallowed"`
}

type PDAOProposeRecurringSpendResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeRecurringSpendUpdateResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	BlockNumber            uint32             `json:"blockNumber"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
	CanPropose             bool               `json:"canPropose"`
	IsRplLockingDisallowed bool               `json:"isRplLockingDisallowed"`
}

type PDAOProposeRecurringSpendUpdateResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeInviteToSecurityCouncilResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	CanPropose             bool               `json:"canPropose"`
	MemberAlreadyExists    bool               `json:"memberAlreadyExists"`
	BlockNumber            uint32             `json:"blockNumber"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
	IsRplLockingDisallowed bool               `json:"isRplLockingDisallowed"`
}
type PDAOProposeInviteToSecurityCouncilResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeKickFromSecurityCouncilResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	BlockNumber            uint32             `json:"blockNumber"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
	CanPropose             bool               `json:"canPropose"`
	IsRplLockingDisallowed bool               `json:"isRplLockingDisallowed"`
}
type PDAOProposeKickFromSecurityCouncilResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeKickMultiFromSecurityCouncilResponse struct {
	Status      string             `json:"status"`
	Error       string             `json:"error"`
	BlockNumber uint32             `json:"blockNumber"`
	GasInfo     rocketpool.GasInfo `json:"gasInfo"`
}
type PDAOProposeKickMultiFromSecurityCouncilResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeReplaceMemberOfSecurityCouncilResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	BlockNumber            uint32             `json:"blockNumber"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
	CanPropose             bool               `json:"canPropose"`
	IsRplLockingDisallowed bool               `json:"isRplLockingDisallowed"`
}

type PDAOProposeReplaceMemberOfSecurityCouncilResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type BondClaimResult struct {
	ProposalID        uint64   `json:"proposalId"`
	IsProposer        bool     `json:"isProposer"`
	UnlockableIndices []uint64 `json:"unlockableIndices"`
	RewardableIndices []uint64 `json:"rewardableIndices"`
	UnlockAmount      *big.Int `json:"unlockAmount"`
	RewardAmount      *big.Int `json:"rewardAmount"`
}

type PDAOGetClaimableBondsResponse struct {
	Status         string            `json:"status"`
	Error          string            `json:"error"`
	ClaimableBonds []BondClaimResult `json:"claimableBonds"`
}

type PDAOCanClaimBondsResponse struct {
	Status       string             `json:"status"`
	Error        string             `json:"error"`
	IsProposer   bool               `json:"isProposer"`
	CanClaim     bool               `json:"canClaim"`
	DoesNotExist bool               `json:"doesNotExist"`
	InvalidState bool               `json:"invalidState"`
	GasInfo      rocketpool.GasInfo `json:"gasInfo"`
}
type PDAOClaimBondsResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type PDAOCanDefeatProposalResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	CanDefeat              bool               `json:"canDefeat"`
	DoesNotExist           bool               `json:"doesNotExist"`
	AlreadyDefeated        bool               `json:"alreadyDefeated"`
	StillInChallengeWindow bool               `json:"stillInChallengeWindow"`
	InvalidChallengeState  bool               `json:"invalidChallengeState"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
}
type PDAODefeatProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type PDAOCanFinalizeProposalResponse struct {
	Status           string             `json:"status"`
	Error            string             `json:"error"`
	CanFinalize      bool               `json:"canFinalize"`
	DoesNotExist     bool               `json:"doesNotExist"`
	InvalidState     bool               `json:"invalidState"`
	AlreadyFinalized bool               `json:"alreadyFinalized"`
	GasInfo          rocketpool.GasInfo `json:"gasInfo"`
}
type PDAOFinalizeProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type PDAOCanSetVotingDelegateResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}

type PDAOSetVotingDelegateResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type PDAOCurrentVotingDelegateResponse struct {
	Status         string         `json:"status"`
	Error          string         `json:"error"`
	AccountAddress common.Address `json:"accountAddress"`
	VotingDelegate common.Address `json:"votingDelegate"`
}

type PDAOCanInitializeVotingWithDelegateResponse struct {
	Status            string             `json:"status"`
	Error             string             `json:"error"`
	VotingInitialized bool               `json:"votingInitialized"`
	GasInfo           rocketpool.GasInfo `json:"gasInfo"`
}

type PDAOInitializeVotingWithDelegateResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type PDAOCanInitializeVotingResponse struct {
	Status            string             `json:"status"`
	Error             string             `json:"error"`
	VotingInitialized bool               `json:"votingInitialized"`
	GasInfo           rocketpool.GasInfo `json:"gasInfo"`
}

type PDAOInitializeVotingResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type PDAOStatusResponse struct {
	Status                          string         `json:"status"`
	Error                           string         `json:"error"`
	VotingPower                     *big.Int       `json:"votingPower"`
	OnchainVotingDelegate           common.Address `json:"onchainVotingDelegate"`
	OnchainVotingDelegateFormatted  string         `json:"onchainVotingDelegateFormatted"`
	BlockNumber                     uint32         `json:"blockNumber"`
	VerifyEnabled                   bool           `json:"verifyEnabled"`
	IsVotingInitialized             bool           `json:"isVotingInitialized"`
	SnapshotVotingDelegate          common.Address `json:"snapshotVotingDelegate"`
	SnapshotVotingDelegateFormatted string         `json:"snapshotVotingDelegateFormatted"`
	SnapshotResponse                struct {
		Error                   string                 `json:"error"`
		ProposalVotes           []SnapshotProposalVote `json:"proposalVotes"`
		ActiveSnapshotProposals []SnapshotProposal     `json:"activeSnapshotProposals"`
	} `json:"snapshotResponse"`
	IsRPLLockingAllowed     bool           `json:"isRPLLockingAllowed"`
	NodeRPLLocked           *big.Int       `json:"nodeRPLLocked"`
	AccountAddress          common.Address `json:"accountAddress"`
	AccountAddressFormatted string         `json:"accountAddressFormatted"`
	TotalDelegatedVp        *big.Int       `json:"totalDelegateVp"`
	SumVotingPower          *big.Int       `json:"sumVotingPower"`
	IsNodeRegistered        bool           `json:"isNodeRegistered"`
}

type PDAOCanSetSignallingAddressResponse struct {
	Status            string             `json:"status"`
	Error             string             `json:"error"`
	GasInfo           rocketpool.GasInfo `json:"gasInfo"`
	VotingInitialized bool               `json:"votingInitialized"`
	NodeToSigner      common.Address     `json:"nodeToSigner"`
}

type PDAOSetSignallingAddressResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type PDAOCanClearSignallingAddressResponse struct {
	Status            string             `json:"status"`
	Error             string             `json:"error"`
	GasInfo           rocketpool.GasInfo `json:"gasInfo"`
	VotingInitialized bool               `json:"votingInitialized"`
	NodeToSigner      common.Address     `json:"nodeToSigner"`
}

type PDAOClearSignallingAddressResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}
