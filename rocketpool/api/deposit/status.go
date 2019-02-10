package deposit

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
)


// Get the node's current deposit status
func getDepositStatus(c *cli.Context) error {

    // Command setup
    _, rp, nodeContract, message, err := setup(c, []string{"rocketNodeAPI", "rocketNodeSettings"});
    if message != "" {
        fmt.Println(message)
        return nil
    }
    if err != nil {
        return err
    }

    // Get node balances
    etherBalance, rplBalance, err := node.GetBalances(nodeContract)
    if err != nil {
        return err
    }

    // Get node deposit reservation details
    reservation, err := node.GetReservationDetails(nodeContract, rp)
    if err != nil {
        return err
    }

    // Log status & return
    fmt.Println(fmt.Sprintf("Node has a balance of %s ETH and %s RPL", etherBalance.String(), rplBalance.String()))
    if reservation.Exists {
        fmt.Println(fmt.Sprintf(
            "Node has a deposit reservation requiring %s ETH and %s RPL, with a staking duration of %s and expiring at %s",
            reservation.EtherRequired.String(),
            reservation.RplRequired.String(),
            reservation.StakingDurationID,
            reservation.ExpiryTime.Format("2006-01-02, 15:04 -0700 MST")))
    } else {
        fmt.Println("Node does not have a current deposit reservation")
    }
    return nil

}

