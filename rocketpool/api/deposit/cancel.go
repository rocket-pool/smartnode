package deposit

import (
    "errors"
    "fmt"

    "github.com/urfave/cli"
)


// Cancel a node deposit reservation
func cancelDeposit(c *cli.Context) error {

    // Command setup
    am, _, nodeContract, message, err := setup(c, []string{"rocketNodeAPI"})
    if message != "" {
        fmt.Println(message)
        return nil
    }
    if err != nil {
        return err
    }

    // Check node has current deposit reservation
    hasReservation := new(bool)
    err = nodeContract.Call(nil, hasReservation, "getHasDepositReservation")
    if err != nil {
        return errors.New("Error retrieving deposit reservation status: " + err.Error())
    }
    if !*hasReservation {
        fmt.Println("Node does not have a current deposit reservation")
        return nil
    }

    // Get node account transactor
    nodeAccountTransactor, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Cancel deposit reservation
    _, err = nodeContract.Transact(nodeAccountTransactor, "depositReserveCancel")
    if err != nil {
        return errors.New("Error canceling deposit reservation: " + err.Error())
    }

    // Log & return
    fmt.Println("Deposit reservation cancelled successfully")
    return nil

}

