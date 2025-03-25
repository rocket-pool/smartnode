package queue

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func processQueue(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check deposit queue can be processed
	canProcess, err := rp.CanProcessQueue()
	if err != nil {
		return err
	}
	if !canProcess.CanProcess {
		fmt.Println("The deposit queue cannot be processed:")
		if canProcess.AssignDepositsDisabled {
			fmt.Println("Deposit assignments are currently disabled.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canProcess.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Do you accept this gas fee?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Process deposit queue
	response, err := rp.ProcessQueue()
	if err != nil {
		return err
	}

	fmt.Printf("Processing queue...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("The deposit queue was successfully processed.")
	return nil

}
