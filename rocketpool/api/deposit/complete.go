package deposit

import (
    "errors"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
    cliutils "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// RocketPool PoolCreated event
type PoolCreated struct {
    Address common.Address
    DurationID [32]byte
    Created *big.Int
}


// Complete a node deposit
func completeDeposit(c *cli.Context) error {

    // Command setup
    am, client, rp, nodeContractAddress, nodeContract, message, err := setup(c, []string{"rocketMinipoolSettings", "rocketNodeAPI", "rocketNodeSettings", "rocketPool", "rocketPoolToken"})
    if message != "" {
        fmt.Println(message)
        return nil
    }
    if err != nil {
        return err
    }

    // Status channels
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

    // Receive status
    for received := 0; received < 3; {
        select {
            case <-successChannel:
                received++
            case msg := <-messageChannel:
                fmt.Println(msg)
                return nil
            case err := <-errorChannel:
                return err
        }
    }

    // Balance channels
    accountBalancesChannel := make(chan *node.Balances)
    nodeBalancesChannel := make(chan *node.Balances)
    requiredBalancesChannel := make(chan *node.Balances)

    // Get node account balances
    go (func() {
        accountBalances, err := node.GetAccountBalances(am.GetNodeAccount().Address, client, rp)
        if err != nil {
            errorChannel <- err
        } else {
            accountBalancesChannel <- accountBalances
        }
    })()

    // Get node balances
    go (func() {
        nodeBalances, err := node.GetBalances(nodeContract)
        if err != nil {
            errorChannel <- err
        } else {
            nodeBalancesChannel <- nodeBalances
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

    // Receive balances
    var accountBalances *node.Balances
    var nodeBalances *node.Balances
    var requiredBalances *node.Balances
    for received := 0; received < 3; {
        select {
            case accountBalances = <-accountBalancesChannel:
                received++
            case nodeBalances = <-nodeBalancesChannel:
                received++
            case requiredBalances = <-requiredBalancesChannel:
                received++
            case err := <-errorChannel:
                return err
        }
    }

    // Get node account transactor
    nodeAccountTransactor, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Check node ether balance and get required deposit transaction value
    depositTransactionValueWei := new(big.Int)
    if nodeBalances.EtherWei.Cmp(requiredBalances.EtherWei) < 0 {

        // Get remaining ether balance required
        remainingEtherRequiredWei := new(big.Int)
        remainingEtherRequiredWei.Sub(requiredBalances.EtherWei, nodeBalances.EtherWei)

        // Check node account balance
        if accountBalances.EtherWei.Cmp(remainingEtherRequiredWei) < 0 {
            fmt.Println(fmt.Sprintf("Node balance of %.2f ETH plus account balance of %.2f ETH is not enough to cover requirement of %.2f ETH", eth.WeiToEth(nodeBalances.EtherWei), eth.WeiToEth(accountBalances.EtherWei), eth.WeiToEth(requiredBalances.EtherWei)))
            return nil
        }

        // Confirm transfer of remaining required ether
        response := cliutils.Prompt(fmt.Sprintf("Node contract requires another %.2f ETH to complete deposit, would you like to pay now from your node account? [y/n]", eth.WeiToEth(remainingEtherRequiredWei)), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            fmt.Println("Deposit not completed")
            return nil
        }

        // Set deposit transaction value
        depositTransactionValueWei.Set(remainingEtherRequiredWei)

    }

    // Check node RPL balance and transfer remaining required RPL
    if nodeBalances.RplWei.Cmp(requiredBalances.RplWei) < 0 {

        // Get remaining RPL balance required
        remainingRplRequiredWei := new(big.Int)
        remainingRplRequiredWei.Sub(requiredBalances.RplWei, nodeBalances.RplWei)

        // Check node account balance
        if accountBalances.RplWei.Cmp(remainingRplRequiredWei) < 0 {
            fmt.Println(fmt.Sprintf("Node balance of %.2f RPL plus account balance of %.2f RPL is not enough to cover requirement of %.2f RPL", eth.WeiToEth(nodeBalances.RplWei), eth.WeiToEth(accountBalances.RplWei), eth.WeiToEth(requiredBalances.RplWei)))
            return nil
        }

        // Confirm transfer of remaining required RPL
        response := cliutils.Prompt(fmt.Sprintf("Node contract requires another %.2f RPL to complete deposit, would you like to pay now from your node account? [y/n]", eth.WeiToEth(remainingRplRequiredWei)), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            fmt.Println("Deposit not completed")
            return nil
        }

        // Transfer remaining required RPL
        nodeAccountTransactor.Value = big.NewInt(0)
        _, err = rp.Contracts["rocketPoolToken"].Transact(nodeAccountTransactor, "transfer", nodeContractAddress, remainingRplRequiredWei)
        if err != nil {
            return errors.New("Error transferring RPL to node contract: " + err.Error())
        }

    }

    // Complete deposit
    nodeAccountTransactor.Value = depositTransactionValueWei
    tx, err := nodeContract.Transact(nodeAccountTransactor, "deposit")
    if err != nil {
        return errors.New("Error completing deposit: " + err.Error())
    }

    // Get minipool created event
    minipoolCreatedEvents, err := eth.GetTransactionEvents(client, tx, rp.Addresses["rocketPool"], rp.Abis["rocketPool"], "PoolCreated", new(PoolCreated))
    if err != nil {
        return errors.New("Error retrieving deposit transaction minipool created event: " + err.Error())
    }
    if len(minipoolCreatedEvents) == 0 {
        return errors.New("Could not retrieve deposit transaction minipool created event")
    }
    minipoolCreatedEvent := (minipoolCreatedEvents[0]).(*PoolCreated)

    // Log & return
    fmt.Println("Deposit completed successfully, minipool created at", minipoolCreatedEvent.Address.Hex())
    return nil

}

