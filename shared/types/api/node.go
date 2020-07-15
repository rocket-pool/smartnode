package api


type NodeStatusResponse struct {
    Status string           `json:"status"`
    Error string            `json:"error"`
    PasswordExists bool     `json:"passwordExists"`
    AccountExists bool      `json:"accountExists"`
    Registered bool         `json:"registered"`
    Trusted bool            `json:"trusted"`
    AccountAddress string   `json:"accountAddress"`
    TimezoneLocation string `json:"timezoneLocation"`
    EthBalance string       `json:"ethBalance"`
    NethBalance string      `json:"nethBalance"`
    MinipoolCounts struct {
        Total int               `json:"total"`
        Initialized int         `json:"initialized"`
        Prelaunch int           `json:"prelaunch"`
        Staking int             `json:"staking"`
        Exited int              `json:"exited"`
        Withdrawable int        `json:"withdrawable"`
        Dissolved int           `json:"dissolved"`
        Refundable int          `json:"refundable"`
    }                       `json:"minipoolCounts"`
}


type RegisterNodeResponse struct {
    Status string           `json:"status"`
    Error string            `json:"error"`
}


type SetNodeTimezoneResponse struct {
    Status string           `json:"status"`
    Error string            `json:"error"`
}


type NodeDepositResponse struct {
    Status string           `json:"status"`
    Error string            `json:"error"`
}


type NodeSendResponse struct {
    Status string           `json:"status"`
    Error string            `json:"error"`
}


type NodeBurnResponse struct {
    Status string           `json:"status"`
    Error string            `json:"error"`
}

