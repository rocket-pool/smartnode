package minipool

import (
    "bytes"
    "errors"
    "fmt"
    "math/big"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Get the node's minipool statuses
func getMinipoolStatus(c *cli.Context) error {

    // Command setup
    if message, err := setup(c, []string{"utilAddressSetStorage"}, []string{}); message != "" {
        fmt.Println(message)
        return nil
    } else if err != nil {
        return err
    }

    // Get node minipool list key
    minipoolListKey := eth.KeccakBytes(bytes.Join([][]byte{[]byte("minipools"), []byte("list.node"), am.GetNodeAccount().Address.Bytes()}, []byte{}))

    // Get node minipool count
    minipoolCount := new(*big.Int)
    if err := cm.Contracts["utilAddressSetStorage"].Call(nil, minipoolCount, "getCount", minipoolListKey); err != nil {
        return errors.New("Error retrieving node minipool count: " + err.Error())
    }

    // Log status & return
    fmt.Println("")
    return nil

}

