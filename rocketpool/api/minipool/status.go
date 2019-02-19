package minipool

import (
    "bytes"
    "errors"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Get the node's minipool statuses
func getMinipoolStatus(c *cli.Context) error {

    // Command setup
    if message, err := setup(c, []string{"rocketPoolToken", "utilAddressSetStorage"}, []string{"rocketMinipool"}); message != "" {
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

    // Minipool data channels
    addressChannels := make([]chan *common.Address, minipoolCount)
    detailsChannels := make([]chan *minipool.Details, minipoolCount)
    errorChannel := make(chan error)

    // Get minipool addresses
    for mi := int64(0); mi < minipoolCount; mi++ {
        addressChannels[mi] = make(chan *common.Address)
        go (func(mi int64) {
            minipoolAddress := new(common.Address)
            if err := cm.Contracts["utilAddressSetStorage"].Call(nil, minipoolAddress, "getItem", minipoolListKey, big.NewInt(mi)); err != nil {
                errorChannel <- errors.New("Error retrieving node minipool address: " + err.Error())
            } else {
                addressChannels[mi] <- minipoolAddress
            }
        })(mi)
    }

    // Receive minipool addresses
    minipoolAddresses := make([]*common.Address, minipoolCount)
    for mi := int64(0); mi < minipoolCount; mi++ {
        select {
            case address := <-addressChannels[mi]:
                minipoolAddresses[mi] = address
            case err := <-errorChannel:
                return err
        }
    }

    // Get minipool details
    for mi := int64(0); mi < minipoolCount; mi++ {
        detailsChannels[mi] = make(chan *minipool.Details)
        go (func(mi int64) {
            if details, err := minipool.GetDetails(cm, minipoolAddresses[mi]); err != nil {
                errorChannel <- err
            } else {
                detailsChannels[mi] <- details
            }
        })(mi)
    }

    // Receive minipool details
    minipoolDetails := make([]*minipool.Details, minipoolCount)
    for mi := int64(0); mi < minipoolCount; mi++ {
        select {
            case details := <-detailsChannels[mi]:
                minipoolDetails[mi] = details
            case err := <-errorChannel:
                return err
        }
    }

    // Log status & return
    fmt.Println("")
    return nil

}

