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
    NodeAccountBalanceEtherWei *big.Int     `json:"nodeAccountBalanceEtherWei"`
    NodeAccountBalanceRplWei *big.Int       `json:"nodeAccountBalanceRplWei"`
    NodeContractBalanceEtherWei *big.Int    `json:"nodeContractBalanceEtherWei"`
    NodeContractBalanceRplWei *big.Int      `json:"nodeContractBalanceRplWei"`

}


// Get deposit status
func GetDepositStatus(p *services.Provider) (*DepositStatusResponse, error) {

    // Response
    response := &DepositStatusResponse{}

    // Status channels
    accountBalancesChannel := make(chan *node.Balances)
    nodeBalancesChannel := make(chan *node.Balances)
    reservationChannel := make(chan *node.ReservationDetails)
    errorChannel := make(chan error)

    // Get node account balances
    go (func() {
        nodeAccount, _ := p.AM.GetNodeAccount()
        if accountBalances, err := node.GetAccountBalances(nodeAccount.Address, p.Client, p.CM); err != nil {
            errorChannel <- err
        } else {
            accountBalancesChannel <- accountBalances
        }
    })()

    // Get node balances
    go (func() {
        if nodeBalances, err := node.GetBalances(p.NodeContract); err != nil {
            errorChannel <- err
        } else {
            nodeBalancesChannel <- nodeBalances
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
    for received := 0; received < 3; {
        select {
            case accountBalances := <- accountBalancesChannel:
                response.NodeAccountBalanceEtherWei = accountBalances.EtherWei
                response.NodeAccountBalanceRplWei = accountBalances.RplWei
                received++
            case nodeBalances := <-balancesChannel:
                response.NodeContractBalanceEtherWei = nodeBalances.EtherWei
                response.NodeContractBalanceRplWei = nodeBalances.RplWei
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

