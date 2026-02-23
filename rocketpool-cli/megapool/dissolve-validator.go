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

func dissolveValidator(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	var validatorId uint64

	if c.IsSet("validator-id") {
		validatorId = c.Uint64("validator-id")
	} else {
		// Get Megapool status
		status, err := rp.MegapoolStatus(false)
		if err != nil {
			return err
		}

		validatorsInPrestake := []api.MegapoolValidatorDetails{}

		for _, validator := range status.Megapool.Validators {
			if validator.InPrestake {
				validatorsInPrestake = append(validatorsInPrestake, validator)
			}
		}
		if len(validatorsInPrestake) > 0 {

			options := make([]string, len(validatorsInPrestake))
			for vi, v := range validatorsInPrestake {
				options[vi] = fmt.Sprintf("ID: %d - Pubkey: 0x%s (Last ETH assignment: %s)", v.ValidatorId, v.PubKey.String(), v.LastAssignmentTime.Format(TimeFormat))
			}
			selected, _ := prompt.Select("Please select a validator to DISSOLVE:", options)

			// Get validators
			validatorId = uint64(validatorsInPrestake[selected].ValidatorId)

		} else {
			fmt.Println("No validators can be dissolved at the moment")
			return nil
		}
	}

	// Check megapool validator can be dissolved
	canDissolve, err := rp.CanDissolveValidator(validatorId)
	if err != nil {
		return err
	}

	if !canDissolve.CanDissolve {
		if canDissolve.NotInPrestake {
			fmt.Printf("Validator %d is not in the prestake status.", validatorId)
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canDissolve.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to DISSOLVE megapool validator ID: %d?", validatorId))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Dissolve
	response, err := rp.DissolveValidator(validatorId)
	if err != nil {
		return err
	}

	fmt.Printf("Dissolving megapool validator...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully dissolved megapool validator ID: %d.\n", validatorId)
	return nil

}
