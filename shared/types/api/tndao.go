package api

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/dao"
    tn "github.com/rocket-pool/rocketpool-go/dao/trustednode"
)


type TNDAOMembersResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    Members []tn.MemberDetails      `json:"members"`
}


type TNDAOProposalsResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    Proposals []dao.ProposalDetails `json:"proposals"`
}


type CanProposeTNDAOInviteResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanPropose bool                 `json:"canPropose"`
    ProposalCooldownActive bool     `json:"proposalCooldownActive"`
    MemberAlreadyExists bool        `json:"memberAlreadyExists"`
}
type ProposeTNDAOInviteResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    ProposalId uint64               `json:"proposalId"`
    TxHash common.Hash              `json:"txHash"`
}


type CanProposeTNDAOLeaveResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanPropose bool                 `json:"canPropose"`
    ProposalCooldownActive bool     `json:"proposalCooldownActive"`
    InsufficientMembers bool        `json:"insufficientMembers"`
}
type ProposeTNDAOLeaveResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    ProposalId uint64               `json:"proposalId"`
    TxHash common.Hash              `json:"txHash"`
}


type CanProposeTNDAOReplaceResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanPropose bool                 `json:"canPropose"`
    ProposalCooldownActive bool     `json:"proposalCooldownActive"`
    MemberAlreadyExists bool        `json:"memberAlreadyExists"`
}
type ProposeTNDAOReplaceResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    ProposalId uint64               `json:"proposalId"`
    TxHash common.Hash              `json:"txHash"`
}


type CanProposeTNDAOKickResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanPropose bool                 `json:"canPropose"`
    ProposalCooldownActive bool     `json:"proposalCooldownActive"`
    InsufficientRplBond bool        `json:"insufficientRplBond"`
}
type ProposeTNDAOKickResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    ProposalId uint64               `json:"proposalId"`
    TxHash common.Hash              `json:"txHash"`
}


type CanCancelTNDAOProposalResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanCancel bool                  `json:"canCancel"`
    InvalidState bool               `json:"invalidState"`
    InvalidProposer bool            `json:"invalidProposer"`
}
type CancelTNDAOProposalResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanVoteOnTNDAOProposalResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanVote bool                    `json:"canVote"`
    InvalidState bool               `json:"invalidState"`
    JoinedAfterCreated bool         `json:"joinedAfterCreated"`
    AlreadyVoted bool               `json:"alreadyVoted"`
}
type VoteOnTNDAOProposalResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanExecuteTNDAOProposalResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanExecute bool                 `json:"canExecute"`
    InvalidState bool               `json:"invalidState"`
}
type ExecuteTNDAOProposalResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanJoinTNDAOResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanJoin bool                    `json:"canJoin"`
    ProposalExpired bool            `json:"proposalExpired"`
    InsufficientRplBalance bool     `json:"insufficientRplBalance"`
}
type JoinTNDAOResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanLeaveTNDAOResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanLeave bool                   `json:"canLeave"`
    ProposalExpired bool            `json:"proposalExpired"`
    InsufficientMembers bool        `json:"insufficientMembers"`
}
type LeaveTNDAOResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanReplaceTNDAOPositionResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanReplace bool                 `json:"canReplace"`
    ProposalExpired bool            `json:"proposalExpired"`
}
type ReplaceTNDAOPositionResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}

