package api


type NodeStatusResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    AccountAddress string       `json:"accountAddress"`
    Registered bool             `json:"registered"`
    Trusted bool                `json:"trusted"`
    TimezoneLocation string     `json:"timezoneLocation"`
    EthBalance string           `json:"ethBalance"`
    NethBalance string          `json:"nethBalance"`
    MinipoolCounts struct {
        Total int                   `json:"total"`
        Initialized int             `json:"initialized"`
        Prelaunch int               `json:"prelaunch"`
        Staking int                 `json:"staking"`
        Exited int                  `json:"exited"`
        Withdrawable int            `json:"withdrawable"`
        Dissolved int               `json:"dissolved"`
        Refundable int              `json:"refundable"`
    }                           `json:"minipoolCounts"`
}


type CanRegisterNodeResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    CanRegister bool            `json:"canRegister"`
    AlreadyRegistered bool      `json:"alreadyRegistered"`
    RegistrationDisabled bool   `json:"registrationDisabled"`
}


type RegisterNodeResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
}


type SetNodeTimezoneResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
}


type NodeDepositResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
}


type NodeSendResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
}


type NodeBurnResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
}

