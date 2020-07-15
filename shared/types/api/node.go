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
    MinipoolCount int       `json:"minipoolCount"`
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

