package api

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao"
	"github.com/rocket-pool/smartnode/bindings/dao/security"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
)

type SecurityStatusResponse struct {
	APIResponse
	IsMember       bool   `json:"isMember"`
	CanJoin        bool   `json:"canJoin"`
	CanLeave       bool   `json:"canLeave"`
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

type SecurityMembersResponse struct {
	APIResponse
	Members []security.SecurityDAOMemberDetails `json:"members"`
}

type SecurityProposalsResponse struct {
	APIResponse
	Proposals []dao.ProposalDetails `json:"proposals"`
}

type SecurityProposalResponse struct {
	APIResponse
	Proposal dao.ProposalDetails `json:"proposal"`
}

type SecurityCanProposeInviteResponse struct {
	APIResponse
	CanPropose          bool            `json:"canPropose"`
	MemberAlreadyExists bool            `json:"memberAlreadyExists"`
	GasLimits           gaslimit.Limits `json:"gasLimits"`
}
type SecurityProposeInviteResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type SecurityCanProposeLeaveResponse struct {
	APIResponse
	CanPropose        bool            `json:"canPropose"`
	MemberDoesntExist bool            `json:"memberDoesntExist"`
	GasLimits         gaslimit.Limits `json:"gasLimits"`
}
type SecurityProposeLeaveResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type SecurityCanProposeKickResponse struct {
	APIResponse
	CanPropose bool            `json:"canPropose"`
	GasLimits  gaslimit.Limits `json:"gasLimits"`
}
type SecurityProposeKickResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type SecurityCanProposeKickMultiResponse struct {
	APIResponse
	CanPropose bool            `json:"canPropose"`
	GasLimits  gaslimit.Limits `json:"gasLimits"`
}
type SecurityProposeKickMultiResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type SecurityCanProposeSettingResponse struct {
	APIResponse
	CanPropose bool            `json:"canPropose"`
	GasLimits  gaslimit.Limits `json:"gasLimits"`
}
type SecurityProposeSettingResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type SecurityCanProposeReplaceResponse struct {
	APIResponse
	CanPropose             bool            `json:"canPropose"`
	OldMemberDoesntExist   bool            `json:"oldMemberDoesntExist"`
	NewMemberAlreadyExists bool            `json:"newMemberAlreadyExists"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
}
type SecurityProposeReplaceResponse struct {
	APIResponse
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type SecurityCanCancelProposalResponse struct {
	APIResponse
	CanCancel       bool            `json:"canCancel"`
	DoesNotExist    bool            `json:"doesNotExist"`
	InvalidState    bool            `json:"invalidState"`
	InvalidProposer bool            `json:"invalidProposer"`
	GasLimits       gaslimit.Limits `json:"gasLimits"`
}
type SecurityCancelProposalResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type SecurityCanVoteOnProposalResponse struct {
	APIResponse
	CanVote            bool            `json:"canVote"`
	DoesNotExist       bool            `json:"doesNotExist"`
	InvalidState       bool            `json:"invalidState"`
	JoinedAfterCreated bool            `json:"joinedAfterCreated"`
	AlreadyVoted       bool            `json:"alreadyVoted"`
	GasLimits          gaslimit.Limits `json:"gasLimits"`
}
type SecurityVoteOnProposalResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type SecurityCanExecuteProposalResponse struct {
	APIResponse
	CanExecute   bool            `json:"canExecute"`
	DoesNotExist bool            `json:"doesNotExist"`
	InvalidState bool            `json:"invalidState"`
	GasLimits    gaslimit.Limits `json:"gasLimits"`
}
type SecurityExecuteProposalResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type SecurityCanJoinResponse struct {
	APIResponse
	CanJoin         bool            `json:"canJoin"`
	ProposalExpired bool            `json:"proposalExpired"`
	AlreadyMember   bool            `json:"alreadyMember"`
	GasLimits       gaslimit.Limits `json:"gasLimits"`
}
type SecurityJoinResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type SecurityCanLeaveResponse struct {
	APIResponse
	CanLeave        bool            `json:"canLeave"`
	ProposalExpired bool            `json:"proposalExpired"`
	GasLimits       gaslimit.Limits `json:"gasLimits"`
}
type SecurityLeaveResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}
