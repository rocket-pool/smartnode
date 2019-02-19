package minipool

import (
    "bytes"
    "context"
    "errors"
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"
)


// Withdraw node deposit from a minipool
func withdrawMinipool(c *cli.Context, minipoolAddressStr string) error {

    // Command setup
    if message, err := setup(c, []string{"rocketNodeAPI"}, []string{"rocketMinipool", "rocketNodeContract"}); message != "" {
        fmt.Println(message)
        return nil
    } else if err != nil {
        return err
    }

    // Check node is registered (contract exists)
    nodeContractAddress := new(common.Address)
    if err := cm.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", am.GetNodeAccount().Address); err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    } else if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        fmt.Println("Node is not registered with Rocket Pool, please register with `rocketpool node register`")
        return nil
    }

    // Initialise node contract
    nodeContract, err := cm.NewContract(nodeContractAddress, "rocketNodeContract")
    if err != nil {
        return errors.New("Error initialising node contract: " + err.Error())
    }

    // Get minipool address
    minipoolAddress := common.HexToAddress(minipoolAddressStr)

    // Check contract code at minipool address
    if code, err := client.CodeAt(context.Background(), minipoolAddress, nil); err != nil {
        return errors.New("Error retrieving contract code at minipool address: " + err.Error())
    } else if len(code) == 0 {
        return errors.New("No contract code found at minipool address")
    }

    // Initialise minipool contract
    minipoolContract, err := cm.NewContract(&minipoolAddress, "rocketMinipool")
    if err != nil {
        return errors.New("Error initialising minipool contract: " + err.Error())
    }

    // Status channels
    successChannel := make(chan bool)
    messageChannel := make(chan string)
    errorChannel := make(chan error)

    // Check minipool node owner
    go (func() {
        nodeOwner := new(common.Address)
        if err := minipoolContract.Call(nil, nodeOwner, "getNodeOwner"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool node owner: " + err.Error())
        } else if bytes.Equal(nodeOwner.Bytes(), am.GetNodeAccount().Address.Bytes()) {
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

    // Withdraw node deposit
    if txor, err := am.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        txor.GasLimit = 600000 // Gas estimates on this method are incorrect
        if _, err := nodeContract.Transact(txor, "withdrawMinipoolDeposit", minipoolAddress); err != nil {
            return errors.New("Error withdrawing deposit from minipool: " + err.Error())
        }
    }

    // Log & return
    fmt.Println("Successfully withdrew deposit from minipool at", minipoolAddress.Hex())
    return nil

}

