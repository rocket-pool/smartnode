package node

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

const (
	registerTimezoneFlag string = "timezone"
)

func registerNode(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Prompt for timezone location
	var timezoneLocation string
	if c.String(registerTimezoneFlag) != "" {
		timezoneLocation = c.String(registerTimezoneFlag)
	} else {
		timezoneLocation = promptTimezone()
	}

	// Build the TX
	response, err := rp.Api.Node.Register(timezoneLocation)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanRegister {
		fmt.Println("The node cannot be registered:")
		if response.Data.AlreadyRegistered {
			fmt.Println("The node is already registered with Rocket Pool.")
		}
		if response.Data.RegistrationDisabled {
			fmt.Println("Node registrations are currently disabled.")
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to register this node?",
		"node registration",
		"Registering node...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("The node was successfully registered with Rocket Pool.")
	return nil
}
