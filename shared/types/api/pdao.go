package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
	"github.com/rocket-pool/smartnode/bindings/types"
)

type PDAOProposalWithNodeVoteDirection struct {
	protocol.ProtocolDaoProposalDetails
	NodeVoteDirection     types.VoteDirection `json:"nodeVoteDirection"`
	DelegateVoteDirection types.VoteDirection `json:"delegateVoteDirection"`
}

type PDAOProposalsResponse struct {
	APIResponse
	Proposals []PDAOProposalWithNodeVoteDirection `json:"proposals"`
}

type PDAOProposalResponse struct {
	APIResponse
	Proposal PDAOProposalWithNodeVoteDirection `json:"proposal"`
}

type CanCancelPDAOProposalResponse struct {
	APIResponse
	CanCancel       bool            `json:"canCancel"`
	DoesNotExist    bool            `json:"doesNotExist"`
	InvalidState    bool            `json:"invalidState"`
	InvalidProposer bool            `json:"invalidProposer"`
	GasLimits       gaslimit.Limits `json:"gasLimits"`
}
type CancelPDAOProposalResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanVoteOnPDAOProposalResponse struct {
	APIResponse
	CanVote           bool            `json:"canVote"`
	DoesNotExist      bool            `json:"doesNotExist"`
	InvalidState      bool            `json:"invalidState"`
	InsufficientPower bool            `json:"insufficientPower"`
	AlreadyVoted      bool            `json:"alreadyVoted"`
	VotingPower       *big.Int        `json:"votingPower"`
	GasLimits         gaslimit.Limits `json:"gasLimits"`
}
type VoteOnPDAOProposalResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanExecutePDAOProposalResponse struct {
	APIResponse
	CanExecute   bool            `json:"canExecute"`
	DoesNotExist bool            `json:"doesNotExist"`
	InvalidState bool            `json:"invalidState"`
	GasLimits    gaslimit.Limits `json:"gasLimits"`
}
type ExecutePDAOProposalResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type GetPDAOSettingsResponse struct {
	APIResponse
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
		ExpressQueueRate                       uint64   `json:"expressQueueRate"`
		ExpressQueueTicketsBaseProvision       uint64   `json:"expressQueueTicketsBaseProvision"`
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
		MaximumPenaltyCount         uint64        `json:"maximumPenaltyCount"`
	} `json:"minipool"`

	Network struct {
		OracleDaoConsensusThreshold             *big.Int         `json:"oracleDaoConsensusThreshold"`
		NodePenaltyThreshold                    *big.Int         `json:"nodePenaltyThreshold"`
		PerPenaltyRate                          *big.Int         `json:"perPenaltyRate"`
		IsSubmitBalancesEnabled                 bool             `json:"isSubmitBalancesEnabled"`
		SubmitBalancesFrequency                 time.Duration    `json:"submitBalancesFrequency"`
		IsSubmitPricesEnabled                   bool             `json:"isSubmitPricesEnabled"`
		SubmitPricesFrequency                   time.Duration    `json:"submitPricesFrequency"`
		MinimumNodeFee                          *big.Int         `json:"minimumNodeFee"`
		TargetNodeFee                           *big.Int         `json:"targetNodeFee"`
		MaximumNodeFee                          *big.Int         `json:"maximumNodeFee"`
		NodeFeeDemandRange                      *big.Int         `json:"nodeFeeDemandRange"`
		TargetRethCollateralRate                *big.Int         `json:"targetRethCollateralRate"`
		IsSubmitRewardsEnabled                  bool             `json:"isSubmitRewardsEnabled"`
		NodeCommissionShare                     *big.Int         `json:"nodeCommissionShare"`
		NodeCommissionShareSecurityCouncilAdder *big.Int         `json:"nodeCommissionShareSecurityCouncilAdder"`
		VoterShare                              *big.Int         `json:"voterShare"`
		ProtocolDAOShare                        *big.Int         `json:"protocolDAOShare"`
		MaxNodeShareSecurityCouncilAdder        *big.Int         `json:"maxNodeCommissionShareCouncilAdder"`
		MaxRethBalanceDelta                     *big.Int         `json:"maxRethBalanceDelta"`
		AllowListedControllers                  []common.Address `json:"allowListedControllers"`
		RethDepositDelay                        uint64           `json:"rethDepositDelay"`
	} `json:"network"`

	Node struct {
		IsRegistrationEnabled              bool          `json:"isRegistrationEnabled"`
		IsSmoothingPoolRegistrationEnabled bool          `json:"isSmoothingPoolRegistrationEnabled"`
		IsDepositingEnabled                bool          `json:"isDepositingEnabled"`
		AreVacantMinipoolsEnabled          bool          `json:"areVacantMinipoolsEnabled"`
		MinimumLegacyRplStake              *big.Int      `json:"minimumLegacyRplStake"`
		ReducedBond                        float64       `json:"reducedBond"`
		NodeUnstakingPeriod                time.Duration `json:"nodeUnstakingPeriod"`
		WithdrawalCooldown                 time.Duration `json:"withdrawalCooldown"`
		MaximumStakeForVotingPower         *big.Int      `json:"maximumStakeForVotingPower"`
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
		UpgradeVetoQuorum   *big.Int      `json:"upgradeVetoQuorum"`
		UpgradeDelay        time.Duration `json:"upgradeDelay"`
	} `json:"security"`

	Megapool struct {
		TimeBeforeDissolve               time.Duration `json:"timeBeforeDissolve"`
		MaximumEthPenalty                *big.Int      `json:"maximumEthPenalty"`
		NotifyThreshold                  uint64        `json:"notifyThreshold"`
		LateNotifyFine                   *big.Int      `json:"lateNotifyFine"`
		DissolvePenalty                  *big.Int      `json:"dissolvePenalty"`
		UserDistributeDelay              uint64        `json:"userDistributeDelay"`
		UserDistributeDelayWithShortfall uint64        `json:"userDistributeDelayWithShortfall"`
		PenaltyThreshold                 *big.Int      `json:"penaltyThreshold"`
	} `json:"megapool"`
}

type CanProposePDAOSettingResponse struct {
	APIResponse
	CanPropose             bool            `json:"canPropose"`
	InsufficientRpl        bool            `json:"proposalCooldownActive"`
	StakedRpl              *big.Int        `json:"stakedRpl"`
	LockedRpl              *big.Int        `json:"lockedRpl"`
	ProposalBond           *big.Int        `json:"proposalBond"`
	BlockNumber            uint32          `json:"blockNumber"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
	IsRplLockingDisallowed bool            `json:"isRplLockingDisallowed"`
}
type ProposePDAOSettingResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOBatchSetting struct {
	Contract string `json:"contract"`
	Setting  string `json:"setting"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value"`
}

type CanProposePDAOSettingMultiResponse struct {
	APIResponse
	CanPropose             bool            `json:"canPropose"`
	InsufficientRpl        bool            `json:"proposalCooldownActive"`
	StakedRpl              *big.Int        `json:"stakedRpl"`
	LockedRpl              *big.Int        `json:"lockedRpl"`
	ProposalBond           *big.Int        `json:"proposalBond"`
	BlockNumber            uint32          `json:"blockNumber"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
	IsRplLockingDisallowed bool            `json:"isRplLockingDisallowed"`
}

type ProposePDAOSettingMultiResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOGetRewardsPercentagesResponse struct {
	APIResponse
	Node        *big.Int `json:"node"`
	OracleDao   *big.Int `json:"odao"`
	ProtocolDao *big.Int `json:"pdao"`
}

type PDAOCanProposeRewardsPercentagesResponse struct {
	APIResponse
	BlockNumber            uint32          `json:"blockNumber"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
	CanPropose             bool            `json:"canPropose"`
	IsRplLockingDisallowed bool            `json:"isRplLockingDisallowed"`
}

type PDAOProposeRewardsPercentagesResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeOneTimeSpendResponse struct {
	APIResponse
	BlockNumber            uint32          `json:"blockNumber"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
	CanPropose             bool            `json:"canPropose"`
	IsRplLockingDisallowed bool            `json:"isRplLockingDisallowed"`
}
type PDAOProposeOneTimeSpendResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeRecurringSpendResponse struct {
	APIResponse
	BlockNumber            uint32          `json:"blockNumber"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
	CanPropose             bool            `json:"canPropose"`
	IsRplLockingDisallowed bool            `json:"isRplLockingDisallowed"`
}

type PDAOProposeRecurringSpendResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeRecurringSpendUpdateResponse struct {
	APIResponse
	BlockNumber            uint32          `json:"blockNumber"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
	CanPropose             bool            `json:"canPropose"`
	IsRplLockingDisallowed bool            `json:"isRplLockingDisallowed"`
}

type PDAOProposeRecurringSpendUpdateResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeInviteToSecurityCouncilResponse struct {
	APIResponse
	CanPropose             bool            `json:"canPropose"`
	MemberAlreadyExists    bool            `json:"memberAlreadyExists"`
	BlockNumber            uint32          `json:"blockNumber"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
	IsRplLockingDisallowed bool            `json:"isRplLockingDisallowed"`
}
type PDAOProposeInviteToSecurityCouncilResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeKickFromSecurityCouncilResponse struct {
	APIResponse
	BlockNumber            uint32          `json:"blockNumber"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
	CanPropose             bool            `json:"canPropose"`
	IsRplLockingDisallowed bool            `json:"isRplLockingDisallowed"`
}
type PDAOProposeKickFromSecurityCouncilResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeKickMultiFromSecurityCouncilResponse struct {
	APIResponse
	BlockNumber uint32          `json:"blockNumber"`
	GasLimits   gaslimit.Limits `json:"gasLimits"`
}
type PDAOProposeKickMultiFromSecurityCouncilResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type PDAOCanProposeReplaceMemberOfSecurityCouncilResponse struct {
	APIResponse
	BlockNumber            uint32          `json:"blockNumber"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
	CanPropose             bool            `json:"canPropose"`
	IsRplLockingDisallowed bool            `json:"isRplLockingDisallowed"`
}

type PDAOProposeReplaceMemberOfSecurityCouncilResponse struct {
	APIResponse
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
	APIResponse
	ClaimableBonds []BondClaimResult `json:"claimableBonds"`
}

type PDAOCanClaimBondsResponse struct {
	APIResponse
	IsProposer   bool            `json:"isProposer"`
	CanClaim     bool            `json:"canClaim"`
	DoesNotExist bool            `json:"doesNotExist"`
	InvalidState bool            `json:"invalidState"`
	GasLimits    gaslimit.Limits `json:"gasLimits"`
}
type PDAOClaimBondsResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type PDAOCanDefeatProposalResponse struct {
	APIResponse
	CanDefeat              bool            `json:"canDefeat"`
	DoesNotExist           bool            `json:"doesNotExist"`
	AlreadyDefeated        bool            `json:"alreadyDefeated"`
	StillInChallengeWindow bool            `json:"stillInChallengeWindow"`
	InvalidChallengeState  bool            `json:"invalidChallengeState"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
}
type PDAODefeatProposalResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type PDAOCanFinalizeProposalResponse struct {
	APIResponse
	CanFinalize      bool            `json:"canFinalize"`
	DoesNotExist     bool            `json:"doesNotExist"`
	InvalidState     bool            `json:"invalidState"`
	AlreadyFinalized bool            `json:"alreadyFinalized"`
	GasLimits        gaslimit.Limits `json:"gasLimits"`
}
type PDAOFinalizeProposalResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type PDAOCanSetVotingDelegateResponse struct {
	APIResponse
	GasLimits gaslimit.Limits `json:"gasLimits"`
}

type PDAOSetVotingDelegateResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type PDAOCurrentVotingDelegateResponse struct {
	APIResponse
	AccountAddress common.Address `json:"accountAddress"`
	VotingDelegate common.Address `json:"votingDelegate"`
}

type PDAOCanInitializeVotingWithDelegateResponse struct {
	APIResponse
	VotingInitialized bool            `json:"votingInitialized"`
	GasLimits         gaslimit.Limits `json:"gasLimits"`
}

type PDAOInitializeVotingWithDelegateResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type PDAOCanInitializeVotingResponse struct {
	APIResponse
	VotingInitialized bool            `json:"votingInitialized"`
	GasLimits         gaslimit.Limits `json:"gasLimits"`
}

type PDAOInitializeVotingResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type PDAOIsVotingInitializedResponse struct {
	APIResponse
	VotingInitialized bool `json:"votingInitialized"`
}

type PDAOStatusResponse struct {
	APIResponse
	VotingPower                    *big.Int               `json:"votingPower"`
	OnchainVotingDelegate          common.Address         `json:"onchainVotingDelegate"`
	OnchainVotingDelegateFormatted string                 `json:"onchainVotingDelegateFormatted"`
	BlockNumber                    uint32                 `json:"blockNumber"`
	VerifyEnabled                  bool                   `json:"verifyEnabled"`
	SnapshotResponse               SnapshotResponseStruct `json:"snapshotResponse"`
	IsRPLLockingAllowed            bool                   `json:"isRPLLockingAllowed"`
	NodeRPLLocked                  *big.Int               `json:"nodeRPLLocked"`
	AccountAddress                 common.Address         `json:"accountAddress"`
	AccountAddressFormatted        string                 `json:"accountAddressFormatted"`
	TotalDelegatedVp               *big.Int               `json:"totalDelegateVp"`
	SumVotingPower                 *big.Int               `json:"sumVotingPower"`
	IsNodeRegistered               bool                   `json:"isNodeRegistered"`
	SignallingAddress              common.Address         `json:"signallingAddress"`
	SignallingAddressFormatted     string                 `json:"SignallingAddressFormatted"`
}

type PDAOCanSetSignallingAddressResponse struct {
	APIResponse
	GasLimits    gaslimit.Limits `json:"gasLimits"`
	NodeToSigner common.Address  `json:"nodeToSigner"`
}

type PDAOSetSignallingAddressResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type PDAOCanClearSignallingAddressResponse struct {
	APIResponse
	GasLimits         gaslimit.Limits `json:"gasLimits"`
	VotingInitialized bool            `json:"votingInitialized"`
	NodeToSigner      common.Address  `json:"nodeToSigner"`
}

type PDAOClearSignallingAddressResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type PDAOACanProposeAllowListedControllersResponse struct {
	APIResponse
	BlockNumber            uint32          `json:"blockNumber"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
	CanPropose             bool            `json:"canPropose"`
	IsRplLockingDisallowed bool            `json:"isRplLockingDisallowed"`
}
type PDAOProposeAllowListedControllersResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
