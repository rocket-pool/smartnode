package deposit

import (
    "errors"
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Complete a node deposit
func completeDeposit(c *cli.Context) error {

    // Command setup
    am, rp, nodeContract, message, err := setup(c, []string{"rocketMinipoolSettings", "rocketNodeAPI", "rocketNodeSettings"})
    if message != "" {
        fmt.Println(message)
        return nil
    }
    if err != nil {
        return err
    }

    // Status channels
    balancesChannel := make(chan *node.Balances)
    requiredBalancesChannel := make(chan *node.Balances)
    successChannel := make(chan bool)
    messageChannel := make(chan string)
    errorChannel := make(chan error)

    // Check node has current deposit reservation
    go (func() {
        hasReservation := new(bool)
        err := nodeContract.Call(nil, hasReservation, "getHasDepositReservation")
        if err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation status: " + err.Error())
        } else if !*hasReservation {
            messageChannel <- "Node does not have a current deposit reservation, please make one with `rocketpool deposit reserve durationID`"
        } else {
            successChannel <- true
        }
    })()

    // Check node deposits are enabled
    go (func() {
        depositsAllowed := new(bool)
        err := rp.Contracts["rocketNodeSettings"].Call(nil, depositsAllowed, "getDepositAllowed")
        if err != nil {
            errorChannel <- errors.New("Error checking node deposits enabled status: " + err.Error())
        } else if !*depositsAllowed {
            messageChannel <- "Node deposits are currently disabled in Rocket Pool"
        } else {
            successChannel <- true
        }
    })()

    // Check minipool creation is enabled
    go (func() {
        minipoolCreationAllowed := new(bool)
        err := rp.Contracts["rocketMinipoolSettings"].Call(nil, minipoolCreationAllowed, "getMinipoolCanBeCreated")
        if err != nil {
            errorChannel <- errors.New("Error checking minipool creation enabled status: " + err.Error())
        } else if !*minipoolCreationAllowed {
            messageChannel <- "Minipool creation is currently disabled in Rocket Pool"
        } else {
            successChannel <- true
        }
    })()

    // Get node balances
    go (func() {
        balances, err := node.GetBalances(nodeContract)
        if err != nil {
            errorChannel <- err
        } else {
            balancesChannel <- balances
        }
    })()

    // Get node balance requirements
    go (func() {
        requiredBalances, err := node.GetRequiredBalances(nodeContract)
        if err != nil {
            errorChannel <- err
        } else {
            requiredBalancesChannel <- requiredBalances
        }
    })()

    // Receive status
    var balances *node.Balances
    var requiredBalances *node.Balances
    for received := 0; received < 5; {
        select {
            case balances = <-balancesChannel:
                received++
            case requiredBalances = <-requiredBalancesChannel:
                received++
            case <-successChannel:
                received++
            case msg := <-messageChannel:
                fmt.Println(msg)
                return nil
            case err := <-errorChannel:
                return err
        }
    }

    // Check node balances
    if balances.EtherWei.Cmp(requiredBalances.EtherWei) < 0 {
        fmt.Println(fmt.Sprintf("Node balance of %.2f ETH is not enough to cover requirement of %.2f ETH", eth.WeiToEth(balances.EtherWei), eth.WeiToEth(requiredBalances.EtherWei)))
        return nil
    }
    if balances.RplWei.Cmp(requiredBalances.RplWei) < 0 {
        fmt.Println(fmt.Sprintf("Node balance of %.2f RPL is not enough to cover requirement of %.2f RPL", eth.WeiToEth(balances.RplWei), eth.WeiToEth(requiredBalances.RplWei)))
        return nil
    }

    // Get node account transactor
    nodeAccountTransactor, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Complete deposit
    _, err = nodeContract.Transact(nodeAccountTransactor, "deposit")
    if err != nil {
        return errors.New("Error canceling deposit reservation: " + err.Error())
    }

    // Log & return
    fmt.Println("Deposit completed successfully")
    return nil

}

