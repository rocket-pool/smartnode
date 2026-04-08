package megapool

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func setUseLatestDelegateMegapool(setting *bool, yes bool) error {
	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get megapool status
	status, err := rp.MegapoolStatus(false)
	if err != nil {
		return err
	}

	// Return if megapool isn't deployed
	if !status.Megapool.Deployed {
		fmt.Println("The node does not have a megapool.")
		return nil
	}

	// If no flag was provided, prompt the user based on the current setting
	if setting == nil {
		currentSetting := status.Megapool.UseLatestDelegate
		var desired bool
		if currentSetting {
			fmt.Println("Your megapool currently has automatic delegate upgrades enabled.")
			if !prompt.Confirm("Would you like to disable automatic delegate upgrades?") {
				fmt.Println("No changes made.")
				return nil
			}
			desired = false
		} else {
			fmt.Println("Your megapool currently has automatic delegate upgrades disabled.")
			if !prompt.Confirm("Would you like to enable automatic delegate upgrades?") {
				fmt.Println("No changes made.")
				return nil
			}
			desired = true
		}
		setting = &desired
	}

	megapoolAddress := status.Megapool.Address

	// Print message we're updating the setting
	if *setting {
		fmt.Printf("Updating the use-latest-delegate setting for megapool %s to enabled...\n", megapoolAddress.Hex())
	} else {
		fmt.Printf("Updating the use-latest-delegate setting for megapool %s to disabled...\n", megapoolAddress.Hex())
	}

	// Get the gas estimate
	canResponse, err := rp.CanSetUseLatestDelegateMegapool(megapoolAddress, *setting)
	if err != nil {
		return fmt.Errorf("error checking if megapool %s could have its use-latest-delegate flag changed: %w", megapoolAddress.Hex(), err)
	}
	if canResponse.MatchesCurrentSetting {
		if *setting {
			fmt.Printf("Could not enable use-latest-delegate on the node's megapool, the setting is already enabled.")
		} else {
			fmt.Printf("Could not disable use-latest-delegate on the node's megapool, the setting is already disabled.")
		}
		fmt.Println()
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(yes || prompt.Confirm("Are you sure you want to change the use-latest-delegate setting for your megapool?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Update flag
	response, err := rp.SetUseLatestDelegateMegapool(megapoolAddress, *setting)
	if err != nil {
		fmt.Printf("Could not set use latest delegate for megapool %s: %s. \n", megapoolAddress.Hex(), err)
		return nil
	}

	// Log and wait for the use-latest-delegate setting update
	fmt.Printf("Updating the use-latest-delegate setting for megapool %s...\n", megapoolAddress.Hex())
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Return
	fmt.Printf("Successfully updated the use-latest-delegate setting for megapool %s.\n", megapoolAddress.Hex())
	return nil

}
