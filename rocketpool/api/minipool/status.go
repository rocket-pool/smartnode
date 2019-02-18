package minipool

import (
    "bytes"
    "errors"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
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
    minipoolCountV := new(*big.Int)
    if err := cm.Contracts["utilAddressSetStorage"].Call(nil, minipoolCountV, "getCount", minipoolListKey); err != nil {
        return errors.New("Error retrieving node minipool count: " + err.Error())
    }
    minipoolCount := (*minipoolCountV).Int64()

    // Get minipool addresses
    addressChannel := make(chan common.Address)
    errorChannel := make(chan error)
    for mi := int64(0); mi < minipoolCount; mi++ {
        go (func(mi int64) {
            minipoolAddress := new(common.Address)
            if err := cm.Contracts["utilAddressSetStorage"].Call(nil, minipoolAddress, "getItem", minipoolListKey, big.NewInt(mi)); err != nil {
                errorChannel <- errors.New("Error retrieving node minipool address: " + err.Error())
            } else {
                addressChannel <- *minipoolAddress
            }
        })(mi)
    }

    // Receive minipool addresses
    addresses := make([]common.Address, 0)
    for mi := int64(0); mi < minipoolCount; mi++ {
        select {
            case address := <-addressChannel:
                addresses = append(addresses, address)
            case err := <-errorChannel:
                return err
        }
    }

    // Log status & return
    fmt.Println("")
    return nil

}

