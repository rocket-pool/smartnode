package deposit

import (
    "errors"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/common"
    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// RocketPool PoolCreated event
type PoolCreated struct {
    Address common.Address
    DurationID [32]byte
    Created *big.Int
}


// Complete a node deposit
func completeDeposit(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        KM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        NodeContract: true,
        LoadContracts: []string{"rocketETHToken", "rocketMinipoolSettings", "rocketNodeAPI", "rocketNodeSettings", "rocketPool", "rocketPoolToken"},
        LoadAbis: []string{"rocketNodeContract"},
        WaitClientSync: true,
    })
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
        if err := p.NodeContract.Call(nil, hasReservation, "getHasDepositReservation"); err != nil {
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
        if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, depositsAllowed, "getDepositAllowed"); err != nil {
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
        if err := p.CM.Contracts["rocketMinipoolSettings"].Call(nil, minipoolCreationAllowed, "getMinipoolCanBeCreated"); err != nil {
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

    // Get deposit reservation validator pubkey
    validatorPubkey := new([]byte)
    if err := p.NodeContract.Call(nil, validatorPubkey, "getDepositReserveValidatorPubkey"); err != nil {
        return errors.New("Error retrieving deposit reservation validator pubkey: " + err.Error())
    }

    // Check for local validator key
    if _, err := p.KM.GetValidatorKey(*validatorPubkey); err != nil {
        return errors.New("Local validator key matching deposit reservation validator pubkey not found")
    }

    // Balance channels
    accountBalancesChannel := make(chan *node.Balances)
    nodeBalancesChannel := make(chan *node.Balances)
    requiredBalancesChannel := make(chan *node.Balances)

    // Get node account balances
    go (func() {
        if accountBalances, err := node.GetAccountBalances(p.AM.GetNodeAccount().Address, p.Client, p.CM); err != nil {
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

    // Get node balance requirements
    go (func() {
        if requiredBalances, err := node.GetRequiredBalances(p.NodeContract); err != nil {
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
    txor, err := p.AM.GetNodeAccountTransactor()
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
        response := cliutils.Prompt(fmt.Sprintf("Node contract requires %.2f ETH to complete deposit, would you like to pay now from your node account? [y/n]", eth.WeiToEth(remainingEtherRequiredWei)), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
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
        response := cliutils.Prompt(fmt.Sprintf("Node contract requires %.2f RPL to complete deposit, would you like to pay now from your node account? [y/n]", eth.WeiToEth(remainingRplRequiredWei)), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            fmt.Println("Deposit not completed")
            return nil
        }

        // Transfer remaining required RPL
        txor.Value = big.NewInt(0)
        fmt.Println("Transferring RPL to node contract...")
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.CM.Addresses["rocketPoolToken"], p.CM.Abis["rocketPoolToken"], "transfer", p.NodeContractAddress, remainingRplRequiredWei); err != nil {
            return errors.New("Error transferring RPL to node contract: " + err.Error())
        }

    }

    // Complete deposit
    txor.Value = depositTransactionValueWei
    fmt.Println("Completing deposit...")
    txReceipt, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "deposit")
    if err != nil {
        return errors.New("Error completing deposit: " + err.Error())
    }

    // Get minipool created event
    minipoolCreatedEvents, err := eth.GetTransactionEvents(p.Client, txReceipt, p.CM.Addresses["rocketPool"], p.CM.Abis["rocketPool"], "PoolCreated", PoolCreated{})
    if err != nil {
        return errors.New("Error retrieving deposit transaction minipool created event: " + err.Error())
    } else if len(minipoolCreatedEvents) == 0 {
        return errors.New("Could not retrieve deposit transaction minipool created event")
    }
    minipoolCreatedEvent := (minipoolCreatedEvents[0]).(*PoolCreated)

    // Log & return
    fmt.Println("Deposit completed successfully, minipool created at", minipoolCreatedEvent.Address.Hex())
    return nil

}

