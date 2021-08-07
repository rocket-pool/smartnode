package api

import (
	"math/big"
)


type NodeFeeResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    NodeFee float64                 `json:"nodeFee"`
    MinNodeFee float64              `json:"minNodeFee"`
    TargetNodeFee float64           `json:"targetNodeFee"`
    MaxNodeFee float64              `json:"maxNodeFee"`
}


type RplPriceResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    RplPrice *big.Int               `json:"rplPrice"`
    RplPriceBlock uint64            `json:"rplPriceBlock"`
    MinPerMinipoolRplStake *big.Int `json:"minPerMinipoolRplStake"`
    MaxPerMinipoolRplStake *big.Int `json:"maxPerMinipoolRplStake"`
}


type NetworkStatsResponse struct {
    Status string                       `json:"status"`
    Error string                        `json:"error"`
    TotalValueLocked float64            `json:"totalValueLocked"`
    DepositPoolBalance float64          `json:"depositPoolBalance"`
    MinipoolCapacity float64            `json:"minipoolCapacity"`
    StakerUtilization float64           `json:"stakerUtilization"`
    CommissionRate float64              `json:"commissionRate"`
    NodeCount uint64                    `json:"nodeCount"`
    InitializedMinipoolCount uint64     `json:"initializedMinipoolCount"`
    PrelaunchMinipoolCount uint64       `json:"prelaunchMinipoolCount"`
    StakingMinipoolCount uint64         `json:"stakingMinipoolCount"`
    WithdrawableMinipoolCount uint64    `json:"withdrawableMinipoolCount"`
    DissolvedMinipoolCount uint64       `json:"dissolvedMinipoolCount"`
    FinalizedMinipoolCount uint64       `json:"finalizedMinipoolCount"`
    RplPrice float64                    `json:"rplPrice"`
    TotalRplStaked float64              `json:"totalRplStaked"`
    EffectiveRplStaked float64          `json:"effectiveRplStaked"`
    RethPrice float64                   `json:"rethPrice"`
}
