package gas

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/gas/etherchain"
	"github.com/rocket-pool/smartnode/shared/services/gas/etherscan"
	rpsvc "github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

const colorReset string = "\033[0m"
const colorYellow string = "\033[33m"
const colorBlue string = "\033[36m"

// DefaultPriorityFeeGwei is the default priority fee in gwei used for automatic transactions
const DefaultPriorityFeeGwei float64 = 0.01

type Gas struct {
	maxFeeGwei         float64
	maxPriorityFeeGwei float64
	gasLimit           uint64
}

func AssignMaxFeeAndLimit(gasInfo rocketpool.GasInfo, rp *rpsvc.Client, headless bool) error {
	g, err := GetMaxFeeAndLimit(gasInfo, rp, headless)
	if err != nil {
		return err
	}
	g.Assign(rp)
	return nil
}

func (g *Gas) Assign(rp *rpsvc.Client) {
	rp.AssignGasSettings(g.maxFeeGwei, g.maxPriorityFeeGwei, g.gasLimit)
	return
}

func GetMaxFeeAndLimit(gasInfo rocketpool.GasInfo, rp *rpsvc.Client, headless bool) (Gas, error) {

	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return Gas{}, fmt.Errorf("Error getting Rocket Pool configuration: %w", err)
	}
	if isNew {
		return Gas{}, fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smartnode.")
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
			maxFeeWei, err := GetHeadlessMaxFeeWei(cfg)
			if err != nil {
				return Gas{}, err
			}
			maxFeeGwei = eth.WeiToGwei(maxFeeWei)
		} else {
			// Try to get the latest gas prices from Etherchain
			etherchainData, err := etherchain.GetGasPrices(cfg)
			if err == nil {
				// Print the Etherchain data and ask for an amount
				maxFeeGwei = handleEtherchainGasPrices(etherchainData, gasInfo, maxPriorityFeeGwei, gasLimit)

			} else {
				// Fallback to Etherscan
				// Etherscan does not have a hoodi endpoint. Fallback uses the mainnet url in all cases.
				fmt.Printf("%sWarning: couldn't get gas estimates from Etherchain - %s\nFalling back to Etherscan%s\n", colorYellow, err.Error(), colorReset)
				etherscanData, err := etherscan.GetGasPrices()
				if err == nil {
					// Print the Etherscan data and ask for an amount
					maxFeeGwei = handleEtherscanGasPrices(etherscanData, gasInfo, maxPriorityFeeGwei, gasLimit)
				} else {
					return Gas{}, fmt.Errorf("Error getting gas price suggestions: %w", err)
				}
			}
		}
		fmt.Printf("%sUsing a max fee of %.3f gwei and a priority fee of %.3f gwei.\n%s", colorBlue, maxFeeGwei, maxPriorityFeeGwei, colorReset)
	}

	// Use the requested gas limit if provided
	if gasLimit != 0 {
		fmt.Printf("Using the requested gas limit of %d units.\n%sNOTE: if you set this too low, your transaction may fail but you will still have to pay the gas fee!%s\n", gasLimit, colorYellow, colorReset)
	}

	if maxPriorityFeeGwei > maxFeeGwei {
		return Gas{}, fmt.Errorf("Priority fee cannot be greater than max fee.")
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
		return Gas{}, fmt.Errorf("Your node has %.6f ETH in its wallet, which is not enough to pay for this transaction with a max fee of %.4f gwei; you require at least %.6f more ETH.", eth.WeiToEth(response.Balance), maxFeeGwei, eth.WeiToEth(big.NewInt(0).Sub(ethRequired, response.Balance)))
	}
	return Gas{maxFeeGwei, maxPriorityFeeGwei, gasLimit}, nil

}

// Get the suggested max fee for service operations
func GetHeadlessMaxFeeWei(cfg *config.RocketPoolConfig) (*big.Int, error) {
	etherchainData, err := etherchain.GetGasPrices(cfg)
	if err == nil {
		return etherchainData.RapidWei, nil
	} else {
		// Etherscan does not have a hoodi endpoint. Fallback uses the mainnet url in all cases.
		fmt.Printf("%sWarning: couldn't get gas estimates from Etherchain - %s\nFalling back to Etherscan%s\n", colorYellow, err.Error(), colorReset)
		etherscanData, err := etherscan.GetGasPrices()
		if err == nil {
			return eth.GweiToWei(etherscanData.FastGwei), nil
		} else {
			return nil, fmt.Errorf("Error getting gas price suggestions: %w", err)
		}
	}
}

func handleEtherchainGasPrices(gasSuggestion etherchain.GasFeeSuggestion, gasInfo rocketpool.GasInfo, priorityFee float64, gasLimit uint64) float64 {

	rapidGwei := eth.WeiToGwei(gasSuggestion.RapidWei) + priorityFee
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

	fastGwei := eth.WeiToGwei(gasSuggestion.FastWei) + priorityFee
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

	standardGwei := eth.WeiToGwei(gasSuggestion.StandardWei) + priorityFee
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

	slowGwei := eth.WeiToGwei(gasSuggestion.SlowWei) + priorityFee
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

	fmt.Printf("%s+================ Suggested Gas Prices ================+\n", colorBlue)
	fmt.Println("| Avg Wait Time |   Max Fee    |     Total Gas Cost     |")
	fmt.Printf("| %-13s | %-9s | %.5f to %.5f ETH |\n",
		gasSuggestion.RapidTime, fmt.Sprintf("%.5f gwei", rapidGwei), rapidLowLimit, rapidHighLimit)
	fmt.Printf("| %-13s | %-9s | %.5f to %.5f ETH |\n",
		gasSuggestion.FastTime, fmt.Sprintf("%.5f gwei", fastGwei), fastLowLimit, fastHighLimit)
	fmt.Printf("| %-13s | %-9s | %.5f to %.5f ETH |\n",
		gasSuggestion.StandardTime, fmt.Sprintf("%.5f gwei", standardGwei), standardLowLimit, standardHighLimit)
	fmt.Printf("| %-13s | %-9s | %.5f to %.5f ETH |\n",
		gasSuggestion.SlowTime, fmt.Sprintf("%.5f gwei", slowGwei), slowLowLimit, slowHighLimit)
	fmt.Printf("+======================================================+\n\n%s", colorReset)

	fmt.Printf("These prices include a maximum priority fee of %.3f gwei.\n", priorityFee)

	for {
		desiredPrice := prompt.Prompt(
			fmt.Sprintf("Please enter your max fee (including the priority fee) or leave blank for the default of %.5f gwei:", fastGwei),
			"^(?:[1-9]\\d*|0)?(?:\\.\\d+)?$",
			"Not a valid gas price, try again:")

		if desiredPrice == "" {
			return fastGwei
		}

		desiredPriceFloat, err := strconv.ParseFloat(desiredPrice, 64)
		if err != nil {
			fmt.Printf("Not a valid gas price (%s), try again.", err.Error())
			fmt.Println("")
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

	fastGwei := gasSuggestion.FastGwei + priorityFee
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

	standardGwei := gasSuggestion.StandardGwei + priorityFee
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

	slowGwei := gasSuggestion.SlowGwei + priorityFee
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

	fmt.Printf("%s+============== Suggested Gas Prices ===============+\n", colorBlue)
	fmt.Println("|   Speed   |    Max Fee   |      Total Gas Cost     |")
	fmt.Printf("| Fast      | %-9s | %.6f to %.6f ETH |\n",
		fmt.Sprintf("%.5f gwei", fastGwei), fastLowLimit, fastHighLimit)
	fmt.Printf("| Standard  | %-9s | %.6f to %.6f ETH |\n",
		fmt.Sprintf("%.5f gwei", standardGwei), standardLowLimit, standardHighLimit)
	fmt.Printf("| Slow      | %-9s | %.6f to %.6f ETH |\n",
		fmt.Sprintf("%.5f gwei", slowGwei), slowLowLimit, slowHighLimit)
	fmt.Printf("+====================================================+\n\n%s", colorReset)

	fmt.Printf("These prices include a maximum priority fee of %.3f gwei.\n", priorityFee)

	for {
		desiredPrice := prompt.Prompt(
			fmt.Sprintf("Please enter your max fee (including the priority fee) or leave blank for the default of %.5f gwei:", fastGwei),
			"^(?:[1-9]\\d*|0)?(?:\\.\\d+)?$",
			"Not a valid gas price, try again:")

		if desiredPrice == "" {
			return fastGwei
		}

		desiredPriceFloat, err := strconv.ParseFloat(desiredPrice, 64)
		if err != nil {
			fmt.Printf("Not a valid gas price (%s), try again.", err.Error())
			fmt.Println("")
			continue
		}
		if desiredPriceFloat <= 0 {
			fmt.Println("Max fee must be greater than zero.")
			continue
		}

		return desiredPriceFloat
	}

}
