package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
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
    CloseAvailable bool                     `json:"closeAvailable"`
}
type ValidatorDetails struct {
    Exists bool                     `json:"exists"`
    Active bool                     `json:"active"`
    Index uint64                    `json:"index"`
    Balance *big.Int                `json:"balance"`
    NodeBalance *big.Int            `json:"nodeBalance"`
}


type CanRefundMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanRefund bool                  `json:"canRefund"`
    InsufficientRefundBalance bool  `json:"insufficientRefundBalance"`
    GasInfo rocketpool.GasInfo      `json:"gasInfo"`
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
    GasInfo rocketpool.GasInfo      `json:"gasInfo"`
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


type CanProcessWithdrawalResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanWithdraw bool                `json:"canWithdraw"`
    InvalidStatus bool              `json:"invalidStatus"`
    GasInfo rocketpool.GasInfo      `json:"gasInfo"`
}
type ProcessWithdrawalResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanProcessWithdrawalAndDestroyResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanWithdraw bool                `json:"canWithdraw"`
    InvalidStatus bool              `json:"invalidStatus"`
    GasInfo rocketpool.GasInfo      `json:"gasInfo"`
}
type ProcessWithdrawalAndDestroyResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanCloseMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanClose bool                   `json:"canClose"`
    InvalidStatus bool              `json:"invalidStatus"`
    InConsensus bool                `json:"inConsensus"`
    GasInfo rocketpool.GasInfo      `json:"gasInfo"`
}
type CloseMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanDestroyMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    GasInfo rocketpool.GasInfo      `json:"gasInfo"`
}
type DestroyMinipoolResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanDelegateUpgradeResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    GasInfo rocketpool.GasInfo      `json:"gasInfo"`
}
type DelegateUpgradeResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanDelegateRollbackResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    GasInfo rocketpool.GasInfo      `json:"gasInfo"`
}
type DelegateRollbackResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanSetUseLatestDelegateResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    GasInfo rocketpool.GasInfo      `json:"gasInfo"`
}
type SetUseLatestDelegateResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type GetUseLatestDelegateResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    Setting bool                    `json:"setting"`
}


type GetDelegateResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    Address common.Address          `json:"address"`
}


type GetPreviousDelegateResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    Address common.Address          `json:"address"`
}


type GetEffectiveDelegateResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    Address common.Address          `json:"address"`
}

