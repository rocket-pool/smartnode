package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	promptcli "github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func setEnsName(c *cli.Context, name string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	fmt.Printf("This will confirm the node's ENS name as '%s'.\n\n%sNOTE: to confirm your name, you must first register it with the ENS application at https://app.ens.domains.\nWe recommend using a hardware wallet as the base domain, and registering your node as a subdomain of it.%s\n\n", name, colorYellow, colorReset)

	// Get gas estimate
	estimateGasSetName, err := rp.EstimateGasSetEnsName(name)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(estimateGasSetName.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	if !promptcli.Confirm("Are you sure you want to confirm your node's ENS name?") {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set the name
	response, err := rp.SetEnsName(name)
	if err != nil {
		return err
	}

	fmt.Printf("Setting ENS name...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	fmt.Printf("The ENS name associated with your node account is now '%s'.\n\n", name)
	return nil

}
