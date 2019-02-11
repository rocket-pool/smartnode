package deposit

import (
    "context"
    "errors"
    "fmt"
    "math/big"
    "strings"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
    cliutils "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Complete a node deposit
func completeDeposit(c *cli.Context) error {

    // Command setup
    am, client, rp, nodeContract, message, err := setup(c, []string{"rocketMinipoolSettings", "rocketNodeAPI", "rocketNodeSettings"})
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

    // Check node RPL balance
    if balances.RplWei.Cmp(requiredBalances.RplWei) < 0 {
        fmt.Println(fmt.Sprintf("Node balance of %.2f RPL is not enough to cover requirement of %.2f RPL", eth.WeiToEth(balances.RplWei), eth.WeiToEth(requiredBalances.RplWei)))
        return nil
    }

    // Check node ether balance and get required transaction value
    transactionValueWei := new(big.Int)
    if balances.EtherWei.Cmp(requiredBalances.EtherWei) < 0 {

        // Get remaining ether balance required
        remainingEtherRequiredWei := new(big.Int)
        remainingEtherRequiredWei.Sub(requiredBalances.EtherWei, balances.EtherWei)

        // Get node account balance
        nodeAccountBalance, err := client.BalanceAt(context.Background(), am.GetNodeAccount().Address, nil)
        if err != nil {
            return errors.New("Error retrieving node account balance: " + err.Error())
        }

        // Check node account balance
        if nodeAccountBalance.Cmp(remainingEtherRequiredWei) < 0 {
            fmt.Println(fmt.Sprintf("Node balance of %.2f ETH plus account balance of %.2f ETH is not enough to cover requirement of %.2f ETH", eth.WeiToEth(balances.EtherWei), eth.WeiToEth(nodeAccountBalance), eth.WeiToEth(requiredBalances.EtherWei)))
            return nil
        }

        // Confirm payment of remaining required ether
        response := cliutils.Prompt(fmt.Sprintf("Node contract requires another %.2f ETH to complete deposit, would you like to pay now from your account? [y/n]", eth.WeiToEth(remainingEtherRequiredWei)), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            fmt.Println("Deposit not completed")
            return nil
        }

        // Set transaction value
        transactionValueWei.Set(remainingEtherRequiredWei)

    }

    // Get node account transactor
    nodeAccountTransactor, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }
    nodeAccountTransactor.Value = transactionValueWei

    // Complete deposit
    _, err = nodeContract.Transact(nodeAccountTransactor, "deposit")
    if err != nil {
        return errors.New("Error canceling deposit reservation: " + err.Error())
    }

    // Log & return
    fmt.Println("Deposit completed successfully")
    return nil

}

