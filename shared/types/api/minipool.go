package api

import (
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/rocket-pool/rocketpool-go/types"
)


type MinipoolStatusResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    Minipools []MinipoolDetails     `json:"minipools"`
}
type MinipoolDetails struct {
    Address common.Address                  `json:"address"`
    ValidatorPubkey types.ValidatorPubkey   `json:"validatorPubkey"`
    Status minipool.StatusDetails           `json:"status"`
    DepositType types.MinipoolDeposit       `json:"depositType"`
    Node minipool.NodeDetails               `json:"node"`
    User minipool.UserDetails               `json:"user"`
    Staking minipool.StakingDetails         `json:"staking"`
    Balances tokens.Balances                `json:"balances"`
    Validator ValidatorDetails              `json:"validator"`
    RefundAvailable bool                    `json:"refundAvailable"`
    WithdrawalAvailable bool                `json:"withdrawalAvailable"`
    WithdrawalAvailableInBlocks uint64      `json:"withdrawalAvailableAfterBlock"`
    CloseAvailable bool                     `json:"closeAvailable"`
}
type ValidatorDetails struct {
    Exists bool                     `json:"exists"`
    Active bool                     `json:"active"`
    Balance *big.Int                `json:"balance"`
    NodeBalance *big.Int            `json:"nodeBalance"`
}


type CanRefundMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanRefund bool                  `json:"canRefund"`
    InsufficientRefundBalance bool  `json:"insufficientRefundBalance"`
}
type RefundMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanDissolveMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanDissolve bool                `json:"canDissolve"`
    InvalidStatus bool              `json:"invalidStatus"`
}
type DissolveMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanExitMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanExit bool                    `json:"canExit"`
    InvalidStatus bool              `json:"invalidStatus"`
}
type ExitMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
}


type CanWithdrawMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanWithdraw bool                `json:"canWithdraw"`
    InvalidStatus bool              `json:"invalidStatus"`
    WithdrawalDelayActive bool      `json:"withdrawalDelayActive"`
}
type WithdrawMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanCloseMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanClose bool                   `json:"canClose"`
    InvalidStatus bool              `json:"invalidStatus"`
}
type CloseMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}

