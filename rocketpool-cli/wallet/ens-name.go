package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func setEnsName(c *cli.Context, name string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	if !cliutils.Confirm(fmt.Sprintf("%sWARNING: This will send a transaction from the node wallet to configure it's ENS name as '%s' \n\n%sDo you want to continue?", colorYellow, name, colorReset)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get  gas estimate
	estimateGasSetName, err := rp.EstimateGasSetEnsName(name)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(estimateGasSetName.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Set the name
	response, err := rp.SetEnsName(name)
	if err != nil {
		return err
	}

	fmt.Printf("Waiting on transaction ...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	fmt.Printf("The ENS name associated to the node account is now '%s'.\n\n", name)
	return nil

}
