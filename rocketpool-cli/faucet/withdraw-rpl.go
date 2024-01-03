package faucet

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func withdrawRpl(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check RPL can be withdrawn
	canWithdraw, err := rp.CanFaucetWithdrawRpl()
	if err != nil {
		return err
	}
	if !canWithdraw.CanWithdraw {
		fmt.Println("Cannot withdraw legacy RPL from the faucet:")
		if canWithdraw.InsufficientFaucetBalance {
			fmt.Println("The faucet does not have any legacy RPL for withdrawal")
		}
		if canWithdraw.InsufficientAllowance {
			fmt.Println("You don't have any allowance remaining for the withdrawal period")
		}
		if canWithdraw.InsufficientNodeBalance {
			fmt.Println("You don't have enough testnet ETH to pay the faucet withdrawal fee")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canWithdraw.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to withdraw legacy RPL from the faucet?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Withdraw RPL
	response, err := rp.FaucetWithdrawRpl()
	if err != nil {
		return err
	}

	fmt.Printf("Withdrawing legacy RPL...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully withdrew %.6f legacy RPL from the faucet.\n", math.RoundDown(eth.WeiToEth(response.Amount), 6))
	return nil

}
