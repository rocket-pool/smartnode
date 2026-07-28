package gas

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/gas/etherscan"
	rpsvc "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

// DefaultPriorityFeeGwei is the default priority fee in gwei used for automatic transactions
const DefaultPriorityFeeGwei float64 = 0.01

type Gas struct {
	maxFeeGwei         float64
	maxPriorityFeeGwei float64
	gasLimit           uint64
}

func AssignMaxFeeAndLimit(gasLimits gaslimit.Limits, rp *rpsvc.Client, headless bool) error {
	g, err := GetMaxFeeAndLimit(gasLimits, rp, headless)
	if err != nil {
		return err
	}
	g.Assign(rp)
	return nil
}

func (g *Gas) Assign(rp *rpsvc.Client) {
	rp.AssignGasSettings(g.maxFeeGwei, g.maxPriorityFeeGwei, g.gasLimit)
}

// GetMaxGasCostEth returns the maximum possible gas cost in ETH for the given gas info,
func (g *Gas) GetMaxGasCostEth(gasLimits gaslimit.Limits) float64 {
	limit := uint64(float64(gasLimits.Estimated) * 1.1)
	if g.gasLimit != 0 {
		limit = g.gasLimit
	}
	return g.maxFeeGwei / math.WeiPerGwei * float64(limit)
}

func GetMaxFeeAndLimit(gasLimits gaslimit.Limits, rp *rpsvc.Client, headless bool) (Gas, error) {

	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return Gas{}, fmt.Errorf("Error getting Rocket Pool configuration: %w", err)
	}
	if isNew {
		return Gas{}, fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smart Node.")
	}

	// Get the current settings from the CLI arguments
	maxFeeGwei, maxPriorityFeeGwei, gasLimit := rp.GetGasSettings()

	// Get the max fee - prioritize the CLI arguments, default to the config file setting
	if maxFeeGwei == 0 {
		maxFee := math.GweiToWei(cfg.Smartnode.ManualMaxFee.Value.(float64))
		if maxFee != nil && maxFee.Uint64() != 0 {
			maxFeeGwei = math.WeiToGwei(maxFee)
		}
	}

	// Get the priority fee - prioritize the CLI arguments, default to the config file setting
	if maxPriorityFeeGwei == 0 {
		maxPriorityFee := math.GweiToWei(cfg.Smartnode.PriorityFee.Value.(float64))
		if maxPriorityFee == nil || maxPriorityFee.Uint64() == 0 {
			color.YellowPrintln("NOTE: max priority fee not set or set to 0, defaulting to 2 gwei")
			maxPriorityFeeGwei = 2
		} else {
			maxPriorityFeeGwei = math.WeiToGwei(maxPriorityFee)
		}
	}

	// Use the requested max fee and priority fee if provided
	if maxFeeGwei != 0 {
		color.YellowPrintf("Using the requested max fee of %.2f gwei (including a max priority fee of %.2f gwei).\n", maxFeeGwei, maxPriorityFeeGwei)

		var lowLimit float64
		var highLimit float64
		if gasLimit == 0 {
			lowLimit = maxFeeGwei / math.WeiPerGwei * float64(gasLimits.Estimated)
			highLimit = maxFeeGwei / math.WeiPerGwei * float64(gasLimits.Safe)
		} else {
			lowLimit = maxFeeGwei / math.WeiPerGwei * float64(gasLimit)
			highLimit = lowLimit
		}
		color.YellowPrintf("Total cost: %.4f to %.4f ETH\n", lowLimit, highLimit)

	} else {
		if headless {
			maxFeeWei, err := GetHeadlessMaxFeeWei(cfg)
			if err != nil {
				return Gas{}, err
			}
			maxFeeGwei = math.WeiToGwei(maxFeeWei)
		} else {
			// Try to get the latest gas prices from Etherscan
			gasData, err := etherscan.GetGasPrices()
			if err != nil || cfg.Smartnode.GetChainID() != 1 {
				gasPrice, err := rp.GetGasPriceFromLatestBlock()
				if err != nil {
					return Gas{}, err
				}
				gasData = etherscan.GasFeeSuggestion{
					SlowGwei:     math.WeiToGwei(gasPrice.GasPrice),
					StandardGwei: math.WeiToGwei(gasPrice.GasPrice) * 1.1,
					FastGwei:     math.WeiToGwei(gasPrice.GasPrice) * 1.2,
				}
			}

			// Print the gas data and ask for an amount
			maxFeeGwei = handleEtherscanGasPrices(gasData, gasLimits, maxPriorityFeeGwei, gasLimit)
		}
		color.LightBluePrintf("Using a max fee of %.3f gwei and a priority fee of %.3f gwei.\n", maxFeeGwei, maxPriorityFeeGwei)
	}

	// Use the requested gas limit if provided
	if gasLimit != 0 {
		fmt.Printf("Using the requested gas limit of %d units.\n", gasLimit)
		color.YellowPrintln("NOTE: if you set this too low, your transaction may fail but you will still have to pay the gas fee!")
	}

	if maxPriorityFeeGwei > maxFeeGwei {
		return Gas{}, fmt.Errorf("Priority fee cannot be greater than max fee.")
	}
	// Verify the node has enough ETH to use this max fee
	maxFee := math.GweiToWei(maxFeeGwei)
	ethRequired := big.NewInt(0)
	if gasLimit != 0 {
		ethRequired.Mul(maxFee, big.NewInt(int64(gasLimit)))
	} else {
		ethRequired.Mul(maxFee, big.NewInt(int64(gasLimits.Safe)))
	}
	response, err := rp.GetEthBalance()
	if err != nil {
		color.YellowPrintf("WARNING: couldn't check the ETH balance of the node (%s)\n", err.Error())
		color.YellowPrintln("Please ensure your node wallet has enough ETH to pay for this transaction.")
		fmt.Println()
	} else if response.Balance.Cmp(ethRequired) < 0 {
		return Gas{}, fmt.Errorf("Your node has %.6f ETH in its wallet, which is not enough to pay for this transaction with a max fee of %.4f gwei; you require at least %.6f more ETH.", math.WeiToEth(response.Balance), maxFeeGwei, math.WeiToEth(big.NewInt(0).Sub(ethRequired, response.Balance)))
	}
	return Gas{maxFeeGwei, maxPriorityFeeGwei, gasLimit}, nil

}

// Get the suggested max fee for service operations
func GetHeadlessMaxFeeWei(cfg *config.RocketPoolConfig) (*big.Int, error) {
	return GetHeadlessMaxFeeWeiWithLatestBlock(cfg, nil)
}

// Get the suggested max fee for service operations using the latest block
func GetHeadlessMaxFeeWeiWithLatestBlock(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool) (*big.Int, error) {
	if rp != nil {
		// Getting the latest block to estimate the gas price
		latestBlock, err := rp.Client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			color.YellowPrintf("Warning: couldn't get gas estimates from the latest block: %s\n", err.Error())
			color.YellowPrintln("Using gas oracles")
		}
		// Get the latest block gas + 20%
		gasPrice := big.NewInt(0).Add(latestBlock.BaseFee, big.NewInt(0).Div(big.NewInt(0).Mul(latestBlock.BaseFee, big.NewInt(20)), big.NewInt(100)))
		return gasPrice, nil
	}

	etherscanData, err := etherscan.GetGasPrices()
	if err == nil {
		return math.GweiToWei(etherscanData.FastGwei), nil
	}
	return nil, fmt.Errorf("error getting gas estimates. You can try again later or specify fees manually using --maxFee and --maxPrioFee.")
}

func handleEtherscanGasPrices(gasSuggestion etherscan.GasFeeSuggestion, gasLimits gaslimit.Limits, priorityFee float64, gasLimit uint64) float64 {

	fastGwei := gasSuggestion.FastGwei + priorityFee
	fastEth := fastGwei / math.WeiPerGwei

	var fastLowLimit float64
	var fastHighLimit float64
	if gasLimit == 0 {
		fastLowLimit = fastEth * float64(gasLimits.Estimated)
		fastHighLimit = fastEth * float64(gasLimits.Safe)
	} else {
		fastLowLimit = fastEth * float64(gasLimit)
		fastHighLimit = fastLowLimit
	}

	standardGwei := gasSuggestion.StandardGwei + priorityFee
	standardEth := standardGwei / math.WeiPerGwei

	var standardLowLimit float64
	var standardHighLimit float64
	if gasLimit == 0 {
		standardLowLimit = standardEth * float64(gasLimits.Estimated)
		standardHighLimit = standardEth * float64(gasLimits.Safe)
	} else {
		standardLowLimit = standardEth * float64(gasLimit)
		standardHighLimit = standardLowLimit
	}

	slowGwei := gasSuggestion.SlowGwei + priorityFee
	slowEth := slowGwei / math.WeiPerGwei

	var slowLowLimit float64
	var slowHighLimit float64
	if gasLimit == 0 {
		slowLowLimit = slowEth * float64(gasLimits.Estimated)
		slowHighLimit = slowEth * float64(gasLimits.Safe)
	} else {
		slowLowLimit = slowEth * float64(gasLimit)
		slowHighLimit = slowLowLimit
	}

	color.LightBluePrintln("+============== Suggested Gas Prices ===============+")
	color.LightBluePrintln("|   Speed   |    Max Fee   |      Total Gas Cost     |")
	color.LightBluePrintf("| Fast      | %-9s | %.6f to %.6f ETH |\n",
		fmt.Sprintf("%.5f gwei", fastGwei), fastLowLimit, fastHighLimit)
	color.LightBluePrintf("| Standard  | %-9s | %.6f to %.6f ETH |\n",
		fmt.Sprintf("%.5f gwei", standardGwei), standardLowLimit, standardHighLimit)
	color.LightBluePrintf("| Slow      | %-9s | %.6f to %.6f ETH |\n",
		fmt.Sprintf("%.5f gwei", slowGwei), slowLowLimit, slowHighLimit)
	color.LightBluePrintln("+====================================================+")
	fmt.Println()

	fmt.Printf("These prices include a maximum priority fee of %.4f gwei.\n", priorityFee)

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
