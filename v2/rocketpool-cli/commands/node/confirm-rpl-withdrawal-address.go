package node

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func confirmRplWithdrawalAddress(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Check if the withdrawal address can be confirmed
	response, err := rp.Api.Node.ConfirmRplWithdrawalAddress()
	if err != nil {
		return err
	}

	// Check if it can be set
	if !response.Data.CanConfirm {
		fmt.Println("Cannot confirm new RPL withdrawal address:")
		if response.Data.IncorrectPendingAddress {
			fmt.Println("Your node address is not the new pending RPL withdrawal address. Confirmation can only be done if it is set to your node address.")
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to confirm your node's address as the new RPL withdrawal address?",
		"confirming the RPL withdrawal address",
		"Confirming new RPL withdrawal address...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("The node's RPL withdrawal address was successfully set to the node address.\n")
	return nil
}
