package node

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func confirmPrimaryWithdrawalAddress(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Check if the withdrawal address can be confirmed
	response, err := rp.Api.Node.ConfirmPrimaryWithdrawalAddress()
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanConfirm {
		fmt.Println("Cannot confirm withdrawal address as the node address:")
		if response.Data.IncorrectPendingAddress {
			fmt.Println("The node's pending withdrawal address must be set to the node address.")
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to confirm your node's address as the new primary withdrawal address?",
		"withdrawal address confirmation",
		"Confirming new primary withdrawal address...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("The node's primary withdrawal address was successfully set to the node address.\n")
	return nil
}
