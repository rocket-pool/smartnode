package api


type NodeFeeResponse struct {
    Status string            `json:"status"`
    Error string             `json:"error"`
    NodeFee float64          `json:"nodeFee"`
    MinNodeFee float64       `json:"minNodeFee"`
    TargetNodeFee float64    `json:"targetNodeFee"`
    MaxNodeFee float64       `json:"maxNodeFee"`
    SuggestedNodeFee float64 `json:"suggestedNodeFee"`
}
