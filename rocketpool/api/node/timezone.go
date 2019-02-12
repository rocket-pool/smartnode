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
    am, _, rp, message, err := setup(c, []string{"rocketNodeAPI"}, []string{}, true)
    if message != "" {
        fmt.Println(message)
        return nil
    }
    if err != nil {
        return err
    }

    // Check node is registered (contract exists)
    nodeContractAddress := new(common.Address)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", am.GetNodeAccount().Address)
    if err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    }
    if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        fmt.Println("Node is not registered with Rocket Pool, please register with `rocketpool node register`")
        return nil
    }

    // Prompt user for timezone
    timezone := promptTimezone()

    // Get node account transactor
    nodeAccountTransactor, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Set node timezone
    _, err = rp.Contracts["rocketNodeAPI"].Transact(nodeAccountTransactor, "setTimezoneLocation", timezone)
    if err != nil {
        return errors.New("Error setting node timezone: " + err.Error())
    }

    // Get node timezone
    nodeTimezone := new(string)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeTimezone, "getTimezoneLocation", am.GetNodeAccount().Address)
    if err != nil {
        return errors.New("Error retrieving node timezone: " + err.Error())
    }

    // Log & return
    fmt.Println("Node timezone successfully updated to:", *nodeTimezone)
    return nil

}

