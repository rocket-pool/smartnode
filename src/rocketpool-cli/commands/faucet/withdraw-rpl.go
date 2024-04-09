package faucet

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func withdrawRpl(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Check RPL can be withdrawn
	response, err := rp.Api.Faucet.WithdrawRpl()
	if err != nil {
		return err
	}
	if !response.Data.CanWithdraw {
		fmt.Println("Cannot withdraw legacy RPL from the faucet:")
		if response.Data.InsufficientFaucetBalance {
			fmt.Println("The faucet does not have any legacy RPL for withdrawal")
		}
		if response.Data.InsufficientAllowance {
			fmt.Println("You don't have any allowance remaining for the withdrawal period")
		}
		if response.Data.InsufficientNodeBalance {
			fmt.Println("You don't have enough testnet ETH to pay the faucet withdrawal fee")
		}
		return nil
	}

	// Run the TX
	amount := math.RoundDown(eth.WeiToEth(response.Data.Amount), 6)
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to withdraw %.6f legacy RPL from the faucet?", amount),
		"withdraw of legacy RPL",
		"Withdrawing legacy RPL...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully withdrew %.6f legacy RPL from the faucet.\n", amount)
	return nil
}
