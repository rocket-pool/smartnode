package node

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func setTimezoneLocation(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Prompt for timezone location
	var timezoneLocation string
	if c.String("timezone") != "" {
		timezoneLocation = c.String("timezone")
	} else {
		timezoneLocation = promptTimezone()
	}

	// Get the gas estimate
	canResponse, err := rp.CanSetNodeTimezone(timezoneLocation)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to set your timezone?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set node's timezone location
	response, err := rp.SetNodeTimezone(timezoneLocation)
	if err != nil {
		return err
	}

	fmt.Printf("Setting timezone...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("The node's timezone location was successfully updated to '%s'.\n", timezoneLocation)
	return nil

}
