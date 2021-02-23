package api

import (
    "github.com/ethereum/go-ethereum/common"
)


type TNDAOMembersResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
}


type TNDAOProposalsResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
}


type CanProposeTNDAOInviteResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanPropose bool                 `json:"canPropose"`
}
type ProposeTNDAOInviteResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanProposeTNDAOLeaveResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanPropose bool                 `json:"canPropose"`
}
type ProposeTNDAOLeaveResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanProposeTNDAOReplaceResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanPropose bool                 `json:"canPropose"`
}
type ProposeTNDAOReplaceResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanProposeTNDAOKickResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanPropose bool                 `json:"canPropose"`
}
type ProposeTNDAOKickResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanCancelTNDAOProposalResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanCancel bool                  `json:"canCancel"`
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
}
type ReplaceTNDAOPositionResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}

