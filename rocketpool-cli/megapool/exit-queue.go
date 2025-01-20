package megapool

import (
	"fmt"
	"strconv"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

// Exit the megapool queue
func exitQueue(c *cli.Context) error {

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
	if !saturnResp.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	var validatorIndex uint64

	// Check if the validator id flag is set
	if c.String("validator-index") != "" {
		// Get selected lot index
		validatorIndex, err = strconv.ParseUint(c.String("validator-index"), 10, 64)
		if err != nil {
			return fmt.Errorf("Invalid validator index '%s': %w", c.String("validator-index"), err)
		}
	} else {
		// Ask for validator index
		validatorIndexString := cliutils.Prompt("Which validator index do you want to exit from the queue?", "^\\d+$", "Invalid validator index")
		validatorIndex, err = strconv.ParseUint(validatorIndexString, 0, 64)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid validator index: %w.\n", validatorIndexString, err)
		}
	}

	// Check whether the validator can be exited
	canExit, err := rp.CanExitQueue(uint32(validatorIndex))
	if err != nil {
		return fmt.Errorf("Error checking if validator can be exited: %w", err)
	}

	if !canExit.CanExit {
		return fmt.Errorf("Validator %d cannot be exited from the megapool queue", validatorIndex)
	}

	// Ask for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to request to exit validator index %d from the megapool queue?", validatorIndex))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Request exit from the megapool queue
	response, err := rp.ExitQueue(uint32(validatorIndex))
	if err != nil {
		return fmt.Errorf("Error requesting exit from the megapool queue: %w", err)
	}

	fmt.Printf("Requesting exit from the megapool queue...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	fmt.Printf("Successfully requested exit from the megapool queue for validator ID %d.\n", validatorIndex)
	return nil
}
