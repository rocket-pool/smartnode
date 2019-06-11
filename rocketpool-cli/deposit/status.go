package deposit

import (
    "fmt"

    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)

// Get the node's current deposit status
func getDepositStatus(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        ClientSync:    true,
        CM:            true,
        NodeContract:  true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        LoadAbis:      []string{"rocketNodeContract"},
    })
    if err != nil {
        return err
    }

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
    var balances *node.Balances
    var reservation *node.ReservationDetails
    for received := 0; received < 2; {
        select {
        case balances = <-balancesChannel:
            received++
        case reservation = <-reservationChannel:
            received++
        case err := <-errorChannel:
            return err
        }
    }

    // Log status & return
    fmt.Println(fmt.Sprintf("Node deposit contract has a balance of %.2f ETH and %.2f RPL", eth.WeiToEth(balances.EtherWei), eth.WeiToEth(balances.RplWei)))
    if reservation.Exists {
        fmt.Println(fmt.Sprintf(
            "Node has a deposit reservation requiring %.2f ETH and %.2f RPL, with a staking duration of %s and expiring at %s",
            eth.WeiToEth(reservation.EtherRequiredWei),
            eth.WeiToEth(reservation.RplRequiredWei),
            reservation.StakingDurationID,
            reservation.ExpiryTime.Format("2006-01-02, 15:04 -0700 MST")))
    } else {
        fmt.Println("Node does not have a current deposit reservation")
    }
    return nil

}
