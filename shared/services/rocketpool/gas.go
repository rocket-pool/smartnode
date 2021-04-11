package rocketpool

import (
    "fmt"
    "math/big"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/rocket-pool/smartnode/shared/utils/math"
)


// Print estimated gas cost and any requested gas parameters
func (rp *Client) PrintGasInfo(gasInfo rocketpool.GasInfo) {

    // Print gas price, gas limit and total eth cost as estimated by the network
    gas := new(big.Int).SetUint64(gasInfo.EstGasLimit)
    var gasPrice *big.Int
    if gasInfo.EstGasPrice != nil {
        gasPrice = gasInfo.EstGasPrice
    } else {
        gasPrice = big.NewInt(0)
    }
    totalGasWei := new(big.Int).Mul(gasPrice, gas)
    fmt.Printf("Estimate gas price: %.6f Gwei, Estimate gas: %d, Estimate gas cost: %.6f ETH\n", 
               eth.WeiToGwei(gasPrice), 
               gasInfo.EstGasLimit, 
               math.RoundDown(eth.WeiToEth(totalGasWei), 6))
    
    // Print gas price, gas limit and max gas cost as requested by the user
    var userGasMessage string
    if gasInfo.ReqGasPrice != nil {
        userGasMessage += fmt.Sprintf("Requested gas price: %.6f Gwei", eth.WeiToGwei(gasInfo.ReqGasPrice))
    }
    if gasInfo.ReqGasLimit != 0 {
        if len(userGasMessage) > 0 {
            userGasMessage += ", "
        }
        userGasMessage += fmt.Sprintf("Requested gas limit: %d", gasInfo.ReqGasLimit)
    }

    // Only print out maximum requested gas cost if either gas price or gas limit has been specified
    if len(userGasMessage) > 0 {
        if gasInfo.ReqGasLimit != 0 {
            gas = new(big.Int).SetUint64(gasInfo.ReqGasLimit)
        }
        if gasInfo.ReqGasPrice != nil {
            gasPrice = gasInfo.ReqGasPrice
        }
        totalGasWei = new(big.Int).Mul(gasPrice, gas)
        userGasMessage += fmt.Sprintf(", Maximum requested gas cost: %.6f ETH\n", math.RoundDown(eth.WeiToEth(totalGasWei), 6))
    }
    fmt.Println(userGasMessage)
}

