package megapool

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func stake(c *cli.Context) error {

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

	validatorId := uint64(0)

	// check if the validator-id flag was used
	if c.IsSet("validator-id") {
		validatorId = c.Uint64("validator-id")
	} else {
		// Get Megapool status
		status, err := rp.MegapoolStatus()
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
				options[vi] = fmt.Sprintf("Pubkey: 0x%s (Last ETH assignment: %s)", v.PubKey.String(), v.LastAssignmentTime.Format(TimeFormat))
			}
			selected, _ := cliutils.Select("Please select a validator to stake:", options)

			// Get validators
			validatorId = uint64(validatorsInPrestake[selected].ValidatorId)

		} else {
			fmt.Println("No validators can be staked at the moment")
			return nil
		}

	// Warning reg the time necessary to build the proof
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("The stake operation will build a beacon chain proof that the validator deposit was correct. This will take several seconds to finish.\n Do you want to continue?", validatorId))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Check megapool validator can be staked
	canStake, err := rp.CanStake(validatorId)
	if err != nil {
		return err
	}

	if !canStake.CanStake {
		fmt.Printf("The validator with index %d can't be staked.\n", validatorId)
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canStake.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to stake validator id %d", validatorId))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Stake
	response, err := rp.Stake(validatorId)
	if err != nil {
		return err
	}

	fmt.Printf("Staking megapool validator %d...\n", validatorId)
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully staked megapool validator %d.\n", validatorId)
	return nil

}
