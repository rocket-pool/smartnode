package services

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/etherchain"
	"github.com/rocket-pool/smartnode/shared/services/etherscan"
	rpsvc "github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const colorReset string = "\033[0m"
const colorYellow string = "\033[33m"
const colorBlue string = "\033[36m"

func AssignMaxFeeAndLimit(gasInfo rocketpool.GasInfo, rp *rpsvc.Client, headless bool) (error) {

    cfg, err := rp.LoadGlobalConfig()
    if err != nil {
        return fmt.Errorf("Error getting Rocket Pool configuration: %w", err)
    }

    // Get the current settings from the CLI arguments
    maxFee, maxPriorityFee, gasLimit := rp.GetGasSettings()

    // Get the max fee - prioritize the CLI arguments, default to the config file setting
    if maxFee == nil || maxFee.Uint64() == 0 {
        maxFee, err = cfg.GetMaxFee()
        if err != nil {
            return err
        }
    }

    // Get the priority fee - prioritize the CLI arguments, default to the config file setting
    if maxPriorityFee == nil || maxPriorityFee.Uint64() == 0 {
        maxPriorityFee, err = cfg.GetMaxPriorityFee()
        if err != nil {
            fmt.Printf("%sWARNING: Couldn't get max priority fee - %w\n", colorYellow, err.Error())
            fmt.Printf("Defaulting to a max priority fee of 2 gwei\n%s", colorReset)
            maxPriorityFee = eth.GweiToWei(2)
        }
        if maxPriorityFee == nil || maxPriorityFee.Uint64() == 0 {
            fmt.Printf("%sNOTE: max priority fee not set or set to 0, defaulting to 2 gwei%s\n", colorYellow, colorReset)
            maxPriorityFee = big.NewInt(2)
        }
    }

    // Get the user gas limit
    if gasLimit == 0 {
        gasLimit, err = cfg.GetGasLimit()
        if err != nil {
            return err
        }
    }

    // Use the requested max fee and priority fee if provided
    if maxFee != nil && maxFee.Uint64() != 0 {
        fmt.Printf("Using the requested max fee of %.2f gwei (including a max priority fee of %.2f gwei).\n", 
            eth.WeiToGwei(maxFee),
            eth.WeiToGwei(maxPriorityFee))
    }

    // Use the requested gas limit if provided
    if gasLimit != 0 {
        fmt.Printf("Using the requested gas limit of %d units.\n%sNOTE: if you set this too low, your transaction may fail but you will still have to pay the gas fee!%s\n", gasLimit, colorYellow, colorReset)
    }

    if headless {
        maxFee, err = GetHeadlessMaxFee()
        if err != nil {
            return err
        }
    } else {
        // Try to get the latest gas prices from Etherchain
        etherchainData, err := etherchain.GetGasPrices()
        if err == nil {
            // Print the Etherchain data and ask for an amount
            maxFee = handleEtherchainGasPrices(etherchainData, gasInfo, maxPriorityFee, gasLimit)
            
        } else {
            // Fallback to Etherscan
            fmt.Printf("%sWarning: couldn't get gas estimates from Etherchain - %s\nFalling back to Etherscan%s\n", colorYellow, err.Error(), colorReset)
            etherscanData, err := etherscan.GetGasPrices()
            if err == nil {
                // Print the Etherscan data and ask for an amount
                maxFee = handleEtherscanGasPrices(etherscanData, gasInfo, maxPriorityFee, gasLimit)
            } else {
                return fmt.Errorf("Error getting gas price suggestions: %w", err)
            }
        }
    }

    rp.AssignGasSettings(maxFee, maxPriorityFee, gasLimit)
    return nil

}


// Get the suggested max fee for service operations
func GetHeadlessMaxFee() (*big.Int, error) {
    etherchainData, err := etherchain.GetGasPrices()
    if err == nil {
        return etherchainData.FastWei, nil
    } else {
        fmt.Printf("%sWarning: couldn't get gas estimates from Etherchain - %s\nFalling back to Etherscan%s\n", colorYellow, err.Error(), colorReset)
        etherscanData, err := etherscan.GetGasPrices()
        if err == nil {
            return etherscanData.FastWei, nil
        } else {
            return nil, fmt.Errorf("Error getting gas price suggestions: %w", err)
        }
    }
}


func handleEtherchainGasPrices(gasSuggestion etherchain.GasFeeSuggestion, gasInfo rocketpool.GasInfo, priorityFee *big.Int, gasLimit uint64) (*big.Int) {

    
    rapidGwei := math.RoundUp(eth.WeiToGwei(big.NewInt(0).Add(gasSuggestion.RapidWei, priorityFee)), 0)
    rapidEth := eth.WeiToEth(gasSuggestion.RapidWei)

    var rapidLowLimit float64
    var rapidHighLimit float64
    if gasLimit == 0 {
        rapidLowLimit = rapidEth * float64(gasInfo.EstGasLimit)
        rapidHighLimit = rapidEth * float64(gasInfo.SafeGasLimit) 
    } else {
        rapidLowLimit = rapidEth * float64(gasLimit)
        rapidHighLimit = rapidLowLimit
    }

    fastGwei := math.RoundUp(eth.WeiToGwei(big.NewInt(0).Add(gasSuggestion.FastWei, priorityFee)), 0)
    fastEth := eth.WeiToEth(gasSuggestion.FastWei)

    var fastLowLimit float64
    var fastHighLimit float64
    if gasLimit == 0 {
        fastLowLimit = fastEth * float64(gasInfo.EstGasLimit)
        fastHighLimit = fastEth * float64(gasInfo.SafeGasLimit)
    } else {
        fastLowLimit = fastEth * float64(gasLimit)
        fastHighLimit = fastLowLimit
    }

    standardGwei := math.RoundUp(eth.WeiToGwei(big.NewInt(0).Add(gasSuggestion.StandardWei, priorityFee)), 0)
    standardEth := eth.WeiToEth(gasSuggestion.StandardWei)

    var standardLowLimit float64
    var standardHighLimit float64
    if gasLimit == 0 {
        standardLowLimit = standardEth * float64(gasInfo.EstGasLimit)
        standardHighLimit = standardEth * float64(gasInfo.SafeGasLimit)
    } else {
        standardLowLimit = standardEth * float64(gasLimit)
        standardHighLimit = standardLowLimit
    }

    slowGwei := math.RoundUp(eth.WeiToGwei(big.NewInt(0).Add(gasSuggestion.SlowWei, priorityFee)), 0)
    slowEth := eth.WeiToEth(gasSuggestion.SlowWei)

    var slowLowLimit float64
    var slowHighLimit float64
    if gasLimit == 0 {
        slowLowLimit = slowEth * float64(gasInfo.EstGasLimit)
        slowHighLimit = slowEth * float64(gasInfo.SafeGasLimit)
    } else {
        slowLowLimit = slowEth * float64(gasLimit)
        slowHighLimit = slowLowLimit
    }

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

    fmt.Printf("These prices include a maximum priority fee of %.2f gwei.\n", eth.WeiToGwei(priorityFee))

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


func handleEtherscanGasPrices(gasSuggestion etherscan.GasFeeSuggestion, gasInfo rocketpool.GasInfo, priorityFee *big.Int, gasLimit uint64) (*big.Int) {

    
    fastGwei := math.RoundUp(eth.WeiToGwei(big.NewInt(0).Add(gasSuggestion.FastWei, priorityFee)), 0)
    fastEth := eth.WeiToEth(gasSuggestion.FastWei)

    var fastLowLimit float64
    var fastHighLimit float64
    if gasLimit == 0 {
        fastLowLimit = fastEth * float64(gasInfo.EstGasLimit)
        fastHighLimit = fastEth * float64(gasInfo.SafeGasLimit)
    } else {
        fastLowLimit = fastEth * float64(gasLimit)
        fastHighLimit = fastLowLimit
    }

    standardGwei := math.RoundUp(eth.WeiToGwei(big.NewInt(0).Add(gasSuggestion.StandardWei, priorityFee)), 0)
    standardEth := eth.WeiToEth(gasSuggestion.StandardWei)

    var standardLowLimit float64
    var standardHighLimit float64
    if gasLimit == 0 {
        standardLowLimit = standardEth * float64(gasInfo.EstGasLimit)
        standardHighLimit = standardEth * float64(gasInfo.SafeGasLimit)
    } else {
        standardLowLimit = standardEth * float64(gasLimit)
        standardHighLimit = standardLowLimit
    }

    slowGwei := math.RoundUp(eth.WeiToGwei(big.NewInt(0).Add(gasSuggestion.SlowWei, priorityFee)), 0)
    slowEth := eth.WeiToEth(gasSuggestion.SlowWei)

    var slowLowLimit float64
    var slowHighLimit float64
    if gasLimit == 0 {
        slowLowLimit = slowEth * float64(gasInfo.EstGasLimit)
        slowHighLimit = slowEth * float64(gasInfo.SafeGasLimit)
    } else {
        slowLowLimit = slowEth * float64(gasLimit)
        slowHighLimit = slowLowLimit
    }

    fmt.Printf("%s+====== Suggested Gas Prices ======+\n", colorBlue)
    fmt.Println("|  Max Fee  |    Total Gas Cost    |")
    fmt.Printf("| %-9s | %.4f to %.4f ETH |\n", 
        fmt.Sprintf("%d gwei", int(fastGwei)), fastLowLimit, fastHighLimit)
    fmt.Printf("| %-9s | %.4f to %.4f ETH |\n", 
        fmt.Sprintf("%d gwei", int(standardGwei)), standardLowLimit, standardHighLimit)
    fmt.Printf("| %-9s | %.4f to %.4f ETH |\n", 
        fmt.Sprintf("%d gwei", int(slowGwei)), slowLowLimit, slowHighLimit)
    fmt.Printf("+==================================+\n\n%s", colorReset)

    fmt.Printf("These prices include a maximum priority fee of %.2f gwei.\n", eth.WeiToGwei(priorityFee))

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


