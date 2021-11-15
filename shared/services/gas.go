package services

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/etherchain"
	rpsvc "github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const colorReset string = "\033[0m"
const colorYellow string = "\033[33m"
const colorBlue string = "\033[36m"

func AssignMaxFee(gasInfo rocketpool.GasInfo, rp *rpsvc.Client) (error) {

    cfg, err := rp.LoadGlobalConfig()
    if err != nil {
        return fmt.Errorf("Error getting Rocket Pool configuration: %w", err)
    }

    // Get the max fee
    maxFee, err := cfg.GetMaxFee()
    if err != nil {
        return err
    }

    // Get the priority fee
    priorityFee, err := cfg.GetMaxPriorityFee()
    if err != nil {
        fmt.Printf("%sWARNING: Couldn't get max priority fee - %w\n", colorYellow, err.Error())
        fmt.Printf("Defaulting to a max priority fee of 2 gwei\n%s", colorReset)
        priorityFee = eth.GweiToWei(2)
    }
    if priorityFee == nil {
        priorityFee = big.NewInt(0)
    }

    // Use the requested max fee and priority fee if provided
    if maxFee != nil {
        fmt.Printf("Using the requested max fee of %.2f gwei (including a max priority fee of %.2f gwei).\n", 
            eth.WeiToGwei(maxFee),
            eth.WeiToGwei(priorityFee))
    }

    // Try to get the latest gas prices from Etherchain
    gasPrices, err := etherchain.GetGasPrices()
    if err == nil {
        // Print the Etherscan data and ask for an amount
        maxFee = handleEtherchainGasPrices(gasPrices, gasInfo, priorityFee)
        
    } else {
        // Fall back to Geth
        fmt.Printf("%sWarning: couldn't get gas estimates from Etherchain - %s\n", colorYellow, err.Error())
        fmt.Printf("Falling back to the old suggestion system\n\n%s", colorReset)

        maxFee = printGasInfo(gasInfo, priorityFee)
    }

    rp.AssignMaxFees(maxFee.String(), priorityFee.String())
    return nil

}


func handleEtherchainGasPrices(gasSuggestion etherchain.GasFeeSuggestion, gasInfo rocketpool.GasInfo, priorityFee *big.Int) (*big.Int) {

    
    rapidGwei := math.RoundUp(eth.WeiToGwei(big.NewInt(0).Add(gasSuggestion.RapidWei, priorityFee)), 0)
    rapidEth := eth.WeiToEth(gasSuggestion.RapidWei)
    rapidLowLimit := rapidEth * float64(gasInfo.EstGasLimit)
    rapidHighLimit := rapidEth * float64(gasInfo.SafeGasLimit)

    fastGwei := math.RoundUp(eth.WeiToGwei(big.NewInt(0).Add(gasSuggestion.FastWei, priorityFee)), 0)
    fastEth := eth.WeiToEth(gasSuggestion.FastWei)
    fastLowLimit := fastEth * float64(gasInfo.EstGasLimit)
    fastHighLimit := fastEth * float64(gasInfo.SafeGasLimit)

    standardGwei := math.RoundUp(eth.WeiToGwei(big.NewInt(0).Add(gasSuggestion.StandardWei, priorityFee)), 0)
    standardEth := eth.WeiToEth(gasSuggestion.StandardWei)
    standardLowLimit := standardEth * float64(gasInfo.EstGasLimit)
    standardHighLimit := standardEth * float64(gasInfo.SafeGasLimit)

    slowGwei := math.RoundUp(eth.WeiToGwei(big.NewInt(0).Add(gasSuggestion.SlowWei, priorityFee)), 0)
    slowEth := eth.WeiToEth(gasSuggestion.SlowWei)
    slowLowLimit := slowEth * float64(gasInfo.EstGasLimit)
    slowHighLimit := slowEth * float64(gasInfo.SafeGasLimit)

    fmt.Printf("%s+============== Suggested Gas Prices ==============+\n", colorBlue)
    fmt.Println("| Avg Wait Time |  Max Fee  |    Total Gas Cost    |")
    fmt.Printf("| %-13s | %-9s | %.4f to %.4f ETH |\n", 
        gasSuggestion.RapidTime, fmt.Sprintf("%d gwei", int(rapidGwei)), rapidLowLimit, rapidHighLimit)
    fmt.Printf("| %-13s | %-9s | %.4f to %.4f ETH |\n", 
        gasSuggestion.FastTime, fmt.Sprintf("%d gwei", int(fastGwei)), fastLowLimit, fastHighLimit)
    fmt.Printf("| %-13s | %-9s | %.4f to %.4f ETH |\n", 
        gasSuggestion.StandardTime, fmt.Sprintf("%d gwei", int(standardGwei)), standardLowLimit, standardHighLimit)
    fmt.Printf("| %-13s | %-9s | %.4f to %.4f ETH |\n", 
        gasSuggestion.SlowTime, fmt.Sprintf("%d gwei", int(slowGwei)), slowLowLimit, slowHighLimit)
    fmt.Printf("+==================================================+\n\n%s", colorReset)

    fmt.Printf("These prices include a maximum priority fee of %.2f gwei.\n", priorityFee)

    for {
        desiredPrice := cliutils.Prompt(
            fmt.Sprintf("Please enter your max fee (including the priority fee) or leave blank for the default of %d gwei:", int(fastGwei)),
            "^(?:[1-9]\\d*|0)?(?:\\.\\d+)?$",
            "Not a valid gas price, try again:")

        if desiredPrice == "" {
            return eth.GweiToWei(fastGwei)
        }

        desiredPriceFloat, err := strconv.ParseFloat("desiredPrice", 64)
        if err != nil {
            fmt.Println("Not a valid gas price, try again.")
        }
        return eth.GweiToWei(desiredPriceFloat)
    }

}


// Print estimated gas cost and any requested gas parameters
func printGasInfo(gasInfo rocketpool.GasInfo, priorityFee *big.Int) (*big.Int) {

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
    fmt.Printf("%sSuggested max fee (including %.2f gwei priority fee): %.2f Gwei\nEstimated gas used: %d to %d gas\nEstimated gas cost: %.4f to %.4f ETH\n%s",
               colorYellow,
               priorityFee,
               eth.WeiToGwei(gasPrice), 
               gasInfo.EstGasLimit, 
               gasInfo.SafeGasLimit,
               math.RoundDown(eth.WeiToEth(totalGasWei), 6),
               math.RoundDown(eth.WeiToEth(totalSafeGasWei), 6),
               colorReset)
    
    // Print gas price, gas limit and max gas cost as requested by the user
    var userGasMessage string
    if gasInfo.ReqGasPrice != nil {
        userGasMessage += fmt.Sprintf("\n%sRequested gas price (including %.2f gwei priority fee): %.2f Gwei\n%s", 
                                      colorYellow,
                                      priorityFee,
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

    return gasPrice
}