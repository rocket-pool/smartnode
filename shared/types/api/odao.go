package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao"
	tn "github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

type TNDAOStatusResponse struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
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
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	Members []tn.MemberDetails `json:"members"`
}

type TNDAOProposalsResponse struct {
	Status    string                `json:"status"`
	Error     string                `json:"error"`
	Proposals []dao.ProposalDetails `json:"proposals"`
}

type TNDAOProposalResponse struct {
	Status    string              `json:"status"`
	Error     string              `json:"error"`
	Proposals dao.ProposalDetails `json:"proposal"`
}

type CanProposeTNDAOInviteResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	CanPropose             bool               `json:"canPropose"`
	ProposalCooldownActive bool               `json:"proposalCooldownActive"`
	MemberAlreadyExists    bool               `json:"memberAlreadyExists"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
}
type ProposeTNDAOInviteResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type CanProposeTNDAOLeaveResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	CanPropose             bool               `json:"canPropose"`
	ProposalCooldownActive bool               `json:"proposalCooldownActive"`
	InsufficientMembers    bool               `json:"insufficientMembers"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
}
type ProposeTNDAOLeaveResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type CanProposeTNDAOReplaceResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	CanPropose             bool               `json:"canPropose"`
	ProposalCooldownActive bool               `json:"proposalCooldownActive"`
	MemberAlreadyExists    bool               `json:"memberAlreadyExists"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
}
type ProposeTNDAOReplaceResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type CanProposeTNDAOKickResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	CanPropose             bool               `json:"canPropose"`
	ProposalCooldownActive bool               `json:"proposalCooldownActive"`
	InsufficientRplBond    bool               `json:"insufficientRplBond"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
}
type ProposeTNDAOKickResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type CanCancelTNDAOProposalResponse struct {
	Status          string             `json:"status"`
	Error           string             `json:"error"`
	CanCancel       bool               `json:"canCancel"`
	DoesNotExist    bool               `json:"doesNotExist"`
	InvalidState    bool               `json:"invalidState"`
	InvalidProposer bool               `json:"invalidProposer"`
	GasInfo         rocketpool.GasInfo `json:"gasInfo"`
}
type CancelTNDAOProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanVoteOnTNDAOProposalResponse struct {
	Status             string             `json:"status"`
	Error              string             `json:"error"`
	CanVote            bool               `json:"canVote"`
	DoesNotExist       bool               `json:"doesNotExist"`
	InvalidState       bool               `json:"invalidState"`
	JoinedAfterCreated bool               `json:"joinedAfterCreated"`
	AlreadyVoted       bool               `json:"alreadyVoted"`
	GasInfo            rocketpool.GasInfo `json:"gasInfo"`
}
type VoteOnTNDAOProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanExecuteTNDAOProposalResponse struct {
	Status       string             `json:"status"`
	Error        string             `json:"error"`
	CanExecute   bool               `json:"canExecute"`
	DoesNotExist bool               `json:"doesNotExist"`
	InvalidState bool               `json:"invalidState"`
	GasInfo      rocketpool.GasInfo `json:"gasInfo"`
}
type ExecuteTNDAOProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanJoinTNDAOResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	CanJoin                bool               `json:"canJoin"`
	ProposalExpired        bool               `json:"proposalExpired"`
	AlreadyMember          bool               `json:"alreadyMember"`
	InsufficientRplBalance bool               `json:"insufficientRplBalance"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
}
type JoinTNDAOApproveResponse struct {
	Status        string      `json:"status"`
	Error         string      `json:"error"`
	ApproveTxHash common.Hash `json:"approveTxHash"`
}
type JoinTNDAOJoinResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	JoinTxHash common.Hash `json:"joinTxHash"`
}

type CanLeaveTNDAOResponse struct {
	Status              string             `json:"status"`
	Error               string             `json:"error"`
	CanLeave            bool               `json:"canLeave"`
	ProposalExpired     bool               `json:"proposalExpired"`
	InsufficientMembers bool               `json:"insufficientMembers"`
	GasInfo             rocketpool.GasInfo `json:"gasInfo"`
}
type LeaveTNDAOResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanReplaceTNDAOPositionResponse struct {
	Status              string             `json:"status"`
	Error               string             `json:"error"`
	CanReplace          bool               `json:"canReplace"`
	ProposalExpired     bool               `json:"proposalExpired"`
	MemberAlreadyExists bool               `json:"memberAlreadyExists"`
	GasInfo             rocketpool.GasInfo `json:"gasInfo"`
}
type ReplaceTNDAOPositionResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanProposeTNDAOSettingResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	CanPropose             bool               `json:"canPropose"`
	ProposalCooldownActive bool               `json:"proposalCooldownActive"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
}
type ProposeTNDAOSettingMembersQuorumResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingMembersRplBondResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingMinipoolUnbondedMaxResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingProposalCooldownResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingProposalVoteTimespanResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingProposalVoteDelayTimespanResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingProposalExecuteTimespanResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingProposalActionTimespanResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}
type ProposeTNDAOSettingScrubPeriodResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	ProposalId uint64      `json:"proposalId"`
	TxHash     common.Hash `json:"txHash"`
}

type GetTNDAOMemberSettingsResponse struct {
	Status              string   `json:"status"`
	Error               string   `json:"error"`
	Quorum              float64  `json:"quorum"`
	RPLBond             *big.Int `json:"rplBond"`
	MinipoolUnbondedMax uint64   `json:"minipoolUnbondedMax"`
	ChallengeCooldown   uint64   `json:"challengeCooldown"`
	ChallengeWindow     uint64   `json:"challengeWindow"`
	ChallengeCost       *big.Int `json:"challengeCost"`
}
type GetTNDAOProposalSettingsResponse struct {
	Status        string `json:"status"`
	Error         string `json:"error"`
	Cooldown      uint64 `json:"cooldown"`
	VoteTime      uint64 `json:"voteTime"`
	VoteDelayTime uint64 `json:"voteDelayTime"`
	ExecuteTime   uint64 `json:"executeTime"`
	ActionTime    uint64 `json:"actionTime"`
}
type GetTNDAOMinipoolSettingsResponse struct {
	Status      string `json:"status"`
	Error       string `json:"error"`
	ScrubPeriod uint64 `json:"scrubPeriod"`
}
