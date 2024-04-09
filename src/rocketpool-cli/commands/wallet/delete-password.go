package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/urfave/cli/v2"
)

func deletePassword(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get & check wallet status
	statusResponse, err := rp.Api.Wallet.Status()
	if err != nil {
		return err
	}
	status := statusResponse.Data.WalletStatus

	// Check if it's already set
	if !status.Password.IsPasswordSaved {
		fmt.Println("The node wallet password is not saved to disk.")
		return nil
	}

	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm("Are you sure you want to delete your password from disk? Your node will not be able to submit transactions after a restart until you manually enter the password")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Run it
	_, err = rp.Api.Wallet.DeletePassword()
	if err != nil {
		return fmt.Errorf("error deleting password: %w", err)
	}

	// Log & return
	fmt.Println("The password has been successfully removed from disk storage.")
	return nil
}
