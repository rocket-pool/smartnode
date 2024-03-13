package wallet

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/urfave/cli/v2"
)

func setPassword(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get & check wallet status
	statusResponse, err := rp.Api.Wallet.Status()
	if err != nil {
		return err
	}
	status := statusResponse.Data.WalletStatus

	// Check if it's already set
	if status.HasPassword {
		fmt.Println("The node wallet password has already been set.")
		return nil
	}

	// Get the password
	passwordString := c.String(passwordFlag.Name)
	if passwordString == "" {
		passwordString = promptPassword()
	}
	password, err := input.ValidateNodePassword("password", passwordString)
	if err != nil {
		return fmt.Errorf("error validating password: %w", err)
	}

	// Get the save flag
	savePassword := c.Bool(utils.YesFlag.Name) || utils.Confirm("Would you like to save the password to disk? If you do, your node will be able to handle transactions automatically after a client restart; otherwise, you will have to repeat this command to manually enter the password after each restart.")

	// Run it
	_, err = rp.Api.Wallet.SetPassword(password, savePassword)
	if err != nil {
		return fmt.Errorf("error setting password: %w", err)
	}

	// Log & return
	fmt.Println("The password has been successfully uploaded to the daemon.")
	return nil
}
