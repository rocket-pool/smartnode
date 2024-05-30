package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func setSnapshotAddress(c *cli.Context, snapshotAddress common.Address, signature string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check for Houston
	houston, err := rp.IsHoustonDeployed()
	if err != nil {
		return fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}
	if !houston.IsHoustonDeployed {
		fmt.Println("This command cannot be used until Houston has been deployed.")
		return nil
	}

	// TODO:
	// resp, err := rp.CanSetSnapshotAddress(snapshotAddress, signature)
	// if err != nil {
	// 	return fmt.Errorf("error calling can-set-snapshot-address: %w", err)
	// }

	// // Assign max fees
	// err = gas.AssignMaxFeeAndLimit(resp.GasInfo, rp, c.Bool("yes"))
	// if err != nil {
	// 	return err
	// }

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to set the snapshot address?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set Snapshot Address
	response, err := rp.SetSnapshotAddress(snapshotAddress, signature)
	if err != nil {
		return fmt.Errorf("error calling set-snapshot-address: %w", err)
	}

	fmt.Printf("Setting snapshot address...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return fmt.Errorf("error setting snapshot address: %w", err)
	}

	// Log & Return
	fmt.Println("Successfully set snapshot address.")
	return nil
}
