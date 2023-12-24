package gas

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas/etherchain"
	"github.com/rocket-pool/smartnode/shared/services/gas/etherscan"
	rpsvc "github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const colorReset string = "\033[0m"
const colorYellow string = "\033[33m"
const colorBlue string = "\033[36m"

func AssignMaxFeeAndLimit(gasInfo rocketpool.GasInfo, rp *rpsvc.Client, headless bool) error {

	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error getting Rocket Pool configuration: %w", err)
	}
	if isNew {
		return fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smartnode.")
	}

	// Get the current settings from the CLI arguments
	maxFeeGwei, maxPriorityFeeGwei, gasLimit := rp.GetGasSettings()

	// Get the max fee - prioritize the CLI arguments, default to the config file setting
	if maxFeeGwei == 0 {
		maxFee := eth.GweiToWei(cfg.Smartnode.ManualMaxFee.Value.(float64))
		if maxFee != nil && maxFee.Uint64() != 0 {
			maxFeeGwei = eth.WeiToGwei(maxFee)
		}
	}

	// Get the priority fee - prioritize the CLI arguments, default to the config file setting
	if maxPriorityFeeGwei == 0 {
		maxPriorityFee := eth.GweiToWei(cfg.Smartnode.PriorityFee.Value.(float64))
		if maxPriorityFee == nil || maxPriorityFee.Uint64() == 0 {
			fmt.Printf("%sNOTE: max priority fee not set or set to 0, defaulting to 2 gwei%s\n", colorYellow, colorReset)
			maxPriorityFeeGwei = 2
		} else {
			maxPriorityFeeGwei = eth.WeiToGwei(maxPriorityFee)
		}
	}

	// Use the requested max fee and priority fee if provided
	if maxFeeGwei != 0 {
		fmt.Printf("%sUsing the requested max fee of %.2f gwei (including a max priority fee of %.2f gwei).\n", colorYellow, maxFeeGwei, maxPriorityFeeGwei)

		var lowLimit float64
		var highLimit float64
		if gasLimit == 0 {
			lowLimit = maxFeeGwei / eth.WeiPerGwei * float64(gasInfo.EstGasLimit)
			highLimit = maxFeeGwei / eth.WeiPerGwei * float64(gasInfo.SafeGasLimit)
		} else {
			lowLimit = maxFeeGwei / eth.WeiPerGwei * float64(gasLimit)
			highLimit = lowLimit
		}
		fmt.Printf("Total cost: %.4f to %.4f ETH%s\n", lowLimit, highLimit, colorReset)

	} else {
		if headless {
			maxFeeWei, err := GetHeadlessMaxFeeWei()
			if err != nil {
				return err
			}
			maxFeeGwei = eth.WeiToGwei(maxFeeWei)
		} else {
			// Try to get the latest gas prices from Etherchain
			etherchainData, err := etherchain.GetGasPrices()
			if err == nil {
				// Print the Etherchain data and ask for an amount
				maxFeeGwei = handleEtherchainGasPrices(etherchainData, gasInfo, maxPriorityFeeGwei, gasLimit)

			} else {
				// Fallback to Etherscan
				fmt.Printf("%sWarning: couldn't get gas estimates from Etherchain - %s\nFalling back to Etherscan%s\n", colorYellow, err.Error(), colorReset)
				etherscanData, err := etherscan.GetGasPrices()
				if err == nil {
					// Print the Etherscan data and ask for an amount
					maxFeeGwei = handleEtherscanGasPrices(etherscanData, gasInfo, maxPriorityFeeGwei, gasLimit)
				} else {
					return fmt.Errorf("Error getting gas price suggestions: %w", err)
				}
			}
		}
		fmt.Printf("%sUsing a max fee of %.2f gwei and a priority fee of %.2f gwei.\n%s", colorBlue, maxFeeGwei, maxPriorityFeeGwei, colorReset)
	}

	// Use the requested gas limit if provided
	if gasLimit != 0 {
		fmt.Printf("Using the requested gas limit of %d units.\n%sNOTE: if you set this too low, your transaction may fail but you will still have to pay the gas fee!%s\n", gasLimit, colorYellow, colorReset)
	}

	if maxPriorityFeeGwei > maxFeeGwei {
		return fmt.Errorf("Priority fee cannot be greater than max fee.")
	}

	// Verify the node has enough ETH to use this max fee
	maxFee := eth.GweiToWei(maxFeeGwei)
	ethRequired := big.NewInt(0)
	if gasLimit != 0 {
		ethRequired.Mul(maxFee, big.NewInt(int64(gasLimit)))
	} else {
		ethRequired.Mul(maxFee, big.NewInt(int64(gasInfo.SafeGasLimit)))
	}
	response, err := rp.GetEthBalance()
	if err != nil {
		fmt.Printf("%sWARNING: couldn't check the ETH balance of the node (%s)\nPlease ensure your node wallet has enough ETH to pay for this transaction.%s\n\n", colorYellow, err.Error(), colorReset)
	} else if response.Balance.Cmp(ethRequired) < 0 {
		return fmt.Errorf("Your node has %.6f ETH in its wallet, which is not enough to pay for this transaction with a max fee of %.4f gwei; you require at least %.6f more ETH.", eth.WeiToEth(response.Balance), maxFeeGwei, eth.WeiToEth(big.NewInt(0).Sub(ethRequired, response.Balance)))
	}

	rp.AssignGasSettings(maxFeeGwei, maxPriorityFeeGwei, gasLimit)
	return nil

}

// Get the suggested max fee for service operations
func GetHeadlessMaxFeeWei() (*big.Int, error) {
	etherchainData, err := etherchain.GetGasPrices()
	if err == nil {
		return etherchainData.RapidWei, nil
	}

	fmt.Printf("%sWarning: couldn't get gas estimates from Etherchain - %s\nFalling back to Etherscan%s\n", colorYellow, err.Error(), colorReset)
	etherscanData, err := etherscan.GetGasPrices()
	if err == nil {
		return eth.GweiToWei(etherscanData.FastGwei), nil
	}

	return nil, fmt.Errorf("Error getting gas price suggestions: %w", err)
}

func handleEtherchainGasPrices(gasSuggestion etherchain.GasFeeSuggestion, gasInfo rocketpool.GasInfo, priorityFee float64, gasLimit uint64) float64 {

	rapidGwei := math.RoundUp(eth.WeiToGwei(gasSuggestion.RapidWei)+priorityFee, 0)
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

	fastGwei := math.RoundUp(eth.WeiToGwei(gasSuggestion.FastWei)+priorityFee, 0)
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

	standardGwei := math.RoundUp(eth.WeiToGwei(gasSuggestion.StandardWei)+priorityFee, 0)
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

	slowGwei := math.RoundUp(eth.WeiToGwei(gasSuggestion.SlowWei)+priorityFee, 0)
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

	fmt.Printf("These prices include a maximum priority fee of %.2f gwei.\n", priorityFee)

	for {
		desiredPrice := cliutils.Prompt(
			fmt.Sprintf("Please enter your max fee (including the priority fee) or leave blank for the default of %d gwei:", int(fastGwei)),
			"^(?:[1-9]\\d*|0)?(?:\\.\\d+)?$",
			"Not a valid gas price, try again:")

		if desiredPrice == "" {
			return fastGwei
		}

		desiredPriceFloat, err := strconv.ParseFloat(desiredPrice, 64)
		if err != nil {
			fmt.Printf("Not a valid gas price (%sv, try again.\n", err)
			continue
		}
		if desiredPriceFloat <= 0 {
			fmt.Println("Max fee must be greater than zero.")
			continue
		}

		return desiredPriceFloat
	}

}

func handleEtherscanGasPrices(gasSuggestion etherscan.GasFeeSuggestion, gasInfo rocketpool.GasInfo, priorityFee float64, gasLimit uint64) float64 {

	fastGwei := math.RoundUp(gasSuggestion.FastGwei+priorityFee, 0)
	fastEth := gasSuggestion.FastGwei / eth.WeiPerGwei

	var fastLowLimit float64
	var fastHighLimit float64
	if gasLimit == 0 {
		fastLowLimit = fastEth * float64(gasInfo.EstGasLimit)
		fastHighLimit = fastEth * float64(gasInfo.SafeGasLimit)
	} else {
		fastLowLimit = fastEth * float64(gasLimit)
		fastHighLimit = fastLowLimit
	}

	standardGwei := math.RoundUp(gasSuggestion.StandardGwei+priorityFee, 0)
	standardEth := gasSuggestion.StandardGwei / eth.WeiPerGwei

	var standardLowLimit float64
	var standardHighLimit float64
	if gasLimit == 0 {
		standardLowLimit = standardEth * float64(gasInfo.EstGasLimit)
		standardHighLimit = standardEth * float64(gasInfo.SafeGasLimit)
	} else {
		standardLowLimit = standardEth * float64(gasLimit)
		standardHighLimit = standardLowLimit
	}

	slowGwei := math.RoundUp(gasSuggestion.SlowGwei+priorityFee, 0)
	slowEth := gasSuggestion.SlowGwei / eth.WeiPerGwei

	var slowLowLimit float64
	var slowHighLimit float64
	if gasLimit == 0 {
		slowLowLimit = slowEth * float64(gasInfo.EstGasLimit)
		slowHighLimit = slowEth * float64(gasInfo.SafeGasLimit)
	} else {
		slowLowLimit = slowEth * float64(gasLimit)
		slowHighLimit = slowLowLimit
	}

	fmt.Printf("%s+============ Suggested Gas Prices ============+\n", colorBlue)
	fmt.Println("|   Speed   |  Max Fee  |    Total Gas Cost    |")
	fmt.Printf("| Fast      | %-9s | %.4f to %.4f ETH |\n",
		fmt.Sprintf("%d gwei", int(fastGwei)), fastLowLimit, fastHighLimit)
	fmt.Printf("| Standard  | %-9s | %.4f to %.4f ETH |\n",
		fmt.Sprintf("%d gwei", int(standardGwei)), standardLowLimit, standardHighLimit)
	fmt.Printf("| Slow      | %-9s | %.4f to %.4f ETH |\n",
		fmt.Sprintf("%d gwei", int(slowGwei)), slowLowLimit, slowHighLimit)
	fmt.Printf("+==============================================+\n\n%s", colorReset)

	fmt.Printf("These prices include a maximum priority fee of %.2f gwei.\n", priorityFee)

	for {
		desiredPrice := cliutils.Prompt(
			fmt.Sprintf("Please enter your max fee (including the priority fee) or leave blank for the default of %d gwei:", int(fastGwei)),
			"^(?:[1-9]\\d*|0)?(?:\\.\\d+)?$",
			"Not a valid gas price, try again:")

		if desiredPrice == "" {
			return fastGwei
		}

		desiredPriceFloat, err := strconv.ParseFloat(desiredPrice, 64)
		if err != nil {
			fmt.Printf("Not a valid gas price (%v), try again.\n", err)
			continue
		}
		if desiredPriceFloat <= 0 {
			fmt.Println("Max fee must be greater than zero.")
			continue
		}

		return desiredPriceFloat
	}

}
