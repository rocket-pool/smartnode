package megapool

import (
	"fmt"
	"strconv"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
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

	validatorIndex := uint64(0)

	// check if the validator-id flag was used
	if c.IsSet("validator-index") {
		validatorIndex = c.Uint64("validator-index")
	} else {
		// Ask for validator index
		validatorIndexString := cliutils.Prompt("Which validator index do you want to stake?", "^\\d+$", "Invalid validator index")
		validatorIndex, err = strconv.ParseUint(validatorIndexString, 0, 64)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid validator index: %w.\n", validatorIndexString, err)
		}
	}
	// Check megapool validator can be staked
	canStake, err := rp.CanStake(validatorIndex)
	if err != nil {
		return err
	}

	if !canStake.CanStake {
		fmt.Printf("The validator with index %d can't be staked.\n", validatorIndex)
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canStake.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to stake validator id %d", validatorIndex))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Stake
	response, err := rp.Stake(validatorIndex)
	if err != nil {
		return err
	}

	fmt.Printf("Staking megapool validator %d...\n", validatorIndex)
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully staked megapool validator %d.\n", validatorIndex)
	return nil

}
