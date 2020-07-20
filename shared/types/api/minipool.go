package api

import (
    "time"

    "github.com/rocket-pool/rocketpool-go/minipool"
)


type MinipoolStatusResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    Minipools []struct {
        Address string                          `json:"address"`
        ValidatorPubkey string                  `json:"validatorPubkey"`
        Status minipool.MinipoolStatus          `json:"status"`
        StatusBlock int                         `json:"statusBlock"`
        StatusTime time.Time                    `json:"statusTime"`
        DepositType minipool.MinipoolDeposit    `json:"depositType"`
        NodeFee float64                         `json:"nodeFee"`
        NodeDepositBalance string               `json:"nodeDepositBalance"`
        NodeRefundBalance string                `json:"nodeRefundBalance"`
        NethBalance string                      `json:"nethBalance"`
        UserDepositBalance string               `json:"userDepositBalance"`
        UserDepositAssigned bool                `json:"userDepositAssigned"`
        StakingStartBalance string              `json:"stakingStartBalance"`
        StakingEndBalance string                `json:"stakingEndBalance"`
        StakingStartBlock int                   `json:"stakingStartBlock"`
        StakingUserStartBlock int               `json:"stakingUserStartBlock"`
        StakingEndBlock int                     `json:"stakingEndBlock"`
    }                               `json:"minipools"`
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
    TxHash string                   `json:"txHash"`
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
    TxHash string                   `json:"txHash"`
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
    TxHash string                   `json:"txHash"`
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
    TxHash string                   `json:"txHash"`
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
    TxHash string                   `json:"txHash"`
}

