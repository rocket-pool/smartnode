package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao"
	tn "github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
)

type TNDAOStatusResponse struct {
	APIResponse
	IsMember       bool   `json:"isMember"`
	CanJoin        bool   `json:"canJoin"`
	CanLeave       bool   `json:"canLeave"`
	CanReplace     bool   `json:"canReplace"`
	TotalMembers   uint64 `json:"totalMembers"`
	ProposalCounts struct {
		Total     int `json:"total"`
		Pending   int `json:"pending"`
		Active    int `json:"active"`
		Cancelled int `json:"cancelled"`
		Defeated  int `json:"defeated"`
		Succeeded int `json:"succeeded"`
		Expired   int `json:"expired"`
		Executed  int `json:"executed"`
	} `json:"proposalCounts"`
}

type TNDAOMembersResponse struct {
	APIResponse
	Members []tn.MemberDetails `json:"members"`
}

type TNDAOProposalsResponse struct {
	APIResponse
	Proposals []dao.ProposalDetails `json:"proposals"`
}

type TNDAOProposalResponse struct {
	APIResponse
	Proposal dao.ProposalDetails `json:"proposal"`
}

type CanProposeTNDAOInviteResponse struct {
	APIResponse
	CanPropose             bool            `json:"canPropose"`
	ProposalCooldownActive bool            `json:"proposalCooldownActive"`
	MemberAlreadyExists    bool            `json:"memberAlreadyExists"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
}
type ProposeTNDAOInviteResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type CanProposeTNDAOLeaveResponse struct {
	APIResponse
	CanPropose             bool            `json:"canPropose"`
	ProposalCooldownActive bool            `json:"proposalCooldownActive"`
	InsufficientMembers    bool            `json:"insufficientMembers"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
}
type ProposeTNDAOLeaveResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type CanProposeTNDAOReplaceResponse struct {
	APIResponse
	CanPropose             bool            `json:"canPropose"`
	ProposalCooldownActive bool            `json:"proposalCooldownActive"`
	MemberAlreadyExists    bool            `json:"memberAlreadyExists"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
}
type ProposeTNDAOReplaceResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type CanProposeTNDAOKickResponse struct {
	APIResponse
	CanPropose             bool            `json:"canPropose"`
	ProposalCooldownActive bool            `json:"proposalCooldownActive"`
	InsufficientRplBond    bool            `json:"insufficientRplBond"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
}
type ProposeTNDAOKickResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type CanCancelTNDAOProposalResponse struct {
	APIResponse
	CanCancel       bool            `json:"canCancel"`
	DoesNotExist    bool            `json:"doesNotExist"`
	InvalidState    bool            `json:"invalidState"`
	InvalidProposer bool            `json:"invalidProposer"`
	GasLimits       gaslimit.Limits `json:"gasLimits"`
}
type CancelTNDAOProposalResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanVoteOnTNDAOProposalResponse struct {
	APIResponse
	CanVote            bool            `json:"canVote"`
	DoesNotExist       bool            `json:"doesNotExist"`
	InvalidState       bool            `json:"invalidState"`
	JoinedAfterCreated bool            `json:"joinedAfterCreated"`
	AlreadyVoted       bool            `json:"alreadyVoted"`
	GasLimits          gaslimit.Limits `json:"gasLimits"`
}
type VoteOnTNDAOProposalResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanExecuteTNDAOProposalResponse struct {
	APIResponse
	CanExecute   bool            `json:"canExecute"`
	DoesNotExist bool            `json:"doesNotExist"`
	InvalidState bool            `json:"invalidState"`
	GasLimits    gaslimit.Limits `json:"gasLimits"`
}
type ExecuteTNDAOProposalResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanExecuteTNDAOUpgradeResponse struct {
	APIResponse
	CanExecute         bool            `json:"canExecute"`
	InvalidTrustedNode bool            `json:"invalidTrustedNode"`
	DoesNotExist       bool            `json:"doesNotExist"`
	InvalidState       bool            `json:"invalidState"`
	GasLimits          gaslimit.Limits `json:"gasLimits"`
}
type ExecuteTNDAOUpgradeResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanJoinTNDAOResponse struct {
	APIResponse
	CanJoin                bool            `json:"canJoin"`
	ProposalExpired        bool            `json:"proposalExpired"`
	AlreadyMember          bool            `json:"alreadyMember"`
	InsufficientRplBalance bool            `json:"insufficientRplBalance"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
}
type JoinTNDAOApproveResponse struct {
	APIResponse
	ApproveTxHash common.Hash `json:"approveTxHash"`
}
type JoinTNDAOJoinResponse struct {
	APIResponse
	JoinTxHash common.Hash `json:"joinTxHash"`
}

type CanLeaveTNDAOResponse struct {
	APIResponse
	CanLeave            bool            `json:"canLeave"`
	ProposalExpired     bool            `json:"proposalExpired"`
	InsufficientMembers bool            `json:"insufficientMembers"`
	GasLimits           gaslimit.Limits `json:"gasLimits"`
}
type LeaveTNDAOResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanReplaceTNDAOPositionResponse struct {
	APIResponse
	CanReplace          bool            `json:"canReplace"`
	ProposalExpired     bool            `json:"proposalExpired"`
	MemberAlreadyExists bool            `json:"memberAlreadyExists"`
	GasLimits           gaslimit.Limits `json:"gasLimits"`
}
type ReplaceTNDAOPositionResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanProposeTNDAOSettingResponse struct {
	APIResponse
	CanPropose             bool            `json:"canPropose"`
	ProposalCooldownActive bool            `json:"proposalCooldownActive"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
}
type ProposeTNDAOSettingMembersQuorumResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingMembersRplBondResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingProposalCooldownResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingProposalVoteTimespanResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingProposalVoteDelayTimespanResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingProposalExecuteTimespanResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingProposalActionTimespanResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingScrubPeriodResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingPromotionScrubPeriodResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingScrubPenaltyEnabledResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingBondReductionWindowStartResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingBondReductionWindowLengthResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type GetTNDAOMemberSettingsResponse struct {
	APIResponse
	Quorum            float64  `json:"quorum"`
	RPLBond           *big.Int `json:"rplBond"`
	ChallengeCooldown uint64   `json:"challengeCooldown"`
	ChallengeWindow   uint64   `json:"challengeWindow"`
	ChallengeCost     *big.Int `json:"challengeCost"`
}
type GetTNDAOProposalSettingsResponse struct {
	APIResponse
	Cooldown      uint64 `json:"cooldown"`
	VoteTime      uint64 `json:"voteTime"`
	VoteDelayTime uint64 `json:"voteDelayTime"`
	ExecuteTime   uint64 `json:"executeTime"`
	ActionTime    uint64 `json:"actionTime"`
}
type GetTNDAOMinipoolSettingsResponse struct {
	APIResponse
	ScrubPeriod               uint64 `json:"scrubPeriod"`
	PromotionScrubPeriod      uint64 `json:"promotionScrubPeriod"`
	ScrubPenaltyEnabled       bool   `json:"scrubPenaltyEnabled"`
	BondReductionWindowStart  uint64 `json:"bondReductionWindowStart"`
	BondReductionWindowLength uint64 `json:"bondReductionWindowLength"`
}

type CanPenaliseMegapoolResponse struct {
	APIResponse
	CanPenalise bool            `json:"canPenalise"`
	GasLimits   gaslimit.Limits `json:"gasLimits"`
}
type PenaliseMegapoolResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}
