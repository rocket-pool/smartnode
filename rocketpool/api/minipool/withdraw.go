package minipool

import (
    "bytes"
    "context"
    "errors"
    "fmt"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
)


// Withdraw node deposit from a minipool
func withdrawMinipool(c *cli.Context, minipoolAddressStr string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        LoadAbis: []string{"rocketMinipool", "rocketNodeContract"},
    })
    if err != nil {
        return err
    }

    // Get minipool address
    minipoolAddress := common.HexToAddress(minipoolAddressStr)

    // Check contract code at minipool address
    if code, err := p.Client.CodeAt(context.Background(), minipoolAddress, nil); err != nil {
        return errors.New("Error retrieving contract code at minipool address: " + err.Error())
    } else if len(code) == 0 {
        return errors.New("No contract code found at minipool address")
    }

    // Initialise minipool contract
    minipoolContract, err := p.CM.NewContract(&minipoolAddress, "rocketMinipool")
    if err != nil {
        return errors.New("Error initialising minipool contract: " + err.Error())
    }

    // Status channels
    successChannel := make(chan bool)
    messageChannel := make(chan string)
    errorChannel := make(chan error)

    // Check withdrawals are allowed
    go (func() {
        withdrawalsAllowed := new(bool)
        if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, withdrawalsAllowed, "getWithdrawalAllowed"); err != nil {
            errorChannel <- errors.New("Error checking node withdrawals enabled status: " + err.Error())
        } else if !*withdrawalsAllowed {
            messageChannel <- "Node withdrawals are currently disabled in Rocket Pool"
        } else {
            successChannel <- true
        }
    })()

    // Check minipool node owner
    go (func() {
        nodeOwner := new(common.Address)
        if err := minipoolContract.Call(nil, nodeOwner, "getNodeOwner"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool node owner: " + err.Error())
        } else if bytes.Equal(nodeOwner.Bytes(), p.AM.GetNodeAccount().Address.Bytes()) {
            successChannel <- true
        } else {
            messageChannel <- "Minipool is not owned by this node"
        }
    })()

    // Check minipool status
    go (func() {
        status := new(uint8)
        if err := minipoolContract.Call(nil, status, "getStatus"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status: " + err.Error())
        } else if *status == 0 || *status == 4 || *status == 6 {
            successChannel <- true
        } else {
            messageChannel <- "Minipool is not currently allowing node withdrawals"
        }
    })()

    // Check minipool node deposit exists
    go (func() {
        nodeDepositExists := new(bool)
        if err := minipoolContract.Call(nil, nodeDepositExists, "getNodeDepositExists"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool node deposit status: " + err.Error())
        } else if *nodeDepositExists {
            successChannel <- true
        } else {
            messageChannel <- "Node deposit does not exist in minipool"
        }
    })()

    // Receive status
    for received := 0; received < 4; {
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

    // Withdraw node deposit
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        txor.GasLimit = 800000 // Gas estimates on this method are incorrect
        if tx, err := p.NodeContract.Transact(txor, "withdrawMinipoolDeposit", minipoolAddress); err != nil {
            return errors.New("Error withdrawing deposit from minipool: " + err.Error())
        } else {

            // Wait for transaction to be mined before continuing
            fmt.Println("Deposit withdrawal transaction awaiting mining...")
            if txReceipt, err := bind.WaitMined(context.Background(), p.Client, tx); err != nil {
                return errors.New("Error retrieving deposit withdrawal transaction receipt")
            } else if txReceipt.Status == 0 {
                return errors.New("Deposit withdrawal transaction failed")
            }

        }
    }

    // Log & return
    fmt.Println("Successfully withdrew deposit from minipool at", minipoolAddress.Hex())
    return nil

}

