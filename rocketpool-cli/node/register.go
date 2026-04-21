package node

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func registerNode(timezoneLocation string, yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Prompt for timezone location
	if timezoneLocation == "" {
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
	err = gas.AssignMaxFeeAndLimit(canRegister.GasInfo, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to register this node?") {
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
