package api


type NodeFeeResponse struct {
    Status string       `json:"status"`
    Error string        `json:"error"`
    NodeFee float64     `json:"nodeFee"`
}

