package api

import (
    "math/big"
    "time"
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


// Deposit status response type
type DepositStatusResponse struct {
    ReservationExists bool                  `json:"reservationExists"`
    ReservationStakingDurationID string     `json:"reservationStakingDurationID"`
    ReservationEtherRequiredWei *big.Int    `json:"reservationEtherRequiredWei"`
    ReservationRplRequiredWei *big.Int      `json:"reservationRplRequiredWei"`
    ReservationExpiryTime time.Time         `json:"reservationExpiryTime"`
    NodeBalanceEtherWei *big.Int            `json:"nodeBalanceEtherWei"`
    NodeBalanceRplWei *big.Int              `json:"nodeBalanceRplWei"`
}


// Deposit reservation response type
type DepositReserveResponse struct {
    Success bool                            `json:"success"`
    HasExistingReservation bool             `json:"hasExistingReservation"`
    DepositsEnabled bool                    `json:"depositsEnabled"`
    PubkeyUsed bool                         `json:"pubkeyUsed"`
}

