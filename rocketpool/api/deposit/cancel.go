package deposit

import (
    "errors"
    "fmt"

    "github.com/urfave/cli"
)


// Cancel a node deposit reservation
func cancelDeposit(c *cli.Context) error {

    // Command setup
    if message, err := setup(c, []string{"rocketNodeAPI"}, []string{"rocketNodeContract"}); message != "" {
        fmt.Println(message)
        return nil
    } else if err != nil {
        return err
    }

    // Check node has current deposit reservation
    hasReservation := new(bool)
    if err := nodeContract.Call(nil, hasReservation, "getHasDepositReservation"); err != nil {
        return errors.New("Error retrieving deposit reservation status: " + err.Error())
    } else if !*hasReservation {
        fmt.Println("Node does not have a current deposit reservation")
        return nil
    }

    // Cancel deposit reservation
    if txor, err := am.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        if _, err := nodeContract.Transact(txor, "depositReserveCancel"); err != nil {
            return errors.New("Error canceling deposit reservation: " + err.Error())
        }
    }

    // Log & return
    fmt.Println("Deposit reservation cancelled successfully")
    return nil

}

