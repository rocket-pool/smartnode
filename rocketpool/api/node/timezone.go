package node

import (
    "bytes"
    "errors"
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"
)


// Set the node's timezone
func setNodeTimezone(c *cli.Context) error {

    // Command setup
    if message, err := setup(c, []string{"rocketNodeAPI"}, []string{}, true); message != "" {
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

    // Prompt user for timezone
    timezone := promptTimezone()

    // Set node timezone
    if txor, err := am.GetNodeAccountTransactor(); err != nil {
        return err
    } else if _, err := cm.Contracts["rocketNodeAPI"].Transact(txor, "setTimezoneLocation", timezone); err != nil {
        return errors.New("Error setting node timezone: " + err.Error())
    }

    // Get node timezone
    nodeTimezone := new(string)
    if err := cm.Contracts["rocketNodeAPI"].Call(nil, nodeTimezone, "getTimezoneLocation", am.GetNodeAccount().Address); err != nil {
        return errors.New("Error retrieving node timezone: " + err.Error())
    }

    // Log & return
    fmt.Println("Node timezone successfully updated to:", *nodeTimezone)
    return nil

}

