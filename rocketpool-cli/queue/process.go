package queue

import (
	"fmt"
	"strconv"

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

	// Check if Saturn is already deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}

	var maxValidators uint64

	if saturnResp.IsSaturnDeployed {
		queueLength, err := rp.GetQueueDetails()
		if err != nil {
			return err
		}
		if queueLength.TotalLength == 0 {
			fmt.Println("There are no validators waiting to be processed")
			return nil
		}
		maxValidatorsStr := cliutils.Prompt(fmt.Sprintf("There is a total of %d validators in the queue. How many do you want to process?", queueLength.TotalLength), "^\\d+$", "Invalid number.")
		maxValidators, err = strconv.ParseUint(maxValidatorsStr, 0, 64)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid number: %w.\n", maxValidatorsStr, err)
		}
	}

	// Check deposit queue can be processed
	canProcess, err := rp.CanProcessQueue(uint32(maxValidators))
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
	response, err := rp.ProcessQueue(uint32(maxValidators))
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
