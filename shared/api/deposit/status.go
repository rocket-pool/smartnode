package deposit

import (
    "math/big"
    "time"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
)


// Deposit status response type
type DepositStatusResponse struct {

    // Reservation info
    ReservationExists bool                  `json:"reservationExists"`
    ReservationStakingDurationID string     `json:"reservationStakingDurationID"`
    ReservationEtherRequiredWei *big.Int    `json:"reservationEtherRequiredWei"`
    ReservationRplRequiredWei *big.Int      `json:"reservationRplRequiredWei"`
    ReservationExpiryTime time.Time         `json:"reservationExpiryTime"`

    // Node balance info
    NodeBalanceEtherWei *big.Int            `json:"nodeBalanceEtherWei"`
    NodeBalanceRplWei *big.Int              `json:"nodeBalanceRplWei"`

}


// Get deposit status
func GetDepositStatus(p *services.Provider) (*DepositStatusResponse, error) {

    // Response
    response := &DepositStatusResponse{}

    // Status channels
    balancesChannel := make(chan *node.Balances)
    reservationChannel := make(chan *node.ReservationDetails)
    errorChannel := make(chan error)

    // Get node balances
    go (func() {
        if balances, err := node.GetBalances(p.NodeContract); err != nil {
            errorChannel <- err
        } else {
            balancesChannel <- balances
        }
    })()

    // Get node deposit reservation details
    go (func() {
        if reservation, err := node.GetReservationDetails(p.NodeContract, p.CM); err != nil {
            errorChannel <- err
        } else {
            reservationChannel <- reservation
        }
    })()

    // Receive status
    for received := 0; received < 2; {
        select {
            case balances := <-balancesChannel:
                response.NodeBalanceEtherWei = balances.EtherWei
                response.NodeBalanceRplWei = balances.RplWei
                received++
            case reservation := <-reservationChannel:
                response.ReservationExists = reservation.Exists
                response.ReservationStakingDurationID = reservation.StakingDurationID
                response.ReservationEtherRequiredWei = reservation.EtherRequiredWei
                response.ReservationRplRequiredWei = reservation.RplRequiredWei
                response.ReservationExpiryTime = reservation.ExpiryTime
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return response
    return response, nil

}

