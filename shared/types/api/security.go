package api

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao"
	"github.com/rocket-pool/smartnode/bindings/dao/security"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
)

type SecurityStatusResponse struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
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
	Status  string                              `json:"status"`
	Error   string                              `json:"error"`
	Members []security.SecurityDAOMemberDetails `json:"members"`
}

type SecurityProposalsResponse struct {
	Status    string                `json:"status"`
	Error     string                `json:"error"`
	Proposals []dao.ProposalDetails `json:"proposals"`
}

type SecurityProposalResponse struct {
	Status   string              `json:"status"`
	Error    string              `json:"error"`
	Proposal dao.ProposalDetails `json:"proposal"`
}

type SecurityCanProposeInviteResponse struct {
	Status              string          `json:"status"`
	Error               string          `json:"error"`
	CanPropose          bool            `json:"canPropose"`
	MemberAlreadyExists bool            `json:"memberAlreadyExists"`
	GasLimits           gaslimit.Limits `json:"gasLimits"`
}
type SecurityProposeInviteResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type SecurityCanProposeLeaveResponse struct {
	Status            string          `json:"status"`
	Error             string          `json:"error"`
	CanPropose        bool            `json:"canPropose"`
	MemberDoesntExist bool            `json:"memberDoesntExist"`
	GasLimits         gaslimit.Limits `json:"gasLimits"`
}
type SecurityProposeLeaveResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type SecurityCanProposeKickResponse struct {
	Status     string          `json:"status"`
	Error      string          `json:"error"`
	CanPropose bool            `json:"canPropose"`
	GasLimits  gaslimit.Limits `json:"gasLimits"`
}
type SecurityProposeKickResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type SecurityCanProposeKickMultiResponse struct {
	Status     string          `json:"status"`
	Error      string          `json:"error"`
	CanPropose bool            `json:"canPropose"`
	GasLimits  gaslimit.Limits `json:"gasLimits"`
}
type SecurityProposeKickMultiResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type SecurityCanProposeSettingResponse struct {
	Status     string          `json:"status"`
	Error      string          `json:"error"`
	CanPropose bool            `json:"canPropose"`
	GasLimits  gaslimit.Limits `json:"gasLimits"`
}
type SecurityProposeSettingResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type SecurityCanProposeReplaceResponse struct {
	Status                 string          `json:"status"`
	Error                  string          `json:"error"`
	CanPropose             bool            `json:"canPropose"`
	OldMemberDoesntExist   bool            `json:"oldMemberDoesntExist"`
	NewMemberAlreadyExists bool            `json:"newMemberAlreadyExists"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
}
type SecurityProposeReplaceResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type SecurityCanCancelProposalResponse struct {
	Status          string          `json:"status"`
	Error           string          `json:"error"`
	CanCancel       bool            `json:"canCancel"`
	DoesNotExist    bool            `json:"doesNotExist"`
	InvalidState    bool            `json:"invalidState"`
	InvalidProposer bool            `json:"invalidProposer"`
	GasLimits       gaslimit.Limits `json:"gasLimits"`
}
type SecurityCancelProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type SecurityCanVoteOnProposalResponse struct {
	Status             string          `json:"status"`
	Error              string          `json:"error"`
	CanVote            bool            `json:"canVote"`
	DoesNotExist       bool            `json:"doesNotExist"`
	InvalidState       bool            `json:"invalidState"`
	JoinedAfterCreated bool            `json:"joinedAfterCreated"`
	AlreadyVoted       bool            `json:"alreadyVoted"`
	GasLimits          gaslimit.Limits `json:"gasLimits"`
}
type SecurityVoteOnProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type SecurityCanExecuteProposalResponse struct {
	Status       string          `json:"status"`
	Error        string          `json:"error"`
	CanExecute   bool            `json:"canExecute"`
	DoesNotExist bool            `json:"doesNotExist"`
	InvalidState bool            `json:"invalidState"`
	GasLimits    gaslimit.Limits `json:"gasLimits"`
}
type SecurityExecuteProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type SecurityCanJoinResponse struct {
	Status          string          `json:"status"`
	Error           string          `json:"error"`
	CanJoin         bool            `json:"canJoin"`
	ProposalExpired bool            `json:"proposalExpired"`
	AlreadyMember   bool            `json:"alreadyMember"`
	GasLimits       gaslimit.Limits `json:"gasLimits"`
}
type SecurityJoinResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type SecurityCanLeaveResponse struct {
	Status          string          `json:"status"`
	Error           string          `json:"error"`
	CanLeave        bool            `json:"canLeave"`
	ProposalExpired bool            `json:"proposalExpired"`
	GasLimits       gaslimit.Limits `json:"gasLimits"`
}
type SecurityLeaveResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}
