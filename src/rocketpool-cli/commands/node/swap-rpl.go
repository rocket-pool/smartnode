package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
)

func nodeSwapRpl(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get swap amount
	var amountWei *big.Int
	if c.String(amountFlag) == "all" {
		// Set amount to node's entire fixed-supply RPL balance
		status, err := rp.Api.Node.Status()
		if err != nil {
			return err
		}
		amountWei = status.Data.NodeBalances.Fsrpl
	} else if c.String(amountFlag) != "" {
		// Parse amount
		swapAmount, err := strconv.ParseFloat(c.String("amount"), 64)
		if err != nil {
			return fmt.Errorf("invalid swap amount '%s': %w", c.String("amount"), err)
		}
		amountWei = eth.EthToWei(swapAmount)
	} else {
		// Get entire fixed-supply RPL balance amount
		status, err := rp.Api.Node.Status()
		if err != nil {
			return err
		}
		entireAmount := status.Data.NodeBalances.Fsrpl

		// Prompt for entire amount
		if utils.Confirm(fmt.Sprintf("Would you like to swap your entire old RPL balance (%.6f RPL)?", math.RoundDown(eth.WeiToEth(entireAmount), 6))) {
			amountWei = entireAmount
		} else {
			// Prompt for custom amount
			inputAmount := utils.Prompt("Please enter an amount of old RPL to swap:", "^\\d+(\\.\\d+)?$", "Invalid amount")
			swapAmount, err := strconv.ParseFloat(inputAmount, 64)
			if err != nil {
				return fmt.Errorf("invalid swap amount '%s': %w", inputAmount, err)
			}
			amountWei = eth.EthToWei(swapAmount)
		}
	}

	return SwapRpl(c, rp, amountWei)
}
