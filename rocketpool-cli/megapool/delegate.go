package megapool

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func setUseLatestDelegateMegapool(c *cli.Context, setting bool) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

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
	canSet, err := rp.CanSetUseLatestDelegateMegapool(megapoolAddress, setting)
	if err != nil {
		return fmt.Errorf("error checking if megapool %s could have its use-latest-delegate flag changed: %w", megapoolAddress.Hex(), err)
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canSet.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to change the auto-upgrade setting for your megapool?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Return
	return nil

}
