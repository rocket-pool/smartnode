package api

import (
    "math/big"
)


// Deposit required response type
type DepositRequiredResponse struct {
    Durations []DurationRequirement         `json:"durations"`
}
type DurationRequirement struct {
    DurationId string                       `json:"durationId"`
    EtherAmountWei *big.Int                 `json:"etherAmountWei"`
    RplAmountWei *big.Int                   `json:"rplAmountWei"`
    RplRatioWei *big.Int                    `json:"rplRatioWei"`
    NetworkUtilisationPercentWei *big.Int   `json:"networkUtilisationPercentWei"`
}

