package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/migration"
	"github.com/urfave/cli"
)

func importKey(c *cli.Context, minipoolAddress common.Address) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check and assign the EC status
	err = cliutils.CheckClientStatus(rp)
	if err != nil {
		return err
	}

	fmt.Printf("This will allow you to import the externally-created private key for the validator associated with minipool %s so it can be managed by the Smartnode's Validator Client instead of your externally-managed Validator Client.\n\n", minipoolAddress.Hex())

	success := migration.ImportKey(c, rp, minipoolAddress, "")
	if !success {
		fmt.Println("Your validator's withdrawal credentials have been changed to the minipool address, but importing the key failed.\nYou can try again later by using `rocketpool minipool import-key`.")
		return nil
	}

	// Restart the VC if necessary
	if c.Bool("no-restart") {
		return nil
	}
	if c.Bool("yes") || cliutils.Confirm("Would you like to restart the Smartnode's Validator Client now so it loads your validator's key?") {
		// Restart the VC
		fmt.Print("Restarting Validator Client... ")
		_, err := rp.RestartVc()
		if err != nil {
			fmt.Printf("failed!\n%sWARNING: error restarting validator client: %s\n\nPlease restart it manually so it picks up the new validator key for your minipool.%s", colorYellow, err.Error(), colorReset)
			return nil
		}
		fmt.Println("done!\n")
	}

	return nil
}
