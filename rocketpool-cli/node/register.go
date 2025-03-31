package node

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func registerNode(c *cli.Context) error {

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

	// Check node can be registered
	canRegister, err := rp.CanRegisterNode(timezoneLocation)
	if err != nil {
		return err
	}
	if !canRegister.CanRegister {
		fmt.Println("The node cannot be registered:")
		if canRegister.AlreadyRegistered {
			fmt.Println("The node is already registered with Rocket Pool.")
		}
		if canRegister.RegistrationDisabled {
			fmt.Println("Node registrations are currently disabled.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canRegister.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to register this node?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Register node
	response, err := rp.RegisterNode(timezoneLocation)
	if err != nil {
		return err
	}

	fmt.Printf("Registering node...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("The node was successfully registered with Rocket Pool.")
	return nil

}
