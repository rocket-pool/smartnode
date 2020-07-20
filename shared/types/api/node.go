package api

import (
    "math/big"

    "github.com/ethereum/go-ethereum/common"
)


type NodeStatusResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    AccountAddress common.Address   `json:"accountAddress"`
    Registered bool                 `json:"registered"`
    Trusted bool                    `json:"trusted"`
    TimezoneLocation string         `json:"timezoneLocation"`
    EthBalance *big.Int             `json:"ethBalance"`
    NethBalance *big.Int            `json:"nethBalance"`
    MinipoolCounts struct {
        Total int                       `json:"total"`
        Initialized int                 `json:"initialized"`
        Prelaunch int                   `json:"prelaunch"`
        Staking int                     `json:"staking"`
        Exited int                      `json:"exited"`
        Withdrawable int                `json:"withdrawable"`
        Dissolved int                   `json:"dissolved"`
        Refundable int                  `json:"refundable"`
    }                               `json:"minipoolCounts"`
}


type CanRegisterNodeResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanRegister bool                `json:"canRegister"`
    AlreadyRegistered bool          `json:"alreadyRegistered"`
    RegistrationDisabled bool       `json:"registrationDisabled"`
}
type RegisterNodeResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type SetNodeTimezoneResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanNodeDepositResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanDeposit bool                 `json:"canDeposit"`
    InsufficientBalance bool        `json:"insufficientBalance"`
    InvalidAmount bool              `json:"invalidAmount"`
    DepositDisabled bool            `json:"depositDisabled"`
}
type NodeDepositResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
    MinipoolAddress common.Address  `json:"minipoolAddress"`
}


type CanNodeSendResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanSend bool                    `json:"canSend"`
    InsufficientBalance bool        `json:"insufficientBalance"`
}
type NodeSendResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}


type CanNodeBurnResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanBurn bool                    `json:"canBurn"`
    InsufficientBalance bool        `json:"insufficientBalance"`
    InsufficientCollateral bool     `json:"insufficientCollateral"`
}
type NodeBurnResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash common.Hash              `json:"txHash"`
}

