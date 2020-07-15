package api


type NodeStatusResponse struct {
    Status string           `json:"status"`
    Error string            `json:"error"`
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

