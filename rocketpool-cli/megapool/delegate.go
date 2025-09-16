package megapool

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func delegateUpgradeMegapool(c *cli.Context) error {

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

	// Get megapool status
	status, err := rp.MegapoolStatus()
	if err != nil {
		return err
	}

	if status.Megapool.DelegateAddress == status.LatestDelegate || status.Megapool.UseLatestDelegate {
		fmt.Printf("The node's megapool: %s is already using the latest delegate\n", status.Megapool.Address.Hex())
		return nil
	}

	// Get the gas estimate
	canResponse, err := rp.CanDelegateUpgradeMegapool(status.Megapool.Address)
	if err != nil {
		return fmt.Errorf("error checking if megapool %s can upgrade: %w", status.Megapool.Address.Hex(), err)
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to upgrade your megapool to the latest delegate?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Upgrade megapool
	response, err := rp.DelegateUpgradeMegapool(status.Megapool.Address)
	if err != nil {
		fmt.Printf("Could not upgrade megapool %s: %s. \n", status.Megapool.Address.Hex(), err)
		return nil
	}

	// Log and wait for the auto-upgrade setting update
	fmt.Printf("Upgrading megapool %s...\n", status.Megapool.Address.Hex())
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Return
	fmt.Printf("Successfully upgraded megapool %s.\n", status.Megapool.Address.Hex())
	return nil
}

func setUseLatestDelegateMegapool(c *cli.Context, setting bool) error {
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

	// Get megapool status
	status, err := rp.MegapoolStatus()
	if err != nil {
		return err
	}

	// Return if megapool isn't deployed
	if !status.Megapool.Deployed {
		fmt.Println("The node does not have a megapool.")
		return nil
	}

	megapoolAddress := status.Megapool.Address

	// Get the gas estimate
	canResponse, err := rp.CanSetUseLatestDelegateMegapool(megapoolAddress, setting)
	if err != nil {
		return fmt.Errorf("error checking if megapool %s could have its use-latest-delegate flag changed: %w", megapoolAddress.Hex(), err)
	}
	if canResponse.MatchesCurrentSetting == true {
		if setting == true {
			fmt.Printf("Could not enable use-latest-delegate on the node's megapool, the setting is already enabled.")
		} else {
			fmt.Printf("Could not disable use-latest-delegate on the node's megapool, the setting is already disabled.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to change the auto-upgrade setting for your megapool?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Update flag
	response, err := rp.SetUseLatestDelegateMegapool(megapoolAddress, setting)
	if err != nil {
		fmt.Printf("Could not set use latest delegate for megapool %s: %s. \n", megapoolAddress.Hex(), err)
		return nil
	}

	// Log and wait for the auto-upgrade setting update
	fmt.Printf("Updating the auto-upgrade setting for megapool %s...\n", megapoolAddress.Hex())
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Return
	fmt.Printf("Successfully updated the auto-upgrade setting for megapool %s.\n", megapoolAddress.Hex())
	return nil

}
