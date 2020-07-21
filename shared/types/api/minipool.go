package api

import (
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/common"

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
    Status types.MinipoolStatus             `json:"status"`
    StatusBlock int64                       `json:"statusBlock"`
    StatusTime time.Time                    `json:"statusTime"`
    DepositType types.MinipoolDeposit       `json:"depositType"`
    NodeFee float64                         `json:"nodeFee"`
    NodeDepositBalance *big.Int             `json:"nodeDepositBalance"`
    NodeRefundBalance *big.Int              `json:"nodeRefundBalance"`
    NethBalance *big.Int                    `json:"nethBalance"`
    UserDepositBalance *big.Int             `json:"userDepositBalance"`
    UserDepositAssigned bool                `json:"userDepositAssigned"`
    StakingStartBalance *big.Int            `json:"stakingStartBalance"`
    StakingEndBalance *big.Int              `json:"stakingEndBalance"`
    StakingStartBlock int64                 `json:"stakingStartBlock"`
    StakingUserStartBlock int64             `json:"stakingUserStartBlock"`
    StakingEndBlock int64                   `json:"stakingEndBlock"`
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

