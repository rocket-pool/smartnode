package api

import (
    "math/big"
)


type NodeFeeResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    NodeFee float64             `json:"nodeFee"`
    MinNodeFee float64          `json:"minNodeFee"`
    TargetNodeFee float64       `json:"targetNodeFee"`
    MaxNodeFee float64          `json:"maxNodeFee"`
    SuggestedMinNodeFee float64 `json:"suggestedMinNodeFee"`
}


type RplPriceResponse struct {
    Status string               `json:"status"`
    Error string                `json:"error"`
    RplPrice *big.Int           `json:"rplPrice"`
    RplPriceBlock uint64        `json:"rplPriceBlock"`
}
