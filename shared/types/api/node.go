package api

import (
    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/tokens"
)


type NodeStatusResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    AccountAddress common.Address       `json:"accountAddress"`
    WithdrawalAddress common.Address    `json:"withdrawalAddress"`
    Registered bool                     `json:"registered"`
    Trusted bool                        `json:"trusted"`
    TimezoneLocation string             `json:"timezoneLocation"`
    Balances tokens.Balances            `json:"balances"`
    MinipoolCounts struct {
        Total int                           `json:"total"`
        Initialized int                     `json:"initialized"`
        Prelaunch int                       `json:"prelaunch"`
        Staking int                         `json:"staking"`
        Withdrawable int                    `json:"withdrawable"`
        Dissolved int                       `json:"dissolved"`
        RefundAvailable int                 `json:"refundAvailable"`
        WithdrawalAvailable int             `json:"withdrawalAvailable"`
        CloseAvailable int                  `json:"closeAvailable"`
    }                                   `json:"minipoolCounts"`
}


type CanRegisterNodeResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanRegister bool                    `json:"canRegister"`
    AlreadyRegistered bool              `json:"alreadyRegistered"`
    RegistrationDisabled bool           `json:"registrationDisabled"`
}
type RegisterNodeResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}


type SetNodeTimezoneResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}


type CanNodeDepositResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanDeposit bool                     `json:"canDeposit"`
    InsufficientBalance bool            `json:"insufficientBalance"`
    InvalidAmount bool                  `json:"invalidAmount"`
    DepositDisabled bool                `json:"depositDisabled"`
}
type NodeDepositResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
    MinipoolAddress common.Address      `json:"minipoolAddress"`
}


type CanNodeSendResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanSend bool                        `json:"canSend"`
    InsufficientBalance bool            `json:"insufficientBalance"`
}
type NodeSendResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}


type CanNodeBurnResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    CanBurn bool                        `json:"canBurn"`
    InsufficientBalance bool            `json:"insufficientBalance"`
    InsufficientCollateral bool         `json:"insufficientCollateral"`
}
type NodeBurnResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TxHash common.Hash                  `json:"txHash"`
}

