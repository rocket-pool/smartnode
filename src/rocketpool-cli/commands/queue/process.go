package queue

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func processQueue(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.Queue.Process()
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanProcess {
		fmt.Println("The deposit queue cannot be processed:")
		if response.Data.AssignDepositsDisabled {
			fmt.Println("Deposit assignments are currently disabled.")
		}
		if response.Data.InsufficientDepositBalance {
			fmt.Println("The deposit pool doesn't have enough ETH to assign to minipools.")
		}
		if response.Data.NoMinipoolsAvailable {
			fmt.Println("There are no minipools in the queue.")
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Do you want to process the queue?",
		"processing queue",
		"Processing queue...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("The deposit queue was successfully processed.")
	return nil
}
