package megapool

import (
	"fmt"
	"strconv"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
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

	var validatorId uint64

	// Check if the validator id flag is set
	if c.String("validator-id") != "" {
		// Get selected lot index
		validatorId, err = strconv.ParseUint(c.String("validator-id"), 10, 64)
		if err != nil {
			return fmt.Errorf("Invalid validator id '%s': %w", c.String("validator-id"), err)
		}
	} else {
		// Get Megapool status
		status, err := rp.MegapoolStatus(false)
		if err != nil {
			return err
		}

		validatorsInQueue := []api.MegapoolValidatorDetails{}

		for _, validator := range status.Megapool.Validators {
			if validator.InQueue {
				validatorsInQueue = append(validatorsInQueue, validator)
			}
		}
		if len(validatorsInQueue) > 0 {

			options := make([]string, len(validatorsInQueue))
			for vi, v := range validatorsInQueue {
				options[vi] = fmt.Sprintf("Pubkey: 0x%s", v.PubKey.String())
			}
			selected, _ := prompt.Select("Please select a validator to exit the queue:", options)

			// Get validators
			validatorId = uint64(validatorsInQueue[selected].ValidatorId)

		} else {
			fmt.Println("No validators can exit the queue at the moment")
			return nil
		}
	}

	// Check whether the validator can be exited
	canExit, err := rp.CanExitQueue(uint32(validatorId))
	if err != nil {
		return fmt.Errorf("Error checking if validator can be exited: %w", err)
	}

	if !canExit.CanExit {
		return fmt.Errorf("Validator %d cannot be exited from the megapool queue", validatorId)
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canExit.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Ask for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to request to exit validator index %d from the megapool queue?", validatorId))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Request exit from the megapool queue
	response, err := rp.ExitQueue(uint32(validatorId))
	if err != nil {
		return fmt.Errorf("Error requesting exit from the megapool queue: %w", err)
	}

	fmt.Printf("Requesting exit from the megapool queue...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	fmt.Printf("Successfully exited validator ID %d from the megapool queue.\nYou have received credit for the validator deposit and may withdraw it using the command `rocketpool node withdraw-credit`.", validatorId)
	return nil
}
