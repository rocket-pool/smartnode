package megapool

import (
	"fmt"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func delegateUpgradeMegapool(yes bool) error {
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

	megapoolAddress := status.Megapool.Address
	currentDelegate := status.Megapool.DelegateAddress
	latestDelegate := status.LatestDelegate

	// Already on the latest stored delegate
	if currentDelegate == latestDelegate {
		fmt.Println("The megapool is already using the latest delegate.")
		return nil
	}

	if status.Megapool.UseLatestDelegate {
		fmt.Println("The megapool is set to automatically use the latest delegate.")
		fmt.Printf("Its stored delegate is still %s and can be upgraded to %s.\n",
			color.LightBlue(currentDelegate.Hex()),
			color.LightBlue(latestDelegate.Hex()),
		)
	} else {
		fmt.Printf("The megapool is using delegate %s and can be upgraded to %s.\n",
			color.LightBlue(currentDelegate.Hex()),
			color.LightBlue(latestDelegate.Hex()),
		)
	}

	// Get the gas estimate
	canResponse, err := rp.CanDelegateUpgradeMegapool(megapoolAddress)
	if err != nil {
		return fmt.Errorf("error checking if megapool %s can upgrade its delegate: %w", megapoolAddress.Hex(), err)
	}
	if canResponse.GasLimits.IsBlank() {
		return fmt.Errorf("could not estimate gas for megapool delegate upgrade (the upgrade may not be possible right now)")
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasLimits, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to upgrade the megapool delegate?") {
		fmt.Println("Cancelled.")
		return nil
	}

	// Upgrade
	response, err := rp.DelegateUpgradeMegapool(megapoolAddress)
	if err != nil {
		return fmt.Errorf("could not upgrade megapool %s delegate: %w", megapoolAddress.Hex(), err)
	}

	// Log and wait for the transaction
	fmt.Printf("Upgrading megapool %s delegate...\n", megapoolAddress.Hex())
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	fmt.Printf("Successfully upgraded megapool %s to delegate %s.\n", megapoolAddress.Hex(), latestDelegate.Hex())
	return nil
}

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
	err = gas.AssignMaxFeeAndLimit(canResponse.GasLimits, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to change the use-latest-delegate setting for your megapool?") {
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
