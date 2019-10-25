package deposit

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Get the current deposit status
func getDepositStatus(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        CM: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Response
    response := api.DepositStatusResponse{}

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
                return err
        }
    }

    // Print response & return
    api.PrintResponse(p.Output, response)
    return nil

}

