package node

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func nodeBurn(c *cli.Context, amount float64, token string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get amount in wei
	amountWei := eth.EthToWei(amount)

	// Check tokens can be burned
	canBurn, err := rp.CanNodeBurn(amountWei, token)
	if err != nil {
		return err
	}
	if !canBurn.CanBurn {
		fmt.Println("Cannot burn tokens:")
		if canBurn.InsufficientBalance {
			fmt.Printf("The node's %s balance is insufficient.\n", token)
		}
		if canBurn.InsufficientCollateral {
			fmt.Printf("There is insufficient ETH collateral to trade %s for.\n", token)
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canBurn.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to burn %.6f %s for ETH?", math.RoundDown(eth.WeiToEth(amountWei), 6), token))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Burn tokens
	response, err := rp.NodeBurn(amountWei, token)
	if err != nil {
		return err
	}

	fmt.Printf("Burning tokens...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully burned %.6f %s for ETH.\n", math.RoundDown(eth.WeiToEth(amountWei), 6), token)
	return nil

}
