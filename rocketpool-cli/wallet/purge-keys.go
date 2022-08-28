package wallet

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func purgeKeys(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	if !cliutils.Confirm("WARNING:\n This command will delete all files related to validator keys. Do you want to continue?") {
		fmt.Println("No action was taken.")
		return nil
	}

	// Delete validator keys
	_, err = rp.PurgeKeys()
	if err != nil {
		return err
	}

	fmt.Println("Deleted all validator keys.")
	return nil

}
