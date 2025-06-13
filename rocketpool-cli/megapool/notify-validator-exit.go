package megapool

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func notifyValidatorExit(c *cli.Context) error {

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

	// List the validators that can be exited
	var validatorId uint64

	if c.IsSet("validator-id") {
		validatorId = c.Uint64("validator-id")
	} else {
		// Get Megapool status
		status, err := rp.MegapoolStatus()
		if err != nil {
			return err
		}

		activeValidators := []api.MegapoolValidatorDetails{}

		for _, validator := range status.Megapool.Validators {
			if validator.Activated && !validator.Exiting && !validator.Exited {
				activeValidators = append(activeValidators, validator)
			}
		}
		if len(activeValidators) > 0 {

			options := make([]string, len(activeValidators))
			for vi, v := range activeValidators {
				options[vi] = fmt.Sprintf("ID: %d - Pubkey: 0x%s", v.ValidatorId, v.PubKey.String())
			}
			selected, _ := prompt.Select("Please select a validator to notify the exit:", options)

			// Get validators
			validatorId = uint64(activeValidators[selected].ValidatorId)

		} else {
			fmt.Println("No validators can be exited at the moment")
			return nil
		}
	}

	response, err := rp.CanNotifyValidatorExit(validatorId)
	if err != nil {
		return err
	}

	if !response.CanExit {
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(response.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to notify about the validator id %d exit?", validatorId))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Exit the validator
	resp, err := rp.NotifyValidatorExit(validatorId)
	if err != nil {
		return err
	}

	fmt.Printf("Notifying validator exit...\n")
	cliutils.PrintTransactionHash(rp, resp.TxHash)
	if _, err = rp.WaitForTransaction(resp.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully notified the exit of validator id %d.\n", validatorId)
	return nil

}
