package api


type QueueStatusResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    DepositPoolBalance string       `json:"depositPoolBalance"`
    MinipoolQueueLength int         `json:"minipoolQueueLength"`
    MinipoolQueueCapacity string    `json:"minipoolQueueCapacity"`
}


type CanProcessQueueResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    CanProcess bool                 `json:"canProcess"`
    AssignDepositsDisabled bool     `json:"assignDepositsDisabled"`
    NoMinipoolsAvailable bool       `json:"noMinipoolsAvailable"`
    InsufficientDepositBalance bool `json:"insufficientDepositBalance"`
}
type ProcessQueueResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    TxHash string                   `json:"txHash"`
}

