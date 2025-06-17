package megapool

import (
	"fmt"
	"sort"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func notifyFinalBalance(c *cli.Context) error {

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

	if c.IsSet("validator-id") {
		validatorId = c.Uint64("validator-id")
	} else {
		// Get Megapool status
		status, err := rp.MegapoolStatus()
		if err != nil {
			return err
		}

		exitingValidators := []api.MegapoolValidatorDetails{}

		for _, validator := range status.Megapool.Validators {
			if validator.Exiting {
				exitingValidators = append(exitingValidators, validator)
			}
		}
		if len(exitingValidators) > 0 {
			sort.Sort(ByIndex(exitingValidators))
			options := make([]string, len(exitingValidators))
			for vi, v := range exitingValidators {
				options[vi] = fmt.Sprintf("ID: %d - Index: %d - Pubkey: 0x%s", v.ValidatorId, v.ValidatorIndex, v.PubKey.String())
			}
			selected, _ := prompt.Select("Please select a validator to notify the final balance:", options)

			// Get validators
			validatorId = uint64(exitingValidators[selected].ValidatorId)

		} else {
			fmt.Println("No validators exiting at the moment")
			return nil
		}
	}
	slot := uint64(0)
	fmt.Println("The Smart Node needs to find the slot containing the validator withdrawal. This may take a while. If you know the slot, you can specify it using --slot or wait for the Smart Node to find it. ")
	if c.IsSet("slot") {
		fmt.Println("Using slot:", c.Uint64("slot"))
	}

	response, err := rp.CanNotifyFinalBalance(validatorId, slot)
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
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to notify de final balance for validator id %d exit?", validatorId))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Exit the validator
	resp, err := rp.NotifyFinalBalance(validatorId, slot)
	if err != nil {
		return err
	}

	fmt.Printf("Notifying validator final balance...\n")
	cliutils.PrintTransactionHash(rp, resp.TxHash)
	if _, err = rp.WaitForTransaction(resp.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully notified final balance for validator id %d.\n", validatorId)
	return nil

}
