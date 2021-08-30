package rocketpool

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const colorReset string = "\033[0m"
const colorYellow string = "\033[33m"


// Print a warning about the gas estimate for operations that have multiple transactions
func (rp *Client) PrintMultiTxWarning() {

    fmt.Printf("%sNOTE: This operation requires multiple transactions.\n%s",
        colorYellow,
        colorReset);

}


// Print estimated gas cost and any requested gas parameters
func (rp *Client) PrintGasInfo(gasInfo rocketpool.GasInfo) {

    // Print gas price, gas limit and total eth cost as estimated by the network
    gas := new(big.Int).SetUint64(gasInfo.EstGasLimit)
    safeGas := new(big.Int).SetUint64(gasInfo.SafeGasLimit)
    var gasPrice *big.Int
    if gasInfo.EstGasPrice != nil {
        gasPrice = gasInfo.EstGasPrice
    } else {
        gasPrice = big.NewInt(0)
    }
    totalGasWei := new(big.Int).Mul(gasPrice, gas)
    totalSafeGasWei := new(big.Int).Mul(gasPrice, safeGas)
    fmt.Printf("%sSuggested gas price: %.6f Gwei\nEstimated gas used: %d to %d gas\nEstimated gas cost: %.6f to %.6f ETH\n%s",
               colorYellow, 
               eth.WeiToGwei(gasPrice), 
               gasInfo.EstGasLimit, 
               gasInfo.SafeGasLimit,
               math.RoundDown(eth.WeiToEth(totalGasWei), 6),
               math.RoundDown(eth.WeiToEth(totalSafeGasWei), 6),
               colorReset)
    
    // Print gas price, gas limit and max gas cost as requested by the user
    var userGasMessage string
    if gasInfo.ReqGasPrice != nil {
        userGasMessage += fmt.Sprintf("\n%sRequested gas price: %.6f Gwei\n%s", 
                                      colorYellow,
                                      eth.WeiToGwei(gasInfo.ReqGasPrice),
                                      colorReset)
    }
    if gasInfo.ReqGasLimit != 0 {
        if len(userGasMessage) > 0 {
            userGasMessage += ", "
        }
        userGasMessage += fmt.Sprintf("%sRequested gas limit: %d\n%s",
                                      colorYellow,
                                      gasInfo.ReqGasLimit,
                                      colorReset)
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
        userGasMessage += fmt.Sprintf("%sMaximum requested gas cost: %.6f ETH\n%s",
                                      colorYellow,
                                      math.RoundDown(eth.WeiToEth(totalGasWei), 6),
                                      colorReset)
    }
    fmt.Println(userGasMessage)
}

